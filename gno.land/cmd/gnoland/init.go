package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type initCfg struct {
	chainID string
	rootDir string

	genesisTxsFile      string
	genesisBalancesFile string
	genesisMaxVMCycles  int64
}

func newInitCmd(io *commands.IO) *commands.Command {
	cfg := &initCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "init [flags]",
			ShortHelp:  "Initialize validators's and node's configuration files.",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execInit(cfg, args, io)
		},
	)
}

func (c *initCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.chainID,
		"chain-id",
		"dev",
		"the ID of the chain",
	)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"testdir",
		"directory for config and data",
	)

	fs.StringVar(
		&c.genesisBalancesFile,
		"genesis-balances-file",
		"./genesis/genesis_balances.txt",
		"initial distribution file",
	)

	fs.StringVar(
		&c.genesisTxsFile,
		"genesis-txs-file",
		"./genesis/genesis_txs.txt",
		"initial txs to replay",
	)

	fs.Int64Var(
		&c.genesisMaxVMCycles,
		"genesis-max-vm-cycles",
		10_000_000,
		"set maximum allowed vm cycles per operation. Zero means no limit.",
	)
}

func execInit(c *initCfg, args []string, io *commands.IO) error {
	// logger := log.NewTMLogger(log.NewSyncWriter(io.Out))
	rootDir := c.rootDir
	genesisFile := rootDir + "/config/genesis.json"

	if osm.FileExists(genesisFile) {
		return fmt.Errorf("genesis.json file already exists: %s", genesisFile)
	}

	cfg := config.LoadOrMakeConfigWithOptions(rootDir, func(cfg *config.Config) {
		cfg.Consensus.CreateEmptyBlocks = true
		cfg.Consensus.CreateEmptyBlocksInterval = 0 * time.Second
	})

	// create priv validator first.
	// need it to generate genesis.json
	newPrivValKey := cfg.PrivValidatorKeyFile()
	newPrivValState := cfg.PrivValidatorStateFile()
	priv := privval.LoadOrGenFilePV(newPrivValKey, newPrivValState)

	// write genesis file
	genesisFilePath := filepath.Join(rootDir, cfg.Genesis)
	genDoc := makeGenesisDoc(
		priv.GetPubKey(),
		c.chainID,
		"",
		// c.genesisBalancesFile,
		[]std.Tx{},
		// loadGenesisTxs(c.genesisTxsFile, c.chainID, c.genesisRemote),
	)
	writeGenesisFile(genDoc, genesisFilePath)

	return nil
}
