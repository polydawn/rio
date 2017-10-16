package tartrans

import (
	"context"
	"fmt"
	"io"

	. "github.com/polydawn/go-errcat"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/go-timeless-api/util"
	"go.polydawn.net/rio/fs/nilfs"
)

var (
	_ rio.MirrorFunc = Mirror
)

func Mirror(
	ctx context.Context, // Long-running call.  Cancellable.
	wareID api.WareID, // What wareID to mirror.
	target api.WarehouseAddr, // Warehouse to ensure the ware is mirrored into.
	sources []api.WarehouseAddr, // Warehouses we can try to fetch from.
	mon rio.Monitor, // Optionally: callbacks for progress monitoring.
) (_ api.WareID, err error) {
	defer RequireErrorHasCategory(&err, rio.ErrorCategory(""))
	if mon.Chan != nil {
		defer close(mon.Chan)
	}

	// Pick a warehouse and get a reader.
	reader, err := PickReader(wareID, sources, false, mon)
	if err != nil {
		return api.WareID{}, err
	}
	defer reader.Close()

	// Connect to warehouse, and get write controller opened.
	//  During mirroring, unlike unpacking, we actually *do* know the hash
	//  of what we'll be uploading... but there's nothing dramatically better
	//  we can do with that knowledge.
	wc, err := OpenWriteController(target, wareID.Type, mon)
	if err != nil {
		return api.WareID{}, err
	}
	defer wc.Close()

	// Prepare to scan this as we process.
	//  It would be unfortunate to accidentally foist corrupted or
	//  wrongly identified content onto a mirror.
	reader = flippingReader{reader, wc}
	afs := nilFS.New()

	// "unpack", scanningly.  This drives the copy.
	filt, _ := apiutil.ProcessFilters(api.Filter_NoMutation, apiutil.FilterPurposeUnpack)
	gotWare, err := unpackTar(ctx, afs, filt, reader)
	if err != nil {
		return gotWare, err
	}

	// Check for hash mismatch before returning, because that IS an error,
	//  but also return the hash we got either way.
	if gotWare != wareID {
		return gotWare, ErrorDetailed(
			rio.ErrWareHashMismatch,
			fmt.Sprintf("hash mismatch: expected %q, got %q", wareID, gotWare),
			map[string]string{
				"expected": wareID.String(),
				"actual":   gotWare.String(),
			},
		)
	}
	return gotWare, nil
}

// Proxy read calls, also copying each buffer into another write.
type flippingReader struct {
	read io.ReadCloser
	dup  io.Writer
}

func (fr flippingReader) Read(b []byte) (int, error) {
	n, err := fr.read.Read(b)
	if err == nil || err == io.EOF {
		n2, err2 := fr.dup.Write(b)
		if n2 < n {
			return n, io.ErrShortWrite
		}
		if err2 != nil {
			return n, err2
		}
	}
	return n, err
}

func (fr flippingReader) Close() error {
	return fr.read.Close()
}
