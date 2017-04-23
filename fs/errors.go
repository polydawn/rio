package fs

import "fmt"

/*
	Grouping interface for errors returned from filesystem operations.

	Implementors are assuring you that they will easily `json.Marshal`
	(or other formats of your choice) AND roundtrip unmarshal
	by virtue of exporting all their fields.
*/
type ErrFS interface {
	error
	_errFS()
}

/*
	Catchall error if an underlying layer returned an error we couldn't normalize.
*/
type ErrIOUnknown struct {
	// The `err.Error()` stringification of the original error.
	//
	// We flatten this immediately so that if this struct is serialized,
	// it's round-trippable.
	Msg string
}

func (ErrIOUnknown) _errFS()         {}
func (e ErrIOUnknown) Error() string { return e.Msg }

/*
	The normalization of anything matching `os.NotExists`.
*/
type ErrNotExists struct {
	Path RelPath
}

func (ErrNotExists) _errFS() {}
func (e ErrNotExists) Error() string {
	return fmt.Sprintf("path %q does not exist", e.Path)
}

/*
	Error returned when operating in a confined filesystem slice and an
	operation performed would result in effects outside the area, e.g.
	calling `PlaceFile` with "./reasonable/path" but "./reasonable" happens
	to be a symlink to "../../.." -- the symlink is valid, but to traverse
	it would violate confinement.

	Not all functions which do symlink checks will verify if the symlink target
	is within the operational area; they may return ErrBreakout upon encountering
	any symlink, even if following it would still be within bounds.
	Check the function's documentation for more info on how it behaves.

	Note that any function returning ErrBreakout is, by nature, doing so in a
	best-effort sense: if there are concurrent modifcations to the operational
	area of the filesystem by any other processes, it is *impossible* to
	avoid a TOCTOU violation.
*/
type ErrBreakout struct {
	OpPath     RelPath
	OpArea     AbsolutePath
	LinkPath   RelPath
	LinkTarget string
}

func (ErrBreakout) _errFS() {}
func (e ErrBreakout) Error() string {
	return fmt.Sprintf(
		"breakout error: refusing to traverse symlink at %q->%q while placing %q in %q",
		e.LinkPath, e.LinkTarget, e.OpPath, e.OpArea)
}
