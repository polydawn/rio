package stitch

import (
	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/rio/fs"
)

/*
	Reduce a formula to a slice of unpackSpecs, ready to be used
	invoking Assembler.Run().

	Typically the filters arg will be either `api.Filter_NoMutation` or
	`api.Filter_LowPriv`, depending if you're using repeatr or rio respectively,
	though other values are of course valid.

	Whether the action, outputs, or saveUrls are set is irrelevant;
	they will be ignored completely.
*/
func FormulaToUnpackTree(frm api.Formula, filters api.FilesetFilters) (parts []UnpackSpec) {
	for path, wareID := range frm.Inputs {
		warehouses, _ := frm.FetchUrls[path]
		parts = append(parts, UnpackSpec{
			Path:       fs.MustAbsolutePath(string(path)),
			WareID:     wareID,
			Filters:    filters,
			Warehouses: warehouses,
		})
	}
	return
}
