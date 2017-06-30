package fsOp

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/testutil"
)

func TestPlaceFile(t *testing.T) {
	Convey("PlaceFile suite:", t, func() {
		testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
			Convey("Simple file placements", func() {
				// TODO
			})
			Convey("Simple dir placements", func() {
				// TODO
			})
			Convey("Symlink traversal checks", func() {
				// TODO
			})
		})
	})
}
