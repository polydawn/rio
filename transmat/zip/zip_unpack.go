package ziptrans

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha512"
	"fmt"
	"io"
	"strings"

	"github.com/polydawn/refmt/misc"
	. "github.com/warpfork/go-errcat"

	api "go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fsOp"
	"go.polydawn.net/rio/lib/treewalk"
	"go.polydawn.net/rio/transmat/mixins/buffer"
	"go.polydawn.net/rio/transmat/mixins/filters"
	"go.polydawn.net/rio/transmat/mixins/fshash"
	"go.polydawn.net/rio/transmat/mixins/log"
	"go.polydawn.net/rio/transmat/util"
)

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

	readerAt, closer, err := buffer.SectionReader(ctx, archiveWareID, reader, mon)
	if err != nil {
		return api.WareID{}, api.WareID{}, err
	}
	defer closer.Close()

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
	for _, zf := range zr.File {
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
		err := ZipHdrToMetadata(&zf.FileHeader, &fmeta)
		if err != nil {
			return api.WareID{}, api.WareID{}, err
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
			r, err := zf.Open()
			if err != nil {
				return api.WareID{}, api.WareID{}, Errorf(rio.ErrWareCorrupt, "error while unpacking: %s", err)
			}
			reader := &util.HashingReader{R: r, Hasher: sha512.New384()}
			if err = fsOp.PlaceFile(afs, filteredFmeta, reader, false); err != nil {
				return api.WareID{}, api.WareID{}, Errorf(rio.ErrInoperablePath, "error while unpacking: %s", err)
			}
			prefilterBucket.AddRecord(fmeta, reader.Hasher.Sum(nil))
			filteredBucket.AddRecord(filteredFmeta, reader.Hasher.Sum(nil))
		case fs.Type_Symlink:
			buf := new(bytes.Buffer)
			r, err := zf.Open()
			if err != nil {
				return api.WareID{}, api.WareID{}, Errorf(rio.ErrWareCorrupt, "error while unpacking: %s", err)
			}
			_, err = buf.ReadFrom(r)
			if err != nil {
				return api.WareID{}, api.WareID{}, Errorf(rio.ErrWareCorrupt, "error while unpacking: %s", err)
			}
			fmeta.Linkname = buf.String()
			filteredFmeta.Linkname = fmeta.Linkname
			if err := fsOp.PlaceFile(afs, filteredFmeta, nil, false); err != nil {
				return api.WareID{}, api.WareID{}, Errorf(rio.ErrInoperablePath, "error while unpacking: %s", err)
			}
			prefilterBucket.AddRecord(fmeta, nil)
			filteredBucket.AddRecord(filteredFmeta, nil)
		case fs.Type_Dir:
			dirs[fmeta.Name] = struct{}{}
			if err := fsOp.PlaceFile(afs, filteredFmeta, nil, false); err != nil {
				return api.WareID{}, api.WareID{}, Errorf(rio.ErrInoperablePath, "error while unpacking: %s", err)
			}
			prefilterBucket.AddRecord(fmeta, nil)
			filteredBucket.AddRecord(filteredFmeta, nil)
		default:
			return api.WareID{}, api.WareID{}, Errorf(rio.ErrPackInvalid, "zip pack does not support files of type %v", fmeta.Type)
		}
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
