package osfs

import (
	"go.polydawn.net/rio/fs"
)

func New(basePath fs.AbsolutePath) fs.FS {
	return &osFS{basePath}
}

type osFS struct {
	basePath fs.AbsolutePath
}
