package rioexecclient

import (
	"go.polydawn.net/timeless-api"
	"go.polydawn.net/timeless-api/rio"
)

func UnpackArgs(
	wareID api.WareID,
	path string,
	filters api.FilesetFilters,
	warehouses []api.WarehouseAddr,
	monitor rio.Monitor,
) ([]string, error) {
	// Required args.
	args := []string{
		"unpack", path, wareID.String(),
	}
	// Append filters if specified.
	//  (We could just pass 'em all even when emptystr, but let's be nice to readers of 'ps'.)
	if filters.Uid != "" {
		args = append(args, "--uid="+filters.Uid)
	}
	if filters.Gid != "" {
		args = append(args, "--gid="+filters.Gid)
	}
	if filters.Mtime != "" {
		args = append(args, "--mtime="+filters.Mtime)
	}
	if filters.Sticky {
		args = append(args, "--sticky")
	}
	// Append warehouses.
	//  Giving this argument repeatedly forms a list in the rio CLI.
	for _, wh := range warehouses {
		args = append(args, "--warehouse="+string(wh))
	}
	// Append monitor options.
	//  (Of which there are currently none meaningful implemented.)
	// Done!
	return args, nil
}
