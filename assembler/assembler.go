package assembler

import (
	"go.polydawn.net/rio/fs"
)

type AssemblyPart struct {
	// Path to use as source for mounting or copy.
	//
	// When assembly is invoked, a source root is required; this
	// path will be relative to that.
	// (Typically the source root with be the CAS cache storage area
	// when this is used by repeatr to launch containers.)
	SourcePath fs.RelPath

	// Path to fill via mounting or copy.
	//
	// When assembly is invoked, a target root is required; this
	// path will be relative to that.
	// (Typically the target root is the path a container will be chroot'd
	// into when this is used by repeatr to launch containers.)
	TargetPath fs.RelPath

	// Toggles whether or not any mounts should be writable.
	// Default is false/read-only.
	//
	// If using a "copy" placer, this will be *ignored*: the paths
	// WILL be writable.  Do not use copy placers if this is an
	// issue; they do not have the power to respect this property.
	//
	// If `Writable && BareMount` are both true, then
	// *this will allow modifications to the source path*.
	// (This is used by repeatr to allow mounts through to the host.
	//
	// TODO REVIEW: host mounts require that the source root not be the
	// CAS cache root path :/  maybe we shoud allow a root per entry...?
	//
	// TODO REVIEW: the behavior for copy placers.  I'd rather they errored loudly.
	// It would be a change from historic repeatr, but probably a worthy one.
	Writable bool

	// BareMount requests direct passthrough -- what this means varies based on writability:
	// if writable==false, this means continuing changes in sourcepath are visible realtime;
	// if writable==true, the placer will employ a bind mount *without* COW or isolation,
	// **meaning mutations will be applied to the sourcepath**.
	BareMount bool
}

// TODO REVIEW: overall, above: yeah, maybe we need some more usage-specific enums,
// rather than this hash of bools.

//type Assembler struct {
//	targetBase     fs.AbsolutePath
//	cacheDir       fs.AbsolutePath
//	unpackTool     rio.UnpackFunc
//	placerTool     func( /*todo*/ )
//	fillerDirProps fs.Metadata
//}
