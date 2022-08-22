//go:build darwin || openbsd || netbsd
// +build darwin openbsd netbsd

// We needed linx-specific syscalls not exported by the standard lib in order to get
// chtimes on symlinks with nano precision to work correctly.
// (Stdlib only provides 'chtimes', no 'lchtimes'.)

package osfs

import (
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"

	"github.com/polydawn/rio/fs"
)

const (
	// These are not currently available in syscall
	// https://go.googlesource.com/sys/+/refs/heads/release-branch.go1.13/unix/zsysnum_darwin_amd64.go#428
	sysSetAttrListAt = uintptr(524)
	// https://fergofrog.com/code/cbowser/xnu/bsd/sys/attr.h.html#_M/ATTR_BIT_MAP_COUNT
	attrBitMapCount = uint16(5)
	// https://fergofrog.com/code/cbowser/xnu/bsd/sys/attr.h.html#_M/ATTR_CMN_CRTIME
	attrCmnModTime = uint32(0x400)
	attrCmnAccTime = uint32(0x1000)
)

type attrList struct {
	BitmapCount uint16
	_           uint16
	CommonAttr  uint32
	VolAttr     uint32
	DirAttr     uint32
	FileAttr    uint32
	Forkattr    uint32
}

// Translation to setattrlistat uses a subset of the logic at
// https://github.com/apple/darwin-xnu/blob/master/libsyscall/wrappers/utimensat.c
func prepareTimes(timesIn [2]syscall.Timespec) (uint32, [2]syscall.Timespec, uint32) {
	var attrs uint32
	var timesOutSize uint32
	// TODO: may need to generalize to support UTIME_NOW / UTIME_EMIT support.
	attrs = attrCmnModTime | attrCmnAccTime
	timesOutSize = uint32(unsafe.Sizeof(timesIn[0]) + unsafe.Sizeof(timesIn[1]))
	return attrs, timesIn, timesOutSize
}

func (afs *osFS) SetTimesLNano(path fs.RelPath, mtime time.Time, atime time.Time) error {
	rpath, err := afs.realpath(path, false)
	if err != nil {
		return err
	}

	var a attrList
	var attrbufSize uint32
	var timesIn [2]syscall.Timespec
	var timesOut [2]syscall.Timespec
	timesIn[0] = syscall.NsecToTimespec(mtime.UnixNano())
	timesIn[1] = syscall.NsecToTimespec(atime.UnixNano())

	var _path *byte
	_path, err = syscall.BytePtrFromString(rpath)
	if err != nil { // EINVAL if the path string contains NUL bytes.
		return fs.NormalizeIOError(err)
	}

	a.BitmapCount = attrBitMapCount
	a.CommonAttr, timesOut, attrbufSize = prepareTimes(timesIn)

	AtFDCWD := -100
	FSOptNoFollow := 0x1

	if _, _, err := syscall.Syscall6(sysSetAttrListAt, uintptr(AtFDCWD), uintptr(unsafe.Pointer(_path)), uintptr(unsafe.Pointer(&a)), uintptr(unsafe.Pointer(&timesOut)), uintptr(attrbufSize), uintptr(FSOptNoFollow)); err != 0 {
		return fs.NormalizeIOError(err)
	}

	// rio symlinks will be chmod 777 on mac, to behave equivalently to their linux (which doesn't support lchmod) counterparts
	return unix.Fchmodat(unix.AT_FDCWD, rpath, uint32(0777), unix.AT_SYMLINK_NOFOLLOW)
}

func (afs *osFS) SetTimesNano(path fs.RelPath, mtime time.Time, atime time.Time) error {
	rpath, err := afs.realpath(path, true)
	if err != nil {
		return err
	}

	// Note that this is disambiguated from plain `os.Chtimes` only in that it refuses to fall back to lower precision on old kernels.
	// Like LUtimesNano, it depends on kernel 2.6.22 or newer.
	var utimes [2]syscall.Timespec
	utimes[0] = syscall.NsecToTimespec(atime.UnixNano())
	utimes[1] = syscall.NsecToTimespec(mtime.UnixNano())
	if err := syscall.UtimesNano(rpath, utimes[0:]); err != nil {
		return fs.NormalizeIOError(err)
	}
	return nil
}
