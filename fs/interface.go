package fs

/*
	Interface for all primitive functions we expect to be able to perform
	on a filesystem.

	All paths accepted are RelPath types; typically the FS instance
	is constructed with an AbsolutePath, and all further operations are
	joined with that base path.
*/
type FS interface {
}
