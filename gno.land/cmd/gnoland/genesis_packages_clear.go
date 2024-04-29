package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type packagesClearCfg struct {
	rootCfg *packagesCfg
}

// newPackagesClearCmd creates the genesis packages clear subcommand
func newPackagesClearCmd(rootCfg *packagesCfg, io commands.IO) *commands.Command {
	cfg := &packagesClearCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "clear",
			ShortUsage: "packages clear [flags]",
			ShortHelp:  "clears all the addpkg transactions",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execPackagesClear(cfg, args, io)
		},
	)
}

func (c *packagesClearCfg) RegisterFlags(fs *flag.FlagSet) {}

func execPackagesClear(cfg *packagesClearCfg, args []string, io commands.IO) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	// Load the genesis
	genesis, err := types.GenesisDocFromFile(cfg.rootCfg.genesisPath)
	if err != nil {
		return fmt.Errorf("unable to load genesis, %w", err)
	}

	state := genesis.AppState.(gnoland.GnoGenesisState)

	var txs []std.Tx
	removed := 0
	for _, tx := range state.Txs {
		include := true
		for _, msg := range tx.Msgs {
			if msg.Type() == "add_package" {
				removed++
				include = false

				break
			}
		}
		if include {
			txs = append(txs, tx)
		}
	}

	// Save the txs
	state.Txs = txs
	genesis.AppState = state

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.rootCfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"%d txs removed!",
		removed,
	)

	return nil
}
