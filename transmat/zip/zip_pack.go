package ziptrans

import (
	"archive/zip"
	"context"
	"crypto/sha512"
	"io"

	"github.com/polydawn/refmt/misc"

	api "github.com/polydawn/go-timeless-api"
	"github.com/polydawn/go-timeless-api/rio"
	"github.com/polydawn/rio/fs"
	"github.com/polydawn/rio/fs/osfs"
	"github.com/polydawn/rio/fsOp"
	"github.com/polydawn/rio/transmat/mixins/filters"
	"github.com/polydawn/rio/transmat/mixins/fshash"
	"github.com/polydawn/rio/transmat/util"
	. "github.com/warpfork/go-errcat"
)

var (
	_ rio.PackFunc = Pack
)

// Pack transmutes a defined fileset into a given warehouse.
func Pack(
	ctx context.Context, // Long-running call.  Cancellable.
	packType api.PackType, // The name of pack format.
	pathStr string, // The fileset to scan and pack (absolute path).
	filt api.FilesetPackFilter, // Filters we should apply while packing.
	warehouseAddr api.WarehouseLocation, // Warehouse to save into (or blank to just scan).
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
	path, err := fs.ParseAbsolutePath(pathStr)
	if err != nil {
		return api.WareID{}, Errorf(rio.ErrUsage, "pack must be called with absolute path: %s", err)
	}

	// Short-circuit exit if the path does not exist.
	//  We could let the errors later bubble, but, why bother opening a writeController,
	//  etc, if we're just going to have to rm the resource a millisecond later?
	afs := osfs.New(path)
	_, err = afs.Stat(fs.RelPath{})
	switch Category(err) {
	case nil:
		// pass
	case fs.ErrNotExists:
		return api.WareID{PackType, ""}, nil
	default:
		return api.WareID{}, Errorf(rio.ErrPackInvalid, "cannot read path for packing: %s", err)
	}

	// Connect to warehouse, and get write controller opened.
	wc, err := util.OpenWriteController(warehouseAddr, packType, mon)
	if err != nil {
		return api.WareID{}, err
	}
	defer wc.Close()

	// Construct zip writer.
	zipWriter := zip.NewWriter(wc)

	// Scan and zip!
	wareID, err := packZip(ctx, afs, filt, zipWriter)
	if err != nil {
		return wareID, err
	}
	// Close all the intermediate writer layers to ensure they've flushed.
	err = zipWriter.Close()
	if err != nil {
		return wareID, err
	}

	// If we made it all the way with no errors, commit.
	//  (Otherwise, the write controller will be closed by default by our defers.)
	return wareID, wc.Commit(wareID)
}

func packZip(
	ctx context.Context,
	afs fs.FS,
	filt api.FilesetPackFilter,
	zw *zip.Writer,
) (api.WareID, error) {
	// Allocate bucket for keeping each metadata entry and content hash;
	// the full tree hash will be computed from this at the end.
	bucket := &fshash.MemoryBucket{}

	// Walk the filesystem, emitting entries and filling the bucket as we go.
	zipHeader := &zip.FileHeader{}
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
		//  The filter may reject things by returning an error;
		//   or, instruct us to ignore things by setting the type to invalid.
		if err := filters.ApplyPackFilter(filt, fmeta); err != nil {
			return err
		}
		if fmeta.Type == fs.Type_Invalid {
			return nil // skip it and continue the walk
		}

		// Flip our metadata to zip header format, and flush it.
		zipHeader = new(zip.FileHeader)
		MetadataToZipHdr(fmeta, zipHeader)

		fw, err := zw.CreateHeader(zipHeader)
		if err != nil {
			return Errorf(rio.ErrWarehouseUnwritable, "error while writing pack: %s", err)
		}

		// If it's a file, stream the body into the file while hashing; for all,
		//  record the metadata in the bucket for the total hash.
		if file == nil && fmeta.Type == fs.Type_Symlink {
			hasher := sha512.New384()
			tee := io.MultiWriter(fw, hasher)
			_, err := tee.Write([]byte(fmeta.Linkname))
			if err != nil {
				return err
			}
			bucket.AddRecord(*fmeta, hasher.Sum(nil))
		} else if file == nil {
			fw.Write([]byte{})
			bucket.AddRecord(*fmeta, nil)
		} else {
			defer file.Close()
			hasher := sha512.New384()
			tee := io.MultiWriter(fw, hasher)
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
	return api.WareID{"zip", misc.Base58Encode(hash)}, nil
}
