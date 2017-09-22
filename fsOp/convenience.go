package fsOp

import (
	. "github.com/polydawn/go-errcat"

	"go.polydawn.net/rio/fs"
)

/*
	Makes dirs recursively so the requested path exists, applying the assigned metadata
	to each one that needed to be produced.

	Existing dirs are not mutated.

	Symlinks will be traversed without comment (i.e. this will never emit ErrBreakout).
	(Note that this means this function is *not* used in either transmats nor stitch.)
*/
func MkdirAll(afs fs.FS, path fs.RelPath, perms fs.Perms) error {
	// Check if the path already exists.
	stat, err := afs.Stat(path)
	// Switch on status of the (derefenced) file.
	//  Recurse and mkdir if necessary.
	switch Category(err) {
	case nil:
		if stat.Type == fs.Type_Dir {
			return nil
		}
		return Errorf(fs.ErrNotDir, "%s already exists and is a %s not %s", afs.BasePath().Join(path), stat.Type, fs.Type_Dir)
	case fs.ErrNotExists:
		if path == (fs.RelPath{}) {
			return Errorf(fs.ErrNotExists, "base path %s does not exist!", afs.BasePath())
		}
		if err := MkdirAll(afs, path.Dir(), perms); err != nil {
			return err
		}
		if err := afs.Mkdir(path, perms); err != nil {
			switch Category(err) {
			case fs.ErrAlreadyExists:
				// this seemingly-contradictory message means the path does exist... it's just that stat said it didn't, beacuse it's a dangling symlink.
				return Errorf(fs.ErrNotDir, "%s already exists and is a %s not %s", afs.BasePath().Join(path), fs.Type_Symlink, fs.Type_Dir)
			default:
				return err
			}
		}
		return nil
	case fs.ErrNotDir:
		// Reformat the error a tad to not say "lstat", which is distracting.
		return Errorf(fs.ErrNotDir, "%s has parents which are not a directory", afs.BasePath().Join(path))
	default:
		return err
	}
}

/*
	Records the mtime property currently set on a path and returns a function which
	will force it to that value again.

	The typical use for this is `defer RepairMtime(fs, "./somedir")()` right
	before invoking some funcs which will mutate the contents of that dir.
	When the rest of the changes are done and the function returns, the mtime
	of the dir will be forced back to its previous value, seemingly unchanged.
*/
func RepairMtime(afs fs.FS, path fs.RelPath) func() {
	fmeta, _ := afs.LStat(path)
	return func() {
		afs.SetTimesLNano(path, fmeta.Mtime, fs.DefaultAtime)
	}
}
