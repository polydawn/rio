/*
	The tar transmat packs filesystems into the widely-recognized "tar" format,
	and can use any k/v-styled warehouse for storage.
*/
package tartrans

import (
	"archive/tar"
	"context"
	"io"

	"go.polydawn.net/rio"
	"go.polydawn.net/rio/fs"
)

func Materialize(
	ctx context.Context, // Long-running call.  Cancellable.
	path fs.AbsolutePath, // Where to put a filesystem.
	filters rio.Filters, // Optionally: filters we should apply while unpacking.
	wareID rio.WareID, // What filesystem slice ware to unpack.
	sources []rio.WarehouseAgent, // Warehouses we can talk to.
	monitor rio.MaterializeMonitor, // Optionally: callbacks for progress monitoring.
) (rio.WareID, error) {
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
	err := Extract(ctx, path, filters, tarReader)
	if err != nil {
		return rio.WareID{}, err
	}
	return rio.WareID{}, nil
}
