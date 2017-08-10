/*
	Provides helper functions for checking if we have some functional sets of capabilities.
*/
package caps

import (
	"os"
	"runtime"

	"github.com/syndtr/gocapability/capability"
)

func Scan() *Fulcrum {
	var err error
	f := &Fulcrum{}
	f.onLinux = runtime.GOOS == "linux"
	f.ourUID = os.Getuid()
	if f.onLinux {
		f.ourCaps, err = capability.NewPid(0) // zero means self
		if err != nil {
			panic(err)
		}
	}
	return f
}

type Fulcrum struct {
	onLinux bool
	ourUID  int
	ourCaps capability.Capabilities // valid on linux; nil on mac (causing completely different logic).
}

// Whether we have enough caps to confidently access all of `$RIO_BASE/*`.
// We sum this up as "have CAP_DAC_OVERRIDE";
// or, on mac, is uid==0.
func (f Fulcrum) CanShareIOCache() bool {
	if !f.onLinux {
		return f.ourUID == 0
	}
	return f.ourCaps.Get(capability.EFFECTIVE, capability.CAP_DAC_OVERRIDE)
}

// Whether we have enough caps to confidently use materialize files with ownership info.
// This requires "have CAP_CHOWN", but also "have CAP_FOWNER" (because we need this cap
// in order to be able to set mtimes on files *after having chown'd them*);
// or, on mac, is uid==0.
func (f Fulcrum) CanManageOwnership() bool {
	if !f.onLinux {
		return f.ourUID == 0
	}
	return f.ourCaps.Get(capability.EFFECTIVE, capability.CAP_CHOWN|capability.CAP_FOWNER)
}
