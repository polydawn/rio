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

func (afs *osFS) BasePath() fs.AbsolutePath {
	return afs.basePath
}

func (afs *osFS) OpenFile(path fs.RelPath, flag int, perms fs.Perms) (fs.File, fs.ErrFS) {
	f, err := os.OpenFile(afs.basePath.Join(path).String(), flag, permsToOs(perms))
	return f, ioError(err)
}

func (afs *osFS) Readlink(path fs.RelPath) (string, bool, fs.ErrFS) {
	target, err := os.Readlink(afs.basePath.Join(path).String())
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

func permsToOs(perms fs.Perms) (mode os.FileMode) {
	mode = os.FileMode(perms & 0777)
	if perms&04000 != 0 {
		mode |= os.ModeSetuid
	}
	if perms&02000 != 0 {
		mode |= os.ModeSetgid
	}
	if perms&01000 != 0 {
		mode |= os.ModeSticky
	}
	return mode
}
