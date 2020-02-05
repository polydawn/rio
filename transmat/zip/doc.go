/*Package ziptrans packs filesystems into the ZIP archive format.
 *they can then use any k/v-styled warehouse for storage.
 */
package ziptrans

import (
	api "go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/rio/transmat/util"
)

// PackType defines this as the zip packing type.
const PackType = api.PackType("zip")

var (
	Mirror rio.MirrorFunc = util.CreateMirror(unpackZip)
	Scan   rio.ScanFunc   = util.CreateScanner(PackType, unpackZip)
	Unpack rio.UnpackFunc = util.CreateUnpack(PackType, unpackZip)
)
