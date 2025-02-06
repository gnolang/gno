package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/mattn/go-isatty"
)

const DefaultDomain = "gno.land"

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

var ErrConflictingFileArgs = errors.New("cannot specify `balances-file` or `txs-file` along with `genesis-file`")

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
	noWeb               bool
	webHTML             bool
	webListenerAddr     string
	webRemoteHelperAddr string
	webWithHTML         bool
	webHome             string

	// Resolver
	resolvers varResolver

	// Node Configuration
	logFormat   string
	lazyLoader  bool
	verbose     bool
	noWatch     bool
	noReplay    bool
	maxGas      int64
	chainId     string
	chainDomain string
	unsafeAPI   bool
	interactive bool
	paths       string
}

var defaultDevOptions = devCfg{
	chainId:             "dev",
	logFormat:           "console",
	chainDomain:         DefaultDomain,
	maxGas:              10_000_000_000,
	webListenerAddr:     "127.0.0.1:8888",
	nodeRPCListenerAddr: "127.0.0.1:26657",
	deployKey:           defaultDeployerAddress.String(),
	home:                gnoenv.HomeDir(),
	root:                gnoenv.RootDir(),
	interactive:         isatty.IsTerminal(os.Stdout.Fd()),
	unsafeAPI:           true,
	lazyLoader:          true,

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
			ShortUsage: "gnodev [flags] ",
			ShortHelp:  "Runs an in-memory node and gno.land web server for development purposes.",
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
	*c = defaultCfg // Copy default config

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
		&c.root,
		"root",
		defaultCfg.root,
		"gno root directory",
	)

	fs.BoolVar(
		&c.noWeb,
		"no-web",
		defaultDevOptions.noWeb,
		"disable gnoweb",
	)

	fs.BoolVar(
		&c.webHTML,
		"web-html",
		defaultDevOptions.webHTML,
		"gnoweb: enable unsafe HTML parsing in markdown rendering",
	)

	fs.StringVar(
		&c.webListenerAddr,
		"web-listener",
		defaultCfg.webListenerAddr,
		"gnoweb: web server listener address",
	)

	fs.StringVar(
		&c.webRemoteHelperAddr,
		"web-help-remote",
		defaultCfg.webRemoteHelperAddr,
		"gnoweb: web server help page's remote addr (default to <node-rpc-listener>)",
	)

	fs.BoolVar(
		&c.webWithHTML,
		"web-with-html",
		defaultCfg.webWithHTML,
		"gnoweb: enable HTML parsing in markdown rendering",
	)

	fs.StringVar(
		&c.webHome,
		"web-home",
		defaultCfg.webHome,
		"gnoweb: set default home page, use `/` or `:none:` to use default web home redirect",
	)

	fs.Var(
		&c.resolvers,
		"resolver",
		"list of additional resolvers (`root`, `dir` or `remote`), will be executed in the given order",
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
		defaultCfg.chainDomain,
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

	fs.BoolVar(
		&c.lazyLoader,
		"lazy-loader",
		defaultCfg.lazyLoader,
		"enable lazy loader",
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

	fs.StringVar(
		&c.logFormat,
		"log-format",
		defaultCfg.logFormat,
		"log output format, can be `json` or `console`",
	)

	fs.StringVar(
		&c.paths,
		"paths",
		defaultCfg.paths,
		"additional path(s) to load, separated by comma",
	)

	// Short flags
	fs.StringVar(
		&c.chdir,
		"C",
		defaultCfg.chdir,
		"change directory context before running gnodev",
	)

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

func execDev(cfg *devCfg, args []string, cio commands.IO) error {
	if cfg.chdir != "" {
		if err := os.Chdir(cfg.chdir); err != nil {
			return fmt.Errorf("unable to change directory: %w", err)
		}
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to guess current dir: %w", err)
	}

	// If no resolvers is defined, use gno example as root resolver
	var baseResolvers []packages.Resolver
	if len(cfg.resolvers) == 0 {
		gnoroot, err := gnoenv.GuessRootDir()
		if err != nil {
			return err
		}

		exampleRoot := filepath.Join(gnoroot, "examples")
		baseResolvers = append(baseResolvers, packages.NewFSResolver(exampleRoot))
	}

	// Check if current directory is a valid gno package
	path := guessPath(cfg, dir)
	resolver := packages.NewLocalResolver(path, dir)
	if resolver.IsValid() {
		// Add current directory as local resolver
		baseResolvers = append(baseResolvers, resolver)
		if len(cfg.paths) > 0 {
			cfg.paths += ","
		}
		cfg.paths += resolver.Path
	}
	cfg.resolvers = append(baseResolvers, cfg.resolvers...)

	return runApp(cfg, cio) // else run app without any dir
}
