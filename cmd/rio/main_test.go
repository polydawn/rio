package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/fsOp"
	"go.polydawn.net/rio/testutil"
	"go.polydawn.net/timeless-api/rio"
)

func TestWithoutArgs(t *testing.T) {
	Convey("rio: usage printed to stderr", t, func() {
		args := []string{"rio"}
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		stdin := &bytes.Buffer{}
		ctx := context.Background()
		exitCode := Main(ctx, args, stdin, stdout, stderr)
		t.Log(string(stdout.Bytes()))
		t.Log(string(stderr.Bytes()))
		So(string(stdout.Bytes()), ShouldBeBlank)
		So(string(stderr.Bytes()), ShouldNotBeBlank)
		firstLine, err := stderr.ReadString('\n')
		So(err, ShouldBeNil)
		So(string(firstLine), ShouldContainSubstring, "usage: rio [<flags>] <command> [<args> ...]")
		So(string(stderr.Bytes()), ShouldNotContainSubstring, "usage: rio [<flags>] <command> [<args> ...]")
		So(exitCode, ShouldEqual, rio.ExitUsage)
	})
}

/*
	Tests against pre-generated, known fixtures of tar binary blobs.

	These tests allow us to cover compat with other tar impls, compression, etc.
*/
func TestTarFixtureUnpack(t *testing.T) {
	Convey("rio: unpacking of tar fixtures", t,
		testutil.Requires(testutil.RequiresCanManageOwnership, func() {
			testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
				Convey("Unpack a fixture from gnu tar which includes a base dir", func() {
					ctx := context.Background()
					wareID := "tar:5y6NvK6GBPQ6CcuNyJyWtSrMAJQ4LVrAcZSoCRAzMSk5o53pkTYiieWyRivfvhZwhZ"
					source := "file://../../transmat/tar/fixtures/tar_withBase.tgz"
					args := []string{
						"rio",
						"unpack",
						fmt.Sprintf("--ware=%s", wareID),
						fmt.Sprintf("--source=%s", source),
						fmt.Sprintf("--path=%s", tmpDir.String()),
					}
					stdout := &bytes.Buffer{}
					stderr := &bytes.Buffer{}
					stdin := &bytes.Buffer{}
					exitCode := Main(ctx, args, stdin, stdout, stderr)
					t.Log(string(stdout.Bytes()))
					t.Log(string(stderr.Bytes()))
					So(exitCode, ShouldEqual, rio.ExitSuccess)
					So(string(stdout.Bytes()), ShouldEqual, wareID+"\n")

					Convey("The filesystem contains the correct unpacked fixture", func() {
						var err error
						fmeta, reader, err := fsOp.ScanFile(osfs.New(tmpDir), fs.MustRelPath("ab"))
						So(err, ShouldBeNil)
						So(fmeta.Name, ShouldResemble, fs.MustRelPath("ab"))
						So(fmeta.Type, ShouldResemble, fs.Type_File)
						So(fmeta.Uid, ShouldEqual, 7000)
						So(fmeta.Gid, ShouldEqual, 7000)
						So(fmeta.Mtime.UTC(), ShouldResemble, time.Date(2015, 05, 30, 19, 53, 35, 0, time.UTC))
						body, err := ioutil.ReadAll(reader)
						So(string(body), ShouldResemble, "")

						fmeta, reader, err = fsOp.ScanFile(osfs.New(tmpDir), fs.MustRelPath("bc"))
						So(err, ShouldBeNil)
						So(fmeta.Name, ShouldResemble, fs.MustRelPath("bc"))
						So(fmeta.Type, ShouldResemble, fs.Type_Dir)
						So(fmeta.Mtime.UTC(), ShouldResemble, time.Date(2015, 05, 30, 19, 53, 35, 0, time.UTC))
						So(reader, ShouldBeNil)

						fmeta, reader, err = fsOp.ScanFile(osfs.New(tmpDir), fs.MustRelPath("."))
						So(err, ShouldBeNil)
						So(fmeta.Name, ShouldResemble, fs.MustRelPath("."))
						So(fmeta.Type, ShouldResemble, fs.Type_Dir)
						So(fmeta.Mtime.UTC(), ShouldResemble, time.Date(2015, 05, 30, 19, 53, 35, 0, time.UTC))
						So(reader, ShouldBeNil)
					})
				})
			})
		}),
	)
}
