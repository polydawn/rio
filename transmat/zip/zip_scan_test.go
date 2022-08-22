package ziptrans

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	api "github.com/polydawn/go-timeless-api"
	"github.com/polydawn/go-timeless-api/rio"
	"github.com/polydawn/rio/fs"
	"github.com/polydawn/rio/testutil"
)

func TestZipFixtureScan(t *testing.T) {
	Convey("Zip transmat: scan of fixtures", t, func() {
		testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
			Convey("Scan a fixture from zip3.0 which lacks a base dir", func() {
				gotWareID, err := Scan(
					context.Background(),
					PackType,
					api.FilesetUnpackFilter_Lossless,
					rio.Placement_Direct,
					"file://./fixtures/withbase.zip",
					rio.Monitor{},
				)
				So(err, ShouldBeNil)
				So(gotWareID, ShouldResemble, api.WareID{"zip", "6c1eVnQ9NutqZSMD5gimy72u3gZMcp4mFAVbQhAkpwTvTH1CCnGgL6yvBJ6MNkWUYZ"})
			})
		})
	})
}
