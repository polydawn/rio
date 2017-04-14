package tartrans

import (
	"go.polydawn.net/rio"
	"go.polydawn.net/rio/fs"
)

// TODO ORG: probably makes sense to share these Apply*Filters funcs with other packages,
// but it doesn't belong in the 'fs' package: fs concepts don't know about rio filters.
// (And your strongly worded hint there is it would be an import cycle, too.)

// Mutate the given fmeta handle to apply filters,
// using the default behaviors appropriate for materializer
// (e.g. keep-by-default).
func ApplyMaterializeFilters(fmeta *fs.Metadata, filters rio.Filters) {
	if !filters.FlattenUID.Keep {
		if filters.FlattenUID.Value != nil {
			fmeta.Uid = *filters.FlattenUID.Value
		}
	}
	if !filters.FlattenGID.Keep {
		if filters.FlattenGID.Value != nil {
			fmeta.Gid = *filters.FlattenGID.Value
		}
	}
	if !filters.FlattenMtime.Keep {
		if filters.FlattenMtime.Value != nil {
			fmeta.Mtime = *filters.FlattenMtime.Value
		}
	}
}

// Mutate the given fmeta handle to apply filters,
// using the default behaviors appropriate for scanning
// (e.g. flatten-by-default).
func ApplyScanFilters(fmeta *fs.Metadata, filters rio.Filters) {
	if !filters.FlattenUID.Keep {
		if filters.FlattenUID.Value != nil {
			fmeta.Uid = *filters.FlattenUID.Value
		} else {
			fmeta.Uid = rio.FilterDefaultUid
		}
	}
	if !filters.FlattenGID.Keep {
		if filters.FlattenGID.Value != nil {
			fmeta.Gid = *filters.FlattenGID.Value
		} else {
			fmeta.Gid = rio.FilterDefaultGid
		}
	}
	if !filters.FlattenMtime.Keep {
		if filters.FlattenMtime.Value != nil {
			fmeta.Mtime = *filters.FlattenMtime.Value
		} else {
			fmeta.Mtime = rio.FilterDefaultMtime
		}
	}
}
