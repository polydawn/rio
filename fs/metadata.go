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
	ModTime  time.Time   // modified time
	Xattrs   map[string]string
}
