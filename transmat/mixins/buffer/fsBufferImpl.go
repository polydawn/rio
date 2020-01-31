package buffer

import (
	"context"
	"io"
	"io/ioutil"
	"os"

	. "github.com/warpfork/go-errcat"

	api "go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
	"go.polydawn.net/rio/fsOp"
)

/*Buffer a ware to a locally seekable instance.
 *Some forms of transmat need to know the length and seek through packed data.
 *This buffer bridges the gap between that need at the `io.ReadCloser`
 *supported by `warehouse.BlobstoreController`.
 */
type buffer struct {
	base string
	fs   fs.FS
}

func NewTemporaryBuffer() *buffer {
	base := os.TempDir()
	dir, err := ioutil.TempDir(base, "rio-")
	if err != nil {
		return nil
	}

	p, err := fs.ParseAbsolutePath(dir)
	if err != nil {
		return nil
	}
	baseFS := osfs.New(p)
	return &buffer{dir, baseFS}
}

func (b *buffer) Close() error {
	return os.RemoveAll(b.base)
}

func (b *buffer) SectionReader(
	ctx context.Context,
	wareID api.WareID,
	reader io.Reader,
	monitor rio.Monitor,
) (*io.SectionReader, error) {
	if _, err := os.Stat(b.base); os.IsNotExist(err) {
		return nil, Errorf(rio.ErrInoperablePath, "error buffering %q: %s", wareID, err)
	}

	// Compute buffer location for the wareID
	fmeta := fs.Metadata{
		Name:  fs.MustRelPath(wareID.Hash),
		Type:  fs.Type_File,
		Perms: 0755,
	}

	// If not present, locally buffer.
	if found, _ := b.fs.Stat(fmeta.Name); found == nil {
		// TODO: PlaceFile should take ctx & monitor as the long / abortable action.
		if err := fsOp.PlaceFile(b.fs, fmeta, reader, true); err != nil {
			return nil, Errorf(rio.ErrLocalCacheProblem, "error buffering %q: %s", wareID, err)
		}
	}

	// Regenerate the file metadata to learn the locally buffered size.
	learnedMeta, err := b.fs.Stat(fmeta.Name)
	if err != nil {
		return nil, Errorf(rio.ErrLocalCacheProblem, "error buffering %q: %s", wareID, err)
	}

	// TODO: should temp file permissions be calculated better?
	file, err := b.fs.OpenFile(fmeta.Name, os.O_RDONLY, 0755)
	if err != nil {
		return nil, Errorf(rio.ErrLocalCacheProblem, "error buffering %q: %s", wareID, err)
	}

	return io.NewSectionReader(file, 0, learnedMeta.Size), nil
}
