package placer

import (
	"go.polydawn.net/rio/fs"
)

type Placer func(srcPath, dstPath fs.AbsolutePath, writable bool) error

/*
	The copy placer is always defined and always supported and is never swappable.

	The placer used to handle "mount"-mode placements may vary drastically, however.
*/

/*
	Returns the most reasonable mounting placer implementation available on this platform.

	If the environment var RIO_MOUNT_PLACER is set, we'll either return that or
	an error explaing why it's not available;
	otherwise, autodetection will examine what filesystem drivers and capabilities
	are available, and try to pick the most performant/reliable/sensible thing available.

	For placers that need a working dir, one will be created under RIO_MOUNT_WORKDIR
	if set,	or RIO_BASE/wrk.
*/
func MountPlacer() Placer {
	return nil
}
