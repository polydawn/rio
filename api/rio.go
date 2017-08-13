package api

/*
	This file is all serializable types used in Rio
	to define filesets, WareIDs, packing systems, and storage locations.
*/

import (
	"fmt"
	"strings"

	"github.com/polydawn/refmt/obj/atlas"
)

/*
	Ware IDs are content-addressable, cryptographic hashes which uniquely identify
	a "ware" -- a packed filesystem snapshot.
	A ware contains one or more files and directories, and metadata for each.

	Ware IDs are serialized as a string in two parts, separated by a colon --
	for example like "git:f23ae1829" or "tar:WJL8or32vD".
	The first part communicates which kind of packing system computed the hash,
	and the second part is the hash itself.
*/
type WareID struct {
	Type string
	Hash string
}

func ParseWareID(x string) (WareID, error) {
	ss := strings.SplitN(x, ":", 2)
	if len(ss) < 2 {
		return WareID{}, fmt.Errorf("wareIDs must have contain a colon character (they are of form \"<type>:<hash>\")")
	}
	return WareID{ss[0], ss[1]}, nil
}

func (x WareID) String() string {
	return x.Type + ":" + x.Hash
}

var WareID_AtlasEntry = atlas.BuildEntry(WareID{}).Transform().
	TransformMarshal(atlas.MakeMarshalTransformFunc(
		func(x WareID) (string, error) {
			return x.String(), nil
		})).
	TransformUnmarshal(atlas.MakeUnmarshalTransformFunc(
		func(x string) (WareID, error) {
			return ParseWareID(x)
		})).
	Complete()

type AbsPath string // Identifier for output slots.  Coincidentally, a path.

type (
	/*
		WarehouseAddr strings describe a protocol and dial address for talking to
		a storage warehouse.

		The serial format is an opaque string, though they typically resemble
		(and for internal use, are parsed as) URLs.
	*/
	WarehouseAddr string

	/*
		Configuration details for a warehouse.

		Many warehouses don't *need* any configuration; the addr string
		can tell the whole story.  But if you need auth or other fanciness,
		here's the place to specify it.
	*/
	WarehouseCfg struct {
		Auth     string      // auth info, if needed.  usually points to another file.
		Addr     interface{} // additional addr info, for protocols that require it.
		Priority int         // higher is checked first.
	}

	/*
		A suite of warehouses.  A transmat can take the entire set,
		and will select the ones it knows how to use, sort them,
		ping each in parallel, and start fetching from the most preferred
		one (or, from several, if it's a really smart protocol like that).
	*/
	WorkspaceWarehouseCfg map[WarehouseAddr]WarehouseCfg
)

/*
	FilesetFilters define how certain filesystem metadata should be treated
	when packing or unpacking files.

	They are stored as strings for simplicity of API, but are more like enums:

	UID and GID can be one of:

		- blank -- meaning "default behavior" (differs for pack and unpack;
		   it's like "1000" for pack and "mine" for unpack).
		   (When Repeatr drives Rio, it defaults to "keep" for unpack.)
		- "keep" -- meaning preserve the attributes of the fileset (in packing)
		   or manifest exactly what the ware specifies (in unpacking).
		- "mine" -- which means to ignore the ware attributes and use the current
		   uid/gid instead (this is only valid for unpacking).
		- an integer -- which means to treat everything as having exactly that
		   numeric uid/gid.

	Mtime can be one of:

		- blank -- meaning "default behavior" (differs for pack and unpack;
		   it's like "@25000" for pack and "keep" for unpack).
		- "keep" -- meaning preserve the attributes of the fileset (in packing)
		   or manifest exactly what the ware specifies (in unpacking).
		- "@" + an integer -- which means to set all times to the integer,
		   interpreted as a unix timestamp.
		- an RFC3339 date -- which means to set all times to that date
		   (and note that this will *not* survive serialization as such again;
		   it will be converted to the "@unix" format).

	Sticky is a simple bool: if true, setuid/setgid/sticky bits will be preserved
	on unpack.  The sticky bool has no meaning on pack; those bits are always packed.
	Repeatr always sets the sticky bool to true when using Rio,
	but it defaults to false when using Rio's command line.
	(For comparison, your system tar command tends to do the same:
	sticky bits are not unpacked by default because of the security implications
	if the user is unwary.)
*/
type FilesetFilters struct {
	Uid    string `refmt:"uid,omitempty"`
	Gid    string `refmt:"gid,omitempty"`
	Mtime  string `refmt:"mtime,omitempty"`
	Sticky bool   `refmt:"sticky,omitempty"`
}
