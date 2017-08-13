package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/polydawn/refmt"
	"github.com/polydawn/refmt/json"
	"gopkg.in/alecthomas/kingpin.v2"

	tar "go.polydawn.net/rio/transmat/tar"
	"go.polydawn.net/timeless-api"
	"go.polydawn.net/timeless-api/rio"
)

/*
	Output serialization formats
*/
const (
	FmtJson = "json"
	FmtDumb = "dumb"
)

type baseCLI struct {
	Deadline       string        // Deadline time (RFC3339 or @UNIX)
	Format         string        // Output api format, eg. json
	ProgressEnable bool          // Emit progress notification yes/no
	ProgressRate   time.Duration // How frequently to emit progress notification
	Test           string        // Testmode
	Timeout        time.Duration // Timeout duration (exclusive with deadline) eg. "60s"
	PackCLI        struct {
		Filters             api.FilesetFilters // TODO: file filters for pack/unpack
		Path                string             // Pack target path
		TargetWarehouseAddr string             // Warehouse address to push to
	}
	UnpackCLI struct {
		Filters              api.FilesetFilters
		Path                 string   // Unpack target path
		SourcesWarehouseAddr []string // Warehouse address to push to
		WareID               string   // Ware id string "<kind>:<hash>"
	}
	MirrorCLI struct {
		SourcesWarehouseAddr []string // Warehouse addresses to fetch from
		TargetWarehouseAddr  string   // Warehouse address to push to
		WareID               string   // Ware id string "<kind>:<hash>"
	}
}

func configurePack(cli *baseCLI, appPack *kingpin.CmdClause) {
	// Non-filter flags
	appPack.Flag("path", "Target path").
		StringVar(&cli.PackCLI.Path)
	appPack.Flag("target", "Warehouse in which to place the ware").
		StringVar(&cli.PackCLI.TargetWarehouseAddr)

	// Filter flags
	appPack.Flag("uid", "Set UID filter [keep, <int>]").
		StringVar(&cli.PackCLI.Filters.Uid)
	appPack.Flag("gid", "Set GID filter [keep, <int>]").
		StringVar(&cli.PackCLI.Filters.Gid)
	appPack.Flag("mtime", "Set mtime filter [keep, <@UNIX>, <RFC3339>]. Will be set to a date if not specified.").
		StringVar(&cli.PackCLI.Filters.Mtime)
	// Sticky flag not used for pack
}

func configureUnpack(cli *baseCLI, appUnpack *kingpin.CmdClause) {
	// Non-filter flags
	appUnpack.Flag("path", "Target path").
		StringVar(&cli.UnpackCLI.Path)
	appUnpack.Flag("source", "Warehouses from which to fetch the ware").
		StringsVar(&cli.UnpackCLI.SourcesWarehouseAddr)
	appUnpack.Flag("ware", "Ware ID").
		StringVar(&cli.UnpackCLI.WareID)

	// Filter flags
	appUnpack.Flag("uid", "Set UID filter [keep, mine, <int>]").
		Default("mine").
		StringVar(&cli.UnpackCLI.Filters.Uid)
	appUnpack.Flag("gid", "Set GID filter [keep, mine, <int>]").
		Default("mine").
		StringVar(&cli.UnpackCLI.Filters.Gid)
	appUnpack.Flag("mtime", "Set mtime filter [keep, <@UNIX>, <RFC3339>]").
		Default("keep").
		StringVar(&cli.UnpackCLI.Filters.Mtime)
	appUnpack.Flag("sticky", "Keep setuid, setgid, and sticky bits").
		BoolVar(&cli.UnpackCLI.Filters.Sticky)
}

func configureMirror(cli *baseCLI, appMirror *kingpin.CmdClause) {
	appMirror.Flag("ware", "Ware ID").
		StringVar(&cli.MirrorCLI.WareID)
	appMirror.Flag("target", "Warehouse in which to place the ware").
		StringVar(&cli.MirrorCLI.TargetWarehouseAddr)
	appMirror.Flag("source", "Warehouses from which to fetch the ware").
		StringsVar(&cli.MirrorCLI.SourcesWarehouseAddr)
}

/*
	Blocks until a sigint is received, then calls cancel.
*/
func CancelOnInterrupt(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	cancel()
	close(signalChan)
}

func main() {
	ctx := context.Background()
	exitCode := Main(ctx, os.Args, os.Stdin, os.Stdout, os.Stderr)
	os.Exit(int(exitCode))
}

func Main(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) rio.ExitCode {
	exitCode := rio.ExitSuccess
	ctx, cancel := context.WithCancel(ctx)
	go CancelOnInterrupt(cancel)

	cli := baseCLI{}

	app := kingpin.New("rio", "Repeatable I/O")
	app.HelpFlag.Short('h')

	app.UsageWriter(stderr)
	app.ErrorWriter(stderr)

	app.Flag("deadline", "Deadline (RFC3339)").
		StringVar(&cli.Deadline)
	app.Flag("timeout", "Timeout for command").
		DurationVar(&cli.Timeout)
	app.Flag("format", "Output api format").
		Default(FmtDumb).
		EnumVar(&cli.Format, FmtJson, FmtDumb)
	app.Flag("progress-rate", "How frequently to emit progress notification").
		DurationVar(&cli.ProgressRate)
	app.Flag("progress", "Emit progress notification").
		BoolVar(&cli.ProgressEnable)
	app.Flag("test", "Testmodes").
		StringVar(&cli.Test)

	appPack := app.Command("pack", "pack a fileset into a ware")
	configurePack(&cli, appPack)

	appUnpack := app.Command("unpack", "fetch a ware")
	configureUnpack(&cli, appUnpack)

	appMirror := app.Command("mirror", "mirror a ware to another warehouse")
	configureUnpack(&cli, appMirror)

	cmd, err := Parse(app, args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return rio.ExitUsage
	}
	var wareID api.WareID
	// FIXME: We'll need to support more than tar eventually
	switch cmd {
	case appPack.FullCommand():
		wareID, err = executePack(ctx, cli, rio.PackFunc(tar.Pack))
		SerializeResult(cli.Format, wareID, err, stdout, stderr)
	case appUnpack.FullCommand():
		wareID, err = executeUnpack(ctx, cli, rio.UnpackFunc(tar.Unpack))
		SerializeResult(cli.Format, wareID, err, stdout, stderr)
	case appMirror.FullCommand():
		fmt.Fprintln(stderr, "not implemented")
		return rio.ExitNotImplemented
	}

	return exitCode
}

/*
	Kingpin really expects terminate to terminate.
	So we make it panic and handle it. Passing errors back up.
*/
func Parse(app *kingpin.Application, args []string) (cmd string, err error) {
	defer func() {
		r := recover()
		if e, ok := r.(error); ok {
			err = e
		}
	}()
	app.Terminate(func(status int) {
		termErr := fmt.Errorf("rio: parse error %d\n", status)
		panic(termErr)
	})
	return app.Parse(args[1:])
}

func SerializeResult(format string, wareID api.WareID, resultErr error, stdout io.Writer, stderr io.Writer) {
	ev := rio.Event{Result: &rio.Event_Result{WareID: wareID, Error: resultErr}}
	switch format {
	case FmtJson:
		marshaller := refmt.NewMarshallerAtlased(json.EncodeOptions{}, stdout, rio.Atlas)
		err := marshaller.Marshal(&ev)
		if err != nil {
			panic(err)
		}
	case FmtDumb:
		if resultErr != nil {
			fmt.Fprintln(stderr, resultErr)
		} else {
			fmt.Fprintln(stdout, wareID)
		}
	default:
		panic(fmt.Errorf("rio: invalid format %s", format))
	}
}

func convertWarehouseSlice(slice []string) []api.WarehouseAddr {
	result := make([]api.WarehouseAddr, len(slice))
	for idx, item := range slice {
		result[idx] = api.WarehouseAddr(item)
	}
	return result
}

func executeUnpack(ctx context.Context, cli baseCLI, unpackFn rio.UnpackFunc) (api.WareID, error) {
	monitor := rio.Monitor{}
	wareID, err := api.ParseWareID(cli.UnpackCLI.WareID)
	if err != nil {
		return api.WareID{}, err
	}
	path := cli.UnpackCLI.Path
	warehouses := convertWarehouseSlice(cli.UnpackCLI.SourcesWarehouseAddr)
	return unpackFn(ctx, wareID, path, cli.UnpackCLI.Filters, warehouses, monitor)
}

func executePack(ctx context.Context, cli baseCLI, packFn rio.PackFunc) (api.WareID, error) {
	monitor := rio.Monitor{}
	warehouse := api.WarehouseAddr(cli.PackCLI.TargetWarehouseAddr)

	return packFn(ctx, cli.PackCLI.Path, cli.PackCLI.Filters, warehouse, monitor)
}
