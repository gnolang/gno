package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

const DefaultDomain = "gno.land"

var (
	DefaultDeployerName    = integration.DefaultAccount_Name
	DefaultDeployerAddress = crypto.MustAddressFromString(integration.DefaultAccount_Address)
	DefaultDeployerSeed    = integration.DefaultAccount_Seed
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
	webHome             string

	// Resolver
	resolvers varResolver

	// Node Configuration
	logFormat   string
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
		"list of additional resolvers, will be executed in the given order",
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

	return runApp(cfg, cio, dir)
}
