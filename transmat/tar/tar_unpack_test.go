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
			wareID, err := Unpack(
				context.Background(),
				api.WareID{"tar", "iJoKZAKWRWTCZ6YQDZYnU2eXjKZ93VyJ1eiis2E3UvbfqXKn8APYeWhSxHyXAtQSG"},
				tmpDir.String(),
				api.FilesetFilters{},
				[]api.WarehouseAddr{"file://./fixtures/tar_withBase.tgz"},
				rio.Monitor{},
			)
			So(err, ShouldBeNil)
			So(wareID, ShouldResemble, api.WareID{"tar", "iJoKZAKWRWTCZ6YQDZYnU2eXjKZ93VyJ1eiis2E3UvbfqXKn8APYeWhSxHyXAtQSG"})
		})
	})
}
