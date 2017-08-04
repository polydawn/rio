/*
	The tar transmat packs filesystems into the widely-recognized "tar" format,
	and can use any k/v-styled warehouse for storage.
*/
package tartrans

import (
	"archive/tar"
	"context"
	"io"

	"go.polydawn.net/rio/fs"
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

	// TODO Warehouse APIs need to get more concrete.  You need an `Open()` method that yields an io.ReadCloser.
	// The current 'warehouseAgent' concept doesn't express that, because it's trying to be generalized (probably to a farcical degree, honestly).
	// Do we really want to accept a range of agents here?  Shouldn't the transmat have an earlier opinion about whether that URL can possibly represent something that makes sense to ping?
	// Are WarehouseAgents a read/write duplex, or does it make more sense to keep those separate?
	// Might it make sense to actually have the transmat interface expose dialing methods?
	var reader io.ReadCloser

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
