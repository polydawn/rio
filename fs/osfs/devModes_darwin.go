// +build darwin

// Devmodes in darwin puts major in the upper 8 bits per sys/types.h.
package osfs

func devModesSplit(rdev int32) (major int64, minor int64) {
	// Constants herein are not a joy: they're a workaround for https://github.com/golang/go/issues/8106
	return int64((rdev >> 24) & 0xff), int64(rdev & 0xffffff)
}
