package tartrans

import (
	"io"
	"net/url"

	. "github.com/polydawn/go-errcat"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/rio/transmat/mixins/log"
	"go.polydawn.net/rio/warehouse"
	"go.polydawn.net/rio/warehouse/impl/kvfs"
	"go.polydawn.net/rio/warehouse/impl/kvhttp"
)

// The shared bits of warehouseAddr parse and dial code.

// Pick a warehouse.
//  With K/V warehouses, this takes the form of "pick the first one that answers".
func PickReader(
	wareID api.WareID,
	warehouses []api.WarehouseAddr,
	requireMono bool,
	mon rio.Monitor,
) (_ io.ReadCloser, err error) {
	defer RequireErrorHasCategory(&err, rio.ErrorCategory(""))

	var reader io.ReadCloser
	var anyWarehouses bool // for clarity in final error messages
	for _, addr := range warehouses {
		// REVIEW ... Do I really have to parse this again?  is this sanely encapsulated?
		u, err := url.Parse(string(addr))
		if err != nil {
			return nil, Errorf(rio.ErrUsage, "failed to parse URI: %s", err)
		}
		var whCtrl warehouse.BlobstoreController
		switch u.Scheme {
		case "ca+file":
			if requireMono {
				return nil, Errorf(rio.ErrUsage, "this operation doesn't support %q scheme (a single-ware warehouse is required, not CA-mode)", u.Scheme)
			}
			fallthrough
		case "file":
			whCtrl, err = kvfs.NewController(addr)
		case "ca+http", "ca+https":
			if requireMono {
				return nil, Errorf(rio.ErrUsage, "this operation doesn't support %q scheme (a single-ware warehouse is required, not CA-mode)", u.Scheme)
			}
			fallthrough
		case "http", "https":
			whCtrl, err = kvhttp.NewController(addr)
		default:
			return nil, Errorf(rio.ErrUsage, "this operation doesn't support %q scheme (valid options are 'file', 'ca+file', 'http', 'ca+http', 'https', or 'ca+https')", u.Scheme)
		}
		switch Category(err) {
		case nil:
			anyWarehouses = true
			// pass
		case rio.ErrWarehouseUnavailable:
			if requireMono {
				return nil, err
			}
			log.WarehouseUnavailable(mon, err, addr, wareID, "read")
			continue // okay!  skip to the next one.
		default:
			return nil, err
		}
		reader, err = whCtrl.OpenReader(wareID)
		switch Category(err) {
		case nil:
			// pass
		case rio.ErrWareNotFound:
			log.WareNotFound(mon, err, addr, wareID)
			continue // okay!  skip to the next one.
		default:
			return nil, err
		}
	}
	if !anyWarehouses {
		return nil, Errorf(rio.ErrWarehouseUnavailable, "no warehouses were available!")
	}
	if reader == nil {
		return nil, Errorf(rio.ErrWareNotFound, "none of the available warehouses have ware %q!", wareID)
	}
	return reader, nil
}
