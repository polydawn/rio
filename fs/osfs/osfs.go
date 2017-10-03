package osfs

import (
	"os"
	"strings"
	"syscall"

	. "github.com/polydawn/go-errcat"

	"go.polydawn.net/rio/fs"
)

func init() {
	syscall.Umask(0)
}

func New(basePath fs.AbsolutePath) fs.FS {
	return &osFS{basePath}
}

type osFS struct {
	basePath fs.AbsolutePath
}

func (afs *osFS) BasePath() fs.AbsolutePath {
	return afs.basePath
}

func (afs *osFS) OpenFile(path fs.RelPath, flag int, perms fs.Perms) (fs.File, error) {
	if path.GoesUp() {
		return nil, Errorf(fs.ErrBreakout, "fs: invalid path %q: must not depart basepath", path)
	}
	f, err := os.OpenFile(afs.basePath.Join(path).String(), flag, permsToOs(perms))
	return f, fs.NormalizeIOError(err)
}

func (afs *osFS) Mkdir(path fs.RelPath, perms fs.Perms) error {
	if path.GoesUp() {
		return Errorf(fs.ErrBreakout, "fs: invalid path %q: must not depart basepath", path)
	}
	err := os.Mkdir(afs.basePath.Join(path).String(), permsToOs(perms))
	return fs.NormalizeIOError(err)
}

func (afs *osFS) Mklink(path fs.RelPath, target string) error {
	if path.GoesUp() {
		return Errorf(fs.ErrBreakout, "fs: invalid path %q: must not depart basepath", path)
	}
	err := os.Symlink(target, afs.basePath.Join(path).String())
	return fs.NormalizeIOError(err)
}

func (afs *osFS) Mkfifo(path fs.RelPath, perms fs.Perms) error {
	if path.GoesUp() {
		return Errorf(fs.ErrBreakout, "fs: invalid path %q: must not depart basepath", path)
	}
	err := syscall.Mkfifo(afs.basePath.Join(path).String(), uint32(perms&07777))
	return fs.NormalizeIOError(err)
}

func (afs *osFS) MkdevBlock(path fs.RelPath, major int64, minor int64, perms fs.Perms) error {
	if path.GoesUp() {
		return Errorf(fs.ErrBreakout, "fs: invalid path %q: must not depart basepath", path)
	}
	mode := uint32(perms&07777) | syscall.S_IFBLK
	err := syscall.Mknod(afs.basePath.Join(path).String(), mode, int(devModesJoin(major, minor)))
	return fs.NormalizeIOError(err)
}

func (afs *osFS) MkdevChar(path fs.RelPath, major int64, minor int64, perms fs.Perms) error {
	if path.GoesUp() {
		return Errorf(fs.ErrBreakout, "fs: invalid path %q: must not depart basepath", path)
	}
	mode := uint32(perms&07777) | syscall.S_IFCHR
	err := syscall.Mknod(afs.basePath.Join(path).String(), mode, int(devModesJoin(major, minor)))
	return fs.NormalizeIOError(err)
}

func (afs *osFS) Lchown(path fs.RelPath, uid uint32, gid uint32) error {
	if path.GoesUp() {
		return Errorf(fs.ErrBreakout, "fs: invalid path %q: must not depart basepath", path)
	}
	err := os.Lchown(afs.basePath.Join(path).String(), int(uid), int(gid))
	return fs.NormalizeIOError(err)
}

func (afs *osFS) Chmod(path fs.RelPath, perms fs.Perms) error {
	if path.GoesUp() {
		return Errorf(fs.ErrBreakout, "fs: invalid path %q: must not depart basepath", path)
	}
	err := os.Chmod(afs.basePath.Join(path).String(), permsToOs(perms))
	return fs.NormalizeIOError(err)
}

func (afs *osFS) Stat(path fs.RelPath) (*fs.Metadata, error) {
	if path.GoesUp() {
		return nil, Errorf(fs.ErrBreakout, "fs: invalid path %q: must not depart basepath", path)
	}
	fi, err := os.Stat(afs.basePath.Join(path).String())
	if err != nil {
		return nil, fs.NormalizeIOError(err)
	}
	return afs.convertFileinfo(path, fi)
}

func (afs *osFS) LStat(path fs.RelPath) (*fs.Metadata, error) {
	if path.GoesUp() {
		return nil, Errorf(fs.ErrBreakout, "fs: invalid path %q: must not depart basepath", path)
	}
	fi, err := os.Lstat(afs.basePath.Join(path).String())
	if err != nil {
		return nil, fs.NormalizeIOError(err)
	}
	return afs.convertFileinfo(path, fi)
}

func (afs *osFS) convertFileinfo(path fs.RelPath, fi os.FileInfo) (*fs.Metadata, error) {
	// Copy over the easy 1-to-1 parts.
	fmeta := &fs.Metadata{
		Name:  path,
		Mtime: fi.ModTime(),
	}

	// Munge perms and mode to our types.
	fm := fi.Mode()
	switch fm & (os.ModeType | os.ModeCharDevice) {
	case 0:
		fmeta.Type = fs.Type_File
	case os.ModeDir:
		fmeta.Type = fs.Type_Dir
	case os.ModeSymlink:
		fmeta.Type = fs.Type_Symlink
		// If it's a symlink, get that info.
		//  It's an extra syscall, but we almost always want it.
		if target, _, err := afs.Readlink(path); err == nil {
			fmeta.Linkname = target
		} else {
			return nil, err
		}
	case os.ModeNamedPipe:
		fmeta.Type = fs.Type_NamedPipe
	case os.ModeSocket:
		fmeta.Type = fs.Type_Socket
	case os.ModeDevice:
		fmeta.Type = fs.Type_Device
	case os.ModeDevice | os.ModeCharDevice:
		fmeta.Type = fs.Type_CharDevice
	default:
		panic("unknown file mode")
	}
	fmeta.Perms = fs.Perms(fm.Perm())
	if fm&os.ModeSetuid != 0 {
		fmeta.Perms |= fs.Perms_Setuid
	}
	if fm&os.ModeSetgid != 0 {
		fmeta.Perms |= fs.Perms_Setgid
	}
	if fm&os.ModeSticky != 0 {
		fmeta.Perms |= fs.Perms_Sticky
	}

	// Copy over the size info... but only for file types.
	//  This is "system dependent" for others.  Knowing how many blocks a dir takes
	//  up is very rarely what we want...
	if fmeta.Type == fs.Type_File {
		fmeta.Size = fi.Size()
	}

	// Munge UID and GID bits.  These are platform dependent.
	// Also munge device bits if applicable; also platform dependent.
	if sys, ok := fi.Sys().(*syscall.Stat_t); ok {
		fmeta.Uid = sys.Uid
		fmeta.Gid = sys.Gid
		if fmeta.Type == fs.Type_Device || fmeta.Type == fs.Type_CharDevice {
			fmeta.Devmajor, fmeta.Devminor = devModesSplit(sys.Rdev)
		}
	}

	// Xattrs are not set by this method, because they require an unbounded
	//  number of additional syscalls (1 to list, $n to get values).

	return fmeta, nil
}

func (afs *osFS) ReadDirNames(path fs.RelPath) ([]string, error) {
	if path.GoesUp() {
		return nil, Errorf(fs.ErrBreakout, "fs: invalid path %q: must not depart basepath", path)
	}
	f, err := os.Open(afs.basePath.Join(path).String())
	if err != nil {
		return nil, fs.NormalizeIOError(err)
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return names, fs.NormalizeIOError(err)
	}
	return names, nil
}

func (afs *osFS) Readlink(path fs.RelPath) (string, bool, error) {
	if path.GoesUp() {
		return "", false, Errorf(fs.ErrBreakout, "fs: invalid path %q: must not depart basepath", path)
	}
	target, err := os.Readlink(afs.basePath.Join(path).String())
	switch {
	case err == nil:
		return target, true, nil
	case err.(*os.PathError).Err == syscall.EINVAL:
		// EINVAL means "not a symlink".
		// We return this as false and a nil error because it's frequently useful to use
		// the readlink syscall blindly with an lstat first in order to save a syscall.
		return "", false, nil
	default:
		return "", false, fs.NormalizeIOError(err)
	}
}

func (afs *osFS) ResolveLink(symlink string, startingAt fs.RelPath) (fs.RelPath, error) {
	if startingAt.GoesUp() {
		return startingAt, Errorf(fs.ErrBreakout, "fs: invalid path %q: must not depart basepath", startingAt)
	}
	return afs.resolveLink(symlink, startingAt, map[fs.RelPath]struct{}{})
}
func (afs *osFS) resolveLink(symlink string, startingAt fs.RelPath, seen map[fs.RelPath]struct{}) (fs.RelPath, error) {
	if _, isSeen := seen[startingAt]; isSeen {
		return startingAt, Errorf(fs.ErrRecursion, "cyclic symlinks detected from %q", startingAt)
	}
	seen[startingAt] = struct{}{}
	segments := strings.Split(symlink, "/")
	path := startingAt
	if segments[0] == "" { // rooted
		path = fs.RelPath{}
		segments = segments[1:]
	} else {
		path = startingAt.Dir() // drop the link node itself
	}
	iLast := len(segments) - 1
	for i, s := range segments {
		// Identity segments can simply be skipped.
		if s == "" || s == "." {
			continue
		}
		// Excessive up segements aren't an error; they simply no-op when already at root.
		if s == ".." && path == (fs.RelPath{}) {
			continue
		}
		// Okay, join the segment and peek at it.
		path = path.Join(fs.MustRelPath(s))
		// Bail on cycles before considering recursion!
		if path == startingAt {
			return startingAt, Errorf(fs.ErrRecursion, "cyclic symlinks detected from %q", startingAt)
		}
		// Check if this is a symlink; if so we must recurse on it.
		morelink, isLink, err := afs.Readlink(path)
		if err != nil {
			if i == iLast && Category(err) == fs.ErrNotExists {
				return path, nil
			}
			return startingAt, err
		}
		if isLink {
			path, err = afs.resolveLink(morelink, path, seen)
			if err != nil {
				return startingAt, err
			}
		}
	}
	return path, nil
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

func devModesJoin(major int64, minor int64) uint32 {
	// Constants herein are not a joy: they're a workaround for https://github.com/golang/go/issues/8106
	return uint32(((minor & 0xfff00) << 12) | ((major & 0xfff) << 8) | (minor & 0xff))
}

func devModesSplit(rdev uint64) (major int64, minor int64) {
	// Constants herein are not a joy: they're a workaround for https://github.com/golang/go/issues/8106
	return int64((rdev >> 8) & 0xff), int64((rdev & 0xff) | ((rdev >> 12) & 0xfff00))
}
