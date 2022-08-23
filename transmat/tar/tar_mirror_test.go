package tartrans

import (
	"fmt"
	"testing"

	api "github.com/polydawn/go-timeless-api"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/polydawn/rio/fs"
	"github.com/polydawn/rio/fs/osfs"
	"github.com/polydawn/rio/testutil"
	"github.com/polydawn/rio/transmat/mixins/tests"
)

func TestTarMirror(t *testing.T) {
	Convey("Spec compliance: Tar mirror", t,
		testutil.Requires(testutil.RequiresCanManageOwnership, func() {
			Convey("Populating kvfs warehouse, in content-addressable mode, from kvfs warehouse, in content-addressable mode:", func() {
				testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
					osfs.New(tmpDir).Mkdir(fs.MustRelPath("src"), 0755)
					osfs.New(tmpDir).Mkdir(fs.MustRelPath("dst"), 0755)
					srcAddr := api.WarehouseLocation(fmt.Sprintf("ca+file://%s/src", tmpDir))
					dstAddr := api.WarehouseLocation(fmt.Sprintf("ca+file://%s/dst", tmpDir))

					tests.CheckMirror(PackType, Mirror, Pack, Unpack, dstAddr, srcAddr)
				})
			})
		}),
	)
}
