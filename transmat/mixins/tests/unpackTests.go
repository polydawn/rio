package tests

import (
	"context"
	"fmt"
	"io/ioutil"

	. "github.com/smartystreets/goconvey/convey"

	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/fsOp"
	"go.polydawn.net/rio/testutil"
	"go.polydawn.net/timeless-api"
	"go.polydawn.net/timeless-api/rio"
)

func CheckRoundTrip(pack rio.PackFunc, unpack rio.UnpackFunc, warehouseAddr api.WarehouseAddr) {
	Convey("SPEC: Round-trip pack and unpack of fileset should work...", func() {
		for _, fixture := range AllFixtures {
			Convey(fmt.Sprintf("- Fixture %q", fixture.Name), func() {
				testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
					fixturePath := tmpDir.Join(fs.MustRelPath("fixture"))
					// Set up fixture.
					PlaceFixture(osfs.New(fixturePath), fixture.Files)
					// Pack up into our warehouseaddr.
					wareID, err := pack(
						context.Background(),
						fixturePath.String(),
						api.FilesetFilters{
							Uid:   "keep",
							Gid:   "keep",
							Mtime: "keep",
						},
						warehouseAddr,
						rio.Monitor{},
					)
					So(err, ShouldBeNil)
					// Unpack to a new path.
					unpackPath := tmpDir.Join(fs.MustRelPath("unpack"))
					wareID2, err := unpack(
						context.Background(),
						wareID,
						unpackPath.String(),
						api.FilesetFilters{
							Sticky: true,
						},
						[]api.WarehouseAddr{warehouseAddr},
						rio.Monitor{},
					)
					Convey("...and agree on hash and content", FailureContinues, func() {
						So(err, ShouldBeNil)
						// Should be same hash reported by unpack hashing process.
						So(wareID, ShouldResemble, wareID2)
						// Each file in the fixture should exist and match rescanning.
						afs := osfs.New(unpackPath)
						for _, file := range fixture.Files {
							fmeta, reader, err := fsOp.ScanFile(afs, file.Metadata.Name)
							So(err, ShouldBeNil)
							fmeta.Mtime = fmeta.Mtime.UTC()
							So(*fmeta, ShouldResemble, file.Metadata)
							if file.Metadata.Type == fs.Type_File {
								body, _ := ioutil.ReadAll(reader)
								So(string(body), ShouldResemble, string(file.Body))
							}
						}
					})
				})
			})
		}
	})
}
