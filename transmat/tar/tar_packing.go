package tartrans

import (
	"archive/tar"
	"fmt"
	"io"

	"go.polydawn.net/rio"
	"go.polydawn.net/rio/fs"
)

func Extract(
	tr *tar.Reader,
	destBasePath fs.AbsolutePath,
) rio.Error {
	// Iterate over each tar entry, mutating filesystem as we go.
	for {
		thdr, err := tr.Next()

		// Check for done.
		if err == io.EOF {
			return nil // sucess!  end of archive.
		}
		if err != nil {
			return rio.ErrWareCorrupt{
				Msg: fmt.Sprintf("corrupt tar: %s", err),
			}
		}

		//
		_ = thdr
	}
	return nil
}
