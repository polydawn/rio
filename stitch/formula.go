package stitch

import (
	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/rio/fs"
)

/*
	Reduce a formula to a slice of []UnpackSpec, ready to be used
	invoking Assembler.Run().

	Typically the filters arg will be either `api.Filter_NoMutation` or
	`api.Filter_LowPriv`, depending if you're using repeatr or rio respectively,
	though other values are of course valid.

	Whether the action, outputs, or saveUrls are set is irrelevant;
	they will be ignored completely.
*/
func FormulaToUnpackSpecs(frm api.Formula, filters api.FilesetFilters) (parts []UnpackSpec) {
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

/*
	Reduce a formula to a slice of []PackSpec, ready to be used
	invoking PackMulti().

	Whether the action, inputs, or fetchUrls are set is irrelevant;
	they will be ignored completely.
*/
func FormulaToPackSpecs(frm api.Formula) (parts []PackSpec) {
	for path, output := range frm.Outputs {
		warehouse, _ := frm.SaveUrls[path]
		parts = append(parts, PackSpec{
			Path:      fs.MustAbsolutePath(string(path)),
			PackType:  output.PackType,
			Filters:   output.Filters,
			Warehouse: warehouse,
		})
	}
	return
}
