package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/rawterm"
	"github.com/gnolang/gno/contribs/gnodev/pkg/watcher"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	NodeLogName        = "Node"
	WebLogName         = "GnoWeb"
	KeyPressLogName    = "KeyPress"
	EventServerLogName = "Events"
)

type devCfg struct {
	webListenerAddr          string
	nodeRPCListenerAddr      string
	nodeP2PListenerAddr      string
	nodeProxyAppListenerAddr string

	minimal   bool
	verbose   bool
	hotreload bool
	noWatch   bool
	noReplay  bool
	maxGas    int64
	chainId   string
}

var defaultDevOptions = &devCfg{
	chainId:             "dev",
	maxGas:              10_000_000_000,
	webListenerAddr:     "127.0.0.1:8888",
	nodeRPCListenerAddr: "127.0.0.1:36657",

	// As we have no reason to configure this yet, set this to random port
	// to avoid potential conflict with other app
	nodeP2PListenerAddr:      "tcp://127.0.0.1:0",
	nodeProxyAppListenerAddr: "tcp://127.0.0.1:0",
}

func main() {
	cfg := &devCfg{}

	stdio := commands.NewDefaultIO()
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "gnodev",
			ShortUsage: "gnodev [flags] [path ...]",
			ShortHelp:  "runs an in-memory node and gno.land web server for development purposes.",
			LongHelp: `The gnodev command starts an in-memory node and a gno.land web interface
primarily for realm package development. It automatically loads the 'examples' directory and any
additional specified paths.`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execDev(cfg, args, stdio)
		})

	cmd.Execute(context.Background(), os.Args[1:])
}

func (c *devCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.webListenerAddr,
		"web-listener",
		defaultDevOptions.webListenerAddr,
		"web server listening address",
	)

	fs.StringVar(
		&c.nodeRPCListenerAddr,
		"node-rpc-listener",
		defaultDevOptions.nodeRPCListenerAddr,
		"gnoland rpc node listening address",
	)

	fs.BoolVar(
		&c.minimal,
		"minimal",
		defaultDevOptions.minimal,
		"do not load packages from examples directory",
	)

	fs.BoolVar(
		&c.verbose,
		"verbose",
		defaultDevOptions.verbose,
		"verbose output when deving",
	)

	fs.StringVar(
		&c.chainId,
		"chain-id",
		defaultDevOptions.chainId,
		"set node ChainID",
	)

	fs.BoolVar(
		&c.noWatch,
		"no-watch",
		defaultDevOptions.noWatch,
		"do not watch for files change",
	)

	fs.BoolVar(
		&c.noReplay,
		"no-replay",
		defaultDevOptions.noReplay,
		"do not replay previous transactions on reload",
	)

	fs.Int64Var(
		&c.maxGas,
		"max-gas",
		defaultDevOptions.maxGas,
		"set the maximum gas by block",
	)
}

func execDev(cfg *devCfg, args []string, io commands.IO) error {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	// Guess root dir
	gnoroot := gnoenv.RootDir()

	// Check and Parse packages
	pkgpaths, err := parseArgsPackages(args)
	if err != nil {
		return fmt.Errorf("unable to parse package paths: %w", err)
	}

	if !cfg.minimal {
		examplesDir := filepath.Join(gnoroot, "examples")
		pkgpaths = append(pkgpaths, examplesDir)
	}

	// Setup Raw Terminal
	rt, restore, err := setupRawTerm(io)
	if err != nil {
		return fmt.Errorf("unable to init raw term: %w", err)
	}
	defer restore()

	// Setup trap signal
	osm.TrapSignal(func() {
		restore()
		cancel(nil)
	})

	zapLoggerEvents := NewZapLogger(rt.NamespacedWriter(EventServerLogName), zapcore.DebugLevel)
	loggerEvents := log.ZapLoggerToSlog(zapLoggerEvents)
	emitterServer := emitter.NewServer(loggerEvents)

	// Setup Dev Node
	// XXX: find a good way to export or display node logs
	devNode, err := setupDevNode(ctx, cfg, emitterServer, rt, pkgpaths)
	if err != nil {
		return err
	}
	defer devNode.Close()

	rt.Taskf(NodeLogName, "Listener: %s\n", devNode.GetRemoteAddress())
	rt.Taskf(NodeLogName, "Default Address: %s\n", gnodev.DefaultCreator.String())
	rt.Taskf(NodeLogName, "Chain ID: %s\n", cfg.chainId)

	// Create server
	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    cfg.webListenerAddr,
	}
	defer server.Close()

	// Setup gnoweb
	webhandler := setupGnoWebServer(cfg, devNode, rt)

	// Setup HotReload if needed
	if !cfg.noWatch {
		evtstarget := fmt.Sprintf("%s/_events", server.Addr)
		mux.Handle("/_events", emitterServer)
		mux.Handle("/", emitter.NewMiddleware(evtstarget, webhandler))
	} else {
		mux.Handle("/", webhandler)
	}

	go func() {
		err := server.ListenAndServe()
		cancel(err)
	}()

	rt.Taskf(WebLogName, "Listener: http://%s\n", server.Addr)

	watcher, err := watcher.NewPackageWatcher(loggerEvents, emitterServer)
	if err != nil {
		return fmt.Errorf("unable to setup packages watcher: %w", err)
	}
	defer watcher.Stop()

	// Add node pkgs to watcher
	watcher.AddPackages(devNode.ListPkgs()...)

	// GnoDev should be ready, run event loop
	rt.Taskf("[Ready]", "for commands and help, press `h`")

	// Run the main event loop
	return runEventLoop(ctx, cfg, rt, devNode, watcher)
}

// XXX: Automatize this the same way command does
func printHelper(rt *rawterm.RawTerm) {
	rt.Taskf("Helper", `
Gno Dev Helper:
  H           Help - display this message
  R           Reload - Reload all packages to take change into account.
  Ctrl+R      Reset - Reset application state.
  Ctrl+C      Exit - Exit the application
`)
}

func runEventLoop(
	ctx context.Context,
	cfg *devCfg,
	rt *rawterm.RawTerm,
	dnode *dev.Node,
	watch *watcher.PackageWatcher,
) error {
	nodeOut := rt.NamespacedWriter(NodeLogName)
	keyOut := rt.NamespacedWriter(KeyPressLogName)

	keyPressCh := listenForKeyPress(keyOut, rt)
	for {
		var err error

		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case pkgs, ok := <-watch.PackagesUpdate:
			if !ok {
				return nil
			}

			fmt.Fprintln(nodeOut, "Loading package updates...")
			if err = dnode.UpdatePackages(pkgs.PackagesPath()...); err != nil {
				return fmt.Errorf("unable to update packages: %w", err)
			}

			fmt.Fprintln(nodeOut, "Reloading...")
			err = dnode.Reload(ctx)

			checkForError(rt, err)

		case key, ok := <-keyPressCh:
			if !ok {
				return nil
			}

			if cfg.verbose {
				fmt.Fprintf(keyOut, "<%s>\n", key.String())
			}

			switch key.Upper() {
			case rawterm.KeyH:
				printHelper(rt)
			case rawterm.KeyR:
				fmt.Fprintln(nodeOut, "Reloading all packages...")
				checkForError(nodeOut, dnode.ReloadAll(ctx))
			case rawterm.KeyCtrlR:
				fmt.Fprintln(nodeOut, "Reseting state...")
				checkForError(nodeOut, dnode.Reset(ctx))
			case rawterm.KeyCtrlC:
				return nil
			default:
			}

			// Listen for the next keypress
			keyPressCh = listenForKeyPress(keyOut, rt)
		}
	}
}

func runPkgsWatcher(ctx context.Context, cfg *devCfg, pkgs []gnomod.Pkg, changedPathsCh chan<- []string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("unable to watch files: %w", err)
	}

	if cfg.noWatch {
		// Noop watcher, wait until context has been cancel
		<-ctx.Done()
		return ctx.Err()
	}

	for _, pkg := range pkgs {
		if err := watcher.Add(pkg.Dir); err != nil {
			return fmt.Errorf("unable to watch %q: %w", pkg.Dir, err)
		}
	}

	const timeout = time.Millisecond * 500

	var debounceTimer <-chan time.Time
	pathList := []string{}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case watchErr := <-watcher.Errors:
			return fmt.Errorf("watch error: %w", watchErr)
		case <-debounceTimer:
			changedPathsCh <- pathList
			// Reset pathList and debounceTimer
			pathList = []string{}
			debounceTimer = nil
		case evt := <-watcher.Events:
			if evt.Op != fsnotify.Write {
				continue
			}

			pathList = append(pathList, evt.Name)
			debounceTimer = time.After(timeout)
		}
	}
}

func setupRawTerm(io commands.IO) (rt *rawterm.RawTerm, restore func() error, err error) {
	rt = rawterm.NewRawTerm()

	restore, err = rt.Init()
	if err != nil {
		return nil, nil, err
	}

	// Correctly format output for terminal
	io.SetOut(commands.WriteNopCloser(rt))
	return rt, restore, nil
}

// setupDevNode initializes and returns a new DevNode.
func setupDevNode(
	ctx context.Context,
	cfg *devCfg,
	remitter emitter.Emitter,
	rt *rawterm.RawTerm,
	pkgspath []string,
) (*gnodev.Node, error) {
	nodeOut := rt.NamespacedWriter("Node")
	zapLogger := NewZapLogger(nodeOut, zapcore.ErrorLevel)

	gnoroot := gnoenv.RootDir()

	// configure gnoland node
	config := gnodev.DefaultNodeConfig(gnoroot)
	config.PackagesPathList = pkgspath
	config.TMConfig.RPC.ListenAddress = resolveUnixOrTCPAddr(cfg.nodeRPCListenerAddr)
	config.NoReplay = cfg.noReplay
	config.SkipFailingGenesisTxs = true
	config.MaxGasPerBlock = cfg.maxGas
	config.ChainID = cfg.chainId

	// other listeners
	config.TMConfig.P2P.ListenAddress = defaultDevOptions.nodeP2PListenerAddr
	config.TMConfig.ProxyApp = defaultDevOptions.nodeProxyAppListenerAddr

	return gnodev.NewDevNode(ctx, log.ZapLoggerToSlog(zapLogger), remitter, config)
}

// setupGnowebServer initializes and starts the Gnoweb server.
func setupGnoWebServer(cfg *devCfg, dnode *gnodev.Node, rt *rawterm.RawTerm) http.Handler {
	webConfig := gnoweb.NewDefaultConfig()
	webConfig.RemoteAddr = dnode.GetRemoteAddress()
	webConfig.HelpRemote = dnode.GetRemoteAddress()
	webConfig.HelpChainID = cfg.chainId

	zapLogger := NewZapLogger(rt.NamespacedWriter("GnoWeb"), zapcore.DebugLevel)
	app := gnoweb.MakeApp(log.ZapLoggerToSlog(zapLogger), webConfig)
	return app.Router
}

func parseArgsPackages(args []string) (paths []string, err error) {
	paths = make([]string, len(args))
	for i, arg := range args {
		abspath, err := filepath.Abs(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid path %q: %w", arg, err)
		}

		ppath, err := gnomod.FindRootDir(abspath)
		if err != nil {
			return nil, fmt.Errorf("unable to find root dir of %q: %w", abspath, err)
		}

		paths[i] = ppath
	}

	return paths, nil
}

func listenForKeyPress(w io.Writer, rt *rawterm.RawTerm) <-chan rawterm.KeyPress {
	cc := make(chan rawterm.KeyPress, 1)
	go func() {
		defer close(cc)
		key, err := rt.ReadKeyPress()
		if err != nil {
			fmt.Fprintf(w, "unable to read keypress: %s\n", err.Error())
			return
		}

		cc <- key
	}()

	return cc
}

func checkForError(w io.Writer, err error) {
	if err != nil {
		fmt.Fprintf(w, "[ERROR] - %s\n", err.Error())
		return
	}

	fmt.Fprintln(w, "[DONE]")
}

func resolveUnixOrTCPAddr(in string) (out string) {
	var err error
	var addr net.Addr

	if strings.HasPrefix(in, "unix://") {
		in = strings.TrimPrefix(in, "unix://")
		if addr, err := net.ResolveUnixAddr("unix", in); err == nil {
			return fmt.Sprintf("%s://%s", addr.Network(), addr.String())
		}

		err = fmt.Errorf("unable to resolve unix address `unix://%s`: %w", in, err)
	} else { // don't bother to checking prefix
		in = strings.TrimPrefix(in, "tcp://")
		if addr, err = net.ResolveTCPAddr("tcp", in); err == nil {
			return fmt.Sprintf("%s://%s", addr.Network(), addr.String())
		}

		err = fmt.Errorf("unable to resolve tcp address `tcp://%s`: %w", in, err)
	}

	panic(err)
}

// NewZapLogger creates a zap logger with a console encoder for development use.
func NewZapLogger(w io.Writer, level zapcore.Level) *zap.Logger {
	// Build encoder config
	consoleConfig := zap.NewDevelopmentEncoderConfig()
	consoleConfig.TimeKey = ""
	consoleConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleConfig.EncodeName = zapcore.FullNameEncoder

	// Build encoder
	enc := zapcore.NewConsoleEncoder(consoleConfig)
	return log.NewZapLogger(enc, w, level)
}
