package tartrans

import (
	"archive/tar"
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
	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fsOp"
	"go.polydawn.net/rio/lib/treewalk"
	"go.polydawn.net/rio/transmat/mixins/filters"
	"go.polydawn.net/rio/transmat/mixins/fshash"
	"go.polydawn.net/rio/transmat/mixins/log"
	"go.polydawn.net/rio/transmat/util"
)

func unpackTar(
	ctx context.Context,
	afs fs.FS,
	filt api.FilesetUnpackFilter,
	_ api.WareID,
	reader io.Reader,
	mon rio.Monitor,
) (
	prefilterWareID api.WareID,
	actualWareID api.WareID,
	err error,
) {
	defer RequireErrorHasCategory(&err, rio.ErrorCategory(""))

	// Wrap input stream with decompression as necessary.
	//  Which kind of decompression to use can be autodetected by magic bytes.
	reader2, err := Decompress(reader)
	if err != nil {
		return api.WareID{}, api.WareID{}, Errorf(rio.ErrWareCorrupt, "corrupt tar compression: %s", err)
	}

	// Convert the raw byte reader to a tar stream.
	tr := tar.NewReader(reader2)

	// Allocate bucket for keeping each metadata entry and content hash;
	// the full tree hash will be computed from this at the end.
	// We keep one for the raw ware data as we consume it, so we can verify no fuckery;
	// we keep a second, separate one for the filtered data, which will compute a different hash.
	prefilterBucket := &fshash.MemoryBucket{}
	filteredBucket := &fshash.MemoryBucket{}

	// Also allocate a map for keeping records of which dirs we've created.
	// This is necessary for correct bookkeepping in the face of the tar format's
	// allowance for implicit parent dirs.
	dirs := map[fs.RelPath]struct{}{}

	// Iterate over each tar entry, mutating filesystem as we go.
	for {
		fmeta := fs.Metadata{}
		thdr, err := tr.Next()

		// Check for done.
		if err == io.EOF {
			break // sucess!  end of archive.
		}
		if err != nil {
			return api.WareID{}, api.WareID{}, Errorf(rio.ErrWareCorrupt, "corrupt tar: %s", err)
		}
		if ctx.Err() != nil {
			return api.WareID{}, api.WareID{}, Errorf(rio.ErrCancelled, "cancelled")
		}

		// Reshuffle metainfo to our default format.
		skipMe, haltMe := TarHdrToMetadata(thdr, &fmeta)
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
					{"tarhdr", fmt.Sprintf("%#v", thdr)},
				},
			})
			continue
		}
		if haltMe != nil {
			return api.WareID{}, api.WareID{}, haltMe
		}
		if strings.HasPrefix(fmeta.Name.String(), "..") {
			return api.WareID{}, api.WareID{}, Errorf(rio.ErrWareCorrupt, "corrupt tar: paths that use '../' to leave the base dir are invalid")
		}

		// Infer parents, if necessary.  The tar format allows implicit parent dirs.
		//
		// Note that if any of the implicitly conjured dirs is specified later, unpacking won't notice,
		// but bucket hashing iteration will (correctly) blow up for repeat entries.
		// It may well be possible to construct a tar like that, but it's already well established that
		// tars with repeated filenames are just asking for trouble and shall be rejected without
		// ceremony because they're just a ridiculous idea.
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
			reader := &util.HashingReader{tr, sha512.New384()}
			if err := fsOp.PlaceFile(afs, filteredFmeta, reader, false); err != nil {
				return api.WareID{}, api.WareID{}, Errorf(rio.ErrInoperablePath, "error while unpacking: %s", err)
			}
			prefilterBucket.AddRecord(fmeta, reader.Hasher.Sum(nil))
			filteredBucket.AddRecord(filteredFmeta, reader.Hasher.Sum(nil))
		case fs.Type_Dir:
			dirs[fmeta.Name] = struct{}{}
			fallthrough
		default:
			if err := fsOp.PlaceFile(afs, filteredFmeta, nil, false); err != nil {
				return api.WareID{}, api.WareID{}, Errorf(rio.ErrInoperablePath, "error while unpacking: %s", err)
			}
			prefilterBucket.AddRecord(fmeta, nil)
			filteredBucket.AddRecord(filteredFmeta, nil)
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

	return api.WareID{"tar", prefilterHash}, api.WareID{"tar", filteredHash}, nil
}
