package fs

import (
	"path"
	"strings"
)

// Meta: yep, these *are not* interchangeable.
// It's expected that if you *can* accept an AbsolutePath,
//  then you should normalize to that ASAP;
// and if you can't, then clearly it's correct to use the RelPath,
//  through and through the whole way.

type RelPath struct {
	path      string
	lastSplit int
}

func MustRelPath(p string) RelPath {
	p = path.Clean(p)
	if p[0] == '/' {
		panic("nope")
	}
	if p == "." { // We can't stop people from using the zero value, so, use it.
		return RelPath{}
	}
	return RelPath{p, strings.LastIndexByte(p, '/')}
}
func (p RelPath) String() string {
	if p.path == "" {
		return "."
	} else if p.path[0] == '.' { // a '..' prefix
		return p.path
	} else {
		return "./" + p.path
	}
}
func (p RelPath) Dir() RelPath {
	if p.path == "" {
		return p
	} else if p.lastSplit == -1 {
		return RelPath{}
	} else {
		p2 := p.path[0:p.lastSplit]
		return RelPath{p2, strings.LastIndexByte(p2, '/')}
	}
}
func (p RelPath) Last() string {
	if p.path == "" {
		return "."
	} else if p.lastSplit == -1 {
		return p.path
	} else {
		return p.path[p.lastSplit+1:]
	}
}
func (p RelPath) Join(p2 RelPath) RelPath {
	switch {
	case p2.path == "":
		return p
	case p.path == "":
		return p2
	default:
		return RelPath{p.path + "/" + p2.path, len(p.path) + p2.lastSplit + 1}
	}
}

type AbsolutePath struct {
	path      string
	lastSplit int
}

func MustAbsolutePath(p string) AbsolutePath {
	p = path.Clean(p)
	if p[0] != '/' {
		panic("nope")
	}
	if p == "/" { // We can't stop people from using the zero value, so, use it.
		return AbsolutePath{}
	}
	return AbsolutePath{p, strings.LastIndexByte(p, '/')}
}
func (p AbsolutePath) String() string {
	if p.path == "" {
		return "/"
	}
	return p.path
}
func (p AbsolutePath) Dir() AbsolutePath {
	if p.path == "" {
		return p
	} else if p.lastSplit == 0 {
		return AbsolutePath{}
	} else {
		p2 := p.path[0:p.lastSplit]
		return AbsolutePath{p2, strings.LastIndexByte(p2, '/')}
	}
}
func (p AbsolutePath) Last() string {
	if p.path == "" {
		return "/"
	} else {
		return p.path[p.lastSplit+1:]
	}
}
func (p AbsolutePath) Join(p2 RelPath) AbsolutePath {
	switch {
	case p2.path == "":
		return p
	//case p.path == "": // Comes out the same as the math below.
	//	return AbsolutePath{"/" + p2.path, p2.lastSplit + 1}
	default:
		return AbsolutePath{p.path + "/" + p2.path, len(p.path) + p2.lastSplit + 1}
	}
}
