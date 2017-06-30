package rio

// Types in this file are interface definitions for `rio`'s utilities to implement.

import (
	"context"

	"go.polydawn.net/rio/fs"
)

type Transmat interface {
	// Unpack a ware into a specified filesystem path.
	// Returns a WareID, which, in the case of filters, may be different
	// than the requested WareID.
	Materialize(
		ctx context.Context, // Long-running call.  Cancellable.
		path fs.AbsolutePath, // Where to put a filesystem.
		wareID WareID, // What filesystem slice ware to unpack.
		filters Filters, // Optionally: filters we should apply while unpacking.
		sources []WarehouseAgent, // Warehouses we can talk to.
		monitor MaterializeMonitor, // Optionally: callbacks for progress monitoring.
	) (WareID, error)

	// Traverses the specified filesystem path, hashing it,
	// and optionally packing it into a Warehouse for storage.
	Scan(
		ctx context.Context, // Long-running call.  Cancellable.
		path fs.AbsolutePath, // What path to scan contents of.
		filters Filters, // Optionally: filters we should apply while packing.
		destination WarehouseAgent, // Optionally: Warehouse to upload to.  (Use mirroring later for multiple warehouses.)
		monitor ScanMonitor, // Optionally: callbacks for progress monitoring.
	) (WareID, error)
}
type MaterializeMonitor interface {
	Progress(at, max int)
}
type ScanMonitor interface {
	Progress(at, max int)
}

// A local filesystem area where CAS caching is maintained.
type Depot struct {
	base fs.AbsolutePath

	// REVIEW: we're not really decided where this type should go; or, if it should be concrete.
	// For example, tar transmats and git transmats cache in *very* different ways,
	// because the underlying systems have different granularity to their sharable objects.
}
