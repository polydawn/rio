package fs

import (
	"path"
	"strings"
)

/*
	Relative paths and absolute paths are not interchangeable.
	Using these path structures aids in keeping these ideas separate and
	hopefully prevent certain kinds of errors caused by accidentally using the incorrect path type.
	It is recommended to use absolute paths when possible.
*/
type RelPath struct {
	path      string
	lastSplit int
}

/*
	Converts a string to an relative path struct.
	Will panic if the given path is absolute.
*/
func MustRelPath(p string) RelPath {
	p = path.Clean(p)
	if p[0] == '/' {
		panic("fs: not a relative path")
	}
	if p == "." { // We can't stop people from using the zero value, so, use it.
		return RelPath{}
	}
	return RelPath{p, strings.LastIndexByte(p, '/')}
}

/*
	An relative path string with a zero value of "." for current directory
*/
func (p RelPath) String() string {
	if p.path == "" {
		return "."
	} else if p.path[0] == '.' { // a '..' prefix
		return p.path
	} else {
		return "./" + p.path
	}
}

/*
	Returns the relative path of the directory containing this path
*/
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

/*
	Returns the filename of this path
*/
func (p RelPath) Last() string {
	if p.path == "" {
		return "."
	} else if p.lastSplit == -1 {
		return p.path
	} else {
		return p.path[p.lastSplit+1:]
	}
}

/*
	Creates a new relative path by following the relative path from this path.
*/
func (p RelPath) Join(p2 RelPath) RelPath {
	switch {
	case p2.path == "":
		return p
	case p.path == "":
		return p2
	case p2.path[0] == '.': // '..' prefix requires cleaning again.
		pj := path.Clean(p.path + "/" + p2.path)
		if pj == "." { // We can't stop people from using the zero value, so, use it.
			return RelPath{}
		}
		return RelPath{pj, strings.LastIndexByte(pj, '/')}
	default:
		return RelPath{p.path + "/" + p2.path, len(p.path) + p2.lastSplit + 1}
	}
}

/*
	Relative paths and absolute paths are not interchangeable.
	Using these path structures aids in keeping these ideas separate and
	hopefully prevent certain kinds of errors caused by accidentally using the incorrect path type.
	It is recommended to use absolute paths when possible.
*/
type AbsolutePath struct {
	path      string
	lastSplit int
}

/*
	Converts a string to an absolute path struct.
	Will panic if the given path is not absolute.
*/
func MustAbsolutePath(p string) AbsolutePath {
	p = path.Clean(p)
	if p[0] != '/' {
		panic("fs: not an absolute path")
	}
	if p == "/" { // We can't stop people from using the zero value, so, use it.
		return AbsolutePath{}
	}
	return AbsolutePath{p, strings.LastIndexByte(p, '/')}
}

/*
	An absolute path string with a zero value of "/" for the root directory
*/
func (p AbsolutePath) String() string {
	if p.path == "" {
		return "/"
	}
	return p.path
}

/*
	Returns the absolute path of the directory containing this path
*/
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

/*
	Returns the filename of this path
*/
func (p AbsolutePath) Last() string {
	if p.path == "" {
		return "/"
	} else {
		return p.path[p.lastSplit+1:]
	}
}

/*
	Creates a new absolute path by following the relative path from this path.
*/
func (p AbsolutePath) Join(p2 RelPath) AbsolutePath {
	switch {
	case p2.path == "":
		return p
	//case p.path == "": // Comes out the same as the math in the default case.
	//	return AbsolutePath{"/" + p2.path, p2.lastSplit + 1}
	case p2.path[0] == '.': // '..' prefix requires cleaning again.
		pj := path.Clean(p.path + "/" + p2.path)
		if pj == "/" { // We can't stop people from using the zero value, so, use it.
			return AbsolutePath{}
		}
		return AbsolutePath{pj, strings.LastIndexByte(pj, '/')}
	default:
		return AbsolutePath{p.path + "/" + p2.path, len(p.path) + p2.lastSplit + 1}
	}
}
