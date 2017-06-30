package rio

// Types in this file are all serializable.
// These types cover the "RPC" interface surface area of `rio`.

import (
	"time"
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
//
// FIXME the hitch project has a definition of this too which must be sync'd up
type WareID struct {
	Kind string
	Hash string
}

// A string describing an address/protocol for talking to a storage warehouse.
//
// "http" or "file://" URLs are common examples, as are S3 buckets, etc.
//
// TODO the 'interface{}' type is a placeholder -- *are* these required to be flat strings?
// maybe.
type WarehouseAddr interface{}

// Placeholder: a data object describing why a Warehouse is storing a WareID.
// (This will evolve along with the `hitch` project, and is probably mostly
// hitch's problem to define.)
type Justification struct{}

// For each value:
//   If set: use that number;
//   default for pack is to flatten;
//   default for unpack is to respect packed metadata.
//   To keep during pack: set the keep bool.
// If keep is true, the value must be nil or the filter is invalid.
type Filters struct {
	FlattenUID struct {
		Keep  bool    `json:"keep,omitempty"`
		Value *uint32 `json:"value,omitempty"`
	} `json:"uid"`
	FlattenGID struct {
		Keep  bool    `json:"keep,omitempty"`
		Value *uint32 `json:"value,omitempty"`
	} `json:"gid"`
	FlattenMtime struct {
		Keep  bool       `json:"keep,omitempty"`
		Value *time.Time `json:"value,omitempty"`
	} `json:"mtime"`
}

var (
	FilterDefaultUid   uint32 = 1000
	FilterDefaultGid   uint32 = 1000
	FilterDefaultMtime        = time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)
)
