package fs

import (
	"os"
	"time"


	"go.polydawn.net/go-timeless-api"
)

type Metadata struct {
	Name     RelPath   // filename
	Type     Type      // type enum
	Perms    Perms     // permission bits
	Uid      uint32    // user id of owner
	Gid      uint32    // group id of owner
	Size     int64     // length in bytes
	Linkname string    // if symlink: target name of link
	Devmajor int64     // major number of character or block device
	Devminor int64     // minor number of character or block device
	Mtime    time.Time // modified time
	Xattrs   map[string]string

	// Notably absent fields:
	//  - ctime -- it's pointless to keep; you can't set such a thing in any posix filesystem.
	//  - atime -- similarly pointless; you can set it, but maybe, with asterisks, and it's
	//     almost certain end up trampled again moments later.
}

// Mode computes the expected os.FileMode for the described Metadata Perms and Type.
func (m *Metadata) Mode() os.FileMode {
	mode := os.FileMode(0)
	if m.Type == Type_Dir {
		mode |= os.ModeDir
	}
	if m.Type == Type_Symlink {
		mode |= os.ModeSymlink
	}
	if m.Type == Type_NamedPipe {
		mode |= os.ModeNamedPipe
	}
	if m.Type == Type_Socket {
		mode |= os.ModeSocket
	}
	if m.Type == Type_Device {
		mode |= os.ModeDevice
	}
	if m.Type == Type_CharDevice {
		mode |= os.ModeCharDevice
	}
	mode |= (os.FileMode(m.Perms) & os.ModePerm)
	if m.Perms&Perms_Setuid != 0 {
		mode |= os.ModeSetuid
	}
	if m.Perms&Perms_Setgid != 0 {
		mode |= os.ModeSetgid
	}
	if m.Perms&Perms_Sticky != 0 {
		mode |= os.ModeSticky
	}
	return mode
}

/*
	The usual posix permission bits (0777) plus the linux interpretation
	of the setuid, setgid, and sticky bits.

	See http://man7.org/linux/man-pages/man2/open.2.html for more information;
	specifically, store the setuid, setgid, and sticky bits with the same bits
	as documented for S_ISUID, S_ISGID, and S_ISVTX.
	http://pubs.opengroup.org/onlinepubs/7908799/xsh/sysstat.h.html documents
	that any choice for those three bits is valid as long as they do not conflict
	with the 0777 range, but, why?

	More precisely:

		#define S_ISUID  0004000
		#define S_ISGID  0002000
		#define S_ISVTX  0001000

	So, '01777' is as it is in your linux system chmod: a sticky bit is set.

	Compared to os.FileMode, this conveys the same amount of information as
	`mode & (ModePerm|ModeSetuid|ModeSetgid|ModeSticky)`, but again note that
	we follow the bit layout for those additional modes that is standard in
	the linux headers.
*/
type Perms uint16

const (
	Perms_Setuid Perms = 0004000
	Perms_Setgid Perms = 0002000
	Perms_Sticky Perms = 0001000
)

type Type uint8

const (
	Type_Invalid    Type = 0
	Type_File       Type = 'f'
	Type_Dir        Type = 'd'
	Type_Symlink    Type = 'L'
	Type_NamedPipe  Type = 'p'
	Type_Socket     Type = 'S'
	Type_Device     Type = 'D'
	Type_CharDevice Type = 'c'
	Type_Hardlink   Type = 'h' // Rare, and may only appear contextually.
)

func (t Type) String() string {
	switch t {
	case Type_File:
		return "file"
	case Type_Dir:
		return "dir"
	case Type_Symlink:
		return "symlink"
	case Type_NamedPipe:
		return "fifo"
	case Type_Socket:
		return "socket"
	case Type_Device:
		return "device"
	case Type_CharDevice:
		return "chardev"
	case Type_Hardlink:
		return "hardlink"
	case Type_Invalid:
		fallthrough
	default:
		return "<invalid type>"
	}
}

// Use this for the accessTime attribute when one is needed but no more
// obvious value is at hand, or for modifiedTime when doing things like
// 'create a default dir' when no other value is at hand.
var DefaultTime = time.Unix(api.DefaultTime, 0).UTC()
