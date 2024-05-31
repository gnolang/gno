package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	errInvalidTxsFile     = errors.New("unable to open transactions file")
	errNoTxsFileSpecified = errors.New("no txs file specified")
	errTxsParsingAborted  = errors.New("transaction parsing aborted")
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

		txs, err := getTransactionsFromFile(ctx, file)
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

// getTransactionsFromFile fetches the transactions from the
// specified reader
func getTransactionsFromFile(ctx context.Context, reader io.Reader) ([]std.Tx, error) {
	txs := make([]std.Tx, 0)

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, errTxsParsingAborted
		default:
			// Parse the amino JSON
			var tx std.Tx

			if err := amino.UnmarshalJSON(scanner.Bytes(), &tx); err != nil {
				return nil, fmt.Errorf(
					"unable to unmarshal amino JSON, %w",
					err,
				)
			}

			txs = append(txs, tx)
		}
	}

	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf(
			"error encountered while reading file, %w",
			err,
		)
	}

	return txs, nil
}
