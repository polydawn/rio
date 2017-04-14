package tartrans

import (
	"go.polydawn.net/rio"
	"go.polydawn.net/rio/fs"
)

// ORG: probably makes sense to share this ApplyFilters function with other packages,
// but it doesn't belong in the 'fs' package: fs concepts don't know about rio filters.
// (And your strongly worded hint there is it would be an import cycle, too.)

// Mutate the given fmeta handle to apply filters.
func ApplyFilters(fmeta *fs.Metadata, filters rio.Filters) {
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
