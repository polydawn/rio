package cache

import (
	"fmt"

	"github.com/polydawn/go-timeless-api"
	"github.com/polydawn/rio/fs"
	whutil "github.com/polydawn/rio/warehouse/util"
)

func ShelfFor(wareID api.WareID) fs.RelPath {
	chunk1, chunk2, _ := whutil.ChunkifyHash(wareID)
	return fs.MustRelPath(fmt.Sprintf("%s/fileset/%s/%s/%s",
		wareID.Type,
		chunk1, chunk2, wareID.Hash,
	))
}
