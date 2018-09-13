package stitch

// FIXME : This whole file is frankly more repeatr's concerns than rio's.
// FIXME : We should thus move its contents to the repeatr repo.
// FIXME : It is now commented out here and will be removed when that's done.

//import (
//	"go.polydawn.net/go-timeless-api"
//	"go.polydawn.net/rio/fs"
//)

///*
//	Reduce a formula to a slice of []UnpackSpec, ready to be used
//	invoking Assembler.Run().

//	Typically the filters arg will be either `api.FilesetUnpackFilter_Lossless`
//	or `api.FilesetUnpackFilter_LowPriv`, depending if you're using repeatr or
//	rio respectively, though other values are of course valid.

//	Whether the action, outputs, or saveUrls are set is irrelevant;
//	they will be ignored completely.
//*/
//func FormulaToUnpackSpecs(frm api.Formula, sourcing api.WareSourcing, filters api.FilesetUnpackFilter) (parts []UnpackSpec) {
//	for path, wareID := range frm.Inputs {
//		warehouses, _ := frmCtx.FetchUrls[path]
//		parts = append(parts, UnpackSpec{
//			Path:     fs.MustAbsolutePath(string(path)),
//			WareID:   wareID,
//			Filter:   filters,
//			Sourcing: sourcing,
//		})
//	}
//	return
//}

///*
//	Reduce a formula to a slice of []PackSpec, ready to be used
//	invoking PackMulti().

//	The filters given will be applied to all *unset* fields in the filters
//	already given by the formula outputs; set fields are not changed.
//	Typically the filters arg will be `api.Filter_DefaultFlatten`
//	(both `rio pack` and repeatr outputs default to this),
//	though other values are of course valid.

//	Whether the action, inputs, or fetchUrls are set is irrelevant;
//	they will be ignored completely.
//*/
//func FormulaToPackSpecs(frm api.Formula, frmCtx api.FormulaContext, filters api.FilesetFilter) (parts []PackSpec) {
//	for path, output := range frm.Outputs {
//		warehouse, _ := frmCtx.SaveUrls[path]
//		parts = append(parts, PackSpec{
//			Path:      fs.MustAbsolutePath(string(path)),
//			PackType:  output.PackType,
//			Filter:    apiutil.MergeFilter(output.Filter, filters),
//			Warehouse: warehouse,
//		})
//	}
//	return
//}
