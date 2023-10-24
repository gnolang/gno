package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
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
	errInvalidTxsFile    = errors.New("unable to open transactions file")
	errTxsParsingAborted = errors.New("transaction parsing aborted")
)

type txsAddCfg struct {
	rootCfg *txsCfg

	parseExport string
}

// newTxsAddCmd creates the genesis txs add subcommand
func newTxsAddCmd(txsCfg *txsCfg, io *commands.IO) *commands.Command {
	cfg := &txsAddCfg{
		rootCfg: txsCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "add",
			ShortUsage: "txs add [flags]",
			ShortHelp:  "Imports transactions into the genesis.json",
			LongHelp:   "Imports the transactions from a tx-archive backup to the genesis.json",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execTxsAdd(ctx, cfg, io)
		},
	)
}

func (c *txsAddCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.parseExport,
		"parse-export",
		"",
		"the path to the transactions export containing a list of transactions",
	)
}

func execTxsAdd(ctx context.Context, cfg *txsAddCfg, io *commands.IO) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.rootCfg.genesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Open the transactions file
	file, loadErr := os.Open(cfg.parseExport)
	if loadErr != nil {
		return fmt.Errorf("%w, %w", errInvalidTxsFile, loadErr)
	}

	txs, err := getTransactionsFromFile(ctx, file)
	if err != nil {
		return fmt.Errorf("unable to read file, %w", err)
	}

	// Initialize the app state if it's not present
	if genesis.AppState == nil {
		genesis.AppState = gnoland.GnoGenesisState{}
	}

	state := genesis.AppState.(gnoland.GnoGenesisState)

	// Left merge the transactions
	fileTxStore := txStore(txs)
	genesisTxStore := txStore(state.Txs)

	if err := genesisTxStore.leftMerge(fileTxStore); err != nil {
		return err
	}

	state.Txs = genesisTxStore
	genesis.AppState = state

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.rootCfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"Saved %d transactions to genesis.json",
		len(txs),
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
