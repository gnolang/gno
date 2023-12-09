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
	"time"

	"github.com/fsnotify/fsnotify"
	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
	tmlog "github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

type devCfg struct {
	webListenerAddr string
	minimal         bool
	verbose         bool
	noWatch         bool
}

var defaultDevOptions = &devCfg{
	webListenerAddr: "127.0.0.1:8888",
}

func main() {
	cfg := &devCfg{}

	stdio := commands.NewDefaultIO()
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "gnodev",
			ShortUsage: "gnodev [flags] [path ...]",
			ShortHelp:  "Runs an in-memory node and gno.land web server for development purposes.",
			LongHelp: `The gnodev command starts an in-memory node and a gno.land web interface
primarily for realm package development. It automatically loads the example folder and any
additional specified paths.`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execDev(cfg, args, stdio)
		})

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
func (c *devCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.webListenerAddr,
		"web-listener",
		defaultDevOptions.webListenerAddr,
		"web server listening address",
	)

	fs.BoolVar(
		&c.minimal,
		"minimal",
		defaultDevOptions.verbose,
		"do not load example folder packages",
	)

	fs.BoolVar(
		&c.verbose,
		"verbose",
		defaultDevOptions.verbose,
		"verbose output when deving",
	)

	fs.BoolVar(
		&c.noWatch,
		"no-watch",
		defaultDevOptions.noWatch,
		"do not watch for files change",
	)

}

func execDev(cfg *devCfg, args []string, io commands.IO) error {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	// guess root dir
	gnoroot := gnoenv.RootDir()

	pkgpaths, err := parseArgsPackages(io, args)
	if err != nil {
		return fmt.Errorf("unable to parse package paths: %w", err)
	}

	// XXX: find a good way to export or display node logs
	noopLogger := tmlog.NewNopLogger()

	// RAWTerm setup
	rt := gnodev.NewRawTerm()
	{
		restore, err := rt.Init()
		if err != nil {
			return fmt.Errorf("unable to init raw term for dev: %s", err)
		}
		defer restore()

		// correctly format output for terminal
		io.SetOut(commands.WriteNopCloser(rt))

		// Setup trap signal
		osm.TrapSignal(func() {
			restore()
			cancel(nil)
		})
	}

	nodeOut := rt.NamespacedWriter("Node")
	webOut := rt.NamespacedWriter("GnoWeb")
	keyOut := rt.NamespacedWriter("KeyPress")

	var dnode *gnodev.Node
	{
		var err error
		// XXX: redirect node output to a file
		dnode, err = setupDevNode(ctx, noopLogger, cfg, pkgpaths, gnoroot)
		if err != nil {
			return err // already formated in setupDevNode
		}
		defer dnode.Close()
	}

	fmt.Fprintf(nodeOut, "Listener: %s\n", dnode.GetRemoteAddress())

	// setup files watcher
	w, err := setupPkgsWatcher(cfg, dnode.ListPkgs())
	if err != nil {
		return fmt.Errorf("unable to watch for files change: %w", err)
	}

	pathChangeCh := make(chan []string, 1)
	go func() {
		defer close(pathChangeCh)
		if err := handleDebounce(ctx, w, pathChangeCh); err != nil {
			cancel(err)
		}
	}()

	// Gnoweb setup
	server := setupGnowebServer(cfg, dnode, rt)

	l, err := net.Listen("tcp", cfg.webListenerAddr)
	if err != nil {
		return fmt.Errorf("unable to listen to %q: %w", cfg.webListenerAddr, err)
	}

	go func() {
		var err error
		if srvErr := server.Serve(l); srvErr != nil {
			err = fmt.Errorf("HTTP server stopped with error: %w", srvErr)
		}
		cancel(err)
	}()
	defer server.Close()

	fmt.Fprintf(webOut, "Listener: %s\n", l.Addr())

	// Print basic infos
	fmt.Fprintf(nodeOut, "Default Address: %s\n", gnodev.DefaultCreator.String())
	fmt.Fprintf(nodeOut, "Chain ID: %s\n", dnode.Config().ChainID())

	rt.Taskf("[Ready]", "for commands and help, press `h`")

	for {
		var err error

		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case paths, ok := <-pathChangeCh:
			if !ok {
				cancel(nil)
				continue
			}

			for _, path := range paths {
				rt.Taskf("HotReload", "path %q has been modified", path)
			}

			fmt.Fprintln(nodeOut, "Loading package updates...")
			if err = dnode.UpdatePackages(paths...); err != nil {
				checkForError(rt, err)
				continue
			}

			fmt.Fprintln(nodeOut, "Reloading...")
			err = dnode.Reload(ctx)
			checkForError(rt, err)
		case key, ok := <-listenForKeyPress(keyOut, rt):
			if !ok {
				cancel(nil)
				continue
			}

			if cfg.verbose {
				fmt.Fprintf(keyOut, "<%s>\n", key.String())
			}

			switch key.Upper() {
			case gnodev.KeyH:
				printHelper(rt)
			case gnodev.KeyR:
				fmt.Fprintln(nodeOut, "Reloading all packages...")
				checkForError(nodeOut, dnode.ReloadAll(ctx))
			case gnodev.KeyCtrlR:
				fmt.Fprintln(nodeOut, "Reseting state...")
				checkForError(nodeOut, dnode.Reset(ctx))
			case gnodev.KeyCtrlC:
				cancel(nil)
			default:
			}
		}
	}
}

// XXX: Automatize this the same way command does
func printHelper(rt *gnodev.RawTerm) {
	rt.Taskf("Helper", `
Gno Dev Helper:
  h, H        Help - display this message
  r, R        Reload - Reload all packages to take change into account.
  Ctrl+R      Reset - Reset application state.
  Ctrl+C      Exit - Exit the application
`)
}

func handleDebounce(ctx context.Context, watcher *fsnotify.Watcher, changedPathsCh chan<- []string) error {
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

func setupPkgsWatcher(cfg *devCfg, pkgs []gnomod.Pkg) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("unable to watch files: %w", err)
	}

	if cfg.noWatch {
		// noop watcher
		return watcher, nil
	}

	for _, pkg := range pkgs {
		if err := watcher.Add(pkg.Dir); err != nil {
			return nil, fmt.Errorf("unable to watch %q: %w", pkg.Dir, err)
		}
	}

	return watcher, nil
}

// setupDevNode initializes and returns a new DevNode.
func setupDevNode(ctx context.Context, logger tmlog.Logger, cfg *devCfg, pkgspath []string, gnoroot string) (*gnodev.Node, error) {
	if !cfg.minimal {
		examplesDir := filepath.Join(gnoroot, "examples")
		pkgspath = append(pkgspath, examplesDir)
	}

	return gnodev.NewDevNode(ctx, logger, pkgspath)
}

// setupGnowebServer initializes and starts the Gnoweb server.
func setupGnowebServer(cfg *devCfg, dnode *gnodev.Node, rt *gnodev.RawTerm) *http.Server {
	var server http.Server

	webConfig := gnoweb.NewDefaultConfig()
	webConfig.RemoteAddr = dnode.GetRemoteAddress()

	loggerweb := tmlog.NewTMLogger(rt.NamespacedWriter("GnoWeb"))
	loggerweb.SetLevel(tmlog.LevelDebug)

	app := gnoweb.MakeApp(loggerweb, webConfig)

	server.ReadHeaderTimeout = 60 * time.Second
	server.Handler = app.Router

	return &server
}

func parseArgsPackages(io commands.IO, args []string) (paths []string, err error) {
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

func listenForKeyPress(w io.Writer, rt *gnodev.RawTerm) <-chan gnodev.KeyPress {
	cc := make(chan gnodev.KeyPress, 1)
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
