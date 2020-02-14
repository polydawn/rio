package buffer

import (
	"context"
	"io"
	"io/ioutil"
	"os"

	. "github.com/warpfork/go-errcat"

	api "go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
)

type tempRemover struct {
	file string
}

func (t tempRemover) Close() error {
	return os.Remove(t.file)
}

// SectionReader buffers a ware to a locally seekable instance.
// Some forms of transmat (zip) need to know the length and seek through packed data.
// This buffer bridges the gap between that need at the `io.ReadCloser`
// supported by `warehouse.BlobstoreController`.
func SectionReader(
	ctx context.Context,
	wareID api.WareID,
	reader io.Reader,
	monitor rio.Monitor,
) (*io.SectionReader, io.Closer, error) {
	bufferFile, err := ioutil.TempFile("", "rio-*")
	if err != nil {
		return nil, nil, Errorf(rio.ErrInoperablePath, "error buffering %q: %s", wareID, err)
	}

	size, err := io.Copy(bufferFile, reader)
	if err != nil {
		_ = os.Remove(bufferFile.Name())
		return nil, nil, Errorf(rio.ErrLocalCacheProblem, "error buffering %q: %s", wareID, err)
	}

	return io.NewSectionReader(bufferFile, 0, size), &tempRemover{bufferFile.Name()}, nil
}
