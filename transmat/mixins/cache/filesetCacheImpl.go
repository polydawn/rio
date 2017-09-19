package cache

import (
	"context"

	. "github.com/polydawn/go-errcat"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	cacheapi "go.polydawn.net/rio/cache"
	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fsOp"
)

var ShelfFor = cacheapi.ShelfFor

func Lrn2Cache(cacheFs fs.FS, unpackTool rio.UnpackFunc) rio.UnpackFunc {
	return cache{cacheFs, unpackTool}.Unpack
}

type cache struct {
	afs        fs.FS
	unpackTool rio.UnpackFunc
}

/*
	Proxies most args to the cache's unpack tool, except for placementmode and path,
	which it sets to rio.Placement_Direct and a temporary path in the cache filesystem.
	If unpacking completes successfully, the temp path will be moved to a permanent
	location in the cache, which is specified by the public interface `rio/cache.GetShelf`.

	Any behaviors specified by the placementMode -- copying, mounting, etc -- are enacted
	by this func after the unpack finishes and the temp path committed to the cache.
*/
func (c cache) Unpack(
	ctx context.Context,
	wareID api.WareID,
	path string,
	filt api.FilesetFilters,
	placementMode rio.PlacementMode,
	warehouses []api.WarehouseAddr,
	monitor rio.Monitor,
) (api.WareID, error) {
	// Absolute first thing: Check if we already have the ware in cache and can jump to placement ASAP.
	//  (This must be first because we're willing to read cache even in "direct" mode, but
	//  yet *not* willing to even initialize empty cache dirs in that mode.)
	shelf := ShelfFor(wareID)
	_, err := c.afs.Stat(shelf)
	switch placementMode {
	case rio.Placement_Direct:
		switch err {
		case nil: // If no error and it exists, use copy placer.
			// TODO time to backtrack and fill in placer impls :)
		default: // On any error at all (not just ErrNotExist), ignore the cache, and go direct:
			return c.unpackTool(ctx, wareID, path, filt, rio.Placement_Direct, warehouses, monitor)
		}
	case rio.Placement_None, rio.Placement_Copy, rio.Placement_Mount:
		switch Category(err) {
		case nil: // If no error and it exists, use copy placer.
			// TODO time to backtrack and fill in placer impls :)
		case fs.ErrNotExists: // "not exists" is just a cache miss!  Proceed to fill it in.
			// continue
		default:
			return api.WareID{}, err
		}
	default:
		panic("unreachable")
	}

	// Initialize cache.
	//  Ensure the cache commit root dir exists.
	if err := fsOp.MkdirAll(c.afs, fs.MustRelPath(wareID.Type+"/fileset"), 0700); err != nil {
		return api.WareID{}, Errorf(rio.ExitLocalCacheProblem, "cannot initialize cache dirs: %s", err)
	}

	// Pick a temp path to unpack into.
	var tmpPath fs.RelPath
	tmpPathStr := c.afs.BasePath().Join(tmpPath).String()
	// Delegate!
	resultWareID, err := c.unpackTool(ctx, wareID, tmpPathStr, filt, rio.Placement_Direct, warehouses, monitor)
	if err != nil {
		// Cleanup the tempdir
		// TODO
		return resultWareID, err
	}

	// Successful unpack: commit it to its shelf location.
	// TODO just an mv.

	// Goto placer.
	// TODO
	return resultWareID, nil
}
