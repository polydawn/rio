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
	// Make the destination path exist and be the right type to mount over.
	mkDest(srcPath, dstPath)

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

func mkDest(srcPath, dstPath fs.AbsolutePath) error {
	// Determine desired type.
	srcStat, err := os.Stat(srcPath.String())
	if err != nil {
		return Errorf(rio.ExitLocalCacheProblem, "error placing with bind mount: %s", err)
	}
	mode := srcStat.Mode() & os.ModeType
	switch mode {
	case os.ModeDir, 0: // pass
	default:
		return Errorf(rio.ErrAssemblyInvalid, "placer: source may only be dir or plain file (%s is not)", srcPath)
	}

	// Handle all the cases for existing things at destination.
	dstStat, err := os.Stat(dstPath.String())
	if err == nil {
		// If exists and wrong type, error.
		if dstStat.Mode()&os.ModeType != mode {
			return Errorf(rio.ErrAssemblyInvalid, "placer: destination already exists and is different type than source")
		}
		// If exists and right type, exit early.
		return nil
	}
	// If it doesn't exist, that's fine; any other error, ErrAssembly.
	if !os.IsNotExist(err) {
		return Errorf(rio.ErrAssemblyInvalid, "placer: destination unusable: %s", err)
	}

	// If we made it this far: dest doesn't exist yet.
	// Capture the parent dir mtime, because we're about to disrupt it.
	defer fsOp.RepairMtime(rootFs, dstPath.Dir().CoerceRelative())()

	// Make the dest node, matching type of the source.
	// The perms don't matter; will be shadowed.
	// We assume the parent dirs are all in place because you're almost
	// certainly using this as part of an assembler.
	switch mode {
	case os.ModeDir:
		err = os.Mkdir(dstPath.String(), 0644)
	case 0:
		var f *os.File
		f, err = os.OpenFile(dstPath.String(), os.O_CREATE, 0644)
		f.Close()
	}
	if err != nil {
		return Errorf(rio.ErrAssemblyInvalid, "placer: destination unusable: %s", err)
	}
	return nil
}
