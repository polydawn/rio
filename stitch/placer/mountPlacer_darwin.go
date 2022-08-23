//go:build darwin
// +build darwin

package placer

import (
	"errors"

	"github.com/polydawn/rio/fs"
)

func NewAufsPlacer(workDir fs.AbsolutePath) (Placer, error) {
	return nil, errors.New("unsupported mount placer")
}

func NewOverlayPlacer(workDir fs.AbsolutePath) (Placer, error) {
	return nil, errors.New("unsupported mount placer")
}

func BindPlacer(srcPath, dstPath fs.AbsolutePath, writable bool) (Janitor, error) {
	return nil, errors.New("unsupported mount placer")
}
