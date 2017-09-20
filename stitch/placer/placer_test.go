package placer

import (
	"bytes"
	"io"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/fsOp"
	"go.polydawn.net/rio/testutil"
)

func TestPlacers(t *testing.T) {
	Convey("Copy placer spec tests:", t, func() {
		specPlacerGood(CopyPlacer)
	})
}

func specPlacerGood(placeFunc Placer) {
	testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
		afs := osfs.New(tmpDir)
		Convey("Placement of a dir should work, and maintain parent props", func() {
			mustPlaceFile(afs, fs.Metadata{Name: fs.MustRelPath("srcParent"), Type: fs.Type_Dir, Perms: 0755, Mtime: time.Date(2004, 01, 15, 0, 0, 0, 0, time.UTC)}, nil)
			mustPlaceFile(afs, fs.Metadata{Name: fs.MustRelPath("srcParent/content"), Type: fs.Type_Dir, Perms: 0755, Mtime: time.Date(2005, 01, 15, 0, 0, 0, 0, time.UTC)}, nil)
			mustPlaceFile(afs, fs.Metadata{Name: fs.MustRelPath("srcParent/content/file"), Type: fs.Type_File, Perms: 0640, Mtime: time.Date(2006, 01, 15, 0, 0, 0, 0, time.UTC)}, bytes.NewBuffer([]byte("asdf")))
			mustPlaceFile(afs, fs.Metadata{Name: fs.MustRelPath("dstParent"), Type: fs.Type_Dir, Perms: 0755, Mtime: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)}, nil)

			So(placeFunc(tmpDir.Join(fs.MustRelPath("srcParent")), tmpDir.Join(fs.MustRelPath("dstParent")), true), ShouldBeNil)
		})
	})
}

func mustPlaceFile(afs fs.FS, fmeta fs.Metadata, body io.Reader) {
	if fmeta.Type == fs.Type_File && body == nil {
		body = &bytes.Buffer{}
	}
	if err := fsOp.PlaceFile(afs, fmeta, body, true); err != nil {
		panic(err)
	}
}
