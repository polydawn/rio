package tartrans

import (
	"context"

	. "github.com/warpfork/go-errcat"
	api "go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/rio/fs"
	nilFS "go.polydawn.net/rio/fs/nilfs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/transmat/util"
)

// A "scan" is roughly the same as an unpack to /dev/null,
// but takes a single URL, and *doesn't* require a hash.
//
// It can even populate the CAS cache!
//
// However, note that it's an overall intention to make this feature
// usable only very knowingly and with moderate inconvenience -- because you
// *should not* do it in the middle of a script; you should be doing any scans
// *once* and then tracking the resulting references via a release catalog:
// which keeps the overall process more controlled, auditable, and
// well-defined even in the case of untrusted networks.

var (
	_ rio.ScanFunc = Scan
)

func Scan(
	ctx context.Context, // Long-running call.  Cancellable.
	packType api.PackType, // The name of pack format.
	filt api.FilesetUnpackFilter, // Optionally: filters we should apply while unpacking.
	placementMode rio.PlacementMode, // For scanning only "None" (cache; the default) and "Direct" (don't cache) are valid.
	addr api.WarehouseLocation, // The *one* warehouse to fetch from.  Must be a monowarehouse (not a CA-mode).
	mon rio.Monitor, // Optionally: callbacks for progress monitoring.
) (_ api.WareID, err error) {
	if mon.Chan != nil {
		defer close(mon.Chan)
	}
	defer RequireErrorHasCategory(&err, rio.ErrorCategory(""))

	// Sanitize arguments.
	if packType != PackType {
		return api.WareID{}, Errorf(rio.ErrUsage, "this transmat implementation only supports packtype %q (not %q)", PackType, packType)
	}
	if !filt.IsComplete() {
		return api.WareID{}, Errorf(rio.ErrUsage, "filters must be completely specified")
	}
	if placementMode == "" {
		placementMode = rio.Placement_None
	}

	// TODO FUTURE actually support cache

	// Dial warehouse.
	//  Note how this is a subset of the usual accepted warehouses;
	//  it must be a monowarehouse, not a legit CA storage bucket.
	reader, err := util.PickReader(api.WareID{"tar", "-"}, []api.WarehouseLocation{addr}, true, mon)
	if err != nil {
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
	//  For once we can actually discard the *prefilter* wareID, since we don't have
	//  an expected one to assert against.
	_, unpackedWareID, err := unpackTar(ctx, afs, filt, reader, mon)
	return unpackedWareID, err
}
