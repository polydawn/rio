/*
	The tar transmat packs filesystems into the widely-recognized "tar" format,
	and can use any k/v-styled warehouse for storage.
*/
package tartrans

import (
	"archive/tar"
	"context"
	"io"
	"net/url"

	"go.polydawn.net/rio/fs"
	. "go.polydawn.net/rio/lib/errcat"
	"go.polydawn.net/rio/warehouse/impl/kvfs"
	"go.polydawn.net/timeless-api"
	"go.polydawn.net/timeless-api/rio"
)

var (
	_ rio.UnpackFunc = Unpack
)

func Unpack(
	ctx context.Context, // Long-running call.  Cancellable.
	wareID api.WareID, // What wareID to fetch for unpacking.
	path string, // Where to unpack the fileset (absolute path).
	filters api.FilesetFilters, // Optionally: filters we should apply while unpacking.
	warehouses []api.WarehouseAddr, // Warehouses we can try to fetch from.
	monitor rio.Monitor, // Optionally: callbacks for progress monitoring.
) (api.WareID, error) {
	// Sanitize arguments.
	path2 := fs.MustAbsolutePath(path)

	// Pick a warehouse.
	//  With K/V warehouses, this takes the form of "pick the first one that answers".
	var reader io.ReadCloser
	for _, addr := range warehouses {
		// REVIEW ... Do I really have to parse this again?  is this sanely encapsulated?
		u, err := url.Parse(string(addr))
		if err != nil {
			return api.WareID{}, Errorf(rio.ErrUsage, "failed to parse URI: %s", err)
		}
		switch u.Scheme {
		case "file", "file+ca":
			whCtrl, err := kvfs.NewController(addr)
			switch Category(err) {
			case rio.ErrWarehouseUnavailable:
				// TODO log something to the monitor
				continue // okay!  skip to the next one.
			default:
				return api.WareID{}, err
			}
			reader, err = whCtrl.OpenReader(wareID)
			switch Category(err) {
			case rio.ErrWareNotFound:
				// TODO log something to the monitor
				continue // okay!  skip to the next one.
			default:
				return api.WareID{}, err
			}
		default:
			return api.WareID{}, Errorf(rio.ErrUsage, "tar unpack doesn't support %q scheme (valid options are 'file' or 'file+ca')", u.Scheme)
		}
	}
	if reader == nil { // aka if no warehouses available:
		return api.WareID{}, Errorf(rio.ErrWarehouseUnavailable, "no warehouses were available!")
	}

	// Wrap input stream with decompression as necessary.
	//  Which kind of decompression to use can be autodetected by magic bytes.
	// TODO

	// Convert the raw byte reader to a tar stream.
	tarReader := tar.NewReader(reader)

	// Extract.
	err := Extract(ctx, path2, filters, tarReader)
	if err != nil {
		return api.WareID{}, err
	}
	return api.WareID{}, nil
}
