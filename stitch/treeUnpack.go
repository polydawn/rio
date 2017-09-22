package stitch

import (
	"sort"

	. "github.com/polydawn/go-errcat"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fsOp"
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

type assembler struct {
	cache          fs.FS
	unpackTool     rio.UnpackFunc
	placerTool     func( /*todo*/ )
	fillerDirProps fs.Metadata
}

func (a assembler) Run(targetFs fs.FS, parts []UnpackSpec) error {
	sort.Sort(UnpackSpecByPath(parts))

	// Fan out materialization into cache paths.
	// TODO

	// Zip up all placements, in order.
	//  Parent dirs are made as necessary along the way.
	for _, part := range parts {
		path := part.Path.CoerceRelative()
		// Ensure parent dirs.
		for _, parentPath := range path.Dir().Split() {
			target, isSymlink, err := targetFs.Readlink(parentPath)
			if isSymlink {
				return fs.NewBreakoutError(
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
					return err
				}
			} else {
				// Halt assembly attempt for any unhandlable errors that come up during parent path establishment.
				return err
			}
		}
		// Invoke placer.
		// TODO
	}
	return nil
}
