package defaults

import (
	"time"
)

var (
	FilterDefaultUid   uint32 = 1000
	FilterDefaultGid   uint32 = 1000
	FilterDefaultMtime        = time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)
)
