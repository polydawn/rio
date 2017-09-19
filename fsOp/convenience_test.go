package fsOp

import (
	"io"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/testutil"
)

func TestMkdirAll(t *testing.T) {
	// Note that all of these are assuming PlaceFile already works just fine.
	Convey("MkdirAll:", t, func() {
		testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
			afs := osfs.New(tmpDir)
			Convey("MkdirAll on an existing path should work...", func() {
				mustPlaceFile(afs, fs.Metadata{Name: fs.MustRelPath("dir"), Type: fs.Type_Dir, Perms: 0755}, nil)

				So(MkdirAll(afs, fs.MustRelPath("dir"), 0755), ShouldBeNil)
			})
			Convey("MkdirAll creating one node should work...", func() {
				So(MkdirAll(afs, fs.MustRelPath("dir"), 0755), ShouldBeNil)
				stat, err := afs.LStat(fs.MustRelPath("dir"))
				So(err, ShouldBeNil)
				So(stat.Type, ShouldEqual, fs.Type_Dir)
			})
			Convey("MkdirAll creating several nodes should work...", func() {
				So(MkdirAll(afs, fs.MustRelPath("dir/2/3"), 0755), ShouldBeNil)
				stat, err := afs.LStat(fs.MustRelPath("dir/2/3"))
				So(err, ShouldBeNil)
				So(stat.Type, ShouldEqual, fs.Type_Dir)
			})
		})
	})
}

func mustPlaceFile(afs fs.FS, fmeta fs.Metadata, body io.Reader) {
	if err := PlaceFile(afs, fmeta, body, true); err != nil {
		panic(err)
	}
}
