package ziptrans

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha512"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/polydawn/refmt/misc"
	. "github.com/warpfork/go-errcat"

	api "go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/rio/config"
	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/fsOp"
	"go.polydawn.net/rio/lib/treewalk"
	"go.polydawn.net/rio/transmat/mixins/buffer"
	"go.polydawn.net/rio/transmat/mixins/cache"
	"go.polydawn.net/rio/transmat/mixins/filters"
	"go.polydawn.net/rio/transmat/mixins/fshash"
	"go.polydawn.net/rio/transmat/mixins/log"
	"go.polydawn.net/rio/transmat/util"
)

var (
	_ rio.UnpackFunc = Unpack
)

func Unpack(
	ctx context.Context, // Long-running call.  Cancellable.
	wareID api.WareID, // What wareID to fetch for unpacking.
	path string, // Where to unpack the fileset (absolute path).
	filt api.FilesetUnpackFilter, // Optionally: filters we should apply while unpacking.
	placementMode rio.PlacementMode, // Optionally: a placement mode (default is "copy").
	warehouses []api.WarehouseLocation, // Warehouses we can try to fetch from.
	mon rio.Monitor, // Optionally: callbacks for progress monitoring.
) (_ api.WareID, err error) {
	if mon.Chan != nil {
		defer close(mon.Chan)
	}
	defer RequireErrorHasCategory(&err, rio.ErrorCategory(""))

	// Sanitize arguments.
	if wareID.Type != PackType {
		return api.WareID{}, Errorf(rio.ErrUsage, "this transmat implementation only supports packtype %q (not %q)", PackType, wareID.Type)
	}
	if !filt.IsComplete() {
		return api.WareID{}, Errorf(rio.ErrUsage, "filters must be completely specified")
	}
	if placementMode == "" {
		placementMode = rio.Placement_Copy
	}
	// Wrap the direct unpack func with cache behavior; call that.
	return cache.Lrn2Cache(
		osfs.New(config.GetCacheBasePath()),
		unpack,
	)(ctx, wareID, path, filt, placementMode, warehouses, mon)
}

func unpack(
	ctx context.Context,
	wareID api.WareID,
	path string,
	filt api.FilesetUnpackFilter,
	placementMode rio.PlacementMode,
	warehouses []api.WarehouseLocation,
	mon rio.Monitor,
) (_ api.WareID, err error) {
	defer RequireErrorHasCategory(&err, rio.ErrorCategory(""))

	// Sanitize arguments.
	path2 := fs.MustAbsolutePath(path)

	// Pick a warehouse and get a reader.
	reader, err := util.PickReader(wareID, warehouses, false, mon)
	if err != nil {
		return api.WareID{}, err
	}
	defer reader.Close()

	// Construct filesystem wrapper to use for all our ops.
	afs := osfs.New(path2)

	// Extract.
	prefilterWareID, unpackWareID, err := unpackZip(ctx, afs, filt, wareID, reader, mon)
	if err != nil {
		return unpackWareID, err
	}

	// Check for hash mismatch before returning, because that IS an error,
	//  but also return the hash we got either way.
	if prefilterWareID != wareID {
		return unpackWareID, ErrorDetailed(
			rio.ErrWareHashMismatch,
			fmt.Sprintf("hash mismatch: expected %q, got %q (filtered %q)", wareID, prefilterWareID, unpackWareID),
			map[string]string{
				"expected": wareID.String(),
				"actual":   prefilterWareID.String(),
				"filtered": unpackWareID.String(),
			},
		)
	}
	return unpackWareID, nil
}

func unpackZip(
	ctx context.Context,
	afs fs.FS,
	filt api.FilesetUnpackFilter,
	archiveWareID api.WareID,
	reader io.Reader,
	mon rio.Monitor,
) (
	prefilterWareID api.WareID,
	actualWareID api.WareID,
	err error,
) {
	defer RequireErrorHasCategory(&err, rio.ErrorCategory(""))

	b := buffer.NewTemporaryBuffer()
	defer b.Close()
	readerAt, err := b.SectionReader(ctx, archiveWareID, reader, mon)
	if err != nil {
		return api.WareID{}, api.WareID{}, err
	}

	// Convert the raw byte reader to a zip stream.
	zr, err := zip.NewReader(readerAt, readerAt.Size())
	if err != nil {
		return api.WareID{}, api.WareID{}, err
	}

	// Allocate bucket for keeping each metadata entry and content hash;
	// the full tree hash will be computed from this at the end.
	// We keep one for the raw ware data as we consume it, so we can verify no fuckery;
	// we keep a second, separate one for the filtered data, which will compute a different hash.
	prefilterBucket := &fshash.MemoryBucket{}
	filteredBucket := &fshash.MemoryBucket{}

	// Also allocate a map for keeping records of which dirs we've created.
	dirs := map[fs.RelPath]struct{}{}

	// Iterate over each entry, mutating filesystem as we go.
	for _, f := range zr.File {
		fmeta := fs.Metadata{}

		// Check for done.
		if err == io.EOF {
			break // sucess!  end of archive.
		}
		if err != nil {
			return api.WareID{}, api.WareID{}, Errorf(rio.ErrWareCorrupt, "corrupt zip: %s", err)
		}
		if ctx.Err() != nil {
			return api.WareID{}, api.WareID{}, Errorf(rio.ErrCancelled, "cancelled")
		}

		// Reshuffle metainfo to our default format.
		skipMe, haltMe := ZipHdrToMetadata(&f.FileHeader, &fmeta)
		if skipMe != nil {
			// n.b. every time this happens in practice so far, it's a `g` header,
			//  and it's got a git commit hash in the paxrecords -- purely advisory info.
			//  It would be nice to turn those down into a debug message, and react
			//  a little more startledly to any other values.  Until then, "warn"
			//  even though every time we've seen this it's harmless.
			mon.Send(rio.Event_Log{
				Time:  time.Now(),
				Level: rio.LogWarn,
				Msg:   fmt.Sprintf("unpacking: skipping an entry: %s", skipMe),
				Detail: [][2]string{
					{"path", fmeta.Name.String()},
					{"skipreason", skipMe.Error()},
					{"ziphdr", fmt.Sprintf("%#v", f)},
				},
			})
			continue
		}
		if haltMe != nil {
			return api.WareID{}, api.WareID{}, haltMe
		}
		if strings.HasPrefix(fmeta.Name.String(), "..") {
			return api.WareID{}, api.WareID{}, Errorf(rio.ErrWareCorrupt, "corrupt zip: paths that use '../' to leave the base dir are invalid")
		}

		// Infer parents, if necessary.  The zip format should not allow implicit dirs, but we allow
		// it for tars, so why not here.
		for _, parent := range fmeta.Name.SplitParent() {
			// If we already initialized this parent, superb; move along.
			if _, exists := dirs[parent]; exists {
				continue
			}
			// If we're missing a dir, conjure a node with defaulted values.
			log.DirectoryInferred(mon, parent, fmeta.Name)
			conjuredFmeta := fshash.DefaultDirMetadata()
			conjuredFmeta.Name = parent
			prefilterBucket.AddRecord(conjuredFmeta, nil)
			filters.ApplyUnpackFilter(filt, &conjuredFmeta)
			filteredBucket.AddRecord(conjuredFmeta, nil)
			dirs[conjuredFmeta.Name] = struct{}{}
			if err := fsOp.PlaceFile(afs, conjuredFmeta, nil, false); err != nil {
				return api.WareID{}, api.WareID{}, Errorf(rio.ErrInoperablePath, "error while unpacking: %s", err)
			}
		}

		// Apply filters.
		//  ... uck, to one copy of the meta.  We can't add either to their buckets
		//  until after the file is placed because we need the content hash.
		filteredFmeta := fmeta
		//  The filter may reject things by returning an error;
		//   or, instruct us to ignore things by setting the type to invalid.
		if err := filters.ApplyUnpackFilter(filt, &filteredFmeta); err != nil {
			return api.WareID{}, api.WareID{}, err
		}
		if fmeta.Type == fs.Type_Invalid {
			// skip placing that file and continue processing...
			//  but *do* still record it in the prefilter bucket for hashing.
			//  (n.b. currently only non-files are ever ejected like this.)
			prefilterBucket.AddRecord(fmeta, nil)
			continue
		}

		// Place the file.
		switch fmeta.Type {
		case fs.Type_File:
			r, err := f.Open()
			if err != nil {
				return api.WareID{}, api.WareID{}, Errorf(rio.ErrInoperablePath, "error while unpacking: %s", err)
			}
			reader := &util.HashingReader{R: r, Hasher: sha512.New384()}
			if err = fsOp.PlaceFile(afs, filteredFmeta, reader, false); err != nil {
				return api.WareID{}, api.WareID{}, Errorf(rio.ErrInoperablePath, "error while unpacking: %s", err)
			}
			prefilterBucket.AddRecord(fmeta, reader.Hasher.Sum(nil))
			filteredBucket.AddRecord(filteredFmeta, reader.Hasher.Sum(nil))
			continue
		case fs.Type_Symlink:
			buf := new(bytes.Buffer)
			r, err := f.Open()
			if err != nil {
				return api.WareID{}, api.WareID{}, Errorf(rio.ErrInoperablePath, "error while unpacking: %s", err)
			}
			buf.ReadFrom(r)
			fmeta.Linkname = buf.String()
			filteredFmeta.Linkname = fmeta.Linkname
		case fs.Type_Dir:
			dirs[fmeta.Name] = struct{}{}
		default:
		}
		if err := fsOp.PlaceFile(afs, filteredFmeta, nil, false); err != nil {
			return api.WareID{}, api.WareID{}, Errorf(rio.ErrInoperablePath, "error while unpacking: %s", err)
		}
		prefilterBucket.AddRecord(fmeta, nil)
		filteredBucket.AddRecord(filteredFmeta, nil)
	}

	// Cleanup dir times with a post-order traversal over the bucket.
	//  Files and dirs placed inside dirs cause the parent's mtime to update, so we have to re-pave them.
	if err := treewalk.Walk(filteredBucket.Iterator(), nil, func(node treewalk.Node) error {
		record := node.(fshash.RecordIterator).Record()
		if record.Metadata.Type != fs.Type_Dir {
			return nil
		}
		return afs.SetTimesNano(record.Metadata.Name, record.Metadata.Mtime, fs.DefaultTime)
	}); err != nil {
		return api.WareID{}, api.WareID{}, Errorf(rio.ErrInoperablePath, "error while unpacking: %s", err)
	}

	// Hash the thing!
	prefilterHash := misc.Base58Encode(fshash.HashBucket(prefilterBucket, sha512.New384))
	filteredHash := misc.Base58Encode(fshash.HashBucket(filteredBucket, sha512.New384))
	if !filt.Altering() {
		// Paranoia check for new feature.
		//  When paranoia reduced, replace with skipping the double computation.
		if prefilterHash != filteredHash {
			panic(fmt.Errorf("prefilterHash %q != filteredHash %q", prefilterHash, filteredHash))
		}
	}

	return api.WareID{"zip", prefilterHash}, api.WareID{"zip", filteredHash}, nil
}
