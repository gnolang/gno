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
	"github.com/gnolang/gno/gno.land/pkg/balances"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	gnolog "github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/muesli/termenv"
)

const (
	NodeLogName        = "Node"
	WebLogName         = "GnoWeb"
	KeyPressLogName    = "KeyPress"
	EventServerLogName = "Event"
)

var (
	DefaultCreatorName    = integration.DefaultAccount_Name
	DefaultCreatorAddress = crypto.MustAddressFromString(integration.DefaultAccount_Address)
	DefaultCreatorSeed    = integration.DefaultAccount_Seed
	DefaultFee            = std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
)

type devCfg struct {
	// Listeners
	webListenerAddr          string
	nodeRPCListenerAddr      string
	nodeP2PListenerAddr      string
	nodeProxyAppListenerAddr string

	// Users default
	genesisCreator  string
	home            string
	root            string
	additionalUsers varAccounts

	// Node Configuration
	minimal      bool
	verbose      bool
	hotreload    bool
	noWatch      bool
	noReplay     bool
	maxGas       int64
	chainId      string
	serverMode   bool
	balancesFile string
}

var defaultDevOptions = &devCfg{
	chainId:             "dev",
	maxGas:              10_000_000_000,
	webListenerAddr:     "127.0.0.1:8888",
	nodeRPCListenerAddr: "127.0.0.1:36657",
	genesisCreator:      DefaultCreatorAddress.String(),
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
		"web server listening address",
	)

	fs.StringVar(
		&c.nodeRPCListenerAddr,
		"node-rpc-listener",
		defaultDevOptions.nodeRPCListenerAddr,
		"listening address for GnoLand RPC node",
	)

	fs.Var(
		&c.additionalUsers,
		"add-user",
		"pre-add a user",
	)

	fs.StringVar(
		&c.genesisCreator,
		"genesis-creator",
		defaultDevOptions.genesisCreator,
		"name or bech32 address of the genesis creator",
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
		"verbose",
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
}

func execDev(cfg *devCfg, args []string, io commands.IO) (err error) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

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
	kb, err := setupKeybase(cfg, logger)
	if err != nil {
		return fmt.Errorf("unable to load keybase: %w", err)
	}

	// Check and Parse packages
	pkgpaths, err := resolvePackagesPathFromArgs(cfg, kb, args)
	if err != nil {
		return fmt.Errorf("unable to parse package paths: %w", err)
	}

	// Setup Dev Node
	// XXX: find a good way to export or display node logs
	nodeLogger := logger.WithGroup(NodeLogName)
	devNode, err := setupDevNode(ctx, nodeLogger, cfg, emitterServer, kb, pkgpaths)
	if err != nil {
		return err
	}
	defer devNode.Close()

	nodeLogger.Info("node started", "lisn", devNode.GetRemoteAddress(), "chainID", cfg.chainId)

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

	if !cfg.serverMode {
		logger.WithGroup("--- READY").Info("for commands and help, press `h`")
	}

	// Run the main event loop
	return runEventLoop(ctx, logger, kb, rt, devNode, watcher)
}

var helper string = `
A           Accounts - Display known accounts
H           Help - Display this message
R           Reload - Reload all packages to take change into account.
Ctrl+R      Reset - Reset application state.
Ctrl+C      Exit - Exit the application
`

func runEventLoop(
	ctx context.Context,
	logger *slog.Logger,
	kb keys.Keybase,
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
			case rawterm.KeyH: // Helper
				logger.Info("Gno Dev Helper", "helper", helper)
			case rawterm.KeyA: // Accounts
				logAccounts(logger.WithGroup("accounts"), kb, dnode)
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

			case rawterm.KeyCtrlC: // Exit
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
	kb keys.Keybase,
	pkgspath []dev.PackagePath,
) (*gnodev.Node, error) {
	balances, err := generateBalances(kb, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to generate balances: %w", err)
	}
	logger.Debug("balances loaded", "list", balances.List())

	// configure gnoland node
	config := gnodev.DefaultNodeConfig(cfg.root)
	config.BalancesList = balances.List()
	config.PackagesPathList = pkgspath
	config.TMConfig.RPC.ListenAddress = resolveUnixOrTCPAddr(cfg.nodeRPCListenerAddr)
	config.NoReplay = cfg.noReplay
	config.MaxGasPerBlock = cfg.maxGas
	config.ChainID = cfg.chainId

	// other listeners
	config.TMConfig.P2P.ListenAddress = defaultDevOptions.nodeP2PListenerAddr
	config.TMConfig.ProxyApp = defaultDevOptions.nodeProxyAppListenerAddr

	return gnodev.NewDevNode(ctx, logger, remitter, config)
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

func resolvePackagesPathFromArgs(cfg *devCfg, kb keys.Keybase, args []string) ([]dev.PackagePath, error) {
	paths := make([]dev.PackagePath, len(args))

	if cfg.genesisCreator == "" {
		return nil, fmt.Errorf("default genesis creator cannot be empty")
	}

	defaultKey, err := kb.GetByNameOrAddress(cfg.genesisCreator)
	if err != nil {
		return nil, fmt.Errorf("unable to get genesis creator %q: %w", cfg.genesisCreator, err)
	}

	for i, arg := range args {
		path, err := dev.ResolvePackagePathQuery(kb, arg)
		if err != nil {
			return nil, fmt.Errorf("invalid package path/query %q: %w", arg, err)
		}

		// Assign a default creator if user haven't specified it.
		if path.Creator.IsZero() {
			path.Creator = defaultKey.GetAddress()
		}

		paths[i] = path
	}

	// Add examples folder if minimal is set to false
	if !cfg.minimal {
		paths = append(paths, gnodev.PackagePath{
			Path:    filepath.Join(cfg.root, "examples"),
			Creator: defaultKey.GetAddress(),
			Deposit: nil,
		})
	}

	return paths, nil
}

func generateBalances(kb keys.Keybase, cfg *devCfg) (balances.Balances, error) {
	bls := balances.New()
	unlimitedFund := std.Coins{std.NewCoin("ugnot", 10e12)}

	keys, err := kb.List()
	if err != nil {
		return nil, fmt.Errorf("unable to list keys from keybase: %w", err)
	}

	// Automatically set every key from keybase to unlimited found (or pre
	// defined found if specified)
	for _, key := range keys {
		found := unlimitedFund
		if preDefinedFound, ok := cfg.additionalUsers[key.GetName()]; ok && preDefinedFound != nil {
			found = preDefinedFound
		}

		address := key.GetAddress()
		bls[address] = gnoland.Balance{Amount: found, Address: address}
	}

	if cfg.balancesFile == "" {
		return bls, nil
	}

	file, err := os.Open(cfg.balancesFile)
	if err != nil {
		return nil, fmt.Errorf("unable to open balance file %q: %w", cfg.balancesFile, err)
	}

	blsFile, err := balances.GetBalancesFromSheet(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read balances file %q: %w", cfg.balancesFile, err)
	}

	// Left merge keybase balance into loaded file balance
	blsFile.LeftMerge(bls)
	return blsFile, nil
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

// createAccount creates a new account with the given name and adds it to the keybase.
func createAccount(kb keys.Keybase, accountName string) (keys.Info, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, fmt.Errorf("error creating entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, fmt.Errorf("error generating mnemonic: %w", err)
	}

	return kb.CreateAccount(accountName, mnemonic, "", "", 0, 0)
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

func setupKeybase(cfg *devCfg, logger *slog.Logger) (keys.Keybase, error) {
	kb := keys.NewInMemory()
	if cfg.home != "" {
		// load home keybase into our inMemory keybase
		kbHome, err := keys.NewKeyBaseFromDir(cfg.home)
		if err != nil {
			return nil, fmt.Errorf("unable to load keybae from dir %q: %w", cfg.home, err)
		}

		keys, err := kbHome.List()
		if err != nil {
			return nil, fmt.Errorf("unable to list keys from keybase %q: %w", cfg.home, err)
		}

		for _, key := range keys {
			name := key.GetName()
			armor, err := kbHome.Export(key.GetName())
			if err != nil {
				return nil, fmt.Errorf("unable to export key %q: %w", name, err)
			}

			if err := kb.Import(name, armor); err != nil {
				return nil, fmt.Errorf("unable to import key %q: %w", name, err)
			}
		}
	}

	// Add additional users to our keybase
	for user := range cfg.additionalUsers {
		info, err := createAccount(kb, user)
		if err != nil {
			return nil, fmt.Errorf("unable to create user %q: %w", user, err)
		}

		logger.Info("additional user", "name", info.GetName(), "addr", info.GetAddress())
	}

	// Next, make sure that we have a default address to load packages
	var info keys.Info
	var err error

	info, err = kb.GetByNameOrAddress(cfg.genesisCreator)
	switch {
	case err == nil: // user already have a default user
	case keyerror.IsErrKeyNotFound(err):
		// if the key isn't found, create a default one
		creatorName := fmt.Sprintf("_default#%.10s", DefaultCreatorAddress.String())
		if ok, _ := kb.HasByName(creatorName); ok {
			// if a collision happen here, someone really want to not run.
			return nil, fmt.Errorf("unable to create creator account, delete %q from your keybase", creatorName)
		}

		info, err = kb.CreateAccount(creatorName, DefaultCreatorSeed, "", "", 0, 0)
		if err != nil {
			return nil, fmt.Errorf("unable to create default %q account: %w", DefaultCreatorName, err)
		}
	default:
		return nil, fmt.Errorf("unable to get address %q from keybase: %w", info.GetAddress(), err)
	}

	logger.Info("default creator", "name", info.GetName(), "addr", info.GetAddress())
	return kb, nil
}

func logAccounts(logger *slog.Logger, kb keys.Keybase, _ *dev.Node) error {
	keys, err := kb.List()
	if err != nil {
		return fmt.Errorf("unable to get keybase keys list: %w", err)
	}

	accounts := make([]string, len(keys))
	for i, key := range keys {
		if key.GetName() == "" {
			continue // skip empty key name
		}

		address := key.GetAddress()
		qres, err := client.NewLocal().ABCIQuery("auth/accounts/"+address.String(), []byte{})
		if err != nil {
			return fmt.Errorf("unable to querry account %q: %w", address.String(), err)
		}

		var qret struct{ BaseAccount std.BaseAccount }
		if err = amino.UnmarshalJSON(qres.Response.Data, &qret); err != nil {
			return fmt.Errorf("unable to unmarshal query response: %w", err)
		}

		// format name - (address) -> (coins) -> (acct-num) -> (seq)
		accounts[i] = fmt.Sprintf("%s: addr(%s) coins(%s) acct_num(%d)",
			key.GetName(),
			address.String(),
			qret.BaseAccount.GetCoins().String(),
			qret.BaseAccount.GetAccountNumber())
	}

	logger.Info("current accounts", "balances", strings.Join(accounts, "\n"))
	return nil
}

func setuplogger(cfg *devCfg, out io.Writer) *slog.Logger {
	level := slog.LevelInfo
	if cfg.verbose {
		level = slog.LevelDebug
	}

	if cfg.serverMode {
		zaplogger := logger.NewZapLogger(out, level)
		return gnolog.ZapLoggerToSlog(zaplogger)
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
