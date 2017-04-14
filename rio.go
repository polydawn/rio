package rio

import (
	"context"
	"time"

	"go.polydawn.net/rio/fs"
)

// A content-addressable Ware ID.
// Serialized as a string "kind:hash" -- Kinds are a whitelist and hashes also must avoid special chars.
//
// A ware is a packed filesystem.
// It contains one or more files and directories, and metadata for each.
//
// You can expect to enumerate the standard metadata for any ware.
// For some formats, traversing to inner bits of ware for extraction will be efficient;
// for most, the unit of fetch is the entire ware, so asking for a subpath is not consequential.
type WareID struct {
	Kind string
	Hash string
}

type WarehouseAddr struct{}

type WarehouseDialer interface {
	Ping(WarehouseAddr) (writable bool, err error)
	NewAgent(WarehouseAddr) (WarehouseAgent, error)
}

type Justification struct{}

type WarehouseAgent interface {
	Has(WareID) (bool, Justification, error)

	CanWrite() bool

	// Modify the warehouse to contain the requested WareID,
	// tagging it immediately with the given Justification,
	// fetching from any (or all, if the protocols are smart enough)
	// of the sources in the given WarehouseAddr list.
	//
	// Mirroring locally is also just a special case of 'Put':
	//  `dialer.NewAgent("./mirror/").Put(theWare, reasons, publicMirror)`.
	Put(WareID, Justification, sources []WarehouseAgent) error
}

type Transmat interface {
	// Unpack a ware into a specified filesystem path.
	// Returns a WareID, which, in the case of filters, may be different
	// than the requested WareID.
	Materialize(
		ctx context.Context, // Long-running call.  Cancellable.
		path fs.AbsolutePath, // Where to put a filesystem.
		wareID WareID, // What filesystem slice ware to unpack.
		filters *Filters, // Optionally: filters we should apply while unpacking.
		sources []WarehouseAgent, // Warehouses we can talk to.
		monitor MaterializeMonitor, // Optionally: callbacks for progress monitoring.
	) (WareID, error)

	// Traverses the specified filesystem path, hashing it,
	// and optionally packing it into a Warehouse for storage.
	Scan(
		ctx context.Context, // Long-running call.  Cancellable.
		path fs.AbsolutePath, // What path to scan contents of.
		filters *Filters, // Optionally: filters we should apply while packing.
		destination WarehouseAgent, // Warehouse to upload to.  (Use mirroring later for multiple warehouses.)
		monitor ScanMonitor, // Optionally: callbacks for progress monitoring.
	) (WareID, error)
}
type MaterializeMonitor interface {
	Progress(at, max int)
}
type ScanMonitor interface {
	Progress(at, max int)
}

type Filters struct {
	FlattenUID   *int       // If set: use that number; default for pack is to flatten to 1000; default for unpack is to respect packed metadata.
	FlattenGID   *int       // If set: use that number; default for pack is to flatten to 1000; default for unpack is to respect packed metadata.
	FlattenMtime *time.Time // If set: use that time; default for pack is to flatten to Jan 2010; default for unpack is to respect packed metadata.
}

// A local filesystem area where CAS caching is maintained.
type Depot struct {
	base fs.AbsolutePath
}

// Helper that yields the path to the ware,
// getting it directly from the Depot if already present,
// or invoking the transmat to get it into the Depot if necessary.
func (d *Depot) YieldFilesystem(WareID, Transmat) (fs.AbsolutePath, error) {
	return fs.MustAbsolutePath("/"), nil
}
