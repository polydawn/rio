package main

import (
	. "github.com/warpfork/go-errcat"

	"github.com/polydawn/go-timeless-api/rio"
	"github.com/polydawn/rio/transmat/git"
	tartrans "github.com/polydawn/rio/transmat/tar"
	ziptrans "github.com/polydawn/rio/transmat/zip"
)

func demuxPackTool(packType string) (rio.PackFunc, error) {
	switch packType {
	case "tar":
		return tartrans.Pack, nil
	case "zip":
		return ziptrans.Pack, nil
	default:
		return nil, Errorf(rio.ErrUsage, "unsupported packtype %q", packType)
	}
}

func demuxUnpackTool(packType string) (rio.UnpackFunc, error) {
	switch packType {
	case "tar":
		return tartrans.Unpack, nil
	case "git":
		return git.Unpack, nil
	case "zip":
		return ziptrans.Unpack, nil
	default:
		return nil, Errorf(rio.ErrUsage, "unsupported packtype %q", packType)
	}
}

func demuxScanTool(packType string) (rio.ScanFunc, error) {
	switch packType {
	case "tar":
		return tartrans.Scan, nil
	case "zip":
		return ziptrans.Scan, nil
	default:
		return nil, Errorf(rio.ErrUsage, "unsupported packtype %q", packType)
	}
}

func demuxMirrorTool(packType string) (rio.MirrorFunc, error) {
	switch packType {
	case "tar":
		return tartrans.Mirror, nil
	case "zip":
		return ziptrans.Mirror, nil
	default:
		return nil, Errorf(rio.ErrUsage, "unsupported packtype %q", packType)
	}
}
