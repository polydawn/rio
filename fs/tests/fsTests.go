package tests

import (
	"testing"

	"github.com/polydawn/go-errcat"

	"go.polydawn.net/rio/fs"
)

func CheckMkdirLstatRoundtrip(t *testing.T, afs fs.FS) {
	d1 := fs.MustRelPath("d1")
	if err := afs.Mkdir(d1, 0755); err != nil {
		t.Fatalf("mkdir failed: %s", err)
	}
	if _, err := afs.LStat(d1); err != nil {
		t.Fatalf("lstat on just-created dir failed: %s", err)
	}
}

func CheckDeepMkdirError(t *testing.T, afs fs.FS) {
	d1d2 := fs.MustRelPath("d1/d2")
	if err := afs.Mkdir(d1d2, 0755); err == nil {
		t.Fatalf("deep mkdir without parents should have failed: %s", err)
	}
	_, err := afs.LStat(d1d2)
	if errcat.Category(err) != fs.ErrNotExists {
		t.Fatalf("deep mkdir without parents error with category %q: got %q", fs.ErrNotExists, errcat.Category(err))
	}
}
