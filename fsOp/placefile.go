package normalfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"go.polydawn.net/rio/fs"
)

/*
	Places a file on the filesystem.
	Replicates all attributes described in the metadata.

	The path within the filesystem is `hdr.Name` (conventionally, this means
	the filesystem will join the `hdr.Name` with the absolute base path
	it was constructed with).

	No changes are allowed to occur outside of the filesystem's base path.
	Hardlinks may not point outside of the base path.
	Symlinks may *point* at paths outside of the base path (because you
	may be about to chroot into this, in which case absolute link paths
	make perfect sense), and invalid symlinks are acceptable -- however
	symlinks may *not* be traversed during any part of `hdr.Name`; this is
	considered malformed input and will result in a BreakoutError.

	Please note that like all filesystem operations within a lightyear of
	symlinks, all validations are best-effort, but are only capable of
	correctness in the absense of concurrent modifications inside `destBasePath`.

	Device files *will* be created, with their maj/min numbers.
	This may be considered a security concern; you should whitelist inputs
	if using this to provision a sandbox.
*/
func PlaceFile(afs fs.FS, fmeta fs.Metadata, body io.Reader) fs.ErrFS {
	// First, no part of the path may be a symlink.
	for path := fmeta.Name; ; path = path.Dir() {
		target, isSymlink, err := afs.Readlink(path)
		if isSymlink {
			return fs.ErrBreakout{
				OpPath:     fmeta.Name,
				OpArea:     afs.BasePath(),
				LinkPath:   path,
				LinkTarget: target,
			}
		} else if err == nil {
			continue // regular paths are fine.
		} else if _, ok := err.(*fs.ErrNotExists); ok {
			continue // not existing is fine.
		} else {
			return err // any other unknown error means we lack perms or something: reject.
		}
		if path == (fs.RelPath{}) {
			break // success
		}
	}

	// Fill in the content.  (Attribs come later.)
	switch fmeta.Type {
	case fs.Type_Invalid:
		panic(fmt.Errorf("invalid fs.Metadata.Type; partially constructed object?"))
	case fs.Type_File:
		file, err := afs.OpenFile(fmeta.Name, os.O_CREATE|os.O_WRONLY, fmeta.Perms)
		if err != nil {
			return err
		}
		if _, err := io.Copy(file, body); err != nil {
			file.Close()
			return fs.IOError(err)
		}
		file.Close()
	case fs.Type_Dir:
		if fmeta.Name == (fs.RelPath{}) {
			// for the base dir only:
			// the dir may exist; we'll just chown+chmod+chtime it.
			// there is no race-free path through this btw, unless you know of a way to lstat and mkdir in the same syscall.
			if existingFmeta, err := afs.LStat(fmeta.Name); err == nil && existingFmeta.Type == fs.Type_Dir {
				break
			}
		}
		if err := afs.Mkdir(fmeta.Name, fmeta.Perms); err != nil {
			return err
		}
	case fs.Type_Symlink:
		// linkname can be anything you want.  It continues to be a string parameter rather than
		// any of our normalized `fs.*Path` types because it is perfectly valid (if odd)
		// to store the string ".///" as a symlink target.
		if err := afs.Mklink(fmeta.Name, fmeta.Linkname); err != nil {
			return err
		}
	case fs.Type_NamedPipe:
		if err := syscall.Mkfifo(destPath, uint32(hdr.Mode&07777)); err != nil {
			ioError(err)
		}
	case fs.Type_Socket:
		panic("todo unhandlable type error") // REVIEW is it?  we certainly can't make a *live* socket, but we could make the dead socket file exist.
	case fs.Type_Device:
		mode := uint32(hdr.Mode&07777) | syscall.S_IFBLK
		if err := syscall.Mknod(destPath, mode, int(fspatch.Mkdev(hdr.Devmajor, hdr.Devminor))); err != nil {
			ioError(err)
		}
	case fs.Type_CharDevice:
		mode := uint32(hdr.Mode&07777) | syscall.S_IFCHR
		if err := syscall.Mknod(destPath, mode, int(fspatch.Mkdev(hdr.Devmajor, hdr.Devminor))); err != nil {
			ioError(err)
		}
	case fs.Type_Hardlink:
		targetPath := filepath.Join(destBasePath, hdr.Linkname)
		if !strings.HasPrefix(targetPath, destBasePath) {
			panic(BreakoutError.New("invalid hardlink %q -> %q", targetPath, hdr.Linkname))
		}
		if err := os.Link(targetPath, destPath); err != nil {
			ioError(err)
		}
	default:
		panic(errors.ProgrammerError.New("placefile: unhandled file mode %q", fmeta.Type))
	}

	if err := os.Lchown(destPath, hdr.Uid, hdr.Gid); err != nil {
		ioError(err)
	}

	for key, value := range hdr.Xattrs {
		if err := fspatch.Lsetxattr(destPath, key, []byte(value), 0); err != nil {
			ioError(err)
		}
	}

	if fmeta.Type == fs.Type_Symlink {
		// need to use LUtimesNano to avoid traverse symlinks
		if err := fspatch.LUtimesNano(destPath, hdr.AccessTime, hdr.ModTime); err != nil {
			ioError(err)
		}
	} else {
		// do this for everything not a symlink, since there's no such thing as `lchmod` on linux -.-
		if err := os.Chmod(destPath, mode); err != nil {
			ioError(err)
		}
		if err := fspatch.UtimesNano(destPath, hdr.AccessTime, hdr.ModTime); err != nil {
			ioError(err)
		}

	}
}
