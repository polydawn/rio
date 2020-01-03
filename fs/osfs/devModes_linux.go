// +build linux

// Devmodes in most linux follows the gnu_dev_major / gnu_dev_minor library functions.
package osfs

func devModesSplit(rdev uint64) (major int64, minor int64) {
	// Constants herein are not a joy: they're a workaround for https://github.com/golang/go/issues/8106
	return int64((rdev >> 8) & 0xff), int64((rdev & 0xff) | ((rdev >> 12) & 0xfff00))
}
