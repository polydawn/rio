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
							{
								Path:       fs.MustAbsolutePath("/bc"),
								WareID:     api.WareID{"tar", "2jkqXaVWCdH7axj1XW56rxZ6WVQ8f46nqMf2BBX7kjLsU9DsvQCquEoy6GcBcQ1Fqc"},
								Filters:    api.Filter_NoMutation,
								Warehouses: []api.WarehouseAddr{"file://../transmat/tar/fixtures/tar_kitchenSink.tgz"},
							},
						},
					)
					So(err, ShouldBeNil)

					// Root and first file come from the first input:
					So(ShouldStat(afs, fs.MustRelPath(".")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("."), Type: fs.Type_Dir, Uid: 7000, Gid: 7000, Perms: 0755, Mtime: time.Date(2015, 05, 30, 19, 53, 35, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("ab")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("ab"), Type: fs.Type_File, Uid: 7000, Gid: 7000, Perms: 0644, Mtime: time.Date(2015, 05, 30, 19, 53, 35, 0, time.UTC)})
					// This dir exists in the first input, but is shadowed to the second input's root props:
					So(ShouldStat(afs, fs.MustRelPath("bc")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("bc"), Type: fs.Type_Dir, Uid: 7000, Gid: 7000, Perms: 0755, Mtime: time.Date(2017, 9, 27, 18, 27, 6, 0, time.UTC)})
					// These are from the second input:
					So(ShouldStat(afs, fs.MustRelPath("bc/dir/")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("bc/dir/"), Type: fs.Type_Dir, Uid: 7000, Gid: 7000, Perms: 0755, Mtime: time.Date(2017, 9, 27, 18, 26, 35, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("bc/dir/f1")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("bc/dir/f1"), Type: fs.Type_File, Uid: 7000, Gid: 7000, Perms: 0750, Mtime: time.Date(2017, 9, 27, 18, 25, 55, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("bc/empty/")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("bc/empty/"), Type: fs.Type_Dir, Uid: 7000, Gid: 7000, Perms: 0755, Mtime: time.Date(2017, 9, 27, 18, 26, 2, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("bc/f2")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("bc/f2"), Type: fs.Type_File, Uid: 4000, Gid: 5000, Perms: 0644, Size: 3, Mtime: time.Date(2017, 9, 27, 18, 26, 39, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("bc/deep/")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("bc/deep/"), Type: fs.Type_Dir, Uid: 7000, Gid: 7000, Perms: 0755, Mtime: time.Date(2017, 9, 27, 18, 27, 10, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("bc/deep/tree/")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("bc/deep/tree/"), Type: fs.Type_Dir, Uid: 7000, Gid: 7000, Perms: 0755, Mtime: time.Date(2017, 9, 27, 18, 27, 19, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("bc/deep/tree/f3")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("bc/deep/tree/f3"), Type: fs.Type_File, Uid: 7000, Gid: 7000, Perms: 0644, Size: 7, Mtime: time.Date(2017, 9, 27, 18, 27, 19, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("bc/lnkdangle")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("bc/lnkdangle"), Type: fs.Type_Symlink, Uid: 7000, Gid: 7000, Perms: 0777, Linkname: "nonexistent", Mtime: time.Date(2017, 9, 27, 18, 26, 14, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("bc/lnkfile")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("bc/lnkfile"), Type: fs.Type_Symlink, Uid: 7000, Gid: 7000, Perms: 0777, Linkname: "f2", Mtime: time.Date(2017, 9, 27, 18, 26, 49, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("bc/lnkdir")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("bc/lnkdir"), Type: fs.Type_Symlink, Uid: 7000, Gid: 7000, Perms: 0777, Linkname: "dir/", Mtime: time.Date(2017, 9, 27, 18, 26, 22, 0, time.UTC)})

					So(cleanupFunc(), ShouldBeNil)
				})
				Convey("Unpack plus implicit parent dir creation should work:", func() {
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
							{
								Path:       fs.MustAbsolutePath("/mk/dir/"),
								WareID:     api.WareID{"tar", "5y6NvK6GBPQ6CcuNyJyWtSrMAJQ4LVrAcZSoCRAzMSk5o53pkTYiieWyRivfvhZwhZ"},
								Filters:    api.Filter_NoMutation,
								Warehouses: []api.WarehouseAddr{"file://../transmat/tar/fixtures/tar_withBase.tgz"},
							},
						},
					)
					So(err, ShouldBeNil)

					// From the first input:
					So(ShouldStat(afs, fs.MustRelPath(".")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("."), Type: fs.Type_Dir, Uid: 7000, Gid: 7000, Perms: 0755, Mtime: time.Date(2015, 05, 30, 19, 53, 35, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("ab")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("ab"), Type: fs.Type_File, Uid: 7000, Gid: 7000, Perms: 0644, Mtime: time.Date(2015, 05, 30, 19, 53, 35, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("bc")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("bc"), Type: fs.Type_Dir, Uid: 7000, Gid: 7000, Perms: 0755, Mtime: time.Date(2015, 05, 30, 19, 53, 35, 0, time.UTC)})
					// From implicit dir creation:
					So(ShouldStat(afs, fs.MustRelPath("mk")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("mk"), Type: fs.Type_Dir, Uid: 0, Gid: 0, Perms: 0755, Mtime: time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)})
					// From the second input:
					So(ShouldStat(afs, fs.MustRelPath("mk/dir")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("mk/dir"), Type: fs.Type_Dir, Uid: 7000, Gid: 7000, Perms: 0755, Mtime: time.Date(2015, 05, 30, 19, 53, 35, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("mk/dir/ab")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("mk/dir/ab"), Type: fs.Type_File, Uid: 7000, Gid: 7000, Perms: 0644, Mtime: time.Date(2015, 05, 30, 19, 53, 35, 0, time.UTC)})
					So(ShouldStat(afs, fs.MustRelPath("mk/dir/bc")), ShouldResemble,
						fs.Metadata{Name: fs.MustRelPath("mk/dir/bc"), Type: fs.Type_Dir, Uid: 7000, Gid: 7000, Perms: 0755, Mtime: time.Date(2015, 05, 30, 19, 53, 35, 0, time.UTC)})

					So(cleanupFunc(), ShouldBeNil)
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
