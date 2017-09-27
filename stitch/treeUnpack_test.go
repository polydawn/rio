package stitch

import (
	"context"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
	. "go.polydawn.net/rio/testutil"
	"go.polydawn.net/rio/transmat/tar"
)

func TestTreeUnpack(t *testing.T) {
	// We lean on the existence and sanity of the tar transmat extensively here.
	//  If we don't have a working unpack tool around for us, it's awfully hard to
	//  exercise much of tree unpack!

	Convey("Tree unpack scenarios:", t,
		Requires(RequiresCanManageOwnership, RequiresCanMountAny, func() {
			WithTmpdir(func(tmpDir fs.AbsolutePath) {
				// Bonk our own config env vars to isolate cache.
				tmpBase := tmpDir.Join(fs.MustRelPath("rio-base"))
				os.Setenv("RIO_BASE", tmpBase.String())

				// Set up the utils.
				assembler, err := NewAssembler(tartrans.Unpack)
				So(err, ShouldBeNil)

				Convey("Single-entry unpack should work:", func() {
					afs := osfs.New(tmpDir.Join(fs.MustRelPath("tree")))
					cleanupFunc, err := assembler.Run(
						context.Background(),
						afs,
						[]UnpackSpec{
							{
								Path:       fs.MustAbsolutePath("/"),
								WareID:     api.WareID{"tar", "5y6NvK6GBPQ6CcuNyJyWtSrMAJQ4LVrAcZSoCRAzMSk5o53pkTYiieWyRivfvhZwhZ"},
								Filters:    api.Filter_NoMutation,
								Warehouses: []api.WarehouseAddr{"file://../transmat/tar/fixtures/tar_withBase.tgz"},
							},
						},
					)
					So(err, ShouldBeNil)

					So(ShouldStat(afs, fs.MustRelPath(".")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("."), Type: fs.Type_Dir, Uid: 7000, Gid: 7000, Perms: 0755, Mtime: time.Date(2015, 05, 30, 19, 53, 35, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("ab")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("ab"), Type: fs.Type_File, Uid: 7000, Gid: 7000, Perms: 0644, Mtime: time.Date(2015, 05, 30, 19, 53, 35, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("bc")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("bc"), Type: fs.Type_Dir, Uid: 7000, Gid: 7000, Perms: 0755, Mtime: time.Date(2015, 05, 30, 19, 53, 35, 0, time.UTC)})

					So(cleanupFunc(), ShouldBeNil)
				})
				Convey("Multi-entry unpack should work:", func() {
					// TODO
				})
				Convey("Unpack plus implicit parent dir creation should work:", func() {
					// TODO
				})
				Convey("Unpack plus mounts should work:", func() {
					// TODO
				})
				Convey("Invalid mounts should fail:", func() {
					// TODO
				})
				Convey("Unpack with no explicit root should work:", func() {
					// TODO
				})
			})
		}),
	)
}
