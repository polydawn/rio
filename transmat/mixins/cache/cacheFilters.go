package cache

import (
	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/util"
)

/*
	Predicate to check whether a filter will cause hash-alterating properties
	changes during an unpack (i.e., it's true if anything isn't "keep" mode).

	Filters which are unpack-hash-altering may result in more complex caching
	code (or, caching simply always missing in primitive implementations).
*/
func isUnpackAltering(filters api.FilesetFilters) bool {
	filt, err := apiutil.ProcessFilters(filters, apiutil.FilterPurposeUnpack)
	if err != nil {
		return true
	}
	if filt.Uid != apiutil.FilterKeep {
		return true
	}
	if filt.Gid != apiutil.FilterKeep {
		return true
	}
	if filt.Mtime != nil {
		return true
	}
	if !filt.Sticky {
		return true
	}
	return false
}
