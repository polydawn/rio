package filters

import (
	"os"
	"strconv"

	"github.com/warpfork/go-errcat"
	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/rio/fs"
)

var (
	myUid = uint32(os.Getuid())
	myGid = uint32(os.Getgid())
)

/*
	ApplyPackFilter mutates the given fs.Metadata to be filtered.

	An error is returned if any of of the metadata is matched by filters
	set to "reject".

	The fmeta.Type will be set to fs.Type_Invalid in the case of a node which
	is matched by filters set to "ignore" if the filter is for a critical
	property of the node (e.g. "dev=ignore" will trigger this; whereas
	"sticky=ignore" will simply strip the sticky bit from the perm bits).
	This means you should just skip creation of this node entirely.

	Note that if the filter has any fields which are set of "unspecified",
	they will be ignored; you should probably check `IsComplete` on the filter
	before calling this.  (We don't do it inside of here because you're
	probably using this in a loop of large cardinality.)
*/
func ApplyPackFilter(ff api.FilesetPackFilter, fmeta *fs.Metadata) error {
	if keep, setTo := ff.Uid(); !keep {
		fmeta.Uid = uint32(setTo)
	}
	if keep, setTo := ff.Gid(); !keep {
		fmeta.Gid = uint32(setTo)
	}
	if keep, setTo := ff.Mtime(); !keep {
		fmeta.Mtime = setTo
	}
	if keep := ff.Sticky(); !keep {
		fmeta.Perms &= ^fs.Perms_Sticky
	}
	if keep, reject := ff.Setid(); reject {
		if fmeta.Perms&(fs.Perms_Setuid|fs.Perms_Setgid) != 0 {
			return errcat.ErrorDetailed(
				rio.ErrFilterRejection,
				"filter rejection: setid bits",
				map[string]string{
					"path":  fmeta.Name.String(),
					"perms": "0" + strconv.FormatUint(uint64(fmeta.Perms), 8),
				},
			)
		}
	} else if !keep {
		fmeta.Perms &= ^(fs.Perms_Setuid | fs.Perms_Setgid)
	}
	if keep, reject := ff.Dev(); reject {
		if fmeta.Type == fs.Type_Device || fmeta.Type == fs.Type_CharDevice {
			return errcat.ErrorDetailed(
				rio.ErrFilterRejection,
				"filter rejection: device node",
				map[string]string{
					"path":   fmeta.Name.String(),
					"type":   string(fmeta.Type),
					"majmin": strconv.FormatInt(fmeta.Devmajor, 10) + "," + strconv.FormatInt(fmeta.Devminor, 10),
				},
			)
		}
	} else if !keep {
		fmeta.Type = fs.Type_Invalid
	}
	return nil
}

/*
	ApplyUnpackFilter mutates the given fs.Metadata to be filtered.

	An error is returned if any of of the metadata is matched by filters
	set to "reject".

	The fmeta.Type will be set to fs.Type_Invalid in the case of a node which
	is matched by filters set to "ignore" if the filter is for a critical
	property of the node (e.g. "dev=ignore" will trigger this; whereas
	"sticky=ignore" will simply strip the sticky bit from the perm bits).
	This means you should just skip creation of this node entirely.

	Note that if the filter has any fields which are set of "unspecified",
	they will be ignored; you should probably check `IsComplete` on the filter
	before calling this.  (We don't do it inside of here because you're
	probably using this in a loop of large cardinality.)
*/
func ApplyUnpackFilter(ff api.FilesetUnpackFilter, fmeta *fs.Metadata) error {
	if follow, mine, setTo := ff.Uid(); mine {
		fmeta.Uid = myUid
	} else if !follow {
		fmeta.Uid = uint32(setTo)
	}
	if follow, mine, setTo := ff.Gid(); mine {
		fmeta.Gid = myGid
	} else if !follow {
		fmeta.Gid = uint32(setTo)
	}
	if follow, now, setTo := ff.Mtime(); now {
		panic("unpack filter mtime=now not yet supported")
	} else if !follow {
		fmeta.Mtime = setTo
	}
	if follow := ff.Sticky(); !follow {
		fmeta.Perms &= ^fs.Perms_Sticky
	}
	if follow, reject := ff.Setid(); reject {
		if fmeta.Perms&(fs.Perms_Setuid|fs.Perms_Setgid) != 0 {
			return errcat.ErrorDetailed(
				rio.ErrFilterRejection,
				"filter rejection: setid bits",
				map[string]string{
					"path":  fmeta.Name.String(),
					"perms": "0" + strconv.FormatUint(uint64(fmeta.Perms), 8),
				},
			)
		}
	} else if !follow {
		fmeta.Perms &= ^(fs.Perms_Setuid | fs.Perms_Setgid)
	}
	if follow, reject := ff.Dev(); reject {
		if fmeta.Type == fs.Type_Device || fmeta.Type == fs.Type_CharDevice {
			return errcat.ErrorDetailed(
				rio.ErrFilterRejection,
				"filter rejection: device node",
				map[string]string{
					"path":   fmeta.Name.String(),
					"type":   string(fmeta.Type),
					"majmin": strconv.FormatInt(fmeta.Devmajor, 10) + "," + strconv.FormatInt(fmeta.Devminor, 10),
				},
			)
		}
	} else if !follow {
		fmeta.Type = fs.Type_Invalid
	}
	return nil
}
