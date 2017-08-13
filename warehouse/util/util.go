package util

import (
	"go.polydawn.net/timeless-api"
)

/*
	Return a first, second, and remaining chunk of a ware's hash as strings.

	These are the first three, second three, and remaining bytes of the string.
	For base58 encoded values, these first two chunks used as dir prefixes are a
	cozy density for storing many many thousands of objects:
*/
func ChunkifyHash(wareID api.WareID) (string, string, string) {
	return wareID.Hash[0:3], wareID.Hash[3:6], wareID.Hash[6:]
}
