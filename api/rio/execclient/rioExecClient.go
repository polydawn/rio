package rioexecclient

import (
	"context"
	"os"
	"os/exec"
	"time"

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
) (api.WareID, error) {
	// Marshal args.
	args, err := UnpackArgs(wareID, path, filters, warehouses, monitor)
	if err != nil {
		return api.WareID{}, err
	}

	// Spawn process.
	cmd := exec.Command("rio", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err) // FIXME needs rpc-breakdown error category
	}
	if err = cmd.Start(); err != nil {
		panic(err) // FIXME needs rpc-breakdown error category
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
	for {
		_ = stdout
		// TODO
	}

	// Wait for process complete.
	//  We don't actually have much use for the exit code,
	//  because we already got the serialized form of error.
	if err := cmd.Wait(); err != nil {
		panic(err) // FIXME needs rpc-breakdown error category
	}
	return api.WareID{}, nil
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
