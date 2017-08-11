package tests

import (
	"context"
	"fmt"

	. "github.com/smartystreets/goconvey/convey"

	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/testutil"
	"go.polydawn.net/timeless-api"
	"go.polydawn.net/timeless-api/rio"
)

func CheckPackProducesConsistentHash(pack rio.PackFunc) {
	Convey("SPEC: Applying the PackFunc to a filesystem twice should produce the same hash", func() {
		for _, fixture := range AllFixtures {
			Convey(fmt.Sprintf("- Fixture %q", fixture.Name), func() {
				testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
					afs := osfs.New(tmpDir)
					// Set up fixture.
					PlaceFixture(afs, fixture.Files)
					// Pack (to /dev/null) once.
					wareID1, err := pack(
						context.Background(),
						tmpDir.String(),
						api.FilesetFilters{},
						"",
						rio.Monitor{},
					)
					So(err, ShouldBeNil)
					// Pack (to /dev/null) from the same path a second time.
					wareID2, err := pack(
						context.Background(),
						tmpDir.String(),
						api.FilesetFilters{},
						"",
						rio.Monitor{},
					)
					So(err, ShouldBeNil)
					// Should be same output.
					//  This is both an assertion that the pack hasher is consistent,
					//  and that it's not making arbitrary mutations during its passage.
					So(wareID1, ShouldResemble, wareID2)
				})
			})
		}
	})
}
