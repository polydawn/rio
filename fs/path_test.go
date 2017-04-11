package fs

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRelPath(t *testing.T) {
	Convey("RelPath stringer suite:", t, func() {
		for _, tr := range []struct {
			title string
			p1    RelPath
			str   string
		}{
			{"zero values",
				RelPath{},
				"."},
			{"dot value",
				MustRelPath("."),
				"."},
			{"short value",
				MustRelPath("aa"),
				"./aa"},
			{"long value",
				MustRelPath("a/bb/ccc"),
				"./a/bb/ccc"},
		} {
			Convey(tr.title, func() {
				v := fmt.Sprintf("%s", tr.p1)
				So(v, ShouldResemble, tr.str)
			})
		}
	})
}

func TestRelPathDir(t *testing.T) {
	Convey("RelPath.Dir suite:", t, func() {
		for _, tr := range []struct {
			title string
			p1    RelPath
			pdir  RelPath
		}{
			{"zero values",
				RelPath{},
				RelPath{}},
			{"dot value",
				MustRelPath("."),
				RelPath{}},
			{"short value",
				MustRelPath("aa"),
				RelPath{}},
			{"long value",
				MustRelPath("a/bb/ccc"),
				MustRelPath("a/bb")},
		} {
			Convey(tr.title, func() {
				v := tr.p1.Dir()
				So(v, ShouldResemble, tr.pdir)
			})
		}
	})
}

func TestRelPathLast(t *testing.T) {
	Convey("RelPath.Last suite:", t, func() {
		for _, tr := range []struct {
			title string
			p1    RelPath
			last  string
		}{
			{"zero values",
				RelPath{},
				"."},
			{"dot value",
				MustRelPath("."),
				"."},
			{"short value",
				MustRelPath("aa"),
				"aa"},
			{"long value",
				MustRelPath("a/bb/ccc"),
				"ccc"},
		} {
			Convey(tr.title, func() {
				v := tr.p1.Last()
				So(v, ShouldResemble, tr.last)
			})
		}
	})
}

func TestRelPathJoins(t *testing.T) {
	Convey("RelPath.Join suite:", t, func() {
		for _, tr := range []struct {
			title  string
			p1, p2 RelPath
			pj     RelPath
		}{
			{"zero values",
				RelPath{}, RelPath{},
				RelPath{}},
			{"regular values",
				MustRelPath("rel"), MustRelPath("pth"),
				MustRelPath("rel/pth")},
			{"zero,short",
				MustRelPath("."), MustRelPath("pth"),
				MustRelPath("pth")},
			{"long,zero",
				MustRelPath("a/bb/ccc"), MustRelPath("."),
				MustRelPath("a/bb/ccc")},
			{"long,long",
				MustRelPath("a/bb/ccc"), MustRelPath("dd/e"),
				MustRelPath("a/bb/ccc/dd/e")},
		} {
			Convey(tr.title, func() {
				v := tr.p1.Join(tr.p2)
				So(v, ShouldResemble, tr.pj)
			})
		}
	})
}
