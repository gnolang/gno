package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/contribs/gnodev/pkg/rawterm"
	"github.com/gnolang/gno/contribs/gnodev/pkg/watcher"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

type App struct {
	cfg    *devCfg
	io     commands.IO
	logger *slog.Logger

	webHome       string
	paths         []string
	devNode       *gnodev.Node
	emitterServer *emitter.Server
	watcher       *watcher.PackageWatcher
	book          *address.Book
	exportPath    string

	// XXX: move this
	exported uint
}

func runApp(cfg *devCfg, cio commands.IO, dirs ...string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var rt *rawterm.RawTerm

	var out io.Writer
	if cfg.interactive {
		var err error
		var restore func() error

		// Setup raw terminal for interaction
		rt, restore, err = setupRawTerm(cfg, cio)
		if err != nil {
			return fmt.Errorf("unable to init raw term: %w", err)
		}
		defer restore()

		osm.TrapSignal(func() {
			cancel()
			restore()
		})

		out = rt
	} else {
		osm.TrapSignal(cancel)
		out = cio.Out()
	}

	logger, err := setuplogger(cfg, out)
	if err != nil {
		return fmt.Errorf("unable to setup logger: %w", err)
	}

	app := NewApp(logger, cfg, cio)
	if err := app.Setup(ctx, dirs...); err != nil {
		return err
	}

	if rt != nil {
		go func() {
			app.RunInteractive(ctx, rt)
			cancel()
		}()
	}

	return app.RunServer(ctx, rt)
}

func NewApp(logger *slog.Logger, cfg *devCfg, io commands.IO) *App {
	return &App{
		logger: logger,
		cfg:    cfg,
		io:     io,
	}
}

func (ds *App) Setup(ctx context.Context, dirs ...string) error {
	var err error

	if err = ds.cfg.validateConfigFlags(); err != nil {
		return fmt.Errorf("validate error: %w", err)
	}

	loggerEvents := ds.logger.WithGroup(EventServerLogName)
	ds.emitterServer = emitter.NewServer(loggerEvents)

	// XXX: it would be nice to having this hardcoded
	examplesDir := filepath.Join(ds.cfg.root, "examples")

	resolver, localPaths := setupPackagesResolver(ds.logger.WithGroup(LoaderLogName), ds.cfg, dirs...)
	loader := packages.NewGlobLoader(examplesDir, resolver)

	// Setup default web home realm, fallback on first local path
	switch webHome := ds.cfg.webHome; webHome {
	case "":
		if len(localPaths) > 0 {
			ds.webHome = strings.TrimPrefix(localPaths[0], ds.cfg.chainDomain)
		}
	case "/": // skip
	default:
		ds.webHome = webHome
	}

	// generate paths
	paths := strings.Split(ds.cfg.paths, ",")
	ds.paths = append(localPaths, paths...)

	ds.book, err = setupAddressBook(ds.logger.WithGroup(AccountsLogName), ds.cfg)
	if err != nil {
		return fmt.Errorf("unable to load keybase: %w", err)
	}

	balances, err := generateBalances(ds.book, ds.cfg)
	if err != nil {
		return fmt.Errorf("unable to generate balances: %w", err)
	}
	ds.logger.Debug("balances loaded", "list", balances.List())

	nodeLogger := ds.logger.WithGroup(NodeLogName)
	nodeCfg := setupDevNodeConfig(ds.cfg, nodeLogger, ds.emitterServer, balances, loader)
	ds.devNode, err = setupDevNode(ctx, ds.cfg, nodeCfg, ds.paths...)
	if err != nil {
		return err
	}

	ds.watcher, err = watcher.NewPackageWatcher(loggerEvents, ds.emitterServer)
	if err != nil {
		return fmt.Errorf("unable to setup packages watcher: %w", err)
	}

	ds.watcher.UpdatePackagesWatch(ds.devNode.ListPkgs()...)

	return nil
}

func (ds *App) setupHandlers() http.Handler {
	mux := http.NewServeMux()
	webhandler := setupGnoWebServer(ds.logger.WithGroup(WebLogName), ds.cfg, ds.devNode)

	if ds.webHome != "" {
		serveWeb := webhandler.ServeHTTP
		webhandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "" || r.URL.Path == "/" {
				http.Redirect(w, r, ds.webHome, http.StatusFound)
			} else {
				serveWeb(w, r)
			}
		})
	}

	// Setup unsage api
	if ds.cfg.unsafeAPI {
		mux.HandleFunc("/reset", func(res http.ResponseWriter, req *http.Request) {
			if err := ds.devNode.Reset(req.Context()); err != nil {
				ds.logger.Error("failed to reset", slog.Any("err", err))
				res.WriteHeader(http.StatusInternalServerError)
			}
		})

		mux.HandleFunc("/reload", func(res http.ResponseWriter, req *http.Request) {
			if err := ds.devNode.Reload(req.Context()); err != nil {
				ds.logger.Error("failed to reload", slog.Any("err", err))
				res.WriteHeader(http.StatusInternalServerError)
			}
		})
	}

	if !ds.cfg.noWatch {
		evtstarget := fmt.Sprintf("%s/_events", ds.cfg.webListenerAddr)
		mux.Handle("/_events", ds.emitterServer)
		mux.Handle("/", emitter.NewMiddleware(evtstarget, webhandler))
	} else {
		mux.Handle("/", webhandler)
	}

	return mux
}

func (ds *App) RunServer(ctx context.Context, term *rawterm.RawTerm) error {
	ctx, cancelWith := context.WithCancelCause(ctx)
	defer cancelWith(nil)

	addr := ds.cfg.webListenerAddr
	ds.logger.WithGroup(WebLogName).Info("gnoweb started", "lisn", fmt.Sprintf("http://%s", addr))

	server := &http.Server{
		Handler:           ds.setupHandlers(),
		Addr:              ds.cfg.webListenerAddr,
		ReadHeaderTimeout: time.Second * 60,
	}

	go func() {
		err := server.ListenAndServe()
		cancelWith(err)
	}()

	if ds.cfg.interactive {
		ds.logger.WithGroup("--- READY").Info("for commands and help, press `h`")
	} else {
		ds.logger.Info("node is ready")
	}

	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case _, ok := <-ds.watcher.PackagesUpdate:
			if !ok {
				return nil
			}
			ds.logger.WithGroup(NodeLogName).Info("reloading...")
			if err := ds.devNode.Reload(ctx); err != nil {
				ds.logger.WithGroup(NodeLogName).Error("unable to reload node", "err", err)
			}
			ds.watcher.UpdatePackagesWatch(ds.devNode.ListPkgs()...)
		}
	}
}

func (ds *App) RunInteractive(ctx context.Context, term *rawterm.RawTerm) {
	var keyPressCh <-chan rawterm.KeyPress
	if ds.cfg.interactive {
		keyPressCh = listenForKeyPress(ds.logger.WithGroup(KeyPressLogName), term)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case key, ok := <-keyPressCh:
			if !ok {
				return
			}

			if key == rawterm.KeyCtrlC {
				return
			}

			ds.handleKeyPress(ctx, key)
			keyPressCh = listenForKeyPress(ds.logger.WithGroup(KeyPressLogName), term)
		}
	}
}

var helper string = `For more in-depth documentation, visit the GNO Tooling CLI documentation:
https://docs.gno.land/gno-tooling/cli/gno-tooling-gnodev

P           Previous TX  - Go to the previous tx
N           Next TX      - Go to the next tx
E           Export       - Export the current state as genesis doc
A           Accounts     - Display known accounts and balances
H           Help         - Display this message
R           Reload       - Reload all packages to take change into account.
Ctrl+S      Save State   - Save the current state
Ctrl+R      Reset        - Reset application to it's initial/save state.
Ctrl+C      Exit         - Exit the application
`

func (ds *App) handleKeyPress(ctx context.Context, key rawterm.KeyPress) {
	var err error
	ds.logger.WithGroup(KeyPressLogName).Debug(fmt.Sprintf("<%s>", key.String()))

	switch key.Upper() {
	case rawterm.KeyH: // Helper
		ds.logger.Info("Gno Dev Helper", "helper", helper)

	case rawterm.KeyA: // Accounts
		logAccounts(ds.logger.WithGroup(AccountsLogName), ds.book, ds.devNode)

	case rawterm.KeyR: // Reload
		ds.logger.WithGroup(NodeLogName).Info("reloading...")
		if err = ds.devNode.ReloadAll(ctx); err != nil {
			ds.logger.WithGroup(NodeLogName).Error("unable to reload node", "err", err)
		}

	case rawterm.KeyCtrlR: // Reset
		ds.logger.WithGroup(NodeLogName).Info("reseting node state...")
		if err = ds.devNode.Reset(ctx); err != nil {
			ds.logger.WithGroup(NodeLogName).Error("unable to reset node state", "err", err)
		}

	case rawterm.KeyCtrlS: // Save
		ds.logger.WithGroup(NodeLogName).Info("saving state...")
		if err := ds.devNode.SaveCurrentState(ctx); err != nil {
			ds.logger.WithGroup(NodeLogName).Error("unable to save node state", "err", err)
		}

	case rawterm.KeyE: // Export
		// Create a temporary export dir
		if ds.exported == 0 {
			ds.exportPath, err = os.MkdirTemp("", "gnodev-export")
			if err != nil {
				ds.logger.WithGroup(NodeLogName).Error("unable to create `export` directory", "err", err)
				return
			}
		}
		ds.exported++

		ds.logger.WithGroup(NodeLogName).Info("exporting state...")
		doc, err := ds.devNode.ExportStateAsGenesis(ctx)
		if err != nil {
			ds.logger.WithGroup(NodeLogName).Error("unable to export node state", "err", err)
			return
		}

		docfile := filepath.Join(ds.exportPath, fmt.Sprintf("export_%d.jsonl", ds.exported))
		if err := doc.SaveAs(docfile); err != nil {
			ds.logger.WithGroup(NodeLogName).Error("unable to save genesis", "err", err)
		}

		ds.logger.WithGroup(NodeLogName).Info("node state exported", "file", docfile)

	case rawterm.KeyN: // Next tx
		ds.logger.Info("moving forward...")
		if err := ds.devNode.MoveToNextTX(ctx); err != nil {
			ds.logger.WithGroup(NodeLogName).Error("unable to move forward", "err", err)
		}

	case rawterm.KeyP: // Previous tx
		ds.logger.Info("moving backward...")
		if err := ds.devNode.MoveToPreviousTX(ctx); err != nil {
			ds.logger.WithGroup(NodeLogName).Error("unable to move backward", "err", err)
		}
	default:
	}
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

// func resolvePackagesPathFromArgs(cfg *devCfg, bk *address.Book, args []string) ([]gnodev.PackageModifier, error) {
// 	modifiers := make([]gnodev.PackageModifier, 0, len(args))

// 	if cfg.deployKey == "" {
// 		return nil, fmt.Errorf("default deploy key cannot be empty")
// 	}

// 	defaultKey, _, ok := bk.GetFromNameOrAddress(cfg.deployKey)
// 	if !ok {
// 		return nil, fmt.Errorf("unable to get deploy key %q", cfg.deployKey)
// 	}

// 	if len(args) == 0 {
// 		args = append(args, ".") // add current dir if none are provided
// 	}

// 	for _, arg := range args {
// 		path, err := gnodev.ResolvePackageModifierQuery(bk, arg)
// 		if err != nil {
// 			return nil, fmt.Errorf("invalid package path/query %q: %w", arg, err)
// 		}

// 		// Assign a default creator if user haven't specified it.
// 		if path.Creator.IsZero() {
// 			path.Creator = defaultKey
// 		}

// 		modifiers = append(modifiers, path)
// 	}

// 	// Add examples folder if minimal is set to false
// 	if cfg.minimal {
// 		modifiers = append(modifiers, gnodev.PackageModifier{
// 			Path:    filepath.Join(cfg.root, "examples"),
// 			Creator: defaultKey,
// 			Deposit: nil,
// 		})
// 	}

// 	return modifiers, nil
// }
