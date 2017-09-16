package fs

import (
	"io"
	"time"
)

/*
	Interface for all primitive functions we expect to be able to perform
	on a filesystem.

	All paths accepted are RelPath types; typically the FS instance
	is constructed with an AbsolutePath, and all further operations are
	joined with that base path.
*/
type FS interface {
	// The basepath this filesystem was constructed with.
	// This may be useful for debugging messages, but should not otherwise
	// be used; there's no reason, since all methods are consistently
	// using it internally.
	// May be '/' (the zero AbsolutePath) if not applicable.
	BasePath() AbsolutePath

	OpenFile(path RelPath, flag int, perms Perms) (File, error)

	Mkdir(path RelPath, perms Perms) error

	Mklink(path RelPath, target string) error

	Mkfifo(path RelPath, perms Perms) error

	MkdevBlock(path RelPath, major int64, minor int64, perms Perms) error

	MkdevChar(path RelPath, major int64, minor int64, perms Perms) error

	Lchown(path RelPath, uid uint32, gid uint32) error

	Chmod(path RelPath, perms Perms) error

	SetTimesLNano(path RelPath, mtime time.Time, atime time.Time) error

	SetTimesNano(path RelPath, mtime time.Time, atime time.Time) error

	LStat(path RelPath) (*Metadata, error)

	ReadDirNames(path RelPath) ([]string, error)

	Readlink(path RelPath) (target string, isSymlink bool, err error)
}

type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Writer
	io.WriterAt
}
