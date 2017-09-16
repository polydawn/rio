package stitch

import (
	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/rio/fs"
)

/*
	Struct to gather the args for a single rio.Unpack func call.
	(The context object and monitors are handled in a different band.)

	Note the similar name to a structure in the go-timeless-api packages;
	this one is not serializable, is internal, and
	contains the literal set of warehouses already resolved,
	as well as the path inline rather than in a map key, so we can sort slices.
*/
type UnpackSpec struct {
	Path       fs.AbsolutePath
	WareID     api.WareID
	Filters    api.FilesetFilters
	Warehouses []api.WarehouseAddr
}

// Cast slices to this type to sort by target path (which is effectively mountability order).
type UnpackSpecByPath []UnpackSpec

func (a UnpackSpecByPath) Len() int           { return len(a) }
func (a UnpackSpecByPath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a UnpackSpecByPath) Less(i, j int) bool { return a[i].Path.String() < a[j].Path.String() }
