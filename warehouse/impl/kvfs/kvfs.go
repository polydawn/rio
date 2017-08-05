package kvfs

import (
	"io"
	"net/url"
	"os"
	"path/filepath"

	"go.polydawn.net/rio/fs"
	. "go.polydawn.net/rio/lib/errcat"
	"go.polydawn.net/rio/lib/guid"
	"go.polydawn.net/rio/warehouse/util"
	"go.polydawn.net/timeless-api"
	"go.polydawn.net/timeless-api/rio"
)

type Controller struct {
	addr     api.WarehouseAddr // user's string retained for messages
	basePath fs.AbsolutePath
	ctntAddr bool
}

/*
	Initialize a new warehouse controller that operates on a local filesystem.

	If any errors are returned, they're of category `rio.ErrUsage`.
	Other potential errors regarding readability and writability are raised
	at the time of usage.
*/
func NewController(addr api.WarehouseAddr) (whCtrl Controller, err error) {
	// Stamp out a warehouse handle.
	//  More values will be accumulated in shortly.
	whCtrl.addr = addr

	// Verify that the addr is sensible up front, and extract features.
	//  - We parse things mostly like URLs.
	//  - We extract whether or not it's content-addressible mode here;
	//  - and extract the filesystem path, and normalize it to its absolute form.
	u, err := url.Parse(string(addr))
	if err != nil {
		return whCtrl, Errorf(rio.ErrUsage, "failed to parse URI: %s", err)
	}
	switch u.Scheme {
	case "file":
	case "file+ca":
		whCtrl.ctntAddr = true
	default:
		return whCtrl, Errorf(rio.ErrUsage, "unsupported scheme in warehouse addr: %q (valid options are 'file' or 'file+ca'", u.Scheme)
	}
	absPth, err := filepath.Abs(filepath.Join(u.Host, u.Path))
	if err != nil {
		panic(err)
	}
	whCtrl.basePath = fs.MustAbsolutePath(absPth)
	return
}

func (whCtrl Controller) OpenWriter() (wc WriteController, err error) {
	// Pick a random upload path.
	if whCtrl.ctntAddr {
		tmpName := fs.MustRelPath(".tmp.upload." + guid.New())
		wc.stagePath = whCtrl.basePath.Join(tmpName)
	} else {
		// In non-CA mode, "base" path isn't really "base"; it's the final destination.
		tmpName := fs.MustRelPath(".tmp.upload." + whCtrl.basePath.Last() + "." + guid.New())
		wc.stagePath = whCtrl.basePath.Dir().Join(tmpName)
	}
	// Open file the file for write.
	file, err := os.OpenFile(wc.stagePath.String(), os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if err != nil {
		return wc, Errorf(rio.ErrWarehouseUnwritable, "failed to reserve temp space in warehouse: %s", err)
	}
	wc.stream = file
	// Return the controller -- which has methods to either commit+close, or cancel+close.
	return
}

type WriteController struct {
	stream    io.WriteCloser  // Write to this.
	whCtrl    Controller      // Needed for the final move-into-place.
	stagePath fs.AbsolutePath // Needed for the final move-into-place.
}

/*
	Cancel the current write.  Close the stream, and remove any temporary files.
*/
func (wc *WriteController) Close() error {
	wc.stream.Close()
	return os.Remove(wc.stagePath.String())
}

/*
	Commit the current data as the given hash.
	Caller must be an adult and specify the hash truthfully.
	Closes the writer and invalidates any future use.
*/
func (wc *WriteController) Commit(wareID api.WareID) error {
	// Close the file.
	if err := wc.stream.Close(); err != nil {
		return Errorf(rio.ErrWarehouseUnwritable, "failed to commit to file: %s", err)
	}
	// Compute final path.
	// Make parent dirs if necessary in content-addr mode.
	finalPath := wc.whCtrl.basePath
	if wc.whCtrl.ctntAddr {
		chunkA, chunkB, _ := util.ChunkifyHash(wareID)
		finalPath = finalPath.Join(fs.MustRelPath(chunkA))
		if err := os.Mkdir(finalPath.String(), 0755); err != nil {
			return Errorf(rio.ErrWarehouseUnwritable, "failed to commit to file: %s", err)
		}
		finalPath = finalPath.Join(fs.MustRelPath(chunkB))
		if err := os.Mkdir(finalPath.String(), 0755); err != nil {
			return Errorf(rio.ErrWarehouseUnwritable, "failed to commit to file: %s", err)
		}
		finalPath = finalPath.Join(fs.MustRelPath(wareID.Hash))
	}
	// Move into place.
	if err := os.Rename(wc.stagePath.String(), finalPath.String()); err != nil {
		return Errorf(rio.ErrWarehouseUnwritable, "failed to commit to file: %s", err)
	}
	return nil
}
