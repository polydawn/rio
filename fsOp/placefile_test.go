package fsOp

import (
	"bytes"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/testutil"
)

func TestPlaceFile(t *testing.T) {
	Convey("PlaceFile suite:", t, func() {
		testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
			Convey("Simple file placements", func() {
				afs := osfs.New(tmpDir)
				err := PlaceFile(afs, fs.Metadata{
					Name: fs.MustRelPath("thing"),
					Type: fs.Type_File,
				}, bytes.NewBuffer([]byte("abc\n")), true)
				So(err, ShouldBeNil)
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
