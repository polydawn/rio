package tartrans

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.polydawn.net/rio/testutil"
	"go.polydawn.net/rio/transmat/mixins/tests"
)

func TestTarPack(t *testing.T) {
	Convey("Spec compliance: Tar pack", t,
		testutil.Requires(testutil.RequiresCanManageOwnership, func() {
			tests.CheckPackProducesConsistentHash(Pack)
			tests.CheckPackHashVariesOnVariations(Pack)
		}),
	)
}
