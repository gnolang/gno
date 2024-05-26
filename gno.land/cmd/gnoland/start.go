package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/telemetry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const defaultNodeDir = "gnoland-data"

var errMissingGenesis = errors.New("missing genesis.json")

var startGraphic = strings.ReplaceAll(`
                    __             __
  ___ ____  ___    / /__ ____  ___/ /
 / _ '/ _ \/ _ \_ / / _ '/ _ \/ _  /
 \_, /_//_/\___(_)_/\_,_/_//_/\_,_/
/___/
`, "'", "`")

type startCfg struct {
	gnoRootDir            string // TODO: remove as part of https://github.com/gnolang/gno/issues/1952
	skipFailingGenesisTxs bool   // TODO: remove as part of https://github.com/gnolang/gno/issues/1952
	genesisBalancesFile   string // TODO: remove as part of https://github.com/gnolang/gno/issues/1952
	genesisTxsFile        string // TODO: remove as part of https://github.com/gnolang/gno/issues/1952
	genesisRemote         string // TODO: remove as part of https://github.com/gnolang/gno/issues/1952
	genesisFile           string
	chainID               string
	dataDir               string
	genesisMaxVMCycles    int64
	config                string
	lazyInit              bool

	logLevel  string
	logFormat string
}

func newStartCmd(io commands.IO) *commands.Command {
	cfg := &startCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "start",
			ShortUsage: "start [flags]",
			ShortHelp:  "starts the Gnoland blockchain node",
			LongHelp:   "Starts the Gnoland blockchain node, with accompanying setup",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execStart(cfg, io)
		},
	)
}

func (c *startCfg) RegisterFlags(fs *flag.FlagSet) {
	gnoroot := gnoenv.RootDir()
	defaultGenesisBalancesFile := filepath.Join(gnoroot, "gno.land", "genesis", "genesis_balances.txt")
	defaultGenesisTxsFile := filepath.Join(gnoroot, "gno.land", "genesis", "genesis_txs.jsonl")

	fs.BoolVar(
		&c.skipFailingGenesisTxs,
		"skip-failing-genesis-txs",
		false,
		"don't panic when replaying invalid genesis txs",
	)

	fs.StringVar(
		&c.genesisBalancesFile,
		"genesis-balances-file",
		defaultGenesisBalancesFile,
		"initial distribution file",
	)

	fs.StringVar(
		&c.genesisTxsFile,
		"genesis-txs-file",
		defaultGenesisTxsFile,
		"initial txs to replay",
	)

	fs.StringVar(
		&c.genesisFile,
		"genesis",
		"genesis.json",
		"the path to the genesis.json",
	)

	fs.StringVar(
		&c.chainID,
		"chainid",
		"dev",
		"the ID of the chain",
	)

	fs.StringVar(
		&c.gnoRootDir,
		"gnoroot-dir",
		gnoroot,
		"the root directory of the gno repository",
	)

	fs.StringVar(
		&c.dataDir,
		"data-dir",
		defaultNodeDir,
		"the path to the node's data directory",
	)

	fs.StringVar(
		&c.genesisRemote,
		"genesis-remote",
		"localhost:26657",
		"replacement for '%%REMOTE%%' in genesis",
	)

	fs.Int64Var(
		&c.genesisMaxVMCycles,
		"genesis-max-vm-cycles",
		100_000_000,
		"set maximum allowed vm cycles per operation. Zero means no limit.",
	)

	fs.StringVar(
		&c.config,
		flagConfigFlag,
		"",
		"the flag config file (optional)",
	)

	fs.StringVar(
		&c.logLevel,
		"log-level",
		zapcore.DebugLevel.String(),
		"log level for the gnoland node,",
	)

	fs.StringVar(
		&c.logFormat,
		"log-format",
		log.ConsoleFormat.String(),
		"log format for the gnoland node",
	)

	fs.BoolVar(
		&c.lazyInit,
		"lazy",
		false,
		"flag indicating if lazy init is enabled. Generates the node secrets, configuration, and genesis.json",
	)
}

func execStart(c *startCfg, io commands.IO) error {
	// Get the absolute path to the node's data directory
	nodeDir, err := filepath.Abs(c.dataDir)
	if err != nil {
		return fmt.Errorf("unable to get absolute path for data directory, %w", err)
	}

	// Get the absolute path to the node's genesis.json
	genesisPath, err := filepath.Abs(c.genesisFile)
	if err != nil {
		return fmt.Errorf("unable to get absolute path for the genesis.json, %w", err)
	}

	// Initialize the logger
	zapLogger, err := initializeLogger(io.Out(), c.logLevel, c.logFormat)
	if err != nil {
		return fmt.Errorf("unable to initialize zap logger, %w", err)
	}

	// Wrap the zap logger
	logger := log.ZapLoggerToSlog(zapLogger)

	if c.lazyInit {
		if err := lazyInitNodeDir(io, nodeDir); err != nil {
			return fmt.Errorf("unable to lazy-init the node directory, %w", err)
		}
	}

	// Load the configuration
	cfg, err := config.LoadConfig(nodeDir)
	if err != nil {
		return fmt.Errorf("unable to load config, %w", err)
	}

	// Check if the genesis.json exists
	if !osm.FileExists(genesisPath) {
		if !c.lazyInit {
			return errMissingGenesis
		}

		// Load the private validator secrets
		privateKey := privval.LoadFilePV(
			cfg.PrivValidatorKeyFile(),
			cfg.PrivValidatorStateFile(),
		)

		// Init a new genesis.json
		if err := lazyInitGenesis(io, c, genesisPath, privateKey.GetPubKey()); err != nil {
			return fmt.Errorf("unable to initialize genesis.json, %w", err)
		}
	}

	// Initialize telemetry
	if err := telemetry.Init(*cfg.Telemetry); err != nil {
		return fmt.Errorf("unable to initialize telemetry, %w", err)
	}

	// Create application and node
	cfg.LocalApp, err = gnoland.NewApp(nodeDir, c.skipFailingGenesisTxs, logger, c.genesisMaxVMCycles)
	if err != nil {
		return fmt.Errorf("unable to create the Gnoland app, %w", err)
	}

	// Print the starting graphic
	if c.logFormat != string(log.JSONFormat) {
		io.Println(startGraphic)
	}

	// Create a default node, with the given setup
	gnoNode, err := node.DefaultNewNode(cfg, genesisPath, logger)
	if err != nil {
		return fmt.Errorf("unable to create the Gnoland node, %w", err)
	}

	// Start the node
	if err := gnoNode.Start(); err != nil {
		return fmt.Errorf("unable to start the Gnoland node, %w", err)
	}

	osm.TrapSignal(func() {
		if gnoNode.IsRunning() {
			if err := gnoNode.Stop(); err != nil {
				logger.Warn("unable to gracefully stop the Gnoland node", "err", err)
			}
		}

		// Sync the logger before exiting
		_ = zapLogger.Sync()
	})

	// Run forever
	select {}
}

// lazyInitNodeDir initializes new secrets, and a default configuration
// in the given node directory, if not present
func lazyInitNodeDir(io commands.IO, nodeDir string) error {
	var (
		configPath  = constructConfigPath(nodeDir)
		secretsPath = constructSecretsPath(nodeDir)
	)

	// Check if the configuration already exists
	if !osm.FileExists(configPath) {
		// Create the gnoland config options
		cfg := &configInitCfg{
			configCfg: configCfg{
				configPath: constructConfigPath(nodeDir),
			},
		}

		// Run gnoland config init
		if err := execConfigInit(cfg, io); err != nil {
			return fmt.Errorf("unable to initialize config, %w", err)
		}

		io.Printfln("WARN: Initialized default node config at %q", filepath.Dir(cfg.configPath))
		io.Println()
	}

	// Create the gnoland secrets options
	secrets := &secretsInitCfg{
		commonAllCfg: commonAllCfg{
			dataDir: secretsPath,
		},
		forceOverwrite: false, // existing secrets shouldn't be pruned
	}

	// Run gnoland secrets init
	err := execSecretsInit(secrets, []string{}, io)
	if err == nil {
		io.Printfln("WARN: Initialized default node secrets at %q", secrets.dataDir)

		return nil
	}

	// Check if the error is valid
	if errors.Is(err, errOverwriteNotEnabled) {
		// No new secrets were generated
		return nil
	}

	return fmt.Errorf("unable to initialize secrets, %w", err)
}

// lazyInitGenesis a new genesis.json file, with a signle validator
func lazyInitGenesis(
	io commands.IO,
	c *startCfg,
	genesisPath string,
	publicKey crypto.PubKey,
) error {
	// Check if the genesis.json is present
	if osm.FileExists(genesisPath) {
		return nil
	}

	// Generate the new genesis.json file
	if err := generateGenesisFile(genesisPath, publicKey, c); err != nil {
		return fmt.Errorf("unable to generate genesis file, %w", err)
	}

	io.Printfln("WARN: Initialized genesis.json at %q", genesisPath)

	return nil
}

// initializeLogger initializes the zap logger using the given format and log level,
// outputting to the given IO
func initializeLogger(io io.WriteCloser, logLevel, logFormat string) (*zap.Logger, error) {
	// Initialize the log level
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return nil, fmt.Errorf("unable to parse log level, %w", err)
	}

	// Initialize the log format
	format := log.Format(strings.ToLower(logFormat))

	// Initialize the zap logger
	return log.GetZapLoggerFn(format)(io, level), nil
}

func generateGenesisFile(genesisFile string, pk crypto.PubKey, c *startCfg) error {
	gen := &bft.GenesisDoc{}
	gen.GenesisTime = time.Now()
	gen.ChainID = c.chainID
	gen.ConsensusParams = abci.ConsensusParams{
		Block: &abci.BlockParams{
			// TODO: update limits.
			MaxTxBytes:   1_000_000,   // 1MB,
			MaxDataBytes: 2_000_000,   // 2MB,
			MaxGas:       100_000_000, // 100M gas
			TimeIotaMS:   100,         // 100ms
		},
	}

	gen.Validators = []bft.GenesisValidator{
		{
			Address: pk.Address(),
			PubKey:  pk,
			Power:   10,
			Name:    "testvalidator",
		},
	}

	// Load balances files
	balances, err := gnoland.LoadGenesisBalancesFile(c.genesisBalancesFile)
	if err != nil {
		return fmt.Errorf("unable to load genesis balances file %q: %w", c.genesisBalancesFile, err)
	}

	// Load examples folder
	examplesDir := filepath.Join(c.gnoRootDir, "examples")
	test1 := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	defaultFee := std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
	pkgsTxs, err := gnoland.LoadPackagesFromDir(examplesDir, test1, defaultFee, nil)
	if err != nil {
		return fmt.Errorf("unable to load examples folder: %w", err)
	}

	// Load Genesis TXs
	genesisTxs, err := gnoland.LoadGenesisTxsFile(c.genesisTxsFile, c.chainID, c.genesisRemote)
	if err != nil {
		return fmt.Errorf("unable to load genesis txs file: %w", err)
	}

	genesisTxs = append(pkgsTxs, genesisTxs...)

	// Construct genesis AppState.
	gen.AppState = gnoland.GnoGenesisState{
		Balances: balances,
		Txs:      genesisTxs,
	}

	// Write genesis state
	if err := gen.SaveAs(genesisFile); err != nil {
		return fmt.Errorf("unable to write genesis file %q: %w", genesisFile, err)
	}

	return nil
}
