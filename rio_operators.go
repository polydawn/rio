package rio

// Types in this file are workhorses.  None of them are serializable.

// A WarehouseAgent factory -- will ping or establish connections to a WarehouseAddr,
// and return info or objects with methods for further communication.
type WarehouseDialer interface {
	Ping(WarehouseAddr) (writable bool, err error)
	NewAgent(WarehouseAddr) (WarehouseAgent, error)
}

// Communicates with a Warehouse, handling read and write operations.
//
// The WarehouseAgent API operates at a very high level: it understands wares
// and handles data by WareID.
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
