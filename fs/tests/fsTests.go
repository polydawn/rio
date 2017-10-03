package tests

import (
	"os"
	"testing"

	"github.com/polydawn/go-errcat"

	"go.polydawn.net/rio/fs"
)

func CheckMkdirLstatRoundtrip(t *testing.T, afs fs.FS) {
	d1 := fs.MustRelPath("d1")
	if err := afs.Mkdir(d1, 0755); err != nil {
		t.Fatalf("mkdir failed: %s", err)
	}
	stat, err := afs.LStat(d1)
	if err != nil {
		t.Fatalf("lstat on just-created dir failed: %s", err)
	}
	if stat.Type != fs.Type_Dir {
		t.Errorf("lstat on just-created dir returned bogus type: %s", stat.Type)
	}
}

func CheckDeepMkdirError(t *testing.T, afs fs.FS) {
	d1d2 := fs.MustRelPath("d1/d2")
	if err := afs.Mkdir(d1d2, 0755); err == nil {
		t.Fatalf("deep mkdir without parents should have failed: %s", err)
	}
	_, err := afs.LStat(d1d2)
	if errcat.Category(err) != fs.ErrNotExists {
		t.Errorf("deep mkdir without parents error with category %q: got %q", fs.ErrNotExists, errcat.Category(err))
	}
}

func CheckMklinkLstatRoundtrip(t *testing.T, afs fs.FS) {
	l1 := fs.MustRelPath("l1")
	if err := afs.Mklink(l1, "./target"); err != nil {
		t.Fatalf("mklink failed: %s", err)
	}
	stat, err := afs.LStat(l1)
	if err != nil {
		t.Fatalf("lstat on just-created symlink failed: %s", err)
	}
	if stat.Type != fs.Type_Symlink {
		t.Errorf("lstat on just-created symlink returned bogus type: %s", stat.Type)
	}
}

func CheckSymlinks(t *testing.T, afs fs.FS) {
	t.Run("symlink resolve", func(t *testing.T) {
		t.Run("symlinks to files resolve correctly", func(t *testing.T) {
			t.Run("short relative case", func(t *testing.T) {
				l1 := fs.MustRelPath("l1")
				targetStr := "./target"
				target := fs.MustRelPath(targetStr)

				must(t, afs.Mklink(l1, targetStr))
				mustFile(t, afs, target, "body")

				resolved, err := afs.ResolveLink(targetStr, l1)
				if err != nil {
					t.Errorf("resolve should not have errored: %s", err)
				}
				if resolved != target {
					t.Errorf("incorrect resolve: %q", resolved)
				}
			})
		})
	})
}

func mustFile(t *testing.T, afs fs.FS, path fs.RelPath, body string) {
	t.Helper()
	f, err := afs.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	must(t, err)
	defer f.Close()
	_, err = f.Write([]byte(body))
	must(t, err)
}
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("setup step failed: %s", err)
	}
}
