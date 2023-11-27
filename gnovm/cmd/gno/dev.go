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
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	tmlog "github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
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

	gnoroot := gnoenv.RootDir()
	pkgpaths, err := parseArgumentsPath(args)
	if err != nil {
		return fmt.Errorf("unable to parse package paths: %w", err)
	}

	// Setup trap signal
	osm.TrapSignal(func() {
		cancel(nil)
	})

	logger := tmlog.NewNopLogger()

	// logger := log.NewTMLogger(log.NewSyncWriter(io.Out))

	dnode, err := setupDevNode(ctx, logger, cfg, pkgpaths, gnoroot)
	if err != nil {
		return err // already formated in setupDevNode
	}
	defer dnode.Close()

	io.Printf("dev-node listening: %s\n", dnode.GetRemoteAddress())

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

	w, err := setupFileWatcher(cfg, io, gnoroot, pkgpaths)
	if err != nil {
		return fmt.Errorf("unable to watch for files change: %w", err)
	}

	ccreload := make(chan struct{}, 1)
	go func() {
		for {
			var evt fsnotify.Event
			select {
			case <-ctx.Done():
				return
			case evt = <-w.Events:
			case err := <-w.Errors:
				cancel(fmt.Errorf("watch errors: %w", err))
				return
			}

			// Only catch write operation
			if evt.Op != fsnotify.Write {
				continue
			}

			io.Printf("%q updated\n", evt.Name)
			select {
			case ccreload <- struct{}{}:
			default:
				// Skip this event, if multiple write are made,
				// we only want to reload the node once
			}
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
	for {
		var err error

		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-ccreload:
			printActionf(io, "file-update", "Reload...\t")
			if err = dnode.Reload(ctx); err != nil {
				io.Printf("error: %s", err.Error())
			} else {
				io.Println("[DONE]")
			}
		case key := <-listenForKeyPress(io, rt):
			var err error
			switch key {
			case 'r', 'R':
				printActionf(io, key.String(), "Reload...\t")
				err = dnode.Reload(ctx)
			case KeyCtrlR:
				printActionf(io, key.String(), "Reset...\t")
				err = dnode.Reset()
			case KeyCtrlE:
				printActionf(io, key.String(), "Forcing error...\t")
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
				io.Printf("\terror: %s\n", err.Error())
			} else {
				io.Println("\t[DONE]")
			}
		}
	}
}

func printActionf(io commands.IO, action string, message string, args ...interface{}) {
	format := fmt.Sprintf(message, args...)
	io.Printf("%-10s %s", "<"+action+">", format)
}

func setupFileWatcher(cfg *devCfg, io commands.IO, gnoroot string, pkgspath []string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("unable to watch files: %w", err)
	}

	if cfg.noWatch {
		// noop watcher
		return watcher, nil
	}

	if !cfg.minimal {
		exampleFolder := filepath.Join(gnoroot, "examples")

		pkgs, err := gnomod.ListPkgs(exampleFolder)
		if err != nil {
			return nil, fmt.Errorf("unable to list pkgs in %q: %w", exampleFolder, err)
		}

		io.Printf("watching example pkgs: %q\n", exampleFolder)
		for _, pkg := range pkgs {
			if err := watcher.Add(pkg.Dir); err != nil {
				return nil, fmt.Errorf("unable to add %q: %w", pkg.Dir, err)
			}
		}

		watcher.Add(exampleFolder)
	}

	for _, path := range pkgspath {
		// list all packages from target path
		pkgs, err := gnomod.ListPkgs(path)
		if err != nil {
			return nil, fmt.Errorf("listing gno packages: %w", err)
		}

		io.Printf("watching custom pkg path: %q\n", path)
		for _, pkg := range pkgs {
			if err := watcher.Add(pkg.Dir); err != nil {
				return nil, fmt.Errorf("unable to add %q: %w", pkg.Dir, err)
			}
		}
	}

	return watcher, nil
}

// setupDevNode initializes and returns a new DevNode.
func setupDevNode(ctx context.Context, logger tmlog.Logger, cfg *devCfg, pkgspaths []string, gnoroot string) (*DevNode, error) {
	rootuser := crypto.MustAddressFromString(integration.DefaultAccount_Address)
	// defaultFee := std.NewFee(50000, std.MustParseCoin("1000000ugnot"))

	var err error

	genesisTxs := []std.Tx{}
	if !cfg.minimal {
		pkgstxs, err := loadDefaultPackages(rootuser, gnoroot)
		if err != nil {
			return nil, fmt.Errorf("unable to load default packages: %w", err)
		}
		genesisTxs = append(genesisTxs, pkgstxs...)
	}

	genesis := gnoland.GnoGenesisState{
		Balances: []gnoland.Balance{
			{
				Address: rootuser, // test1
				Amount:  std.MustParseCoins("10000000000000ugnot"),
			},
		},
		Txs: genesisTxs,
	}

	dnode, err := NewDevNode(logger, gnoroot, genesis)
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
		paths[i], err = filepath.Abs(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid path %q: %w", arg, err)
		}
	}

	return paths, nil
}
