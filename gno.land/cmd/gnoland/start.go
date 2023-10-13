package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
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
}

func newStartCmd(io *commands.IO) *commands.Command {
	cfg := &startCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "start",
			ShortUsage: "start [flags]",
			ShortHelp:  "Run the full node",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execStart(cfg, args, io)
		},
	)
}

func (c *startCfg) RegisterFlags(fs *flag.FlagSet) {
	gnoroot := gnoland.MustGuessGnoRootDir()
	defaultGenesisBalancesFile := filepath.Join(gnoroot, "gno.land/genesis/genesis_balances.txt")
	defaultGenesisTxsFile := filepath.Join(gnoroot, "gno.land/genesis/genesis_txs.txt")

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
		"config",
		"",
		"config file (optional)",
	)

	// XXX(deprecated): use data-dir instead
	fs.StringVar(
		&c.dataDir,
		"root-dir",
		"testdir",
		"deprecated: use data-dir instead - directory for config and data",
	)
}

func execStart(c *startCfg, args []string, io *commands.IO) error {
	logger := log.NewTMLogger(log.NewSyncWriter(io.Out))
	dataDir := c.dataDir

	cfg := config.LoadOrMakeConfigWithOptions(dataDir, func(cfg *config.Config) {
		cfg.Consensus.CreateEmptyBlocks = true
		cfg.Consensus.CreateEmptyBlocksInterval = 0 * time.Second
	})

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

	io.ErrPrintln("Node created.")

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
			MaxTxBytes:   1000000,  // 1MB,
			MaxDataBytes: 2000000,  // 2MB,
			MaxGas:       10000000, // 10M gas
			TimeIotaMS:   100,      // 100ms
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
	examplePkgs := gnoland.PackagePath{
		Path:    examplesDir,
		Creator: test1,
		Fee:     std.NewFee(50000, std.MustParseCoin("1000000ugnot")),
		Deposit: nil,
	}

	pkgsTxs, err := examplePkgs.Load()
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
