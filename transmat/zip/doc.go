/*Package ziptrans packs filesystems into the ZIP archive format.
 *they can then use any k/v-styled warehouse for storage.
 */
package ziptrans

import (
	api "go.polydawn.net/go-timeless-api"
)

// PackType defines this as the zip packing type.
const PackType = api.PackType("zip")
