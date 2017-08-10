package tartrans

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/testutil"
	"go.polydawn.net/timeless-api"
	"go.polydawn.net/timeless-api/rio"
)

/*
	Tests against pre-generated, known fixtures of tar binary blobs.

	These tests allow us to cover compat with other tar impls, compression, etc.
*/
func TestTarFixtureUnpack(t *testing.T) {
	Convey("Tar transmat: unpacking of fixtures", t, func() {
		testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
			Convey("Unpack a fixture from gnu tar which includes a base dir", func() {
				wareID := api.WareID{"tar", "5y6NvK6GBPQ6CcuNyJyWtSrMAJQ4LVrAcZSoCRAzMSk5o53pkTYiieWyRivfvhZwhZ"}
				gotWareID, err := Unpack(
					context.Background(),
					wareID,
					tmpDir.String(),
					api.FilesetFilters{},
					[]api.WarehouseAddr{"file://./fixtures/tar_withBase.tgz"},
					rio.Monitor{},
				)
				So(err, ShouldBeNil)
				So(gotWareID, ShouldResemble, wareID)
			})
			Convey("Unpack a fixture from gnu tar which lacks a base dir", func() {
				wareID := api.WareID{"tar", "2RLHdc3am6tMCFy56vfcHm5kWLoAtYBfiaQcq17vDm1tEzQn9CC6tcF2yzpAJvehPC"}
				gotWareID, err := Unpack(
					context.Background(),
					wareID,
					tmpDir.String(),
					api.FilesetFilters{},
					[]api.WarehouseAddr{"file://./fixtures/tar_sansBase.tgz"},
					rio.Monitor{},
				)
				So(err, ShouldBeNil)
				So(gotWareID, ShouldResemble, wareID)
			})
		})
	})
}
