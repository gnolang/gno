package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
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
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/file"
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/null"
	eventstorecfg "github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
	"go.uber.org/zap/zapcore"
)

type startCfg struct {
	gnoRootDir            string
	skipFailingGenesisTxs bool
	skipStart             bool
	genesisBalancesFile   string
	genesisTxsFile        string
	chainID               string
	genesisRemote         string
	dataDir               string
	genesisMaxVMCycles    int64
	config                string

	txEventStoreType string
	txEventStorePath string
	nodeConfigPath   string

	logLevel  string
	logFormat string
}

func newStartCmd(io commands.IO) *commands.Command {
	cfg := &startCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "start",
			ShortUsage: "start [flags]",
			ShortHelp:  "run the full node",
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

	fs.BoolVar(
		&c.skipStart,
		"skip-start",
		false,
		"quit after initialization, don't start the node",
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

	// XXX: Use home directory for this
	fs.StringVar(
		&c.dataDir,
		"data-dir",
		"testdir",
		"directory for config and data",
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
		10_000_000,
		"set maximum allowed vm cycles per operation. Zero means no limit.",
	)

	fs.StringVar(
		&c.config,
		flagConfigFlag,
		"",
		"the flag config file (optional)",
	)

	fs.StringVar(
		&c.nodeConfigPath,
		"config-path",
		"",
		"the node TOML config file path (optional)",
	)

	fs.StringVar(
		&c.txEventStoreType,
		"tx-event-store-type",
		null.EventStoreType,
		fmt.Sprintf(
			"type of transaction event store [%s]",
			strings.Join(
				[]string{
					null.EventStoreType,
					file.EventStoreType,
				},
				", ",
			),
		),
	)

	fs.StringVar(
		&c.txEventStorePath,
		"tx-event-store-path",
		"",
		fmt.Sprintf("path for the file tx event store (required if event store is '%s')", file.EventStoreType),
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

	// XXX(deprecated): use data-dir instead
	fs.StringVar(
		&c.dataDir,
		"root-dir",
		"testdir",
		"deprecated: use data-dir instead - directory for config and data",
	)
}

func execStart(c *startCfg, io commands.IO) error {
	dataDir := c.dataDir

	var (
		cfg        *config.Config
		loadCfgErr error
	)

	// Set the node configuration
	if c.nodeConfigPath != "" {
		// Load the node configuration
		// from the specified path
		cfg, loadCfgErr = config.LoadConfigFile(c.nodeConfigPath)
	} else {
		// Load the default node configuration
		cfg, loadCfgErr = config.LoadOrMakeConfigWithOptions(dataDir)
	}

	if loadCfgErr != nil {
		return fmt.Errorf("unable to load node configuration, %w", loadCfgErr)
	}

	// Initialize the log level
	logLevel, err := zapcore.ParseLevel(c.logLevel)
	if err != nil {
		return fmt.Errorf("unable to parse log level, %w", err)
	}

	// Initialize the log format
	logFormat := log.Format(strings.ToLower(c.logFormat))

	// Initialize the zap logger
	zapLogger := log.GetZapLoggerFn(logFormat)(io.Out(), logLevel)

	// Wrap the zap logger
	logger := log.ZapLoggerToSlog(zapLogger)

	// Write genesis file if missing.
	genesisFilePath := filepath.Join(dataDir, cfg.Genesis)

	if !osm.FileExists(genesisFilePath) {
		// Create priv validator first.
		// Need it to generate genesis.json
		newPrivValKey := cfg.PrivValidatorKeyFile()
		newPrivValState := cfg.PrivValidatorStateFile()
		priv := privval.LoadOrGenFilePV(newPrivValKey, newPrivValState)
		pk := priv.GetPubKey()

		// Generate genesis.json file
		if err := generateGenesisFile(genesisFilePath, pk, c); err != nil {
			return fmt.Errorf("unable to generate genesis file: %w", err)
		}
	}

	// Initialize the indexer config
	txEventStoreCfg, err := getTxEventStoreConfig(c)
	if err != nil {
		return fmt.Errorf("unable to parse indexer config, %w", err)
	}
	cfg.TxEventStore = txEventStoreCfg

	// Create application and node.
	gnoApp, err := gnoland.NewApp(dataDir, c.skipFailingGenesisTxs, logger, c.genesisMaxVMCycles)
	if err != nil {
		return fmt.Errorf("error in creating new app: %w", err)
	}
	cfg.LocalApp = gnoApp

	gnoNode, err := node.DefaultNewNode(cfg, logger)
	if err != nil {
		return fmt.Errorf("error in creating node: %w", err)
	}

	fmt.Fprintln(io.Err(), "Node created.")

	if c.skipStart {
		io.ErrPrintln("'--skip-start' is set. Exiting.")
		return nil
	}

	if err := gnoNode.Start(); err != nil {
		return fmt.Errorf("error in start node: %w", err)
	}

	osm.TrapSignal(func() {
		if gnoNode.IsRunning() {
			_ = gnoNode.Stop()
		}

		// Sync the logger before exiting
		_ = zapLogger.Sync()
	})

	// Run forever
	select {}
}

func generateGenesisFile(genesisFile string, pk crypto.PubKey, c *startCfg) error {
	gen := &bft.GenesisDoc{}
	gen.GenesisTime = time.Now()
	gen.ChainID = c.chainID
	gen.ConsensusParams = abci.ConsensusParams{
		Block: &abci.BlockParams{
			// TODO: update limits.
			MaxTxBytes:   1_000_000,  // 1MB,
			MaxDataBytes: 2_000_000,  // 2MB,
			MaxGas:       10_0000_00, // 10M gas
			TimeIotaMS:   100,        // 100ms
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

// getTxEventStoreConfig constructs an event store config from provided user options
func getTxEventStoreConfig(c *startCfg) (*eventstorecfg.Config, error) {
	var cfg *eventstorecfg.Config

	switch c.txEventStoreType {
	case file.EventStoreType:
		if c.txEventStorePath == "" {
			return nil, errors.New("unspecified file transaction indexer path")
		}

		// Fill out the configuration
		cfg = &eventstorecfg.Config{
			EventStoreType: file.EventStoreType,
			Params: map[string]any{
				file.Path: c.txEventStorePath,
			},
		}
	default:
		cfg = eventstorecfg.DefaultEventStoreConfig()
	}

	return cfg, nil
}
