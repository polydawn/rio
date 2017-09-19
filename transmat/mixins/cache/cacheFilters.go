package cache

import (
	"go.polydawn.net/go-timeless-api"
)

/*
	Predicate to check whether a filter will cause hash-alterating properties
	changes during an unpack (i.e., it's true if anything isn't "keep" mode).

	Filters which are unpack-hash-altering may result in more complex caching
	code (or, caching simply always missing in primitive implementations).
*/
func isUnpackAltering(filters api.FilesetFilters) bool {
	if filters.Uid != "keep" {
		return true
	}
	if filters.Gid != "keep" {
		return true
	}
	if filters.Mtime != "" && filters.Mtime != "keep" {
		return true
	}
	if !filters.Sticky {
		return true
	}
	return false
}
