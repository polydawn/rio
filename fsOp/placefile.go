package normalfs

import (
	"fmt"
	"io"
	"os"

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
		file, err := afs.OpenFile(fmeta.Name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, fmeta.Perms)
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
				if err := afs.Chmod(fmeta.Name, fmeta.Perms); err != nil {
					return err
				}
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
		// There is no chmod call here, because there is no such thing as 'lchmod' on linux :I
	case fs.Type_NamedPipe:
		if err := afs.Mkfifo(fmeta.Name, fmeta.Perms); err != nil {
			return err
		}
	case fs.Type_Socket:
		panic("todo unhandlable type error") // REVIEW is it?  we certainly can't make a *live* socket, but we could make the dead socket file exist.
	case fs.Type_Device:
		if err := afs.MkdevBlock(fmeta.Name, fmeta.Devmajor, fmeta.Devminor, fmeta.Perms); err != nil {
			return err
		}
	case fs.Type_CharDevice:
		if err := afs.MkdevChar(fmeta.Name, fmeta.Devmajor, fmeta.Devminor, fmeta.Perms); err != nil {
			return err
		}
	case fs.Type_Hardlink:
		panic("todo hardlines not handled")
	default:
		panic(fmt.Sprintf("placefile: unhandled file mode %q", fmeta.Type))
	}

	// Set the UID and GID for all file and dir types.
	if err := afs.Lchown(fmeta.Name, fmeta.Uid, fmeta.Gid); err != nil {
		return err
	}

	// Skipping on xattrs for the moment.
	//	for key, value := range hdr.Xattrs {
	//		if err := fspatch.Lsetxattr(destPath, key, []byte(value), 0); err != nil {
	//			ioError(err)
	//		}
	//	}

	// Last of all, set times.  (All the earlier mutations like chown would alter them again.)
	// We split behavior based whether or not target is a symlink, because it broadens
	//  our platform support: Mac doesn't support the 'L' version of this call, so refraining
	//  from using it unless absolutely necessary means we can support unpacking a filesystem
	//  on Macs as long as it doesn't include symlinks.  (Eyeroll.)
	switch fmeta.Type {
	case fs.Type_Symlink:
		if err := afs.SetTimesLNano(fmeta.Name, fmeta.Mtime, fs.DefaultAtime); err != nil {
			return err
		}
	default:
		if err := afs.SetTimesNano(fmeta.Name, fmeta.Mtime, fs.DefaultAtime); err != nil {
			return err
		}
	}

	// Success!
	return nil
}
