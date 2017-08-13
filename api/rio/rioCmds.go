/*
	Interfaces of rio commands.

	The heuristic for the rio callable library API is that essentially
	all information must be racked up in the call already: the assumption
	is that the side doing the real work might not share a filesystem with
	you (well, in rio's case, it probably does!  but it might be a subset,
	translated through chroots and bind mounts), doesn't share env vars, etc.
	So, general rule of thumb: the caller is going to have already handled
	all config loading and parsing, and those objects are params in this funcs.
*/
package rio

import (
	"context"

	"go.polydawn.net/timeless-api"
)

type UnpackFunc func(
	ctx context.Context, // Long-running call.  Cancellable.
	wareID api.WareID, // What wareID to fetch for unpacking.
	path string, // Where to unpack the fileset (absolute path).
	filters api.FilesetFilters, // Optionally: filters we should apply while unpacking.
	warehouses []api.WarehouseAddr, // Warehouses we can try to fetch from.
	monitor Monitor, // Optionally: callbacks for progress monitoring.
) (api.WareID, error)

type PackFunc func(
	ctx context.Context, // Long-running call.  Cancellable.
	path string, // The fileset to scan and pack (absolute path).
	filters api.FilesetFilters, // Optionally: filters we should apply while unpacking.
	warehouse api.WarehouseAddr, // Warehouse to save into (or blank to just scan).
	monitor Monitor, // Optionally: callbacks for progress monitoring.
) (api.WareID, error)

type MirrorFunc func(
	ctx context.Context, // Long-running call.  Cancellable.
	wareID api.WareID, // What wareID to mirror.
	target api.WarehouseAddr, // Warehouse to ensure the ware is mirrored into.
	sources []api.WarehouseAddr, // Warehouses we can try to fetch from.
	monitor Monitor, // Optionally: callbacks for progress monitoring.
) (api.WareID, error)

/*
	Monitoring configuration structs, and message types used.
*/
type (
	// REVIEW ... it's rather generalizing to use the same monitor and event union
	//  for all these different functions, isn't it?

	/*
		Configuration for what intermediate progress reports a process should send,
		and slot for the channel the caller wishes them to be sent to.
	*/
	Monitor struct {
		// FUTURE: may add options for how many things we'd like to be sent to us

		// Channel to which events will be sent as the process proceeds.
		// The channel will be closed when the process is done or cancelled.
		// A nil channel will disable all intermediate progress reporting.
		Chan chan<- Event
	}

	/*
		A "union" type of all the kinds of event that may be generated in the
		course of any of the functions.

		The "Result" message is never sent to Monitor.Chan --
		its values are converted into the function returns --
		but *is* seen in the serial form on the wire.

		(This type may be replaced by an interface in the future when the refmt
		library's union message support becomes available.)
	*/
	Event struct {
		Progress *Event_Progress `refmt:"prog,omitempty"`
		Result   *Event_Result   `refmt:"result,omitempty"`
	}

	/*
		Notifications about progress updates.

		Imagine it being used to draw the following:

			Frobnozing (145/290kb): [=====>    ]  50%

		The 'totalProg' and 'totalWork' ints are expected to be a percentage;
		when they equal, a "done" state should be up next.
		A value of 'totalProg' greater than 'totalWork' is nonsensical.

		The 'phase' and 'desc' args are freetext;
		Typically, 'phase' will remain the same for many calls in a row, while
		'desc' is used to communicate a more specific contextual info
		than the 'total*' ints and like the ints may likely change on each call.
	*/
	Event_Progress struct {
		Phase, Desc          string
		TotalProg, TotalWork int
	}

	Event_Result struct {
		WareID api.WareID
		Error  error // FIXME resolution needed on how we want to serialize these... mainly because... do we really want to import errcat in our api packages?  not so much.
	}
)

type ErrorCategory string
type ExitCode int

const (
	ExitSuccess                                       = ExitCode(0)
	ExitUsage, ErrUsage                               = ExitCode(1), ErrorCategory("rio-usage-error")           // Indicates some piece of user input to a command was invalid and unrunnable.
	ExitPanic                                         = ExitCode(2)                                             // Placeholder.  We don't use this.  '2' happens when golang exits due to panic.
	ExitWarehouseUnavailable, ErrWarehouseUnavailable = ExitCode(3), ErrorCategory("rio-warehouse-unavailable") // Warehouse 404.
	ExitWarehouseUnwritable, ErrWarehouseUnwritable   = ExitCode(4), ErrorCategory("rio-warehouse-unwritable")  // Indicates a warehouse failed to accept a write operation.  The Warehouse is having a bad day.  ("unauthorized" is a different error category.)
	ExitWareNotFound, ErrWareNotFound                 = ExitCode(5), ErrorCategory("rio-ware-not-found")        // Ware 404 -- warehouse appeared online, but the requested ware isn't in it.
	ExitWareCorrupt, ErrWareCorrupt                   = ExitCode(6), ErrorCategory("rio-ware-corrupt")          // Incidates a ware retreival started, but during unpacking it was found to be malformed.
	ExitHashMismatch, ErrWareHashMismatch             = ExitCode(7), ErrorCategory("rio-hash-mismatch")         // Returned when fetching and unpacking a ware gets results in a different content hash than we requested.  (This is distinct from ErrWareCorrupt because a full fileset *was* able to be unpacked; it's just not the one we asked for.)
	ExitCancelled, ErrCancelled                       = ExitCode(8), ErrorCategory("rio-cancelled")             // The operation timed out or was cancelled
	ExitNotImplemented, ErrNotImplemented             = ExitCode(9), ErrorCategory("rio-not-implemented")       // The operation is not implemented, PRs welcome
	ExitRPCBreakdown, ErrRPCBreakdown                 = ExitCode(120), ErrorCategory("rio-rpc-breakdown")       // Raised when running a remote rio process and the control channel is lost, the process fails to start, or unrecognized messages are received.
	ExitTODO                                          = ExitCode(254)                                           // This exit code should be replaced with something more specific
)
