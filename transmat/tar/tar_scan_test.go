package tartrans

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/polydawn/go-timeless-api"
	"github.com/polydawn/go-timeless-api/rio"
	"github.com/polydawn/rio/fs"
	"github.com/polydawn/rio/testutil"
)

func TestTarFixtureScan(t *testing.T) {
	Convey("Tar transmat: scan of fixtures", t, func() {
		testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
			Convey("Scan a fixture from gnu tar which includes a base dir", func() {
				gotWareID, err := Scan(
					context.Background(),
					PackType,
					api.FilesetUnpackFilter_Lossless,
					rio.Placement_Direct,
					"file://./fixtures/tar_withBase.tgz",
					rio.Monitor{},
				)
				So(err, ShouldBeNil)
				So(gotWareID, ShouldResemble, api.WareID{"tar", "5y6NvK6GBPQ6CcuNyJyWtSrMAJQ4LVrAcZSoCRAzMSk5o53pkTYiieWyRivfvhZwhZ"})
			})
			Convey("Scan a fixture from gnu tar which lacks a base dir", func() {
				gotWareID, err := Scan(
					context.Background(),
					PackType,
					api.FilesetUnpackFilter_Lossless,
					rio.Placement_Direct,
					"file://./fixtures/tar_sansBase.tgz",
					rio.Monitor{},
				)
				So(err, ShouldBeNil)
				So(gotWareID, ShouldResemble, api.WareID{"tar", "2RLHdc3am6tMCFy56vfcHm5kWLoAtYBfiaQcq17vDm1tEzQn9CC6tcF2yzpAJvehPC"})
			})
			Convey("Scan a kitchen sink fixture tar", func() {
				gotWareID, err := Scan(
					context.Background(),
					PackType,
					api.FilesetUnpackFilter_Lossless,
					rio.Placement_Direct,
					"file://./fixtures/tar_kitchenSink.tgz",
					rio.Monitor{},
				)
				So(err, ShouldBeNil)
				So(gotWareID, ShouldResemble, api.WareID{"tar", "2jkqXaVWCdH7axj1XW56rxZ6WVQ8f46nqMf2BBX7kjLsU9DsvQCquEoy6GcBcQ1Fqc"})
			})
		})
	})
}
