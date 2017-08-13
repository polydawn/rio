/*
	"Hal, prena meme!"
	"I'm sorry Dave, I can't do that"
*/
package halprenameme

import (
	"fmt"
	"strconv"
	"time"

	"go.polydawn.net/timeless-api"
)

type UnfuckMode bool

const (
	PackMode   UnfuckMode = false
	UnpackMode UnfuckMode = true
)

const (
	Keep = -1
	Mine = -2
)

var (
	DefaultUid   int = 1000
	DefaultGid   int = 1000
	DefaultMtime     = time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)
)

func UnfuckFilters(ff api.FilesetFilters, mode UnfuckMode) (uf UsableFilters, err error) {
	// Parse UID.
	switch ff.Uid {
	case "":
		switch mode {
		case PackMode:
			uf.Uid = DefaultUid
		case UnpackMode:
			uf.Uid = Mine
		}
	case "keep":
		uf.Uid = Keep
	case "mine":
		switch mode {
		case PackMode:
			return uf, fmt.Errorf("filter UID cannot use 'mine' mode: only makes sense when unpacking")
		case UnpackMode:
			uf.Uid = Mine
		}
	default:
		uf.Uid, err = strconv.Atoi(ff.Uid)
		if err != nil || uf.Uid < 0 {
			return uf, fmt.Errorf("filter UID must be one of 'keep', 'mine', or a positive int")
		}
	}

	// Parse GID.
	switch ff.Gid {
	case "":
		switch mode {
		case PackMode:
			uf.Gid = DefaultGid
		case UnpackMode:
			uf.Gid = Mine
		}
	case "keep":
		uf.Gid = Keep
	case "mine":
		switch mode {
		case PackMode:
			return uf, fmt.Errorf("filter GID cannot use 'mine' mode: only makes sense when unpacking")
		case UnpackMode:
			uf.Gid = Mine
		}
	default:
		uf.Gid, err = strconv.Atoi(ff.Gid)
		if err != nil || uf.Gid < 0 {
			return uf, fmt.Errorf("filter GID must be one of 'keep', 'mine', or a positive int")
		}
	}

	// Parse time.
	switch {
	case ff.Mtime == "":
		switch mode {
		case PackMode:
			uf.Mtime = &DefaultMtime
		case UnpackMode:
			uf.Mtime = nil // 'keep'
		}
	case ff.Mtime == "keep":
		uf.Mtime = nil
	case ff.Mtime[1] == '@':
		ut, err := strconv.Atoi(ff.Mtime[1:])
		if err != nil {
			return uf, fmt.Errorf("filter mtime parameter starting with '@' must be unix timestamp integer")
		}
		*uf.Mtime = time.Unix(int64(ut), 0)
	default:
		*uf.Mtime, err = time.Parse(time.RFC3339, ff.Mtime)
		if err != nil {
			return uf, fmt.Errorf("filter mtime parameter must be either 'keep', a unix timestamp integer beginning with '@', or an RFC3339 date string")
		}
	}

	// Sticky, mercy me, is simple.
	uf.Sticky = ff.Sticky

	return
}

type UsableFilters struct {
	Uid    int        // -1 for "keep", -2 for "mine"
	Gid    int        // -1 for "keep", -2 for "mine"
	Mtime  *time.Time // nil for "keep"
	Sticky bool
}
