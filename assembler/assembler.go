package assembler

import (
	"context"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/rio/fs"
)

type Assembler struct {
	targetBase     fs.AbsolutePath
	cacheDir       fs.AbsolutePath
	unpackTool     rio.UnpackFunc
	placerTool     func( /*todo*/ )
	fillerDirProps fs.Metadata
}

/*
	Unpacking with an assembler proxies the unpack command,
	but transforms the unpack path into one relative to the assembler's configured base path,
	then runs the unpack in its own goroutine,
	performs the unpack into a CAS cache area,
	and finally uses a 'placer' to get the now-cached content
	to appear in the (translated) target path.

	Since the work is done in parallel, the wareID returned is always zero,
	and any errors will be usage and args errors;
	the caller must check the error returned by `Wait`,
	which will return any other gathered errors.

	If the unpack path is deep, the assembler will create parent dirs as necessary.
	TODO finish defining ordering behavior and mounts and such.
*/
func (a *Assembler) Unpack(
	ctx context.Context, // Long-running call.  Cancellable.
	wareID api.WareID, // What wareID to fetch for unpacking.
	path string, // Where to unpack the fileset (absolute path).
	filters api.FilesetFilters, // Optionally: filters we should apply while unpacking.
	warehouses []api.WarehouseAddr, // Warehouses we can try to fetch from.
	monitor rio.Monitor, // Optionally: callbacks for progress monitoring.
) (api.WareID, error) {
	return api.WareID{}, nil
}

func (a *Assembler) Wait() error {
	return nil
}
