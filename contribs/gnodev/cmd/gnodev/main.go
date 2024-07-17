package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/rawterm"
	"github.com/gnolang/gno/contribs/gnodev/pkg/watcher"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

const (
	NodeLogName        = "Node"
	WebLogName         = "GnoWeb"
	KeyPressLogName    = "KeyPress"
	EventServerLogName = "Event"
	AccountsLogName    = "Accounts"
)

var ErrConflictingFileArgs = errors.New("cannot specify `balances-file` or `txs-file` along with `genesis-file`")

var (
	DefaultDeployerName    = integration.DefaultAccount_Name
	DefaultDeployerAddress = crypto.MustAddressFromString(integration.DefaultAccount_Address)
	DefaultDeployerSeed    = integration.DefaultAccount_Seed
)

type devCfg struct {
	// Listeners
	nodeRPCListenerAddr      string
	nodeP2PListenerAddr      string
	nodeProxyAppListenerAddr string

	// Users default
	deployKey       string
	home            string
	root            string
	premineAccounts varPremineAccounts

	// Files
	balancesFile string
	genesisFile  string
	txsFile      string

	// Web Configuration
	webListenerAddr     string
	webRemoteHelperAddr string

	// Node Configuration
	minimal    bool
	verbose    bool
	noWatch    bool
	noReplay   bool
	maxGas     int64
	chainId    string
	serverMode bool
	unsafeAPI  bool
}

var defaultDevOptions = &devCfg{
	chainId:             "dev",
	maxGas:              10_000_000_000,
	webListenerAddr:     "127.0.0.1:8888",
	nodeRPCListenerAddr: "127.0.0.1:26657",
	deployKey:           DefaultDeployerAddress.String(),
	home:                gnoenv.HomeDir(),
	root:                gnoenv.RootDir(),

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
			LongHelp:   `The gnodev command starts an in-memory node and a gno.land web interface primarily for realm package development. It automatically loads the 'examples' directory and any additional specified paths.`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execDev(cfg, args, stdio)
		})

	cmd.Execute(context.Background(), os.Args[1:])
}

func (c *devCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.home,
		"home",
		defaultDevOptions.home,
		"user's local directory for keys",
	)

	fs.StringVar(
		&c.root,
		"root",
		defaultDevOptions.root,
		"gno root directory",
	)

	fs.StringVar(
		&c.webListenerAddr,
		"web-listener",
		defaultDevOptions.webListenerAddr,
		"web server listener address",
	)

	fs.StringVar(
		&c.webRemoteHelperAddr,
		"web-help-remote",
		defaultDevOptions.webRemoteHelperAddr,
		"web server help page's remote addr (default to <node-rpc-listener>)",
	)

	fs.StringVar(
		&c.nodeRPCListenerAddr,
		"node-rpc-listener",
		defaultDevOptions.nodeRPCListenerAddr,
		"listening address for GnoLand RPC node",
	)

	fs.Var(
		&c.premineAccounts,
		"add-account",
		"add (or set) a premine account in the form `<bech32|name>[=<amount>]`, can be used multiple time",
	)

	fs.StringVar(
		&c.balancesFile,
		"balance-file",
		defaultDevOptions.balancesFile,
		"load the provided balance file (refer to the documentation for format)",
	)

	fs.StringVar(
		&c.txsFile,
		"txs-file",
		defaultDevOptions.txsFile,
		"load the provided transactions file (refer to the documentation for format)",
	)

	fs.StringVar(
		&c.genesisFile,
		"genesis",
		defaultDevOptions.genesisFile,
		"load the given genesis file",
	)

	fs.StringVar(
		&c.deployKey,
		"deploy-key",
		defaultDevOptions.deployKey,
		"default key name or Bech32 address for deploying packages",
	)

	fs.BoolVar(
		&c.minimal,
		"minimal",
		defaultDevOptions.minimal,
		"do not load packages from the examples directory",
	)

	fs.BoolVar(
		&c.serverMode,
		"server-mode",
		defaultDevOptions.serverMode,
		"disable interaction, and adjust logging for server use.",
	)

	fs.BoolVar(
		&c.verbose,
		"v",
		defaultDevOptions.verbose,
		"enable verbose output for development",
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
		"do not watch for file changes",
	)

	fs.BoolVar(
		&c.noReplay,
		"no-replay",
		defaultDevOptions.noReplay,
		"do not replay previous transactions upon reload",
	)

	fs.Int64Var(
		&c.maxGas,
		"max-gas",
		defaultDevOptions.maxGas,
		"set the maximum gas per block",
	)

	fs.BoolVar(
		&c.unsafeAPI,
		"unsafe-api",
		defaultDevOptions.unsafeAPI,
		"enable /reset and /reload endpoints which are not safe to expose publicly",
	)
}

func (c *devCfg) validateConfigFlags() error {
	if (c.balancesFile != "" || c.txsFile != "") && c.genesisFile != "" {
		return ErrConflictingFileArgs
	}

	return nil
}

func execDev(cfg *devCfg, args []string, io commands.IO) (err error) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	if err := cfg.validateConfigFlags(); err != nil {
		return fmt.Errorf("validate error: %w", err)
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

	// load keybase
	book, err := setupAddressBook(logger.WithGroup(AccountsLogName), cfg)
	if err != nil {
		return fmt.Errorf("unable to load keybase: %w", err)
	}

	// Check and Parse packages
	pkgpaths, err := resolvePackagesPathFromArgs(cfg, book, args)
	if err != nil {
		return fmt.Errorf("unable to parse package paths: %w", err)
	}

	// generate balances
	balances, err := generateBalances(book, cfg)
	if err != nil {
		return fmt.Errorf("unable to generate balances: %w", err)
	}
	logger.Debug("balances loaded", "list", balances.List())

	// Setup Dev Node
	// XXX: find a good way to export or display node logs
	nodeLogger := logger.WithGroup(NodeLogName)
	nodeCfg := setupDevNodeConfig(cfg, logger, emitterServer, balances, pkgpaths)
	devNode, err := setupDevNode(ctx, cfg, nodeCfg)
	if err != nil {
		return err
	}
	defer devNode.Close()

	nodeLogger.Info("node started", "lisn", devNode.GetRemoteAddress(), "chainID", cfg.chainId)

	// Create server
	mux := http.NewServeMux()
	server := http.Server{
		Handler:           mux,
		Addr:              cfg.webListenerAddr,
		ReadHeaderTimeout: time.Second * 60,
	}
	defer server.Close()

	// Setup gnoweb
	webhandler := setupGnoWebServer(logger.WithGroup(WebLogName), cfg, devNode)

	// Setup unsafe APIs if enabled
	if cfg.unsafeAPI {
		mux.HandleFunc("/reset", func(res http.ResponseWriter, req *http.Request) {
			if err := devNode.Reset(req.Context()); err != nil {
				logger.Error("failed to reset", slog.Any("err", err))
				res.WriteHeader(http.StatusInternalServerError)
			}
		})

		mux.HandleFunc("/reload", func(res http.ResponseWriter, req *http.Request) {
			if err := devNode.Reload(req.Context()); err != nil {
				logger.Error("failed to reload", slog.Any("err", err))
				res.WriteHeader(http.StatusInternalServerError)
			}
		})
	}

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

	if !cfg.serverMode {
		logger.WithGroup("--- READY").Info("for commands and help, press `h`")
	}

	// Run the main event loop
	return runEventLoop(ctx, logger, book, rt, devNode, watcher)
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

func runEventLoop(
	ctx context.Context,
	logger *slog.Logger,
	bk *address.Book,
	rt *rawterm.RawTerm,
	dnode *gnodev.Node,
	watch *watcher.PackageWatcher,
) error {
	// XXX: move this in above, but we need to have a proper struct first
	// XXX: make this configurable
	var exported uint
	path, err := os.MkdirTemp("", "gnodev-export")
	if err != nil {
		return fmt.Errorf("unable to create `export` directory: %w", err)
	}

	defer func() {
		if exported == 0 {
			_ = os.RemoveAll(path)
		}
	}()

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
			case rawterm.KeyH: // Helper
				logger.Info("Gno Dev Helper", "helper", helper)

			case rawterm.KeyA: // Accounts
				logAccounts(logger.WithGroup(AccountsLogName), bk, dnode)

			case rawterm.KeyR: // Reload
				logger.WithGroup(NodeLogName).Info("reloading...")
				if err = dnode.ReloadAll(ctx); err != nil {
					logger.WithGroup(NodeLogName).
						Error("unable to reload node", "err", err)
				}

			case rawterm.KeyCtrlR: // Reset
				logger.WithGroup(NodeLogName).Info("reseting node state...")
				if err = dnode.Reset(ctx); err != nil {
					logger.WithGroup(NodeLogName).
						Error("unable to reset node state", "err", err)
				}

			case rawterm.KeyCtrlS: // Save
				logger.WithGroup(NodeLogName).Info("saving state...")
				if err := dnode.SaveCurrentState(ctx); err != nil {
					logger.WithGroup(NodeLogName).
						Error("unable to save node state", "err", err)
				}

			case rawterm.KeyE:
				logger.WithGroup(NodeLogName).Info("exporting state...")
				doc, err := dnode.ExportStateAsGenesis(ctx)
				if err != nil {
					logger.WithGroup(NodeLogName).
						Error("unable to export node state", "err", err)
					continue
				}

				docfile := filepath.Join(path, fmt.Sprintf("export_%d.jsonl", exported))
				if err := doc.SaveAs(docfile); err != nil {
					logger.WithGroup(NodeLogName).
						Error("unable to save genesis", "err", err)
				}
				exported++

				logger.WithGroup(NodeLogName).Info("node state exported", "file", docfile)

			case rawterm.KeyN: // Next tx
				logger.Info("moving forward...")
				if err := dnode.MoveToNextTX(ctx); err != nil {
					logger.WithGroup(NodeLogName).
						Error("unable to move forward", "err", err)
				}

			case rawterm.KeyP: // Next tx
				logger.Info("moving backward...")
				if err := dnode.MoveToPreviousTX(ctx); err != nil {
					logger.WithGroup(NodeLogName).
						Error("unable to move backward", "err", err)
				}

			case rawterm.KeyCtrlC: // Exit
				return nil

			default:
			}

			// Reset listen for the next keypress
			keyPressCh = listenForKeyPress(logger.WithGroup(KeyPressLogName), rt)
		}
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

func resolvePackagesPathFromArgs(cfg *devCfg, bk *address.Book, args []string) ([]gnodev.PackagePath, error) {
	paths := make([]gnodev.PackagePath, 0, len(args))

	if cfg.deployKey == "" {
		return nil, fmt.Errorf("default deploy key cannot be empty")
	}

	defaultKey, _, ok := bk.GetFromNameOrAddress(cfg.deployKey)
	if !ok {
		return nil, fmt.Errorf("unable to get deploy key %q", cfg.deployKey)
	}

	for _, arg := range args {
		path, err := gnodev.ResolvePackagePathQuery(bk, arg)
		if err != nil {
			return nil, fmt.Errorf("invalid package path/query %q: %w", arg, err)
		}

		// Assign a default creator if user haven't specified it.
		if path.Creator.IsZero() {
			path.Creator = defaultKey
		}

		paths = append(paths, path)
	}

	// Add examples folder if minimal is set to false
	if !cfg.minimal {
		paths = append(paths, gnodev.PackagePath{
			Path:    filepath.Join(cfg.root, "examples"),
			Creator: defaultKey,
			Deposit: nil,
		})
	}

	return paths, nil
}
