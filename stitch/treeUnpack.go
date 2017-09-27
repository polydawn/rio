package stitch

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	. "github.com/polydawn/go-errcat"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/rio/cache"
	"go.polydawn.net/rio/config"
	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/fsOp"
	"go.polydawn.net/rio/stitch/placer"
)

/*
	Struct to gather the args for a single rio.Unpack func call.
	(The context object and monitors are handled in a different band.)

	Note the similar name to a structure in the go-timeless-api packages;
	this one is not serializable, is internal, and
	contains the literal set of warehouses already resolved,
	as well as the path inline rather than in a map key, so we can sort slices.
*/
type UnpackSpec struct {
	Path       fs.AbsolutePath
	WareID     api.WareID
	Filters    api.FilesetFilters
	Warehouses []api.WarehouseAddr
}

// Cast slices to this type to sort by target path (which is effectively mountability order).
type UnpackSpecByPath []UnpackSpec

func (a UnpackSpecByPath) Len() int           { return len(a) }
func (a UnpackSpecByPath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a UnpackSpecByPath) Less(i, j int) bool { return a[i].Path.String() < a[j].Path.String() }

type unpackResult struct {
	Path  fs.AbsolutePath // cache path or mount source path
	Error error
}

type Assembler struct {
	cache          fs.FS
	unpackTool     rio.UnpackFunc
	placerTool     placer.Placer
	fillerDirProps fs.Metadata
}

func NewAssembler(unpackTool rio.UnpackFunc) (*Assembler, error) {
	placerTool, err := placer.GetMountPlacer()
	if err != nil {
		return nil, err
	}
	return &Assembler{
		cache:      osfs.New(config.GetCacheBasePath()),
		unpackTool: unpackTool,
		placerTool: placerTool,
		fillerDirProps: fs.Metadata{
			Type: fs.Type_Dir, Perms: 0755, Uid: 0, Gid: 0, Mtime: fs.DefaultAtime,
		},
	}, nil
}

func (a *Assembler) Run(ctx context.Context, targetFs fs.FS, parts []UnpackSpec) (func() error, error) {
	sort.Sort(UnpackSpecByPath(parts))

	// Fan out materialization into cache paths.
	unpackResults := make([]unpackResult, len(parts))
	var wg sync.WaitGroup
	wg.Add(len(parts))
	for i, part := range parts {
		go func(i int, part UnpackSpec) {
			defer wg.Done()
			res := &unpackResults[i]
			// If it's a mount, shortcut.
			if part.WareID.Type == "mount" {
				res.Path, res.Error = fs.ParseAbsolutePath(part.WareID.Hash)
				return
			}
			// Unpack with placement=none to populate cache.
			resultWareID, err := a.unpackTool(
				ctx, // TODO fork em out
				part.WareID,
				"-",
				part.Filters,
				rio.Placement_None,
				part.Warehouses,
				rio.Monitor{},
			)
			// Yield the cache path.
			res.Path = config.GetCacheBasePath().Join(cache.ShelfFor(resultWareID))
			res.Error = err
			// TODO if any error, fan out cancellations
		}(i, part)
	}
	wg.Wait()
	// Yield up any errors from individual unpacks.
	for _, result := range unpackResults {
		if result.Error != nil {
			return nil, result.Error
		}
	}

	// Zip up all placements, in order.
	//  Parent dirs are made as necessary along the way.
	hk := &housekeeping{}
	for i, part := range parts {
		path := part.Path.CoerceRelative()

		// Ensure parent dirs.
		for _, parentPath := range path.Dir().Split() {
			target, isSymlink, err := targetFs.Readlink(parentPath)
			if isSymlink {
				return nil, fs.NewBreakoutError(
					targetFs.BasePath(),
					path,
					parentPath,
					target,
				)
			} else if err == nil {
				continue
			} else if Category(err) == fs.ErrNotExists {
				// Make the parent dir if it does not exist.
				a.fillerDirProps.Name = parentPath
				// Could be cleaner: this PlaceFile call rechecks the symlink thing, but it's the shortest call for "make all props right plz".
				if err := fsOp.PlaceFile(targetFs, a.fillerDirProps, nil, false); err != nil {
					return nil, err
				}
			} else {
				// Halt assembly attempt for any unhandlable errors that come up during parent path establishment.
				return nil, err
			}
		}

		// Invoke placer.
		//  Accumulate the individual cleanup funcs into a mega func we'll return.
		//  If errors occur during any placement, fire the cleanups so far before returning.
		var janitor placer.Janitor
		var err error
		switch part.WareID.Type {
		case "mount":
			janitor, err = placer.BindPlacer(unpackResults[i].Path, part.Path, false)
		default:
			janitor, err = a.placerTool(unpackResults[i].Path, part.Path, false)
		}
		if err != nil {
			hk.Teardown()
			return nil, err
		}
		hk.append(janitor)
	}
	return hk.Teardown, nil
}

type housekeeping struct {
	CleanupStack []placer.Janitor
}

func (hk *housekeeping) append(janitor placer.Janitor) {
	hk.CleanupStack = append(hk.CleanupStack, janitor)
}

func (hk housekeeping) Teardown() error {
	progress := make([]string, len(hk.CleanupStack))
	var firstError error
	for i := len(hk.CleanupStack) - 1; i >= 0; i-- {
		janitor := hk.CleanupStack[i]
		if firstError != nil && !janitor.AlwaysTry() {
			progress[i] = "\tskipped: " + janitor.Description()
			continue
		}
		err := hk.CleanupStack[i].Teardown()
		if err != nil {
			if firstError == nil {
				firstError = err
			}
			progress[i] = "\tfailed:  " + janitor.Description()
			continue
		}
		progress[i] = "\tsuccess: " + janitor.Description()
	}
	if firstError != nil {
		// Keep the category of the first one, but also fold in
		//  the string of everything that did or did not get cleaned up.
		cleanupReport := strings.Join(progress, "\n")
		firstError = ErrorDetailed(
			Category(firstError),
			fmt.Sprintf("%s.  The following cleanups were attempted:\n%s", firstError, cleanupReport),
			map[string]string{"cleanupReport": cleanupReport},
		)
	}
	return firstError
}
