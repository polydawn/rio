package rioexecclient_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"go.polydawn.net/rio/fs" // todo problematic for extraction
	"go.polydawn.net/rio/testutil"
	"go.polydawn.net/timeless-api"
	"go.polydawn.net/timeless-api/rio"
	"go.polydawn.net/timeless-api/rio/execclient"
)

// This test is moderately terrifying because it does indeed and really exec.
// This means the rio binary must have already been built.
// (TODO : figure out what the heck that means when we extract this to a freestanding api lib repo?!)
// We set the path to the project's build dir; any commands on your regular host PATH should not interfere.

func Test(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	err = os.Setenv("PATH", filepath.Join(cwd, "../../../bin/"))
	if err != nil {
		panic(err)
	}

	testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
		_, err := rioexecclient.UnpackFunc(
			context.Background(),
			api.WareID{"tar", "5y6NvK6GBPQ6CcuNyJyWtSrMAJQ4LVrAcZSoCRAzMSk5o53pkTYiieWyRivfvhZwhZ"},
			tmpDir.String(),
			api.FilesetFilters{},
			[]api.WarehouseAddr{"file://../../../transmat/tar/fixtures/tar_withBase.tgz"},
			rio.Monitor{},
		)
		if err != nil {
			t.Error(err)
		}
	})
}
