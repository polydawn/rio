/*
	Helper functions for emitting structured logs to the rio.Monitor.

	These functions encompass most common lifecycle events in a transmat,
	and using them A) saves typing and B) keeps the common stuff formatted
	in a common way between transmats.
	Transmats can of course also write their own log events raw; it is freetext.
*/
package log

import (
	"fmt"
	"time"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
)

func CacheHasIt(mon rio.Monitor, ware api.WareID) {
	if mon.Chan == nil {
		return
	}
	mon.Chan <- rio.Event{
		Log: &rio.Event_Log{
			Time:  time.Now(),
			Level: rio.LogInfo,
			Msg:   fmt.Sprintf("cache already has ware %q", ware),
			Detail: [][2]string{
				{"wareID", ware.String()},
			},
		},
	}
}

// Log path for a 'rio.ErrWarehouseUnavailable'; mode is "read" or "write".
func WarehouseUnavailable(mon rio.Monitor, err error, wh api.WarehouseAddr, ware api.WareID, mode string) {
	if mon.Chan == nil {
		return
	}
	mon.Chan <- rio.Event{
		Log: &rio.Event_Log{
			Time:  time.Now(),
			Level: rio.LogWarn,
			Msg:   fmt.Sprintf("%s while dialing warehouse %q for %s: %s", rio.ErrWarehouseUnavailable, wh, mode, err),
			Detail: [][2]string{
				{"warehouse", string(wh)},
				{"wareID", ware.String()},
				{"error", err.Error()},
			},
		},
	}
}

// Log path for a 'rio.ErrWareNotFound'.
func WareNotFound(mon rio.Monitor, err error, wh api.WarehouseAddr, ware api.WareID) {
	if mon.Chan == nil {
		return
	}
	mon.Chan <- rio.Event{
		Log: &rio.Event_Log{
			Time:  time.Now(),
			Level: rio.LogInfo,
			Msg:   fmt.Sprintf("%s from warehouse %q for ware %q", rio.ErrWareNotFound, wh, ware),
			Detail: [][2]string{
				{"warehouse", string(wh)},
				{"wareID", ware.String()},
				{"error", err.Error()},
			},
		},
	}
}
