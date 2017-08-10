package tartrans

import (
	"archive/tar"
	"context"
	"crypto/sha512"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/polydawn/refmt/misc"

	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/fsOp"
	. "go.polydawn.net/rio/lib/errcat"
	"go.polydawn.net/rio/transmat/mixins/fshash"
	"go.polydawn.net/rio/warehouse/impl/kvfs"
	"go.polydawn.net/timeless-api"
	"go.polydawn.net/timeless-api/rio"
)

var (
	_ rio.UnpackFunc = Unpack
)

func Unpack(
	ctx context.Context, // Long-running call.  Cancellable.
	wareID api.WareID, // What wareID to fetch for unpacking.
	path string, // Where to unpack the fileset (absolute path).
	filters api.FilesetFilters, // Optionally: filters we should apply while unpacking.
	warehouses []api.WarehouseAddr, // Warehouses we can try to fetch from.
	monitor rio.Monitor, // Optionally: callbacks for progress monitoring.
) (api.WareID, error) {
	// Sanitize arguments.
	path2 := fs.MustAbsolutePath(path)

	// Pick a warehouse.
	//  With K/V warehouses, this takes the form of "pick the first one that answers".
	var reader io.ReadCloser
	for _, addr := range warehouses {
		// REVIEW ... Do I really have to parse this again?  is this sanely encapsulated?
		u, err := url.Parse(string(addr))
		if err != nil {
			return api.WareID{}, Errorf(rio.ErrUsage, "failed to parse URI: %s", err)
		}
		switch u.Scheme {
		case "file", "file+ca":
			whCtrl, err := kvfs.NewController(addr)
			switch Category(err) {
			case nil:
				// pass
			case rio.ErrWarehouseUnavailable:
				// TODO log something to the monitor
				continue // okay!  skip to the next one.
			default:
				return api.WareID{}, err
			}
			reader, err = whCtrl.OpenReader(wareID)
			switch Category(err) {
			case nil:
				// pass
			case rio.ErrWareNotFound:
				// TODO log something to the monitor
				continue // okay!  skip to the next one.
			default:
				return api.WareID{}, err
			}
		default:
			return api.WareID{}, Errorf(rio.ErrUsage, "tar unpack doesn't support %q scheme (valid options are 'file' or 'file+ca')", u.Scheme)
		}
	}
	if reader == nil { // aka if no warehouses available:
		return api.WareID{}, Errorf(rio.ErrWarehouseUnavailable, "no warehouses were available!")
	}
	defer reader.Close()

	// Wrap input stream with decompression as necessary.
	//  Which kind of decompression to use can be autodetected by magic bytes.
	reader2, err := Decompress(reader)
	if err != nil {
		return api.WareID{}, Errorf(rio.ErrWareCorrupt, "corrupt tar compression: %s", err)
	}

	// Convert the raw byte reader to a tar stream.
	tarReader := tar.NewReader(reader2)

	// Extract.
	return unpackTar(ctx, path2, filters, tarReader)
}

func unpackTar(
	ctx context.Context,
	destBasePath fs.AbsolutePath,
	filters api.FilesetFilters,
	tr *tar.Reader,
) (api.WareID, error) {
	// Allocate bucket for keeping each metadata entry and content hash;
	// the full tree hash will be computed from this at the end.
	bucket := &fshash.MemoryBucket{}

	// Construct filesystem wrapper to use for all our ops.
	afs := osfs.New(destBasePath)

	// Iterate over each tar entry, mutating filesystem as we go.
	for {
		fmeta := fs.Metadata{}
		thdr, err := tr.Next()

		// Check for done.
		if err == io.EOF {
			break // sucess!  end of archive.
		}
		if err != nil {
			return api.WareID{}, Errorf(rio.ErrWareCorrupt, "corrupt tar: %s", err)
		}
		if ctx.Err() != nil {
			return api.WareID{}, Errorf(rio.ErrCancelled, "cancelled")
		}

		// Reshuffle metainfo to our default format.
		if err := TarHdrToMetadata(thdr, &fmeta); err != nil {
			return api.WareID{}, err
		}
		if strings.HasPrefix(fmeta.Name.String(), "..") {
			return api.WareID{}, Errorf(rio.ErrWareCorrupt, "corrupt tar: paths that use '../' to leave the base dir are invalid")
		}

		// Apply filters.
		ApplyMaterializeFilters(&fmeta, filters)

		// Infer parents, if necessary.  The tar format allows implicit parent dirs.
		//
		// Note that if any of the implicitly conjured dirs is specified later, unpacking won't notice,
		// but bucket hashing iteration will (correctly) blow up for repeat entries.
		// It may well be possible to construct a tar like that, but it's already well established that
		// tars with repeated filenames are just asking for trouble and shall be rejected without
		// ceremony because they're just a ridiculous idea.

		for parent := fmeta.Name.Dir(); parent != (fs.RelPath{}); parent = parent.Dir() {
			_, err := os.Lstat(destBasePath.Join(parent).String())
			// if it already exists, move along; if the error is anything interesting, let PlaceFile decide how to deal with it
			if err == nil || !os.IsNotExist(err) {
				continue
			}
			// if we're missing a dir, conjure a node with defaulted values (same as we do for "./")
			conjuredFmeta := fshash.DefaultDirMetadata()
			conjuredFmeta.Name = parent
			fsOp.PlaceFile(afs, conjuredFmeta, nil, false)
			bucket.AddRecord(conjuredFmeta, nil)
		}
	}

	// Hash the thing!
	hash := fshash.HashBucket(bucket, sha512.New384)
	return api.WareID{"tar", misc.Base58Encode(hash)}, nil
}
