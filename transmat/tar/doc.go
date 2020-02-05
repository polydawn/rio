/*
	The tar transmat packs filesystems into the widely-recognized "tar" format,
	and can use any k/v-styled warehouse for storage.
*/
package tartrans

import (
	api "go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/rio/transmat/util"
)

const PackType = api.PackType("tar")

var (
	Mirror rio.MirrorFunc = util.CreateMirror(unpackTar)
	Scan   rio.ScanFunc   = util.CreateScanner(PackType, unpackTar)
	Unpack rio.UnpackFunc = util.CreateUnpack(PackType, unpackTar)
)
