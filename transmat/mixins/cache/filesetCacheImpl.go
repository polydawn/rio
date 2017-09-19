package cache

import (
	"context"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	cacheapi "go.polydawn.net/rio/cache"
	"go.polydawn.net/rio/fs"
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
	// Initialize cache.
	// TODO

	// Check if we already have it in cache and can return earlier.
	// TODO

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
