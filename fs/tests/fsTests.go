package tests

import (
	"os"

	"github.com/polydawn/go-errcat"
	. "github.com/smartystreets/goconvey/convey"

	"go.polydawn.net/rio/fs"
)

func CheckMkdirLstatRoundtrip(afs fs.FS) {
	Convey("SPEC: mkdir and lstat should roundtrip", func() {
		d1 := fs.MustRelPath("d1")
		So(afs.Mkdir(d1, 0755), ShouldBeNil)
		stat, err := afs.LStat(d1)
		So(err, ShouldBeNil)
		So(stat.Type, ShouldEqual, fs.Type_Dir)
	})
}

func CheckDeepMkdirError(afs fs.FS) {
	Convey("SPEC: deep mkdir should error", func() {
		d1d2 := fs.MustRelPath("d1/d2")
		So(afs.Mkdir(d1d2, 0755), errcat.ErrorShouldHaveCategory, fs.ErrNotExists)
		_, err := afs.LStat(d1d2)
		So(err, errcat.ErrorShouldHaveCategory, fs.ErrNotExists)
	})
}

func CheckMklinkLstatRoundtrip(afs fs.FS) {
	Convey("SPEC: mklink and lstat should roundtrip", func() {
		l1 := fs.MustRelPath("l1")
		So(afs.Mklink(l1, "./target"), ShouldBeNil)
		stat, err := afs.LStat(l1)
		So(err, ShouldBeNil)
		So(stat.Type, ShouldEqual, fs.Type_Symlink)
	})
}

func CheckSymlinks(afs fs.FS) {
	Convey("SPEC: symlink resolve", func() {
		Convey("symlinks to files resolve correctly", func() {
			Convey("short relative case", func() {
				l1 := fs.MustRelPath("l1")
				linkStr := "./target"
				target := fs.MustRelPath(linkStr)

				So(afs.Mklink(l1, linkStr), ShouldBeNil)
				So(makeFile(afs, target, "body"), ShouldBeNil)

				resolved, err := afs.ResolveLink(linkStr, l1)
				So(err, ShouldBeNil)
				So(resolved, ShouldResemble, target)
			})
			Convey("long relative case", func() {
				d1l1 := fs.MustRelPath("d1/l1")
				linkStr := ".././/d2/target"
				target := fs.MustRelPath("d2/target")

				So(afs.Mkdir(d1l1.Dir(), 0755), ShouldBeNil)
				So(afs.Mklink(d1l1, linkStr), ShouldBeNil)
				So(afs.Mkdir(target.Dir(), 0755), ShouldBeNil)
				So(makeFile(afs, target, "body"), ShouldBeNil)

				resolved, err := afs.ResolveLink(linkStr, d1l1)
				So(err, ShouldBeNil)
				So(resolved, ShouldResemble, target)
			})
			Convey("dotdot-overload case", func() {
				d1l1 := fs.MustRelPath("d1/l1")
				linkStr := ".././/../../../d2/target"
				target := fs.MustRelPath("d2/target")

				So(afs.Mkdir(d1l1.Dir(), 0755), ShouldBeNil)
				So(afs.Mklink(d1l1, linkStr), ShouldBeNil)
				So(afs.Mkdir(target.Dir(), 0755), ShouldBeNil)
				So(makeFile(afs, target, "body"), ShouldBeNil)

				resolved, err := afs.ResolveLink(linkStr, d1l1)
				So(err, ShouldBeNil)
				So(resolved, ShouldResemble, target)
			})
		})
		Convey("dangling symlinks resolve correctly", func() {
			Convey("short relative case", func() {
				l1 := fs.MustRelPath("l1")
				linkStr := "./target"
				target := fs.MustRelPath(linkStr)

				So(afs.Mklink(l1, linkStr), ShouldBeNil)

				resolved, err := afs.ResolveLink(linkStr, l1)
				So(err, ShouldBeNil)
				So(resolved, ShouldResemble, target)
			})
			Convey("long relative case", func() {
				d1l1 := fs.MustRelPath("d1/l1")
				linkStr := ".././/d2/target"
				target := fs.MustRelPath("d2/target")

				So(afs.Mkdir(d1l1.Dir(), 0755), ShouldBeNil)
				So(afs.Mklink(d1l1, linkStr), ShouldBeNil)
				So(afs.Mkdir(target.Dir(), 0755), ShouldBeNil)

				resolved, err := afs.ResolveLink(linkStr, d1l1)
				So(err, ShouldBeNil)
				So(resolved, ShouldResemble, target)
			})
		})
	})
}

func makeFile(afs fs.FS, path fs.RelPath, body string) error {
	f, err := afs.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(body))
	return err
}
