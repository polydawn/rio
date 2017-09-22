package placer

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/testutil"
	. "go.polydawn.net/rio/transmat/mixins/tests"
)

func TestPlacers(t *testing.T) {
	Convey("Copy placer spec tests:", t, testutil.Requires(testutil.RequiresCanManageOwnership, func() {
		specPlacerGood(CopyPlacer)
	}))
	Convey("Bind placer spec tests:", t, testutil.Requires(testutil.RequiresCanMountBind, func() {
		specPlacerGood(BindPlacer)
	}))
}

func specPlacerGood(placeFunc Placer) {
	testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
		afs := osfs.New(tmpDir)
		Convey("Placement of a dir should work, and maintain parent props", func() {
			PlaceFixture(afs, []FixtureFile{
				{fs.Metadata{Name: fs.MustRelPath("srcParent"), Type: fs.Type_Dir, Perms: 0755, Mtime: time.Date(2004, 01, 15, 0, 0, 0, 0, time.UTC)}, nil},
				{fs.Metadata{Name: fs.MustRelPath("srcParent/content"), Type: fs.Type_Dir, Uid: 4000, Perms: 0755, Mtime: time.Date(2005, 01, 15, 0, 0, 0, 0, time.UTC)}, nil},
				{fs.Metadata{Name: fs.MustRelPath("srcParent/content/file"), Type: fs.Type_File, Perms: 0640, Mtime: time.Date(2006, 01, 15, 0, 0, 0, 0, time.UTC)}, []byte("asdf")},
				{fs.Metadata{Name: fs.MustRelPath("dstParent"), Type: fs.Type_Dir, Perms: 0755, Mtime: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)}, nil},
			})

			cleanupFunc, err := placeFunc(tmpDir.Join(fs.MustRelPath("srcParent/content")), tmpDir.Join(fs.MustRelPath("dstParent/content")), true)
			So(err, ShouldBeNil)

			// First check the content files and dirs.
			So(shouldStat(afs, fs.MustRelPath("dstParent/content")), ShouldResemble, fs.Metadata{Name: fs.MustRelPath("dstParent/content"), Type: fs.Type_Dir, Uid: 4000, Perms: 0755, Mtime: time.Date(2005, 01, 15, 0, 0, 0, 0, time.UTC)})
			So(shouldStat(afs, fs.MustRelPath("dstParent/content/file")), ShouldResemble, fs.Metadata{Name: fs.MustRelPath("dstParent/content/file"), Type: fs.Type_File, Perms: 0640, Mtime: time.Date(2006, 01, 15, 0, 0, 0, 0, time.UTC), Size: 4})
			// Last (because you're most likely to screw this up) check the parent dir didn't get boinked.
			So(shouldStat(afs, fs.MustRelPath("dstParent")), ShouldResemble, fs.Metadata{Name: fs.MustRelPath("dstParent"), Type: fs.Type_Dir, Perms: 0755, Mtime: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)})

			So(cleanupFunc(), ShouldBeNil)
		})
	})
}

func shouldStat(afs fs.FS, path fs.RelPath) fs.Metadata {
	stat, err := afs.LStat(path)
	So(err, ShouldBeNil)
	stat.Mtime = stat.Mtime.UTC()
	return *stat
}
