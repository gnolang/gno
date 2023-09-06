package main

import (
	"context"
	"flag"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/commands"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
)

type addGnoGenesisMessageCfg struct {
	rootDir string

	genesisTxsFile string
}

func newAddGnoGenesisMessageCmd(io *commands.IO) *commands.Command {
	cfg := &addGnoGenesisMessageCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "add-gno-genesis-message",
			ShortUsage: "add-gno-genesis-message [flags]",
			ShortHelp:  "add-gno-genesis-message gno genesis import txs",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execAddGnoGenesisMessage(cfg, args, io)
		},
	)
}

func (c *addGnoGenesisMessageCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"testdir",
		"directory for config and data",
	)

	fs.StringVar(
		&c.genesisTxsFile,
		"genesis-txs-file",
		"./genesis/genesis_txs.txt",
		"addGnoGenesisMessageial txs to replay",
	)
}

func execAddGnoGenesisMessage(c *addGnoGenesisMessageCfg, args []string, io *commands.IO) error {
	// logger := log.NewTMLogger(log.NewSyncWriter(io.Out))
	rootDir := c.rootDir
	genesisFile := rootDir + "/config/genesis.json"

	gen, err := bft.GenesisDocFromFile(genesisFile)
	if err != nil {
		return err
	}

	appState, ok := gen.AppState.(gnoland.GnoGenesisState)
	if !ok {
		panic("failed to parse genesis state")
	}
	_ = appState

	return nil
}
