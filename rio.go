package rio

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
	Ping(WarehouseAddr) error
	NewAgent(WarehouseAddr) (WarehouseAgent, error)
}

type Justification struct{}

type WarehouseAgent interface {
	Has(WareID) (bool, Justification, error)

	// Modify the warehouse to contain the requested WareID,
	// tagging it immediately with the given Justification,
	// fetching from any (or all, if the protocols are smart enough)
	// of the sources in the given WarehouseAddr list.
	//
	// Mirroring locally is also just a special case of 'Put':
	//  `dialer.NewAgent("./mirror/").Put(theWare, reasons, publicMirror)`.
	Put(WareID, Justification, sources []WarehouseAddr) error
}
