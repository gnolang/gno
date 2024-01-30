package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

const (
	onValue  = "on"
	offValue = "off"
)

var errInvalidToggleValue = errors.New("invalid toggle value")

type configBaseCfg struct {
	commonEditCfg

	rootDir                 string
	proxyApp                string
	moniker                 string
	dbBackend               string
	dbPath                  string
	genesis                 string
	privValidatorKey        string
	privValidatorState      string
	privValidatorListenAddr string
	nodeKey                 string
	abci                    string
	profListenAddress       string
	fastSyncMode            string
	filterPeers             string
}

// newConfigBaseCmd creates the new config base command
func newConfigBaseCmd(io commands.IO) *commands.Command {
	cfg := &configBaseCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "base",
			ShortUsage: "config base [flags]",
			ShortHelp:  "Edits the Gno node's base configuration",
			LongHelp:   "Edits the Gno node's base configuration locally",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execConfigBase(cfg, io)
		},
	)

	return cmd
}

func (c *configBaseCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonEditCfg.RegisterFlags(fs)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"the root directory for all data",
	)

	fs.StringVar(
		&c.proxyApp,
		"proxy-app",
		"",
		"the TCP or UNIX socket address of the ABCI application (or name)",
	)

	fs.StringVar(
		&c.moniker,
		"moniker",
		"",
		"a custom human readable name for this node",
	)

	fs.StringVar(
		&c.fastSyncMode,
		"fast-sync",
		onValue,
		fmt.Sprintf(
			"value indicating if the node should perform 'fast sync' upon startup: %s | %s",
			onValue,
			offValue,
		),
	)

	fs.StringVar(
		&c.dbBackend,
		"db-backend",
		"",
		fmt.Sprintf(
			"the database backend: %s | %s | %s",
			config.LevelDBName,
			config.ClevelDBName,
			config.BoltDBName,
		),
	)

	fs.StringVar(
		&c.dbPath,
		"db-path",
		"",
		"the database directory",
	)

	fs.StringVar(
		&c.genesis,
		"genesis-file",
		"",
		"the path to the genesis.json",
	)

	fs.StringVar(
		&c.privValidatorKey,
		"validator-key-file",
		"",
		"the path to the validator's private key",
	)

	fs.StringVar(
		&c.privValidatorState,
		"validator-state-file",
		"",
		"the path to the last validator's sign state",
	)

	fs.StringVar(
		&c.privValidatorListenAddr,
		"validator-laddr",
		"",
		"the TCP or UNIX socket address for Tendermint to listen on for",
	)

	fs.StringVar(
		&c.nodeKey,
		"node-key",
		"",
		"the path to the validator's P2P key",
	)

	fs.StringVar(
		&c.abci,
		"abci",
		"",
		fmt.Sprintf(
			"the mechanism to connect to the ABCI application: %s | %s",
			config.LocalABCI,
			config.SocketABCI,
		),
	)

	fs.StringVar(
		&c.profListenAddress,
		"prof-laddr",
		"",
		"the TCP or UNIX socket address for the profiling server to listen on",
	)

	fs.StringVar(
		&c.filterPeers,
		"filter-peers",
		offValue,
		fmt.Sprintf(
			"value indicating if new peers should be filtered through the ABCI app: %s | %s",
			onValue,
			offValue,
		),
	)
}

func execConfigBase(cfg *configBaseCfg, io commands.IO) error {
	// Load the config
	loadedCfg, err := config.LoadConfigFile(cfg.configPath)
	if err != nil {
		return fmt.Errorf("unable to load config, %w", err)
	}

	// Set the root dir, if any
	if cfg.rootDir != "" {
		loadedCfg.RootDir = cfg.rootDir
	}

	// Set the proxy app, if any
	if cfg.proxyApp != "" {
		loadedCfg.ProxyApp = cfg.proxyApp
	}

	// Set the moniker, if any
	if cfg.moniker != "" {
		loadedCfg.Moniker = cfg.moniker
	}

	// Set the db backend, if any
	if cfg.dbBackend != "" {
		loadedCfg.DBBackend = cfg.dbBackend
	}

	// Set the db path, if any
	if cfg.dbPath != "" {
		loadedCfg.DBPath = cfg.dbPath
	}

	// Set the genesis.json, if any
	if cfg.genesis != "" {
		loadedCfg.Genesis = cfg.genesis
	}

	// Set the validator key path, if any
	if cfg.privValidatorKey != "" {
		loadedCfg.PrivValidatorKey = cfg.privValidatorKey
	}

	// Set the validator state path, if any
	if cfg.privValidatorState != "" {
		loadedCfg.PrivValidatorState = cfg.privValidatorState
	}

	// Set the validator listen address, if any
	if cfg.privValidatorListenAddr != "" {
		loadedCfg.PrivValidatorListenAddr = cfg.privValidatorListenAddr
	}

	// Set the node p2p key path, if any
	if cfg.nodeKey != "" {
		loadedCfg.NodeKey = cfg.nodeKey
	}

	// Set the abci medium, if any
	if cfg.abci != "" {
		loadedCfg.ABCI = cfg.abci
	}

	// Set the profiling listen address, if any
	if cfg.profListenAddress != "" {
		loadedCfg.ProfListenAddress = cfg.profListenAddress
	}

	// Set the fast sync mode, if any
	syncToggleVal, err := parseToggleValue(cfg.fastSyncMode)
	if err != nil {
		return err
	}

	if syncToggleVal != loadedCfg.FastSyncMode {
		loadedCfg.FastSyncMode = syncToggleVal
	}

	// Set the filter peers flag, if any
	filterPeersToggleVal, err := parseToggleValue(cfg.filterPeers)
	if err != nil {
		return err
	}

	if filterPeersToggleVal != loadedCfg.FilterPeers {
		loadedCfg.FilterPeers = filterPeersToggleVal
	}

	// Make sure the config is now valid
	if err := loadedCfg.ValidateBasic(); err != nil {
		return fmt.Errorf("unable to validate config, %w", err)
	}

	// Save the config
	if err := config.WriteConfigFile(cfg.configPath, loadedCfg); err != nil {
		return fmt.Errorf("unable to save updated config, %w", err)
	}

	io.Printfln("Updated configuration saved at %s", cfg.configPath)

	return nil
}

// parseToggleValue parses the string toggle value into a bool.
// This method exists because the CLI package utilized in the project
// (ffcli) only supports bool values in the form of (is set -> true, not set -> false).
// Given that some default values are true, and false, it would be impossible to
// toggle them through the BoolVar option of the flag.FlagSet
func parseToggleValue(value string) (bool, error) {
	if value != onValue && value != offValue {
		return false, errInvalidToggleValue
	}

	if value == onValue {
		return true, nil
	}

	return value == onValue, nil
}
