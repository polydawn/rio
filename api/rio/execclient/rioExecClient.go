package rioexecclient

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/polydawn/refmt"
	"github.com/polydawn/refmt/json"

	. "go.polydawn.net/rio/lib/errcat"
	"go.polydawn.net/timeless-api"
	"go.polydawn.net/timeless-api/rio"
)

var (
	_ rio.UnpackFunc = UnpackFunc
	_ rio.PackFunc   = PackFunc
	_ rio.MirrorFunc = MirrorFunc
)

func UnpackFunc(
	ctx context.Context,
	wareID api.WareID,
	path string,
	filters api.FilesetFilters,
	warehouses []api.WarehouseAddr,
	monitor rio.Monitor,
) (gotWareID api.WareID, err error) {
	if monitor.Chan != nil {
		defer close(monitor.Chan)
	}

	// Marshal args.
	args, err := UnpackArgs(wareID, path, filters, warehouses, monitor)
	if err != nil {
		return api.WareID{}, err
	}

	// Spawn process.
	cmd := exec.Command("rio", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return api.WareID{}, Errorf(rio.ErrRPCBreakdown, "fork rio: failed to start: %s", err)
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	if err = cmd.Start(); err != nil {
		return api.WareID{}, Errorf(rio.ErrRPCBreakdown, "fork rio: failed to start: %s", err)
	}

	// Set up reaction to ctx.done: send a sig to the child proc.
	//  (No, you couldn't set this up without a goroutine -- you can't select with the IO we're about to do;
	//  and No, you couldn't do it until after cmd.Start -- the Process handle doesn't exist until then.)
	go func() {
		<-ctx.Done()
		cmd.Process.Signal(os.Interrupt)
		time.Sleep(100 * time.Millisecond)
		cmd.Process.Signal(os.Kill)
	}()

	// Consume stdout, converting it to Monitor.Chan sends.
	//  (We're relying on the child proc getting signal'd to close the stdout pipe
	//  and in turn release us here in case of ctx.done.)
	unmarshaller := refmt.NewUnmarshallerAtlased(json.DecodeOptions{}, stdout, rio.Atlas)
	var msgSlot rio.Event
	for {
		// Peel off a message.
		if err := unmarshaller.Unmarshal(&msgSlot); err != nil {
			return api.WareID{}, Errorf(rio.ErrRPCBreakdown, "fork rio: API parse error: %s", err)
		}

		// If it's the final "result" message, prepare to return.
		if msgSlot.Result != nil {
			gotWareID = msgSlot.Result.WareID
			err = msgSlot.Result.Error
			break
		}
		// For all other messages: forward to the monitor channel (if it exists!)
		if monitor.Chan != nil {
			select {
			case <-ctx.Done():
				break
			case monitor.Chan <- msgSlot:
				// continue
			}
		}
	}

	// Wait for process complete.
	//  We don't actually have much use for the exit code,
	//  because we already got the serialized form of error.
	if err := cmd.Wait(); err != nil {
		return api.WareID{}, Errorf(rio.ErrRPCBreakdown, "fork rio: wait error: %s (stderr: %q)", err, stderrBuf.String())
	}
	return
}

func PackFunc(
	ctx context.Context,
	path string,
	filters api.FilesetFilters,
	warehouse api.WarehouseAddr,
	monitor rio.Monitor,
) (api.WareID, error) {
	// TODO all of it
	return api.WareID{}, nil
}

func MirrorFunc(
	ctx context.Context,
	wareID api.WareID,
	target api.WarehouseAddr,
	sources []api.WarehouseAddr,
	monitor rio.Monitor,
) (api.WareID, error) {
	// TODO all of it
	return api.WareID{}, nil
}
