package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	errInvalidTxsFile     = errors.New("unable to open transactions file")
	errNoTxsFileSpecified = errors.New("no txs file specified")
)

// newTxsAddCmd creates the genesis txs add subcommand
func newTxsAddCmd(txsCfg *txsCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "add",
			ShortUsage: "txs add <tx-file ...>",
			ShortHelp:  "imports transactions into the genesis.json",
			LongHelp:   "Imports the transactions from a tx-archive backup to the genesis.json",
		},
		commands.NewEmptyConfig(),
		func(ctx context.Context, args []string) error {
			return execTxsAdd(ctx, txsCfg, io, args)
		},
	)
}

func execTxsAdd(
	ctx context.Context,
	cfg *txsCfg,
	io commands.IO,
	args []string,
) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.genesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Open the transactions files
	if len(args) == 0 {
		return errNoTxsFileSpecified
	}

	parsedTxs := make([]std.Tx, 0)
	for _, file := range args {
		file, loadErr := os.Open(file)
		if loadErr != nil {
			return fmt.Errorf("%w, %w", errInvalidTxsFile, loadErr)
		}

		txs, err := std.ParseTxs(ctx, file)
		if err != nil {
			return fmt.Errorf("unable to read file, %w", err)
		}

		parsedTxs = append(parsedTxs, txs...)
	}

	// Initialize the app state if it's not present
	if genesis.AppState == nil {
		genesis.AppState = gnoland.GnoGenesisState{}
	}

	state := genesis.AppState.(gnoland.GnoGenesisState)

	// Left merge the transactions
	fileTxStore := txStore(parsedTxs)
	genesisTxStore := txStore(state.Txs)

	// The genesis transactions have preference with the order
	// in the genesis.json
	if err := genesisTxStore.leftMerge(fileTxStore); err != nil {
		return err
	}

	// Save the state
	state.Txs = genesisTxStore
	genesis.AppState = state

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"Saved %d transactions to genesis.json",
		len(parsedTxs),
	)

	return nil
}
