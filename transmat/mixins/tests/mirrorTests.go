package tests

import (
	"context"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/polydawn/go-timeless-api"
	"github.com/polydawn/go-timeless-api/rio"
	"github.com/polydawn/rio/fs"
	"github.com/polydawn/rio/fs/osfs"
	"github.com/polydawn/rio/testutil"
)

func CheckMirror(packType api.PackType, mirror rio.MirrorFunc, pack rio.PackFunc, unpack rio.UnpackFunc, target api.WarehouseLocation, source api.WarehouseLocation) {
	testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
		// Initialization: make a pack to test against, put it in source warehouse.
		fixturePath := tmpDir.Join(fs.MustRelPath("fixture"))
		// Set up fixture.
		PlaceFixture(osfs.New(fixturePath), FixtureGamma)
		// Pack up into our warehouseaddr.
		wareID, err := pack(
			context.Background(),
			packType,
			fixturePath.String(),
			api.FilesetPackFilter_Lossless,
			source,
			rio.Monitor{},
		)
		So(err, ShouldBeNil)

		// Okay, now mirror:
		Convey("mirror should succeed", func() {
			mirroredWareID, err := mirror(
				context.Background(),
				wareID,
				target,
				[]api.WarehouseLocation{source},
				rio.Monitor{},
			)
			So(err, ShouldBeNil)
			So(mirroredWareID, ShouldResemble, wareID)

			Convey("unpack from mirror should succeed", func() {
				unpackedWareID, err := unpack(
					context.Background(),
					wareID,
					"-",
					api.FilesetUnpackFilter_Lossless,
					rio.Placement_None,
					[]api.WarehouseLocation{target},
					rio.Monitor{},
				)
				So(err, ShouldBeNil)
				So(unpackedWareID, ShouldResemble, wareID)
			})
		})
	})
}
