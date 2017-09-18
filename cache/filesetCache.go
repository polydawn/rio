package cache

import (
	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/rio/fs"
)

type Cache interface {
	ShelfFor(wareID api.WareID) fs.RelPath
}
