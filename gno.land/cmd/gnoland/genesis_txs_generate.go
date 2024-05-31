package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var errUnableToLoadPackages = errors.New("unable to load packages")

// newTxsGenerateCmd creates the genesis txs generate subcommand
func newTxsGenerateCmd(txsCfg *txsCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "generate",
			ShortUsage: "txs generate <path> [<path>...]",
			ShortHelp:  "generates addpkg txs from dir and add them to genesis.json",
		},
		commands.NewEmptyConfig(),
		func(ctx context.Context, args []string) error {
			return execTxsGenerate(txsCfg, args, io)
		},
	)
}

func execTxsGenerate(cfg *txsCfg, args []string, io commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.genesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	txs, err := gnoland.LoadPackagesFromDirs(args, test1, defaultFee, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", errUnableToLoadPackages, err)
	}

	// append generated addpkg txs to genesis
	if err := appendTxs(genesis, txs); err != nil {
		return err
	}

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"Saved %d transactions to genesis.json",
		len(txs),
	)

	return nil
}
