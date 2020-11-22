/*
	The tar transmat packs filesystems into the widely-recognized "tar" format,
	and can use any k/v-styled warehouse for storage.
*/
package tartrans

import (
	api "github.com/polydawn/go-timeless-api"
	"github.com/polydawn/go-timeless-api/rio"
	"github.com/polydawn/rio/transmat/util"
)

const PackType = api.PackType("tar")

var (
	Mirror rio.MirrorFunc = util.CreateMirror(unpackTar)
	Scan   rio.ScanFunc   = util.CreateScanner(PackType, unpackTar)
	Unpack rio.UnpackFunc = util.CreateUnpack(PackType, unpackTar)
)
