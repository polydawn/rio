package osfs

import (
	"testing"

	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/tests"
	"go.polydawn.net/rio/testutil"
)

func TestAll(t *testing.T) {
	testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
		tfs := New(tmpDir)
		boxPath := fs.MustRelPath("sandbox")
		tfs.Mkdir(boxPath, 0755)
		afs := New(tmpDir.Join(boxPath))

		tests.CheckMkdirLstatRoundtrip(t, afs)
	})
}
