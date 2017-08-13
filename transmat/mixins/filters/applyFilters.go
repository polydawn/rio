package filters

import (
	"go.polydawn.net/rio/fs"
	"go.polydawn.net/timeless-api/util"
)

/*
	Mutate the given fmeta handle to apply filters.

	Since this is the apiutil package's version of FilesetFilters,
	we can trust the values have been validated to reasonable ranges already,
	and defaults (for either pack or unpack mode) have already been mapped in.
*/
func Apply(filters apiutil.FilesetFilters, fmeta *fs.Metadata) {
	// Apply UID.
	switch filters.Uid {
	case apiutil.FilterKeep:
		// pass
	case apiutil.FilterMine:
		// TODO need some magic numberssss
	default:
		fmeta.Uid = uint32(filters.Uid)
	}

	// Apply GID.
	switch filters.Gid {
	case apiutil.FilterKeep:
		// pass
	case apiutil.FilterMine:
		// TODO need some magic numberssss
	default:
		fmeta.Gid = uint32(filters.Gid)
	}

	// Apply Mtime.
	if filters.Mtime != nil {
		fmeta.Mtime = *filters.Mtime
	}

	// Apply Sticky.
	if !filters.Sticky {
		fmeta.Perms &= 0777
	}
}
