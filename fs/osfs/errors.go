package osfs

import (
	"go.polydawn.net/rio/fs"
)

// Attempt to normalize.
func ioError(err error) fs.ErrFS {
	return fs.ErrIOUnknown{err.Error()}
}
