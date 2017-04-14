package fs

import (
	"os"
	"time"
)

type Metadata struct {
	Name     RelPath     // filename
	Mode     os.FileMode // type, permission and mode bits
	Uid      int         // user id of owner
	Gid      int         // group id of owner
	Size     int64       // length in bytes
	Linkname string      // if symlink: target name of link
	Devmajor int64       // major number of character or block device
	Devminor int64       // minor number of character or block device
	Mtime    time.Time   // modified time
	Xattrs   map[string]string

	// Notably absent fields:
	//  - ctime -- it's pointless to keep; you can't set such a thing in any posix filesystem.
	//  - atime -- similarly pointless; you can set it, but maybe, with asterisks, and it's
	//     almost certain end up tramped again moments later.
}
