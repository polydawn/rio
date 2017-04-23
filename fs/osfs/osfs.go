package osfs

import (
	"os"
	"syscall"

	"go.polydawn.net/rio/fs"
)

func New(basePath fs.AbsolutePath) fs.FS {
	return &osFS{basePath}
}

type osFS struct {
	basePath fs.AbsolutePath
}

func (x *osFS) Readlink(path fs.RelPath) (string, bool, fs.ErrFS) {
	target, err := os.Readlink(x.basePath.Join(path).String())
	switch {
	case err == nil:
		return target, true, nil
	case os.IsNotExist(err):
		return "", false, &fs.ErrNotExists{path}
	case err.(*os.PathError).Err == syscall.EINVAL:
		// EINVAL means "not a symlink".
		// We return this as false and a nil error because it's frequently useful to use
		// the readlink syscall blindly with an lstat first in order to save a syscall.
		return "", false, nil
	default:
		return "", false, ioError(err)
	}
}
