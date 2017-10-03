package tests

import (
	"testing"

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
