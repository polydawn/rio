package main

import (
	. "github.com/warpfork/go-errcat"

	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/rio/transmat/git"
	tartrans "go.polydawn.net/rio/transmat/tar"
	ziptrans "go.polydawn.net/rio/transmat/zip"
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
