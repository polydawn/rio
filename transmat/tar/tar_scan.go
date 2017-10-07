package tartrans

import (
	"context"
	"net/url"

	. "github.com/polydawn/go-errcat"
	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/go-timeless-api/util"
	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/nilfs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/warehouse"
	"go.polydawn.net/rio/warehouse/impl/kvfs"
	"go.polydawn.net/rio/warehouse/impl/kvhttp"
)

// A "scan" is roughly the same as an unpack to /dev/null,
// but takes a single URL, and *doesn't* require a hash.
//
// It can even populate the CAS cache!
//
// It's basically a toolchain design intent that we have
// this feature, but also force you to use it only very
// knowingly and with moderate inconvenience -- because
// you *should not* do it in the middle of a script; you
// should be doing it *once* and then referencing hitch.

var (
	_ rio.ScanFunc = Scan
)

func Scan(
	ctx context.Context, // Long-running call.  Cancellable.
	packType api.PackType, // The name of pack format.
	filt api.FilesetFilters, // Optionally: filters we should apply while unpacking.
	placementMode rio.PlacementMode, // For scanning only "None" (cache; the default) and "Direct" (don't cache) are valid.
	addr api.WarehouseAddr, // The *one* warehouse to fetch from.  Must be a monowarehouse (not a CA-mode).
	monitor rio.Monitor, // Optionally: callbacks for progress monitoring.
) (_ api.WareID, err error) {
	defer RequireErrorHasCategory(&err, rio.ErrorCategory(""))

	// Sanitize arguments.
	if packType != PackType {
		return api.WareID{}, Errorf(rio.ErrUsage, "this transmat implementation only supports packtype %q (not %q)", PackType, packType)
	}
	if placementMode == "" {
		placementMode = rio.Placement_None
	}
	filt2, err := apiutil.ProcessFilters(filt, apiutil.FilterPurposeUnpack)
	if err != nil {
		return api.WareID{}, Errorf(rio.ErrUsage, "invalid filter specification: %s", err)
	}

	// TODO FUTURE actually support cache

	// Dial warehouse.
	//  Note how this is a subset of the usual accepted warehouses;
	//  it must be a monowarehouse, not a legit CA storage bucket.
	u, err := url.Parse(string(addr))
	if err != nil {
		return api.WareID{}, Errorf(rio.ErrUsage, "failed to parse URI: %s", err)
	}
	var whCtrl warehouse.BlobstoreController
	switch u.Scheme {
	case "file":
		whCtrl, err = kvfs.NewController(addr)
	case "http", "https":
		whCtrl, err = kvhttp.NewController(addr)
	case "ca+file", "ca+http", "ca+https":
		return api.WareID{}, Errorf(rio.ErrUsage, "tar scan doesn't support %q scheme -- content-addressable warehouses contain many wares; which to scan?", u.Scheme)
	default:
		return api.WareID{}, Errorf(rio.ErrUsage, "tar scan doesn't support %q scheme (valid options are 'file', 'http', or 'https')", u.Scheme)
	}
	switch Category(err) {
	case nil:
		// pass
	case rio.ErrWarehouseUnavailable:
		fallthrough // actually fatal here since it's just the one
	default:
		return api.WareID{}, err
	}
	reader, err := whCtrl.OpenReader(api.WareID{"tar", "-"})
	switch Category(err) {
	case nil:
		// pass
	case rio.ErrWareNotFound:
		fallthrough // actually fatal here since it's just the one
	default:
		return api.WareID{}, err
	}
	defer reader.Close()

	// Construct filesystem wrapper to use for all our ops.
	//  If caching, it's a real fs handle;
	//  if not, it's a bunch of no-op'ing functions.
	var afs fs.FS
	switch placementMode {
	case rio.Placement_None:
		afs = osfs.New(fs.MustAbsolutePath("/nope/nope")) // TODO cache
	case rio.Placement_Direct:
		afs = nilFS.New()
	default:
		panic("unreachable")

	}

	// Extract.
	return unpackTar(ctx, afs, filt2, reader)
}
