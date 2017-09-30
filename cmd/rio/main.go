package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	. "github.com/polydawn/go-errcat"
	"github.com/polydawn/refmt"
	"github.com/polydawn/refmt/json"
	"gopkg.in/alecthomas/kingpin.v2"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/rio"
	tar "go.polydawn.net/rio/transmat/tar"
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
		PackType            string             // Pack type
		Path                string             // Pack target path
		Filters             api.FilesetFilters // TODO: file filters for pack/unpack
		TargetWarehouseAddr string             // Warehouse address to push to
	}
	UnpackCLI struct {
		WareID               string // Ware id string "<kind>:<hash>"
		Path                 string // Unpack target path
		Filters              api.FilesetFilters
		PlacementMode        string   // Placement mode enum
		SourcesWarehouseAddr []string // Warehouse address to push to
	}
	MirrorCLI struct {
		SourcesWarehouseAddr []string // Warehouse addresses to fetch from
		TargetWarehouseAddr  string   // Warehouse address to push to
		WareID               string   // Ware id string "<kind>:<hash>"
	}
}

func configurePack(cli *baseCLI, appPack *kingpin.CmdClause) {
	// Non-filter flags
	appPack.Arg("pack", "Pack type").
		Required().
		StringVar(&cli.PackCLI.PackType)
	appPack.Arg("path", "Target path").
		Required().
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
	appPack.Flag("sticky", "Keep setuid, setgid, and sticky bits [keep, zero]").
		Default("keep").
		EnumVar(&cli.UnpackCLI.Filters.Sticky,
			"keep", "zero")
}

func configureUnpack(cli *baseCLI, appUnpack *kingpin.CmdClause) {
	// Non-filter flags
	appUnpack.Arg("path", "Target path").
		Required().
		StringVar(&cli.UnpackCLI.Path)
	appUnpack.Arg("ware", "Ware ID").
		Required().
		StringVar(&cli.UnpackCLI.WareID)
	appUnpack.Flag("placer", "Placement mode to use [copy, direct, mount, none]").
		EnumVar(&cli.UnpackCLI.PlacementMode,
			string(rio.Placement_Copy), string(rio.Placement_Direct), string(rio.Placement_Mount), string(rio.Placement_None))
	appUnpack.Flag("source", "Warehouses from which to fetch the ware").
		StringsVar(&cli.UnpackCLI.SourcesWarehouseAddr)

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
	appUnpack.Flag("sticky", "Keep setuid, setgid, and sticky bits [keep, zero]").
		Default("zero").
		EnumVar(&cli.UnpackCLI.Filters.Sticky,
			"keep", "zero")
}

func configureMirror(cli *baseCLI, appMirror *kingpin.CmdClause) {
	appMirror.Arg("ware", "Ware ID").
		Required().
		StringVar(&cli.MirrorCLI.WareID)
	appMirror.Arg("target", "Warehouse in which to place the ware").
		Required().
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

	var termErr error
	app.Terminate(func(status int) {
		termErr = fmt.Errorf("parsing error: %d\n", status)
	})
	cmd, err := app.Parse(args[1:])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return rio.ExitUsage
	}
	if termErr != nil {
		fmt.Fprintln(stderr, termErr)
		return rio.ExitUsage
	}
	var wareID api.WareID
	switch cmd {
	case appPack.FullCommand():
		wareID, err = executePack(ctx, cli)
		SerializeResult(cli.Format, wareID, err, stdout, stderr)
	case appUnpack.FullCommand():
		wareID, err = executeUnpack(ctx, cli)
		SerializeResult(cli.Format, wareID, err, stdout, stderr)
	case appMirror.FullCommand():
		fmt.Fprintln(stderr, "not implemented")
		return rio.ExitNotImplemented
	}

	return exitCode
}

func SerializeResult(format string, wareID api.WareID, resultErr error, stdout io.Writer, stderr io.Writer) {
	result := &rio.Event_Result{
		WareID: wareID,
	}
	result.SetError(resultErr)
	ev := rio.Event{Result: result}
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

func executeUnpack(ctx context.Context, cli baseCLI) (api.WareID, error) {
	wareID, err := api.ParseWareID(cli.UnpackCLI.WareID)
	if err != nil {
		return api.WareID{}, err
	}
	var (
		unpackFunc rio.UnpackFunc
	)
	switch wareID.Type {
	case "tar":
		unpackFunc = tar.Unpack
	default:
		return api.WareID{}, Errorf(rio.ErrUsage, "unsupported packtype %q", wareID.Type)
	}
	monitor := rio.Monitor{}
	path := cli.UnpackCLI.Path
	warehouses := convertWarehouseSlice(cli.UnpackCLI.SourcesWarehouseAddr)
	return unpackFunc(ctx, wareID, path, cli.UnpackCLI.Filters, rio.PlacementMode(cli.UnpackCLI.PlacementMode), warehouses, monitor)
}

func executePack(ctx context.Context, cli baseCLI) (api.WareID, error) {
	var (
		packType api.PackType
		packFunc rio.PackFunc
	)
	switch cli.PackCLI.PackType {
	case "tar":
		packType, packFunc = tar.PackType, tar.Pack
	default:
		return api.WareID{}, Errorf(rio.ErrUsage, "unsupported packtype %q", cli.PackCLI.PackType)
	}
	monitor := rio.Monitor{}
	warehouse := api.WarehouseAddr(cli.PackCLI.TargetWarehouseAddr)

	return packFunc(ctx, packType, cli.PackCLI.Path, cli.PackCLI.Filters, warehouse, monitor)
}
