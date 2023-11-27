package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
	tmlog "github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

type devCfg struct {
	bindAddr string
	minimal  bool
	verbose  bool
	noWatch  bool
}

var defaultDevOptions = &devCfg{
	bindAddr: "127.0.0.1:8888",
}

func newDevCmd(io commands.IO) *commands.Command {
	cfg := &devCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "dev",
			ShortUsage: "dev [flags] <path>",
			ShortHelp:  "Devs run a node for dev purpose, it will load the give package path",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execDev(cfg, args, io)
		},
	)
}

func (c *devCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.bindAddr,
		"web-bind",
		defaultDevOptions.bindAddr,
		"verbose output when deving",
	)

	fs.BoolVar(
		&c.minimal,
		"minimal",
		defaultDevOptions.verbose,
		"don't load example folder packages along default transaction",
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
		"watch for files change",
	)

}

func execDev(cfg *devCfg, args []string, io commands.IO) error {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	// Setup trap signal
	osm.TrapSignal(func() {
		cancel(nil)
	})

	// guess root dir
	gnoroot := gnoenv.RootDir()

	pkgpaths, err := parseArgumentsPath(args)
	if err != nil {
		return fmt.Errorf("unable to parse package paths: %w", err)
	}

	// logger := log.NewTMLogger(log.NewSyncWriter(io.Out))
	logger := tmlog.NewNopLogger()

	var dnode *DevNode
	{
		var err error
		dnode, err = setupDevNode(ctx, logger, cfg, pkgpaths, gnoroot)
		if err != nil {
			return err // already formated in setupDevNode
		}
		defer dnode.Close()
		io.Printf("dev-node listening: %s\n", dnode.GetRemoteAddress())
	}

	// RAWTerm setup
	rt := NewRawTerm()
	{
		restore, err := rt.Init()
		if err != nil {
			return fmt.Errorf("unable to init raw term for dev: %s", err)
		}
		defer restore()

		// correctly format output for terminal
		io.SetOut(commands.WriteNopCloser(rt))
	}

	// setup files watcher
	w, err := setupPkgsWatcher(cfg, dnode.ListPkgs())
	if err != nil {
		return fmt.Errorf("unable to watch for files change: %w", err)
	}
	ccpath := make(chan []string, 1)

	go func() {
		defer close(ccpath)

		const debounceTimeout = time.Millisecond * 500

		if err := handleDebounce(ctx, w, ccpath, debounceTimeout); err != nil {
			cancel(err)
		}
	}()

	// Gnoweb setup
	server := setupGnowebServer(cfg, dnode, rt)

	l, err := net.Listen("tcp", cfg.bindAddr)
	if err != nil {
		return fmt.Errorf("unable to listen to %q: %w", cfg.bindAddr, err)
	}

	go func() {
		var err error
		if srvErr := server.Serve(l); srvErr != nil {
			err = fmt.Errorf("HTTP server stopped with error: %w", srvErr)
		}
		cancel(err)
	}()

	io.Printfln("gnoweb listening on %q", l.Addr().String())
	defer server.Close()

	// main loop
	io.Printf("default address: %s\n", defaultCreator.String())
	io.Printf("chainid: %s\n", dnode.node.Config().ChainID())

	cckeypress := listenForKeyPress(io, rt)
	for {
		var err error

		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case paths := <-ccpath:
			for _, path := range paths {
				io.Printf("path %q has been modified\n", path)
			}
			printActionf(io, "file-update", "Reload...")
			if err := dnode.UpdatePackages(paths...); err != nil {
				io.Println("unable to update packages: %s", err.Error())
				continue
			}

			if err = dnode.Reload(ctx); err != nil {
				io.Printf(" [ERROR] - %s\n", err.Error())
			} else {
				io.Println(" [DONE]")
			}
		case key := <-cckeypress:
			var err error
			switch key {
			case 'r', 'R':
				printActionf(io, key.String(), "Reload All...")
				err = dnode.ReloadAll(ctx)
			case KeyCtrlR:
				printActionf(io, key.String(), "Reset...")
				err = dnode.Reset()
			case KeyCtrlE:
				printActionf(io, key.String(), "Forcing error...")
				err = fmt.Errorf("boom")
			case KeyCtrlT:
				printActionf(io, key.String(), "REPL MODE\n")
				termline := rt.TermMode()
				for line := range termline {
					io.Println("line:", line)
				}
				io.Println("<END>")
				continue
			case KeyCtrlC:
				cancel(nil)
				continue
			default:
				continue
			}

			if err != nil {
				io.Printf(" [ERROR] - %s\n", err.Error())
			} else {
				io.Println(" [DONE]")
			}
		}
	}
}

func printActionf(io commands.IO, action string, message string, args ...interface{}) {
	format := fmt.Sprintf(message, args...)
	io.Printf("%-20s %s", "<"+action+">", format)
}

func handleDebounce(ctx context.Context, w *fsnotify.Watcher, ccpaths chan<- []string, timeout time.Duration) error {
	var debounce <-chan time.Time
	pathlist := []string{}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case watchErr := <-w.Errors:
			return fmt.Errorf("watch error: %w", watchErr)
		case <-debounce:
			ccpaths <- pathlist
			// reset list and debounce
			pathlist = []string{}
			debounce = nil
		case evt := <-w.Events:
			fmt.Printf("evts: %s - %s\r\n", evt.Op.String(), evt.Name)
			if evt.Op == fsnotify.Write {
				pathlist = append(pathlist, evt.Name)
				debounce = time.After(timeout)
			}
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
func setupDevNode(ctx context.Context, logger tmlog.Logger, cfg *devCfg, pkgspath []string, gnoroot string) (*DevNode, error) {
	var err error

	if !cfg.minimal {
		examplesDir := filepath.Join(gnoroot, "examples")
		pkgspath = append(pkgspath, examplesDir)
	}

	dnode, err := NewDevNode(logger, gnoroot, pkgspath)
	if err != nil {
		return nil, fmt.Errorf("unable create dev node: %w", err)
	}

	// Wait for readiness
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("unable to wait for node readiness: %w", context.Cause(ctx))
	case <-dnode.WaitForNodeReadiness(): // ok
		return dnode, nil
	}

}

// setupGnowebServer initializes and starts the Gnoweb server.
func setupGnowebServer(cfg *devCfg, dnode *DevNode, rt *RawTerm) *http.Server {
	var server http.Server

	webConfig := gnoweb.NewDefaultConfig()
	webConfig.RemoteAddr = dnode.GetRemoteAddress()

	loggerweb := log.New(rt, "gnoweb: ", log.LstdFlags)
	app := gnoweb.MakeApp(loggerweb, webConfig)

	server.ReadHeaderTimeout = 60 * time.Second
	server.Handler = app.Router

	return &server
}

func parseArgumentsPath(args []string) (paths []string, err error) {
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
