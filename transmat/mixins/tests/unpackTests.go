package tests

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/polydawn/go-timeless-api"
	"github.com/polydawn/go-timeless-api/rio"
	"github.com/polydawn/rio/cache"
	"github.com/polydawn/rio/config"
	"github.com/polydawn/rio/fs"
	"github.com/polydawn/rio/fs/osfs"
	"github.com/polydawn/rio/fsOp"
	"github.com/polydawn/rio/testutil"
)

func CheckRoundTrip(packType api.PackType, pack rio.PackFunc, unpack rio.UnpackFunc, warehouseAddr api.WarehouseLocation) {
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
						packType,
						fixturePath.String(),
						api.FilesetPackFilter_Lossless,
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
						api.FilesetUnpackFilter_Lossless,
						rio.Placement_Direct,
						[]api.WarehouseLocation{warehouseAddr},
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

func CheckCachePopulation(packType api.PackType, pack rio.PackFunc, unpack rio.UnpackFunc, warehouseAddr api.WarehouseLocation) {
	Convey("SPEC: Caching: unpack with 'none' placement should result in cache...", func() {
		testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
			// Bonk our own config env vars to isolate cache.
			tmpBase := tmpDir.Join(fs.MustRelPath("rio-base"))
			os.Setenv("RIO_BASE", tmpBase.String())

			// Set up fixture.
			fixture := FixtureGamma
			fixturePath := tmpDir.Join(fs.MustRelPath("fixture"))
			PlaceFixture(osfs.New(fixturePath), fixture)

			// Pack up into our warehouseaddr.  (Not interesting, but we lack other fixtures.)
			wareID, err := pack(
				context.Background(),
				packType,
				fixturePath.String(),
				api.FilesetPackFilter_Lossless,
				warehouseAddr,
				rio.Monitor{},
			)
			So(err, ShouldBeNil)

			// Unpack to a new path, with Placement_None mode.
			unpackPath := tmpDir.Join(fs.MustRelPath("unpack"))
			wareID2, err := unpack(
				context.Background(),
				wareID,
				unpackPath.String(),
				api.FilesetUnpackFilter_Lossless,
				rio.Placement_None,
				[]api.WarehouseLocation{warehouseAddr},
				rio.Monitor{},
			)
			So(err, ShouldBeNil)
			So(wareID, ShouldResemble, wareID2)

			Convey("cache should exist, and agree on hash and content", FailureContinues, func() {
				// Each file in the fixture should exist and match rescanning.
				shelfPath := config.GetCacheBasePath().Join(cache.ShelfFor(wareID))
				afs := osfs.New(shelfPath)
				for _, file := range fixture {
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
