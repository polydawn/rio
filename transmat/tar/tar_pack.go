package tartrans

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"io"
	"net/url"
	"time"

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
	_ rio.PackFunc = Pack
)

func Pack(
	ctx context.Context, // Long-running call.  Cancellable.
	path string, // The fileset to scan and pack (absolute path).
	filters api.FilesetFilters, // Optionally: filters we should apply while unpacking.
	warehouse api.WarehouseAddr, // Warehouse to save into (or blank to just scan).
	monitor rio.Monitor, // Optionally: callbacks for progress monitoring.
) (api.WareID, error) {
	// Sanitize arguments.
	path2 := fs.MustAbsolutePath(path)

	// Connect to warehouse, and get write controller opened.
	// FUTURE : this write controller abstraction should be a reusable interface.
	var wc *kvfs.WriteController
	// REVIEW ... Do I really have to parse this again?  is this sanely encapsulated?
	u, err := url.Parse(string(warehouse))
	if err != nil {
		return api.WareID{}, Errorf(rio.ErrUsage, "failed to parse URI: %s", err)
	}
	switch u.Scheme {
	case "file", "file+ca":
		whCtrl, err := kvfs.NewController(warehouse)
		switch Category(err) {
		case nil:
			// pass
		case rio.ErrWarehouseUnavailable:
			return api.WareID{}, err
		default:
			return api.WareID{}, err
		}
		wc, err = whCtrl.OpenWriter()
		switch Category(err) {
		case nil:
			// pass
		case rio.ErrWarehouseUnwritable:
			return api.WareID{}, err
		default:
			return api.WareID{}, err
		}
	default:
		return api.WareID{}, Errorf(rio.ErrUsage, "tar pack doesn't support %q scheme (valid options are 'file' or 'file+ca')", u.Scheme)
	}
	defer wc.Close()

	// Wrap writer stream to do compress on the way out.
	//  Note on compression levels: The default is 6; and per http://tukaani.org/lzma/benchmarks.html
	//  this appears quite reasonable: higher levels appear to have minimal size payoffs, but significantly rising compress time costs;
	//  decompression time does not vary with compression level.
	// Save a gzip reference just to close it; tar.Writer doesn't passthru its own close.
	gzWriter := gzip.NewWriter(wc)
	defer gzWriter.Close()

	// Construct tar writer.
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Scan and tarify!
	return packTar(ctx, path2, filters, tarWriter)
}

func packTar(
	ctx context.Context,
	srcBasePath fs.AbsolutePath,
	filters api.FilesetFilters,
	tw *tar.Writer,
) (api.WareID, error) {
	// Allocate bucket for keeping each metadata entry and content hash;
	// the full tree hash will be computed from this at the end.
	bucket := &fshash.MemoryBucket{}

	// Construct filesystem wrapper to use for all our ops.
	afs := osfs.New(srcBasePath)

	// Walk the filesystem, emitting tar entries and filling the bucket as we go.
	var tarHeader *tar.Header
	preVisit := func(filenode *fs.FilewalkNode) error {
		if filenode.Err != nil {
			return filenode.Err
		}

		// Consider cancellation.
		if ctx.Err() != nil {
			return Errorf(rio.ErrCancelled, "cancelled")
		}

		// Open file.
		fmeta, file, err := fsOp.ScanFile(afs, filenode.Info.Name) // FIXME : we already have the full metadata loaded; give ScanFile option to accept it!
		if err != nil {
			return err
		}

		// Apply filters.
		ApplyMaterializeFilters(fmeta, filters)

		// Flatten time to seconds.  The tar writer impl doesn't do subsecond precision.
		//  The writer will always flatten it internally, but we need to do it here as well
		//  so that the hash and the serial form are describing the same thing.
		fmeta.Mtime = fmeta.Mtime.Truncate(time.Second)

		// Flip our metadata to tar header format, and flush it.
		MetadataToTarHdr(fmeta, tarHeader)
		tw.WriteHeader(tarHeader)

		// If it's a file, stream the body into the tar while hashing; for all,
		//  record the metadata in the bucket for the total hash.
		if file == nil {
			bucket.AddRecord(*fmeta, nil)
		} else {
			defer file.Close()
			hasher := sha512.New384()
			tee := io.MultiWriter(tw, hasher)
			_, err := io.Copy(tee, file)
			if err != nil {
				return err
			}
			bucket.AddRecord(*fmeta, hasher.Sum(nil))
		}
		return nil
	}
	if err := fs.Walk(afs, preVisit, nil); err != nil {
		return api.WareID{}, err
	}

	// Hash the thing!
	hash := fshash.HashBucket(bucket, sha512.New384)
	return api.WareID{"tar", misc.Base58Encode(hash)}, nil
}
