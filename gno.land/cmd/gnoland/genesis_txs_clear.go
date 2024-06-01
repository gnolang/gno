package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newTxsClearCmd creates the genesis txs clear subcommand
func newTxsClearCmd(txsCfg *txsCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "clear",
			ShortUsage: "txs clear",
			ShortHelp:  "clears all the transactions",
		},
		commands.NewEmptyConfig(),
		func(ctx context.Context, args []string) error {
			return execTxsClear(txsCfg, args, io)
		},
	)
}

func execTxsClear(cfg *txsCfg, args []string, io commands.IO) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	// Load the genesis
	genesis, err := types.GenesisDocFromFile(cfg.genesisPath)
	if err != nil {
		return fmt.Errorf("unable to load genesis, %w", err)
	}

	state := genesis.AppState.(gnoland.GnoGenesisState)

	totalTxs := len(state.Txs)

	// Remove all txs
	state.Txs = nil
	genesis.AppState = state

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"%d txs removed!",
		totalTxs,
	)

	return nil
}
