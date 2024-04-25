package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
	"github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/logger"
	"github.com/gnolang/gno/contribs/gnodev/pkg/rawterm"
	"github.com/gnolang/gno/contribs/gnodev/pkg/watcher"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	zaplog "github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/muesli/termenv"
)

const (
	NodeLogName        = "Node"
	WebLogName         = "GnoWeb"
	KeyPressLogName    = "KeyPress"
	EventServerLogName = "Event"
)

type devCfg struct {
	webListenerAddr          string
	nodeRPCListenerAddr      string
	nodeP2PListenerAddr      string
	nodeProxyAppListenerAddr string

	minimal    bool
	verbose    bool
	hotreload  bool
	noWatch    bool
	noReplay   bool
	maxGas     int64
	chainId    string
	serverMode bool
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
		&c.serverMode,
		"server-mode",
		defaultDevOptions.serverMode,
		"disable interaction, and adjust logging for server use.",
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

func execDev(cfg *devCfg, args []string, io commands.IO) (err error) {
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
	rt, restore, err := setupRawTerm(cfg, io)
	if err != nil {
		return fmt.Errorf("unable to init raw term: %w", err)
	}
	defer restore()

	// Setup trap signal
	osm.TrapSignal(func() {
		cancel(nil)
		restore()
	})

	logger := setuplogger(cfg, rt)
	loggerEvents := logger.WithGroup(EventServerLogName)
	emitterServer := emitter.NewServer(loggerEvents)

	// Setup Dev Node
	// XXX: find a good way to export or display node logs
	devNode, err := setupDevNode(ctx, logger, cfg, emitterServer, pkgpaths)
	if err != nil {
		return err
	}
	defer devNode.Close()

	logger.WithGroup(NodeLogName).
		Info("node started",
			"lisn", devNode.GetRemoteAddress(),
			"addr", gnodev.DefaultCreator.String(),
			"chainID", cfg.chainId,
		)

	// Create server
	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    cfg.webListenerAddr,
	}
	defer server.Close()

	// Setup gnoweb
	webhandler := setupGnoWebServer(logger.WithGroup(WebLogName), cfg, devNode)

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

	logger.WithGroup(WebLogName).
		Info("gnoweb started",
			"lisn", fmt.Sprintf("http://%s", server.Addr))

	watcher, err := watcher.NewPackageWatcher(loggerEvents, emitterServer)
	if err != nil {
		return fmt.Errorf("unable to setup packages watcher: %w", err)
	}
	defer watcher.Stop()

	// Add node pkgs to watcher
	watcher.AddPackages(devNode.ListPkgs()...)

	logger.WithGroup("--- READY").Info("for commands and help, press `h`")

	// Run the main event loop
	return runEventLoop(ctx, logger, rt, devNode, watcher)
}

var helper string = `
H           Help - display this message
R           Reload - Reload all packages to take change into account.
Ctrl+R      Reset - Reset application state.
Ctrl+C      Exit - Exit the application
`

func runEventLoop(
	ctx context.Context,
	logger *slog.Logger,
	rt *rawterm.RawTerm,
	dnode *dev.Node,
	watch *watcher.PackageWatcher,
) error {

	keyPressCh := listenForKeyPress(logger.WithGroup(KeyPressLogName), rt)
	for {
		var err error

		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case pkgs, ok := <-watch.PackagesUpdate:
			if !ok {
				return nil
			}

			// fmt.Fprintln(nodeOut, "Loading package updates...")
			if err = dnode.UpdatePackages(pkgs.PackagesPath()...); err != nil {
				return fmt.Errorf("unable to update packages: %w", err)
			}

			logger.WithGroup(NodeLogName).Info("reloading...")
			if err = dnode.Reload(ctx); err != nil {
				logger.WithGroup(NodeLogName).
					Error("unable to reload node", "err", err)
			}

		case key, ok := <-keyPressCh:
			if !ok {
				return nil
			}

			logger.WithGroup(KeyPressLogName).Debug(
				fmt.Sprintf("<%s>", key.String()),
			)

			switch key.Upper() {
			case rawterm.KeyH:
				logger.Info("Gno Dev Helper", "helper", helper)
			case rawterm.KeyR:
				logger.WithGroup(NodeLogName).Info("reloading...")
				if err = dnode.ReloadAll(ctx); err != nil {
					logger.WithGroup(NodeLogName).
						Error("unable to reload node", "err", err)

				}

			case rawterm.KeyCtrlR:
				logger.WithGroup(NodeLogName).Info("reseting node state...")
				if err = dnode.Reset(ctx); err != nil {
					logger.WithGroup(NodeLogName).
						Error("unable to reset node state", "err", err)
				}

			case rawterm.KeyCtrlC:
				return nil
			case rawterm.KeyCtrlE:
				panic("NOOOO")
				return nil
			default:
			}

			// Reset listen for the next keypress
			keyPressCh = listenForKeyPress(logger.WithGroup(KeyPressLogName), rt)
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

var noopRestore = func() error { return nil }

func setupRawTerm(cfg *devCfg, io commands.IO) (*rawterm.RawTerm, func() error, error) {
	rt := rawterm.NewRawTerm()
	restore := noopRestore
	if !cfg.serverMode {
		var err error
		restore, err = rt.Init()
		if err != nil {
			return nil, nil, err
		}
	}

	// correctly format output for terminal
	io.SetOut(commands.WriteNopCloser(rt))
	return rt, restore, nil
}

// setupDevNode initializes and returns a new DevNode.
func setupDevNode(
	ctx context.Context,
	logger *slog.Logger,
	cfg *devCfg,
	remitter emitter.Emitter,
	pkgspath []string,
) (*gnodev.Node, error) {
	nodeLogger := logger.WithGroup(NodeLogName)

	gnoroot := gnoenv.RootDir()

	// configure gnoland node
	config := gnodev.DefaultNodeConfig(gnoroot)
	config.PackagesPathList = pkgspath
	config.TMConfig.RPC.ListenAddress = resolveUnixOrTCPAddr(cfg.nodeRPCListenerAddr)
	config.NoReplay = cfg.noReplay
	config.MaxGasPerBlock = cfg.maxGas
	config.ChainID = cfg.chainId

	// other listeners
	config.TMConfig.P2P.ListenAddress = defaultDevOptions.nodeP2PListenerAddr
	config.TMConfig.ProxyApp = defaultDevOptions.nodeProxyAppListenerAddr

	return gnodev.NewDevNode(ctx, nodeLogger, remitter, config)
}

// setupGnowebServer initializes and starts the Gnoweb server.
func setupGnoWebServer(logger *slog.Logger, cfg *devCfg, dnode *gnodev.Node) http.Handler {
	webConfig := gnoweb.NewDefaultConfig()
	webConfig.RemoteAddr = dnode.GetRemoteAddress()
	webConfig.HelpRemote = dnode.GetRemoteAddress()
	webConfig.HelpChainID = cfg.chainId

	app := gnoweb.MakeApp(logger, webConfig)
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

func listenForKeyPress(logger *slog.Logger, rt *rawterm.RawTerm) <-chan rawterm.KeyPress {
	cc := make(chan rawterm.KeyPress, 1)
	go func() {
		defer close(cc)
		key, err := rt.ReadKeyPress()
		if err != nil {
			logger.Error("unable to read keypress", "err", err)
			return
		}

		cc <- key
	}()

	return cc
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

func setuplogger(cfg *devCfg, out io.Writer) *slog.Logger {
	level := slog.LevelInfo
	if cfg.verbose {
		level = slog.LevelDebug
	}

	if cfg.serverMode {
		zaplogger := logger.NewZapLogger(out, level)
		return zaplog.ZapLoggerToSlog(zaplogger)
	}

	// Detect term color profile
	colorProfile := termenv.DefaultOutput().Profile
	clogger := logger.NewColumnLogger(out, level, colorProfile)

	// Register well known group color with system colors
	clogger.RegisterGroupColor(NodeLogName, lipgloss.Color("3"))
	clogger.RegisterGroupColor(WebLogName, lipgloss.Color("4"))
	clogger.RegisterGroupColor(KeyPressLogName, lipgloss.Color("5"))
	clogger.RegisterGroupColor(EventServerLogName, lipgloss.Color("6"))

	return slog.New(clogger)
}
