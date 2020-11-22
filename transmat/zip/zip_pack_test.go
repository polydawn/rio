package ziptrans

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/polydawn/rio/testutil"
	"github.com/polydawn/rio/transmat/mixins/tests"
)

func TestZipPack(t *testing.T) {
	Convey("Spec compliance: zip pack", t,
		testutil.Requires(testutil.RequiresCanManageOwnership, func() {
			tests.CheckPackProducesConsistentHash(PackType, Pack)
			tests.CheckPackHashVariesOnVariations(PackType, Pack)
			tests.CheckPackErrorsGracefully(PackType, Pack)
		}),
	)
}
