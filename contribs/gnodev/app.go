package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/middleware"
	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/contribs/gnodev/pkg/proxy"
	"github.com/gnolang/gno/contribs/gnodev/pkg/rawterm"
	"github.com/gnolang/gno/contribs/gnodev/pkg/watcher"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

const (
	DefaultDeployerName = integration.DefaultAccount_Name
	DefaultDeployerSeed = integration.DefaultAccount_Seed
)

var defaultDeployerAddress = crypto.MustAddressFromString(integration.DefaultAccount_Address)

const (
	NodeLogName        = "Node"
	WebLogName         = "GnoWeb"
	KeyPressLogName    = "KeyPress"
	EventServerLogName = "Event"
	AccountsLogName    = "Accounts"
	LoaderLogName      = "Loader"
	ProxyLogName       = "Proxy"
)

type App struct {
	io          commands.IO
	start       time.Time // Time when the server started
	cfg         *AppConfig
	logger      *slog.Logger
	pathManager *pathManager
	// Contains all the deferred functions of the app.
	// Will be triggered on close for cleanup.
	deferred func()

	webHomePath   string
	paths         []string
	devNode       *gnodev.Node
	emitterServer *emitter.Server
	watcher       *watcher.PackageWatcher
	loader        packages.Loader
	book          *address.Book
	exportPath    string
	proxy         *proxy.PathInterceptor

	// XXX: move this
	exported uint
}

func runApp(cfg *AppConfig, cio commands.IO, dirs ...string) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var rt *rawterm.RawTerm
	var out io.Writer
	if cfg.interactive {
		var restore func() error
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
	defer app.Close()

	if rt != nil {
		go func() {
			app.RunInteractive(ctx, rt)
			cancel()
		}()
	}

	return app.RunServer(ctx, rt)
}

func NewApp(logger *slog.Logger, cfg *AppConfig, io commands.IO) *App {
	return &App{
		start:       time.Now(),
		deferred:    func() {},
		logger:      logger,
		cfg:         cfg,
		io:          io,
		pathManager: newPathManager(),
	}
}

func (ds *App) Defer(fn func()) {
	old := ds.deferred
	ds.deferred = func() {
		defer old()
		fn()
	}
}

func (ds *App) DeferClose(fn func() error) {
	ds.Defer(func() {
		if err := fn(); err != nil {
			ds.logger.Debug("close", "error", err.Error())
		}
	})
}

func (ds *App) Close() {
	ds.deferred()
}

func (ds *App) Setup(ctx context.Context, dirs ...string) (err error) {
	if err := ds.cfg.validateConfigFlags(); err != nil {
		return fmt.Errorf("validate error: %w", err)
	}

	loggerEvents := ds.logger.WithGroup(EventServerLogName)
	ds.emitterServer = emitter.NewServer(loggerEvents)

	// XXX: it would be nice to not have this hardcoded
	examplesDir := filepath.Join(ds.cfg.root, "examples")

	// Setup loader and resolver
	loaderLogger := ds.logger.WithGroup(LoaderLogName)
	resolver, localPaths := setupPackagesResolver(loaderLogger, ds.cfg, dirs...)
	ds.loader = packages.NewGlobLoader(examplesDir, resolver)

	// Get user's address book from local keybase
	accountLogger := ds.logger.WithGroup(AccountsLogName)
	ds.book, err = setupAddressBook(accountLogger, ds.cfg)
	if err != nil {
		return fmt.Errorf("unable to load keybase: %w", err)
	}

	// Generate user's paths using a comma as the delimiter
	qpaths := strings.Split(ds.cfg.paths, ",")

	// Set up the packages modifier and extract paths from queries
	// XXX: This should probably be moved into the setup node configuration
	modifiers, paths, err := resolvePackagesModifier(ds.cfg, ds.book, qpaths)
	if err != nil {
		return fmt.Errorf("unable to resolve paths %v: %w", paths, err)
	}

	// Add the user's paths to the pre-loaded paths
	// Modifiers will be added later to the node config bellow
	ds.paths = append(paths, localPaths...)

	balances, err := generateBalances(ds.book, ds.cfg)
	if err != nil {
		return fmt.Errorf("unable to generate balances: %w", err)
	}
	ds.logger.Debug("balances loaded", "list", balances.List())

	nodeLogger := ds.logger.WithGroup(NodeLogName)
	nodeCfg, err := setupDevNodeConfig(ds.cfg, nodeLogger, ds.emitterServer, balances, ds.loader, ds.book)
	if err != nil {
		return fmt.Errorf("unable to setup node config: %w", err)
	}
	nodeCfg.PackagesModifier = modifiers // add modifiers

	address := resolveUnixOrTCPAddr(nodeCfg.TMConfig.RPC.ListenAddress)

	// Setup lazy proxy
	if ds.cfg.lazyLoader {
		proxyLogger := ds.logger.WithGroup(ProxyLogName)
		ds.proxy, err = proxy.NewPathInterceptor(proxyLogger, address)
		if err != nil {
			return fmt.Errorf("unable to setup proxy: %w", err)
		}
		ds.DeferClose(ds.proxy.Close)

		// Override current rpc listener
		nodeCfg.TMConfig.RPC.ListenAddress = ds.proxy.ProxyAddress()
		proxyLogger.Debug("proxy started",
			"proxy_addr", ds.proxy.ProxyAddress(),
			"target_addr", ds.proxy.TargetAddress(),
		)

		proxyLogger.Info("lazy loading is enabled. packages will be loaded only upon a request via a query or transaction.", "loader", ds.loader.Name())
	} else {
		nodeCfg.TMConfig.RPC.ListenAddress = fmt.Sprintf("%s://%s", address.Network(), address.String())
	}

	ds.devNode, err = setupDevNode(ctx, ds.cfg, nodeCfg, ds.paths...)
	if err != nil {
		return err
	}
	ds.DeferClose(ds.devNode.Close)

	// Setup default web home realm, fallback on first local path
	devNodePaths := ds.devNode.Paths()

	switch webHome := ds.cfg.webHome; webHome {
	case "":
		if len(devNodePaths) > 0 {
			ds.webHomePath = strings.TrimPrefix(devNodePaths[0], ds.cfg.chainDomain)
			ds.logger.WithGroup(WebLogName).Info("using default package", "path", devNodePaths[0])
		}
	case "/", ":none:": // skip
	default:
		ds.webHomePath = webHome
	}

	if !ds.cfg.noWatch {
		ds.watcher, err = watcher.NewPackageWatcher(loggerEvents, ds.emitterServer)
		if err != nil {
			return fmt.Errorf("unable to setup packages watcher: %w", err)
		}

		ds.watcher.UpdatePackagesWatch(ds.devNode.ListPkgs()...)
	} else {
		ds.logger.WithGroup("Watcher").Info("watcher disabled (--no-watch)")
	}

	return nil
}

func (ds *App) setupHandlers(ctx context.Context) (http.Handler, error) {
	mux := http.NewServeMux()
	remote := ds.devNode.GetRemoteAddress()

	if ds.proxy != nil {
		proxyLogger := ds.logger.WithGroup(ProxyLogName)
		remote = ds.proxy.TargetAddress() // update remote address with proxy target address

		// Generate initial paths
		initPaths := map[string]struct{}{}
		for _, pkg := range ds.devNode.ListPkgs() {
			initPaths[pkg.Path] = struct{}{}
		}

		ds.proxy.HandlePath(func(paths ...string) {
			newPath := false
			for _, path := range paths {
				// Check if the path is an initial path.
				if _, ok := initPaths[path]; ok {
					continue
				}

				// Try to resolve the path first.
				// If we are unable to resolve it, ignore and continue

				if _, err := ds.loader.Resolve(path); err != nil {
					proxyLogger.Debug("unable to resolve path",
						"error", err,
						"path", path)
					continue
				}

				// If we already know this path, continue.
				if exist := ds.pathManager.Save(path); exist {
					continue
				}

				proxyLogger.Info("new monitored path",
					"path", path)

				newPath = true
			}

			if !newPath {
				return
			}

			ds.emitterServer.LockEmit()
			defer ds.emitterServer.UnlockEmit()

			ds.devNode.SetPackagePaths(ds.paths...)
			ds.devNode.AddPackagePaths(ds.pathManager.List()...)

			// Check if the node needs to be reloaded
			// XXX: This part can likely be optimized if we believe
			// it significantly impacts performance.
			for _, path := range paths {
				if ds.devNode.HasPackageLoaded(path) {
					continue
				}

				ds.logger.WithGroup(NodeLogName).Debug("some paths aren't loaded yet", "path", path)

				// If the package isn't loaded, attempt to reload the node
				if err := ds.devNode.Reload(ctx); err != nil {
					ds.logger.WithGroup(NodeLogName).Error("unable to reload node", "err", err)
				}

				// Update the watcher list with the currently loaded packages
				if !ds.cfg.noWatch {
					ds.watcher.UpdatePackagesWatch(ds.devNode.ListPkgs()...)
				}

				// Reloading the node once is sufficient, so exit the loop
				return
			}

			ds.logger.WithGroup(NodeLogName).Debug("paths already loaded, skipping reload", "paths", paths)
		})
	}

	// Setup gnoweb
	webhandler, err := setupGnoWebServer(ds.logger.WithGroup(WebLogName), ds.cfg, remote)
	if err != nil {
		return nil, fmt.Errorf("unable to setup gnoweb server: %w", err)
	}

	if ds.webHomePath != "" {
		serveWeb := webhandler.ServeHTTP
		webhandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "" || r.URL.Path == "/" {
				http.Redirect(w, r, ds.webHomePath, http.StatusFound)
			} else {
				serveWeb(w, r)
			}
		})
	}

	// Setup unsafe API
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

		mux.HandleFunc("/next_tx", func(res http.ResponseWriter, req *http.Request) {
			if err := ds.devNode.MoveToNextTX(req.Context()); err != nil {
				ds.logger.Error("failed to move forward", slog.Any("err", err))
				res.WriteHeader(http.StatusInternalServerError)
			}
		})

		mux.HandleFunc("/prev_tx", func(res http.ResponseWriter, req *http.Request) {
			if err := ds.devNode.MoveToPreviousTX(req.Context()); err != nil {
				ds.logger.Error("failed to move backward", slog.Any("err", err))
				res.WriteHeader(http.StatusInternalServerError)
			}
		})

		mux.HandleFunc("/list_accounts", func(res http.ResponseWriter, req *http.Request) {
			if jsonBytes, err := marshalJSONAccounts(req.Context(), ds.book); err != nil {
				ds.logger.Error("failed to list accounts", slog.Any("err", err))
				res.WriteHeader(http.StatusInternalServerError)
			} else {
				res.Header().Set("Content-Type", "application/json")
				res.Write(jsonBytes)
			}
		})
	}

	// Setup scripts to inject into the web pages
	var scripts [][]byte

	// Reload script
	if !ds.cfg.noWatch {
		evtstarget := fmt.Sprintf("%s/_events", ds.cfg.webListenerAddr)
		mux.Handle("/_events", ds.emitterServer)

		reloadScript, err := emitter.GenerateReloadScript(evtstarget)
		if err != nil {
			return nil, fmt.Errorf("unable to generate reload script: %w", err)
		}

		scripts = append(scripts, reloadScript)
	}

	// Custom script
	if ds.cfg.webCustomJS != "" {
		customJS, err := os.ReadFile(ds.cfg.webCustomJS)
		if err != nil {
			return nil, fmt.Errorf("unable to read custom JS file %q: %w", ds.cfg.webCustomJS, err)
		}

		var customScript bytes.Buffer

		// Prepend gnodev infos to the custom JS script.
		customScript.WriteString("const gnodev = {\n")
		customScript.WriteString(fmt.Sprintf("  rpcListenerAddr: '%s',\n", ds.cfg.nodeP2PListenerAddr))
		customScript.WriteString(fmt.Sprintf("  webListenerAddr: '%s',\n", ds.cfg.webListenerAddr))
		customScript.WriteString(fmt.Sprintf("  webHomePath: '%s',\n", ds.webHomePath))
		customScript.WriteString(fmt.Sprintf("  chainID: '%s',\n", ds.cfg.chainId))
		customScript.WriteString("};\n\n")

		customScript.Write(customJS)
		scripts = append(scripts, customScript.Bytes())
	}

	if len(scripts) > 0 {
		mux.Handle("/", middleware.NewInjectorMiddleware(scripts, webhandler))
	} else {
		mux.Handle("/", webhandler)
	}

	return mux, nil
}

func (ds *App) RunServer(ctx context.Context, term *rawterm.RawTerm) error {
	ctx, cancelWith := context.WithCancelCause(ctx)
	defer cancelWith(nil)

	addr := ds.cfg.webListenerAddr
	handlers, err := ds.setupHandlers(ctx)
	if err != nil {
		return fmt.Errorf("unable to setup handlers: %w", err)
	}

	server := &http.Server{
		Handler:           handlers,
		Addr:              addr,
		ReadHeaderTimeout: 60 * time.Second,
	}

	// Serve gnoweb
	if !ds.cfg.noWeb {
		go func() {
			err := server.ListenAndServe()
			cancelWith(err)
		}()

		ds.logger.WithGroup(WebLogName).Info("gnoweb started",
			"lisn", fmt.Sprintf("http://%s", addr))
	}

	if ds.cfg.interactive {
		ds.logger.WithGroup("--- READY").Info("for commands and help, press `h`", "took", time.Since(ds.start))
	} else {
		ds.logger.Info("node is ready", "took", time.Since(ds.start))
	}

	if ds.cfg.noWatch {
		<-ctx.Done()
		return context.Cause(ctx)
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
	ds.logger.WithGroup(KeyPressLogName).Debug("starting interactive mode")
	var keyPressCh <-chan rawterm.KeyPress
	if ds.cfg.interactive {
		keyPressCh = listenForKeyPress(ds.logger.WithGroup(KeyPressLogName), term)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case key, ok := <-keyPressCh:
			ds.logger.WithGroup(KeyPressLogName).Debug("pressed", "key", key.String())
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
https://docs.gno.land/builders/local-dev-with-gnodev

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

	switch key.Upper() {
	case rawterm.KeyH: // Helper
		ds.logger.Info("Gno Dev Helper", "helper", helper)

	case rawterm.KeyA: // Accounts
		logAccounts(ctx, ds.logger.WithGroup(AccountsLogName), ds.book, ds.devNode)

	case rawterm.KeyR: // Reload
		ds.logger.WithGroup(NodeLogName).Info("reloading...")
		if err = ds.devNode.ReloadAll(ctx); err != nil {
			ds.logger.WithGroup(NodeLogName).Error("unable to reload node", "err", err)
		}

	case rawterm.KeyCtrlR: // Reset
		ds.logger.WithGroup(NodeLogName).Info("resetting node state...")
		// Reset paths
		ds.pathManager.Reset()
		ds.devNode.SetPackagePaths(ds.paths...)
		// Reset the node
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

// XXX: packages modifier does not support glob yet
func resolvePackagesModifier(cfg *AppConfig, bk *address.Book, qpaths []string) ([]gnodev.QueryPath, []string, error) {
	modifiers := make([]gnodev.QueryPath, 0, len(qpaths))
	paths := make([]string, 0, len(qpaths))

	for _, path := range qpaths {
		if path == "" {
			continue
		}

		qpath, err := gnodev.ResolveQueryPath(bk, path)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid package path/query %q: %w", path, err)
		}

		modifiers = append(modifiers, qpath)
		paths = append(paths, qpath.Path)
	}

	return slices.Clip(modifiers), slices.Clip(paths), nil
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
