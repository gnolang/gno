package main

import (
	"flag"
	"fmt"
)

// LoadMode defines the package loading strategy
type LoadMode string

const (
	// LoadModeAuto pre-loads current workspace/package only (not examples).
	// If running from examples folder, behaves like lazy mode.
	LoadModeAuto LoadMode = "auto"
	// LoadModeLazy loads packages on-demand as they're accessed.
	LoadModeLazy LoadMode = "lazy"
	// LoadModeFull pre-loads all discovered packages.
	LoadModeFull LoadMode = "full"
)

func (m LoadMode) String() string { return string(m) }

func (m *LoadMode) Set(s string) error {
	switch LoadMode(s) {
	case LoadModeAuto, LoadModeLazy, LoadModeFull:
		*m = LoadMode(s)
		return nil
	default:
		return fmt.Errorf("invalid load mode %q: must be auto, lazy, or full", s)
	}
}

// varResolver is a placeholder for the deprecated resolver flag.
// The new NativeLoader handles package resolution automatically.
type varResolver []string

func (va varResolver) String() string {
	return fmt.Sprintf("%v", []string(va))
}

func (va *varResolver) Set(value string) error {
	*va = append(*va, value)
	return nil
}

type AppConfig struct {
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
	logFormat           string
	loadMode            LoadMode
	verbose             bool
	noWatch             bool
	noReplay            bool
	maxGas              int64
	chainId             string
	chainDomain         string
	unsafeAPI           bool
	interactive         bool
	paths               string
	emptyBlocks         bool
	emptyBlocksInterval int64
}

func (c *AppConfig) RegisterFlagsWith(fs *flag.FlagSet, defaultCfg AppConfig) {
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
		defaultLocalAppConfig.noWeb,
		"disable gnoweb",
	)

	fs.BoolVar(
		&c.webHTML,
		"web-html",
		defaultLocalAppConfig.webHTML,
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
		"[DEPRECATED] this flag is ignored; package resolution is now handled automatically via gnomod.toml and gnowork.toml discovery",
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

	fs.Var(
		&c.loadMode,
		"load",
		"package loading mode: `auto` (pre-load current workspace/package), `lazy` (load on-demand), `full` (pre-load all discovered packages)",
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
		`additional paths to preload in the form of "gno.land/r/my/realm", separated by commas; glob is supported`,
	)

	fs.BoolVar(
		&c.emptyBlocks,
		"empty-blocks",
		defaultCfg.emptyBlocks,
		"enable creation of empty blocks (default: ~1s interval)",
	)

	fs.Int64Var(
		&c.emptyBlocksInterval,
		"empty-blocks-interval",
		defaultCfg.emptyBlocksInterval,
		"set the interval for creating empty blocks (in seconds)",
	)

	fs.BoolVar(
		&c.verbose,
		"v",
		defaultCfg.verbose,
		"enable verbose output for development",
	)
}

func (c *AppConfig) validateConfigFlags() error {
	if (c.balancesFile != "" || c.txsFile != "") && c.genesisFile != "" {
		return ErrConflictingFileArgs
	}

	return nil
}
