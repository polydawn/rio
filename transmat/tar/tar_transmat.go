/*
	The tar transmat packs filesystems into the widely-recognized "tar" format,
	and can use any k/v-styled warehouse for storage.
*/
package tartrans

import (
	"context"

	"go.polydawn.net/rio"
	"go.polydawn.net/rio/fs"
)

func Materialize(
	ctx context.Context, // Long-running call.  Cancellable.
	path fs.AbsolutePath, // Where to put a filesystem.
	wareID rio.WareID, // What filesystem slice ware to unpack.
	sources []rio.WarehouseAgent, // Warehouses we can talk to.
	monitor rio.MaterializeMonitor, // Optionally: callbacks for progress monitoring.
) error {
	return nil
}
