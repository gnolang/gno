package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

func newInitCmd(io *commands.IO) *commands.Command {
	cfg := &initCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "init [flags]",
			ShortHelp:  "Initialize a gnoland configuration on file-system",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execInit(cfg, args, io)
		},
	)
}

type initCfg struct {
	// common
	rootDir string
	config  string

	// gnoland init specific
	chainID               string
	genesisTxsFile        string
	genesisRemote         string
	skipFailingGenesisTxs bool
	genesisBalancesFile   string
	genesisMaxVMCycles    int64
}

func (c *initCfg) RegisterFlags(fs *flag.FlagSet) {
	// common
	fs.StringVar(&c.rootDir, "root-dir", "testdir", "directory for config and data")
	fs.StringVar(&c.config, "config", "", "config file (optional)")

	// init specific
	fs.BoolVar(&c.skipFailingGenesisTxs, "skip-failing-genesis-txs", false, "don't panic when replaying invalid genesis txs")
	fs.StringVar(&c.genesisBalancesFile, "genesis-balances-file", "./genesis/genesis_balances.txt", "initial distribution file")
	fs.StringVar(&c.genesisTxsFile, "genesis-txs-file", "./genesis/genesis_txs.txt", "initial txs to replay")
	fs.StringVar(&c.chainID, "chainid", "dev", "the ID of the chain")
	fs.StringVar(&c.genesisRemote, "genesis-remote", "localhost:26657", "replacement for '%%REMOTE%%' in genesis")
	fs.Int64Var(&c.genesisMaxVMCycles, "genesis-max-vm-cycles", 10_000_000, "set maximum allowed vm cycles per operation. Zero means no limit.")
}

func execInit(c *initCfg, args []string, io *commands.IO) error {
	logger := log.NewTMLogger(log.NewSyncWriter(io.Out))
	rootDir := c.rootDir

	cfg := config.LoadOrMakeConfigWithOptions(rootDir, func(cfg *config.Config) {
		cfg.Consensus.CreateEmptyBlocks = true
		cfg.Consensus.CreateEmptyBlocksInterval = 0 * time.Second
	})

	// create priv validator first.
	// need it to generate genesis.json
	newPrivValKey := cfg.PrivValidatorKeyFile()
	newPrivValState := cfg.PrivValidatorStateFile()
	priv := privval.LoadOrGenFilePV(newPrivValKey, newPrivValState)

	// write genesis file if missing.
	genesisFilePath := filepath.Join(rootDir, cfg.Genesis)
	if !osm.FileExists(genesisFilePath) {
		genDoc := makeGenesisDoc(
			priv.GetPubKey(),
			c.chainID,
			c.genesisBalancesFile,
			loadGenesisTxs(c.genesisTxsFile, c.chainID, c.genesisRemote),
		)
		writeGenesisFile(genDoc, genesisFilePath)
	}

	// create application and node.
	gnoApp, err := gnoland.NewApp(rootDir, c.skipFailingGenesisTxs, logger, c.genesisMaxVMCycles)
	if err != nil {
		return fmt.Errorf("error in creating new app: %w", err)
	}

	cfg.LocalApp = gnoApp

	fmt.Fprintln(io.Err, "Node created.")
	return nil
}
