package util

import (
	"context"
	"fmt"
	"io"

	. "github.com/warpfork/go-errcat"

	api "github.com/polydawn/go-timeless-api"
	"github.com/polydawn/go-timeless-api/rio"
	"github.com/polydawn/rio/config"
	"github.com/polydawn/rio/fs"
	"github.com/polydawn/rio/fs/osfs"
	"github.com/polydawn/rio/transmat/mixins/cache"
)

type unpackFn func(
	ctx context.Context,
	afs fs.FS,
	filt api.FilesetUnpackFilter,
	archiveWareID api.WareID,
	reader io.Reader,
	mon rio.Monitor,
) (
	prefilterWareID api.WareID,
	actualWareID api.WareID,
	err error,
)

// CreateUnpack generates a standard rio.UnpackFunc shared by both zip and tar transmat implementations.
// The basic wrapper caches unpacking, validates types, and manages warehouse behavior.
func CreateUnpack(t api.PackType, unpacker unpackFn) rio.UnpackFunc {
	return func(
		ctx context.Context, // Long-running call.  Cancellable.
		wareID api.WareID, // What wareID to fetch for unpacking.
		path string, // Where to unpack the fileset (absolute path).
		filt api.FilesetUnpackFilter, // Optionally: filters we should apply while unpacking.
		placementMode rio.PlacementMode, // Optionally: a placement mode (default is "copy").
		warehouses []api.WarehouseLocation, // Warehouses we can try to fetch from.
		mon rio.Monitor, // Optionally: callbacks for progress monitoring.
	) (_ api.WareID, err error) {
		if mon.Chan != nil {
			defer close(mon.Chan)
		}
		defer RequireErrorHasCategory(&err, rio.ErrorCategory(""))

		// Sanitize arguments.
		if wareID.Type != t {
			return api.WareID{}, Errorf(rio.ErrUsage, "this transmat implementation only supports packtype %q (not %q)", t, wareID.Type)
		}
		if !filt.IsComplete() {
			return api.WareID{}, Errorf(rio.ErrUsage, "filters must be completely specified")
		}
		if placementMode == "" {
			placementMode = rio.Placement_Copy
		}
		// Wrap the direct unpack func with cache behavior; call that.
		return cache.Lrn2Cache(
			osfs.New(config.GetCacheBasePath()),
			wrapUnpacker(unpacker),
		)(ctx, wareID, path, filt, placementMode, warehouses, mon)
	}
}

func wrapUnpacker(unpacker unpackFn) rio.UnpackFunc {
	return func(
		ctx context.Context,
		wareID api.WareID,
		path string,
		filt api.FilesetUnpackFilter,
		placementMode rio.PlacementMode,
		warehouses []api.WarehouseLocation,
		mon rio.Monitor,
	) (_ api.WareID, err error) {
		defer RequireErrorHasCategory(&err, rio.ErrorCategory(""))

		// Sanitize arguments.
		path2 := fs.MustAbsolutePath(path)

		// Pick a warehouse and get a reader.
		reader, err := PickReader(wareID, warehouses, false, mon)
		if err != nil {
			return api.WareID{}, err
		}
		defer reader.Close()

		// Construct filesystem wrapper to use for all our ops.
		afs := osfs.New(path2)

		// Extract.
		prefilterWareID, unpackWareID, err := unpacker(ctx, afs, filt, wareID, reader, mon)
		if err != nil {
			return unpackWareID, err
		}

		// Check for hash mismatch before returning, because that IS an error,
		//  but also return the hash we got either way.
		if prefilterWareID != wareID {
			return unpackWareID, ErrorDetailed(
				rio.ErrWareHashMismatch,
				fmt.Sprintf("hash mismatch: expected %q, got %q (filtered %q)", wareID, prefilterWareID, unpackWareID),
				map[string]string{
					"expected": wareID.String(),
					"actual":   prefilterWareID.String(),
					"filtered": unpackWareID.String(),
				},
			)
		}
		return unpackWareID, nil
	}
}
