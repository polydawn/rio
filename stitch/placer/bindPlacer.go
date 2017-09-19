package placer

import (
	"os"
	"syscall"

	. "github.com/polydawn/go-errcat"

	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fsOp"
)

var _ Placer = BindPlacer

/*
	Makes files appear in place by use of a bind mount.

	If writable=true, the *source* will be mutable.  If you want the destination
	to be writable, but do not want the source to be mutable, then
	you need a placer like "aufs" or "overlay".
*/
func BindPlacer(srcPath, dstPath fs.AbsolutePath, writable bool) error {
	// Determine desired type.
	srcStat, err := rootFs.LStat(srcPath.CoerceRelative())
	if err != nil {
		return Errorf(rio.ErrLocalCacheProblem, "error placing with bind mount: %s", err)
	}
	switch srcStat.Type {
	case fs.Type_File: // pass
	case fs.Type_Dir: // pass
	default:
		return Errorf(rio.ErrAssemblyInvalid, "placer: source may only be dir or plain file (%s is %s)", srcPath)
	}

	// Make the destination path exist and be the right type to mount over.
	mkDest(dstPath, srcStat.Type)

	// Make mount syscall to bind, and optionally then push it to readonly.
	//  Works the same for dirs or files.
	flags := syscall.MS_BIND | syscall.MS_REC
	if err := syscall.Mount(srcPath.String(), dstPath.String(), "bind", uintptr(flags), ""); err != nil {
		return Errorf(rio.ErrAssemblyInvalid, "error placing with bind mount: %s", err)
	}
	if !writable {
		flags |= syscall.MS_RDONLY | syscall.MS_REMOUNT
		if err := syscall.Mount(srcPath.String(), dstPath.String(), "bind", uintptr(flags), ""); err != nil {
			return Errorf(rio.ErrAssemblyInvalid, "error placing with bind mount: %s", err)
		}
	}
	return nil
}

func mkDest(dstPath fs.AbsolutePath, wantType fs.Type) error {
	// Handle all the cases for existing things at destination.
	dstStat, err := rootFs.LStat(dstPath.CoerceRelative())
	switch Category(err) {
	case nil: // It exists.  But is it the right type?
		if dstStat.Type == wantType {
			return nil // Already matches.  Huzzah.
		}
		return Errorf(rio.ErrAssemblyInvalid, "placer: destination already exists and is different type than source")
	case fs.ErrNotExists:
		// Carry on.  We'll create it.
	default: // Any other error: raise.
		return Errorf(rio.ErrAssemblyInvalid, "placer: destination unusable: %s", err)
	}

	// If we made it this far: dest doesn't exist yet.
	// Capture the parent dir mtime, because we're about to disrupt it.
	defer fsOp.RepairMtime(rootFs, dstPath.Dir().CoerceRelative())()

	// Make the dest node, matching type of the source.
	// The perms don't matter; will be shadowed.
	// We assume the parent dirs are all in place because you're almost
	// certainly using this as part of an assembler.
	switch wantType {
	case fs.Type_File:
		var f *os.File
		f, err = os.OpenFile(dstPath.String(), os.O_CREATE, 0644)
		f.Close()
	case fs.Type_Dir:
		err = os.Mkdir(dstPath.String(), 0644)
	}
	if err != nil {
		return Errorf(rio.ErrAssemblyInvalid, "placer: destination unusable: %s", err)
	}
	return nil
}
