package tartrans

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"go.polydawn.net/rio"
	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/fsOp"
	"go.polydawn.net/rio/transmat/mixins/fshash"
)

func Extract(
	ctx context.Context,
	destBasePath fs.AbsolutePath,
	filters rio.Filters,
	tr *tar.Reader,
) rio.Error {
	// Allocate bucket for keeping each metadata entry and content hash;
	// the full tree hash will be computed from this at the end.
	bucket := fshash.MemoryBucket{}

	// Construct filesystem wrapper to use for all our ops.
	afs := osfs.New(destBasePath)

	// Iterate over each tar entry, mutating filesystem as we go.
	for {
		fmeta := fs.Metadata{}
		thdr, err := tr.Next()

		// Check for done.
		if err == io.EOF {
			return nil // sucess!  end of archive.
		}
		if err != nil {
			return &rio.ErrWareCorrupt{
				Msg: fmt.Sprintf("corrupt tar: %s", err),
			}
		}
		if ctx.Err() != nil {
			return rio.Cancelled{}
		}

		// Reshuffle metainfo to our default format.
		if err := TarHdrToMetadata(thdr, &fmeta); err != nil {
			return err
		}
		if strings.HasPrefix(fmeta.Name.String(), "..") {
			return &rio.ErrWareCorrupt{
				Msg: "corrupt tar: paths that use '../' to leave the base dir are invalid",
			}
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
	return nil
}

// Mutate tar.Header fields to match the given fmeta.
func MetadataToTarHdr(fmeta *fs.Metadata, hdr *tar.Header) {

}

// Mutate fs.Metadata fields to match the given tar header.
// Does not check for names that go above '.'; caller may want to do that.
func TarHdrToMetadata(hdr *tar.Header, fmeta *fs.Metadata) *rio.ErrWareCorrupt {
	fmeta.Name = fs.MustRelPath(hdr.Name) // FIXME should not use the 'must' path
	fmeta.Type = tarTypeToFsType(hdr.Typeflag)
	if fmeta.Type == fs.Type_Invalid {
		return &rio.ErrWareCorrupt{
			Msg: fmt.Sprintf("corrupt tar: %q is not a known file type", hdr.Typeflag),
		}
	}
	fmeta.Perms = fs.Perms(hdr.Mode & 07777)
	fmeta.Uid = uint32(hdr.Uid)
	fmeta.Gid = uint32(hdr.Gid)
	fmeta.Size = hdr.Size
	fmeta.Linkname = hdr.Linkname
	fmeta.Devmajor = hdr.Devmajor
	fmeta.Devminor = hdr.Devminor
	fmeta.Mtime = hdr.ModTime
	fmeta.Xattrs = hdr.Xattrs
	return nil
}

func tarTypeToFsType(tarType byte) fs.Type {
	switch tarType {
	case tar.TypeReg, tar.TypeRegA:
		return fs.Type_File
	case tar.TypeLink:
		return fs.Type_Hardlink
	case tar.TypeSymlink:
		return fs.Type_Symlink
	case tar.TypeChar:
		return fs.Type_CharDevice
	case tar.TypeBlock:
		return fs.Type_Device
	case tar.TypeDir:
		return fs.Type_Dir
	case tar.TypeFifo:
		return fs.Type_NamedPipe
	// Notice that tar does not have a type for socket files
	default:
		return fs.Type_Invalid
	}
}
