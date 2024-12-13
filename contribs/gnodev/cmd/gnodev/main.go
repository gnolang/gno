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
	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
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
	LoaderLogName      = "Loader"
)

var ErrConflictingFileArgs = errors.New("cannot specify `balances-file` or `txs-file` along with `genesis-file`")

var (
	DefaultDeployerName    = integration.DefaultAccount_Name
	DefaultDeployerAddress = crypto.MustAddressFromString(integration.DefaultAccount_Address)
	DefaultDeployerSeed    = integration.DefaultAccount_Seed
)

type devCfg struct {
	chdir string

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
	webWithHTML         bool

	// Resolver
	resolvers varResolver

	// Node Configuration
	minimal     bool
	verbose     bool
	noWatch     bool
	noReplay    bool
	maxGas      int64
	chainId     string
	chainDomain string
	unsafeAPI   bool
	interactive bool
	loadPath    string
}

var defaultDevOptions = devCfg{
	chainId:             "dev",
	chainDomain:         "gno.land",
	maxGas:              10_000_000_000,
	webListenerAddr:     "127.0.0.1:8888",
	nodeRPCListenerAddr: "127.0.0.1:26657",
	deployKey:           DefaultDeployerAddress.String(),
	home:                gnoenv.HomeDir(),
	root:                gnoenv.RootDir(),
	interactive:         true,
	unsafeAPI:           true,

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

	cmd.AddSubCommands(NewStagingCmd(stdio))

	cmd.Execute(context.Background(), os.Args[1:])
}

func (c *devCfg) RegisterFlags(fs *flag.FlagSet) {
	c.registerFlagsWithDefault(defaultDevOptions, fs)
}

func (c *devCfg) registerFlagsWithDefault(defaultCfg devCfg, fs *flag.FlagSet) {
	fs.StringVar(
		&c.home,
		"home",
		defaultCfg.home,
		"user's local directory for keys",
	)

	fs.BoolVar(
		&c.interactive,
		"interactive",
		defaultCfg.interactive,
		"enable gnodev interactive mode",
	)

	fs.StringVar(
		&c.chdir,
		"chdir",
		defaultCfg.chdir,
		"change directory context",
	)

	fs.StringVar(
		&c.root,
		"root",
		defaultCfg.root,
		"gno root directory",
	)

	fs.StringVar(
		&c.webListenerAddr,
		"web-listener",
		defaultDevOptions.webListenerAddr,
		"gnoweb: web server listener address",
	)

	fs.StringVar(
		&c.webRemoteHelperAddr,
		"web-help-remote",
		defaultDevOptions.webRemoteHelperAddr,
		"gnoweb: web server help page's remote addr (default to <node-rpc-listener>)",
	)

	fs.BoolVar(
		&c.webWithHTML,
		"web-with-html",
		defaultDevOptions.webWithHTML,
		"gnoweb: enable HTML parsing in markdown rendering",
	)

	fs.Var(
		&c.resolvers,
		"resolver",
		"list of addtional resolvers, will be exectued in the given order",
	)

	fs.StringVar(
		&c.nodeRPCListenerAddr,
		"node-rpc-listener",
		defaultCfg.nodeRPCListenerAddr,
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
		defaultCfg.balancesFile,
		"load the provided balance file (refer to the documentation for format)",
	)

	fs.StringVar(
		&c.balancesFile,
		"load-path",
		defaultCfg.balancesFile,
		"load given dir (glob supported)",
	)

	fs.StringVar(
		&c.txsFile,
		"txs-file",
		defaultCfg.txsFile,
		"load the provided transactions file (refer to the documentation for format)",
	)

	fs.StringVar(
		&c.genesisFile,
		"genesis",
		defaultCfg.genesisFile,
		"load the given genesis file",
	)

	fs.StringVar(
		&c.deployKey,
		"deploy-key",
		defaultCfg.deployKey,
		"default key name or Bech32 address for deploying packages",
	)

	fs.StringVar(
		&c.chainId,
		"chain-id",
		defaultCfg.chainId,
		"set node ChainID",
	)

	fs.StringVar(
		&c.chainDomain,
		"chain-domain",
		defaultDevOptions.chainDomain,
		"set node ChainDomain",
	)

	fs.BoolVar(
		&c.noWatch,
		"no-watch",
		defaultCfg.noWatch,
		"do not watch for file changes",
	)

	fs.BoolVar(
		&c.noReplay,
		"no-replay",
		defaultCfg.noReplay,
		"do not replay previous transactions upon reload",
	)

	fs.Int64Var(
		&c.maxGas,
		"max-gas",
		defaultCfg.maxGas,
		"set the maximum gas per block",
	)

	fs.BoolVar(
		&c.unsafeAPI,
		"unsafe-api",
		defaultCfg.unsafeAPI,
		"enable /reset and /reload endpoints which are not safe to expose publicly",
	)

	// Short flags
	fs.BoolVar(
		&c.verbose,
		"v",
		defaultCfg.verbose,
		"enable verbose output for development",
	)
}

func (c *devCfg) validateConfigFlags() error {
	if (c.balancesFile != "" || c.txsFile != "") && c.genesisFile != "" {
		return ErrConflictingFileArgs
	}

	return nil
}

type App struct {
	ctx    context.Context
	cfg    *devCfg
	io     commands.IO
	logger *slog.Logger

	devNode       *gnodev.Node
	server        *http.Server
	emitterServer *emitter.Server
	watcher       *watcher.PackageWatcher
	book          *address.Book
	exportPath    string

	// XXX: move this
	exported uint
}

func NewApp(ctx context.Context, logger *slog.Logger, cfg *devCfg, io commands.IO) *App {
	return &App{
		ctx:    ctx,
		logger: logger,
		cfg:    cfg,
		io:     io,
	}
}

func execDev(cfg *devCfg, args []string, io commands.IO) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	rt, restore, err := setupRawTerm(cfg, io)
	if err != nil {
		return fmt.Errorf("unable to init raw term: %w", err)
	}
	defer restore()

	// Setup trap signal
	osm.TrapSignal(func() {
		cancel()
		restore()
	})

	logger := setuplogger(cfg, rt)
	devServer := NewApp(ctx, logger, cfg, io)
	if err := devServer.Setup(); err != nil {
		return err
	}

	return devServer.Run(rt)
}

func (ds *App) Setup() error {
	if err := ds.cfg.validateConfigFlags(); err != nil {
		return fmt.Errorf("validate error: %w", err)
	}

	if ds.cfg.chdir != "" {
		if err := os.Chdir(ds.cfg.chdir); err != nil {
			return fmt.Errorf("unable to change directory: %w", err)
		}
	}

	loggerEvents := ds.logger.WithGroup(EventServerLogName)
	ds.emitterServer = emitter.NewServer(loggerEvents)

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to guess current dir: %w", err)
	}

	path, ok := guessPath(ds.cfg, dir)
	if !ok {
		return fmt.Errorf("unable to guess path from %q", dir)
	}

	resolver := setupPackagesResolver(ds.logger.WithGroup(LoaderLogName), ds.cfg, path, dir)
	loader := packages.NewResolverLoader(resolver)

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
	ds.devNode, err = setupDevNode(ds.ctx, ds.cfg, nodeCfg, path)
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

func (ds *App) Run(term *rawterm.RawTerm) error {
	ctx, cancelWith := context.WithCancelCause(ds.ctx)
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
	}

	keyPressCh := listenForKeyPress(ds.logger.WithGroup(KeyPressLogName), term)
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case _, ok := <-ds.watcher.PackagesUpdate:
			if !ok {
				return nil
			}
			ds.logger.WithGroup(NodeLogName).Info("reloading...")
			if err := ds.devNode.Reload(ds.ctx); err != nil {
				ds.logger.WithGroup(NodeLogName).Error("unable to reload node", "err", err)
			}
			ds.watcher.UpdatePackagesWatch(ds.devNode.ListPkgs()...)

		case key, ok := <-keyPressCh:
			if !ok {
				return nil
			}

			if key == rawterm.KeyCtrlC {
				cancelWith(nil)
				continue
			}

			ds.handleKeyPress(key)
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

func (ds *App) handleKeyPress(key rawterm.KeyPress) {
	var err error
	ds.logger.WithGroup(KeyPressLogName).Debug(fmt.Sprintf("<%s>", key.String()))

	switch key.Upper() {
	case rawterm.KeyH: // Helper
		ds.logger.Info("Gno Dev Helper", "helper", helper)

	case rawterm.KeyA: // Accounts
		logAccounts(ds.logger.WithGroup(AccountsLogName), ds.book, ds.devNode)

	case rawterm.KeyR: // Reload
		ds.logger.WithGroup(NodeLogName).Info("reloading...")
		if err = ds.devNode.ReloadAll(ds.ctx); err != nil {
			ds.logger.WithGroup(NodeLogName).Error("unable to reload node", "err", err)
		}

	case rawterm.KeyCtrlR: // Reset
		ds.logger.WithGroup(NodeLogName).Info("reseting node state...")
		if err = ds.devNode.Reset(ds.ctx); err != nil {
			ds.logger.WithGroup(NodeLogName).Error("unable to reset node state", "err", err)
		}

	case rawterm.KeyCtrlS: // Save
		ds.logger.WithGroup(NodeLogName).Info("saving state...")
		if err := ds.devNode.SaveCurrentState(ds.ctx); err != nil {
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
		doc, err := ds.devNode.ExportStateAsGenesis(ds.ctx)
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
		if err := ds.devNode.MoveToNextTX(ds.ctx); err != nil {
			ds.logger.WithGroup(NodeLogName).Error("unable to move forward", "err", err)
		}

	case rawterm.KeyP: // Previous tx
		ds.logger.Info("moving backward...")
		if err := ds.devNode.MoveToPreviousTX(ds.ctx); err != nil {
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

func resolvePackagesPathFromArgs(cfg *devCfg, bk *address.Book, args []string) ([]gnodev.PackageModifier, error) {
	modifiers := make([]gnodev.PackageModifier, 0, len(args))

	if cfg.deployKey == "" {
		return nil, fmt.Errorf("default deploy key cannot be empty")
	}

	defaultKey, _, ok := bk.GetFromNameOrAddress(cfg.deployKey)
	if !ok {
		return nil, fmt.Errorf("unable to get deploy key %q", cfg.deployKey)
	}

	if len(args) == 0 {
		args = append(args, ".") // add current dir if none are provided
	}

	for _, arg := range args {
		path, err := gnodev.ResolvePackageModifierQuery(bk, arg)
		if err != nil {
			return nil, fmt.Errorf("invalid package path/query %q: %w", arg, err)
		}

		// Assign a default creator if user haven't specified it.
		if path.Creator.IsZero() {
			path.Creator = defaultKey
		}

		modifiers = append(modifiers, path)
	}

	// Add examples folder if minimal is set to false
	if cfg.minimal {
		modifiers = append(modifiers, gnodev.PackageModifier{
			Path:    filepath.Join(cfg.root, "examples"),
			Creator: defaultKey,
			Deposit: nil,
		})
	}

	return modifiers, nil
}
