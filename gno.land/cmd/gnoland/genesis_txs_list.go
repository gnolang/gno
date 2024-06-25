package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newTxsListCmd list all transactions on the specified genesis file
func newTxsListCmd(txsCfg *txsCfg, io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "list",
			ShortUsage: "txs list [flags] [<arg>...]",
			ShortHelp:  "lists transactions existing on genesis.json",
			LongHelp:   "Lists transactions existing on genesis.json",
		},
		commands.NewEmptyConfig(),
		func(ctx context.Context, args []string) error {
			return execTxsListCmd(io, txsCfg)
		},
	)

	return cmd
}

func execTxsListCmd(io commands.IO, cfg *txsCfg) error {
	genesis, err := types.GenesisDocFromFile(cfg.genesisPath)
	if err != nil {
		return fmt.Errorf("unable to load genesis, %w", err)
	}

	gs, ok := genesis.AppState.(gnoland.GnoGenesisState)
	if !ok {
		return fmt.Errorf("genesis state is not using the correct Gno Genesis type")
	}

	je := json.NewEncoder(io.Out())

	je.SetIndent("", "    ")

	return je.Encode(gs.Txs)
}
