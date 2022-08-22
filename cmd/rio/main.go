package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync"

	"github.com/polydawn/refmt"
	"github.com/polydawn/refmt/json"
	. "github.com/warpfork/go-errcat"
	"gopkg.in/alecthomas/kingpin.v2"

	api "github.com/polydawn/go-timeless-api"
	"github.com/polydawn/go-timeless-api/rio"
	"github.com/polydawn/rio/fs"
	"github.com/polydawn/rio/fs/osfs"
	"github.com/polydawn/rio/fsOp"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go CancelOnInterrupt(cancel)
	exitCode := Main(ctx, os.Args, os.Stdin, os.Stdout, os.Stderr)
	os.Exit(exitCode)
}

// Blocks until a sigint is received, then calls cancel.
func CancelOnInterrupt(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	cancel()
}

// Holder type which makes it easier for us to inspect
//  the args parser result in test code before running logic.
type behavior struct {
	parsedArgs interface{}
	action     func() error
}

type format string

const (
	format_Dumb = "dumb"
	format_Json = "json"
)

func Main(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	bhv := Parse(ctx, args, stdin, stdout, stderr)
	err := bhv.action()
	return rio.ExitCodeForError(err)
}

func Parse(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) behavior {
	// CLI boilerplate.
	app := kingpin.New("rio", "Repeatable I/O.")
	app.HelpFlag.Short('h')
	app.UsageWriter(stderr)
	app.ErrorWriter(stderr)
	app.Terminate(func(int) {})

	// Output control helper.
	//  Declared early because we reference it in action thunks;
	//  however its format field may not end up set until much lower in the file.
	oc := &outputController{"", stdout, stderr, nil, sync.WaitGroup{}}

	// Args struct defs and flag declarations.
	bhvs := map[string]*behavior{}
	baseArgs := struct {
		Format string
	}{}
	app.Flag("format", "Output api format").
		Default(format_Dumb).
		EnumVar(&baseArgs.Format,
			format_Dumb, format_Json)
	{
		cmd := app.Command("pack", "Pack a Fileset into a Ware.")
		args := struct {
			PackType                string // Pack type
			Path                    string // Pack target path, abs or rel
			Filter                  string // Filters for pack
			TargetWarehouseLocation string // Warehouse address to push to
		}{}
		cmd.Arg("pack", "Pack type").
			Required().
			StringVar(&args.PackType)
		cmd.Arg("path", "Target path").
			Required().
			StringVar(&args.Path)
		cmd.Flag("target", "Warehouse in which to place the ware").
			StringVar(&args.TargetWarehouseLocation)
		cmd.Flag("filters", "Configure filters for file properties, such as mtime, uid, gid, etc.  By default many of these attribute will be flattened.").
			StringVar(&args.Filter)
		bhvs[cmd.FullCommand()] = &behavior{&args, func() (err error) {
			defer RequireErrorHasCategory(&err, rio.ErrorCategory(""))

			packFunc, err := demuxPackTool(args.PackType)
			if err != nil {
				return err
			}
			path, err := filepath.Abs(args.Path)
			if err != nil {
				return Recategorize(rio.ErrUsage, err)
			}
			filt, err := api.ParseFilesetPackFilter(args.Filter)
			if err != nil {
				return Recategorize(rio.ErrUsage, err)
			}
			filt = filt.Apply(api.FilesetPackFilter_Conservative)
			resultWareID, err := packFunc(
				ctx,
				api.PackType(args.PackType),
				path,
				filt,
				api.WarehouseLocation(args.TargetWarehouseLocation),
				oc.WireMonitor(ctx, rio.Monitor{}),
			)
			if err != nil {
				return err
			}
			oc.EmitResult(resultWareID, nil)
			return nil
		}}
	}
	{
		cmd := app.Command("unpack", "Unpack a Ware into a Fileset on your local filesystem.")
		args := struct {
			WareID                   string   // Ware id string "<kind>:<hash>"
			Path                     string   // Unpack target path, may be abs or rel
			Filter                   string   // Filters for unpack
			PlacementMode            string   // Placement mode enum
			SourcesWarehouseLocation []string // Warehouse address to fetch from
		}{}
		cmd.Arg("ware", "Ware ID").
			Required().
			StringVar(&args.WareID)
		cmd.Arg("path", "Target path").
			Required().
			StringVar(&args.Path)
		cmd.Flag("placer", "Placement mode to use [copy, direct, mount, none]").
			EnumVar(&args.PlacementMode,
				string(rio.Placement_Copy), string(rio.Placement_Direct), string(rio.Placement_Mount), string(rio.Placement_None))
		cmd.Flag("source", "Warehouses from which to fetch the ware").
			StringsVar(&args.SourcesWarehouseLocation)
		cmd.Flag("filters", "Configure filters for file properties, such as mtime, uid, gid, etc.  By default all of these will be kept, except any use of setuid, setgid, and device modes will be rejected.").
			StringVar(&args.Filter)
		bhvs[cmd.FullCommand()] = &behavior{&args, func() (err error) {
			defer RequireErrorHasCategory(&err, rio.ErrorCategory(""))

			wareID, err := api.ParseWareID(args.WareID)
			if err != nil {
				return err
			}
			unpackFunc, err := demuxUnpackTool(string(wareID.Type))
			if err != nil {
				return err
			}
			path, err := filepath.Abs(args.Path)
			if err != nil {
				return Recategorize(rio.ErrInoperablePath, err)
			}
			filt, err := api.ParseFilesetUnpackFilter(args.Filter)
			if err != nil {
				return Recategorize(rio.ErrUsage, err)
			}
			filt = filt.Apply(api.FilesetUnpackFilter_LowPriv)
			err = fsOp.RemoveDirContent(osfs.New(fs.MustAbsolutePath(path)), fs.RelPath{})
			if err != nil {
				return Recategorize(rio.ErrInoperablePath, err)
			}
			resultWareID, err := unpackFunc(
				ctx,
				wareID,
				path,
				filt,
				rio.PlacementMode(args.PlacementMode),
				convertWarehouseSlice(args.SourcesWarehouseLocation),
				oc.WireMonitor(ctx, rio.Monitor{}),
			)
			if err != nil {
				return err
			}
			oc.EmitResult(resultWareID, nil)
			return nil
		}}
	}
	{
		cmd := app.Command("scan", "Scan some existing data stream see if it's a known packed format, and compute its WareID if so.  (Mostly used for importing tars from the interweb.)")
		args := struct {
			PackType                string // Pack type
			Filter                  string // Filters as if unpacking
			SourceWarehouseLocation string // Warehouse address of data to scan
		}{}
		cmd.Arg("pack", "Pack type").
			Required().
			StringVar(&args.PackType)
		cmd.Flag("source", "Address to of the data to scan.").
			StringVar(&args.SourceWarehouseLocation)
		cmd.Flag("filters", "Configure filters for file properties, such as mtime, uid, gid, etc.").
			StringVar(&args.Filter)
		bhvs[cmd.FullCommand()] = &behavior{&args, func() (err error) {
			defer RequireErrorHasCategory(&err, rio.ErrorCategory(""))

			scanFunc, err := demuxScanTool(string(args.PackType))
			if err != nil {
				return err
			}
			filt, err := api.ParseFilesetUnpackFilter(args.Filter)
			if err != nil {
				return Recategorize(rio.ErrUsage, err)
			}
			filt = filt.Apply(api.FilesetUnpackFilter_Conservative)
			resultWareID, err := scanFunc(
				ctx,
				api.PackType(args.PackType),
				filt,
				rio.Placement_Direct,
				api.WarehouseLocation(args.SourceWarehouseLocation),
				oc.WireMonitor(ctx, rio.Monitor{}),
			)
			if err != nil {
				return err
			}
			oc.EmitResult(resultWareID, nil)
			return nil
		}}
	}
	{
		cmd := app.Command("mirror", "Store already-packed wares in one warehouse, copying from other warehouses.")
		args := struct {
			WareID                   string   // WareID to mirror
			TargetWarehouseLocation  string   // Warehouse to mirror into
			SourceWarehouseLocations []string // Warehouses we can fetch from
		}{}
		cmd.Arg("ware", "Ware ID").
			Required().
			StringVar(&args.WareID)
		cmd.Flag("target", "Warehouse in which to place the ware").
			StringVar(&args.TargetWarehouseLocation)
		cmd.Flag("source", "Warehouses from which to fetch the ware").
			StringsVar(&args.SourceWarehouseLocations)
		bhvs[cmd.FullCommand()] = &behavior{&args, func() (err error) {
			defer RequireErrorHasCategory(&err, rio.ErrorCategory(""))

			wareID, err := api.ParseWareID(args.WareID)
			if err != nil {
				return err
			}
			mirrorFunc, err := demuxMirrorTool(string(wareID.Type))
			if err != nil {
				return err
			}
			resultWareID, err := mirrorFunc(
				ctx,
				wareID,
				api.WarehouseLocation(args.TargetWarehouseLocation),
				convertWarehouseSlice(args.SourceWarehouseLocations),
				oc.WireMonitor(ctx, rio.Monitor{}),
			)
			if err != nil {
				return err
			}
			oc.EmitResult(resultWareID, nil)
			return nil
		}}
	}
	// Okay now let's be clear: actually all of these behaviors should, end of day,
	//  actually send their errors through our output control.
	//  We still also return it, both so you can write tests around this
	//  method as a whole, and so the top of the program can choose an exit code!
	for _, bhv := range bhvs {
		_action := bhv.action
		bhv.action = func() error {
			err := _action()
			if err != nil {
				oc.EmitResult(api.WareID{}, err)
			}
			return err
		}
	}

	// Parse!
	parsedCmdStr, err := app.Parse(args[1:])
	oc.format = format(baseArgs.Format)
	if err != nil {
		return behavior{
			parsedArgs: err,
			action: func() error {
				err := Errorf(rio.ErrUsage, "error parsing args: %s", err)
				oc.EmitResult(api.WareID{}, err)
				return err
			},
		}
	}
	// Return behavior named by the command and subcommand strings.
	if bhv, ok := bhvs[parsedCmdStr]; ok {
		return *bhv
	}
	panic("unreachable, cli parser must error on unknown commands")
}

type outputController struct {
	format         format
	stdout, stderr io.Writer
	monChan        chan rio.Event // set up when calling WireMonitor
	monWg          sync.WaitGroup
}

func (oc *outputController) EmitResult(wareID api.WareID, err error) {
	oc.monWg.Wait()
	var evt rio.Event = rio.Event_Result{wareID, rio.ToError(err)}
	switch oc.format {
	case "", format_Dumb:
		if err != nil {
			fmt.Fprintln(oc.stderr, err)
		} else {
			fmt.Fprintln(oc.stdout, wareID)
		}
	case format_Json:
		if err != nil {
			fmt.Fprintln(oc.stderr, err)
		}
		marshaller := refmt.NewMarshallerAtlased(json.EncodeOptions{}, oc.stdout, rio.Atlas)
		err := marshaller.Marshal(&evt)
		if err != nil {
			panic(err)
		}
	default:
		panic(fmt.Errorf("rio: invalid format %s", oc.format))
	}
}

func (oc *outputController) WireMonitor(ctx context.Context, m rio.Monitor) rio.Monitor {
	oc.monChan = make(chan rio.Event)
	oc.monWg.Add(1)
	m.Chan = oc.monChan
	switch oc.format {
	case "", format_Dumb:
		go func() {
			defer oc.monWg.Done()
			for {
				select {
				case evt, ok := <-oc.monChan:
					if !ok {
						return
					}
					switch evt := evt.(type) {
					case rio.Event_Log:
						fmt.Fprintf(oc.stderr, "log: lvl=%s msg=%s\n", evt.Level, evt.Msg)
					case rio.Event_Progress:
						// pass... for now
					case rio.Event_Result:
						// pass
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	case format_Json:
		marshaller := refmt.NewMarshallerAtlased(json.EncodeOptions{}, oc.stdout, rio.Atlas)
		go func() {
			defer oc.monWg.Done()
			for {
				select {
				case evt, ok := <-oc.monChan:
					if !ok {
						return
					}
					err := marshaller.Marshal(&evt)
					oc.stdout.Write([]byte{'\n'})
					if err != nil {
						panic(err)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	default:
		panic(fmt.Errorf("rio: invalid format %s", oc.format))
	}
	return m
}

func convertWarehouseSlice(slice []string) []api.WarehouseLocation {
	result := make([]api.WarehouseLocation, len(slice))
	for idx, item := range slice {
		result[idx] = api.WarehouseLocation(item)
	}
	return result
}
