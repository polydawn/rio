package fs

import "io"

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

	OpenFile(path RelPath, flag int, perms Perms) (File, ErrFS)

	Readlink(path RelPath) (target string, isSymlink bool, err ErrFS)
}

type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Writer
	io.WriterAt
}
