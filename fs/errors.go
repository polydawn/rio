package fs

import "fmt"

/*
	Catchall error
*/
type ErrIO struct {
	Error error // this is not going to be roundtrippable.
}

/*
	Error returned when operating in a confined filesystem slice, but doing the

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

func (e ErrBreakout) Error() string {
	return fmt.Sprintf(
		"breakout error: refusing to traverse symlink at %q->%q while placing %q in %q",
		e.LinkPath, e.LinkTarget, e.OpPath, e.OpArea)
}
