package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	errInvalidTxsPath     = errors.New("invalid transactions path")
	errInvalidTxsFile     = errors.New("unable to open transactions file")
	errNoTxsFileSpecified = errors.New("no txs file specified")
)

var (
	genesisDeployAddress = crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // test1
	genesisDeployFee     = std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
)

// newTxsAddCmd creates the genesis txs add subcommand
func newTxsAddCmd(txsCfg *txsCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "add",
			ShortUsage: "txs add <tx-path ...>",
			ShortHelp:  "imports transactions into the genesis.json",
			LongHelp:   "Imports the transactions from a given transactions sheet, or package directory, to the genesis.json",
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
	for _, argPath := range args {
		// Grab the absolute path
		path, err := filepath.Abs(argPath)
		if err != nil {
			return fmt.Errorf("unable to get absolute path %s, %w", path, err)
		}

		// Grab the file info
		fileInfo, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("%w %s, %w", errInvalidTxsPath, path, err)
		}

		var txs []std.Tx

		if fileInfo.IsDir() {
			// Generate transactions from the packages
			txs, err = loadTxFromDir(path)
		} else {
			// Load the transactions from the transaction sheet
			txs, err = loadTxFromFile(ctx, path)
		}

		if err != nil {
			return fmt.Errorf("unable to load transactions, %w", err)
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

// loadTxFromFile loads the transactions from a transaction sheet
func loadTxFromFile(ctx context.Context, path string) ([]std.Tx, error) {
	file, loadErr := os.Open(path)
	if loadErr != nil {
		return nil, fmt.Errorf("%w, %w", errInvalidTxsFile, loadErr)
	}
	defer file.Close()

	txs, err := std.ParseTxs(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("unable to parse file, %w", err)
	}

	return txs, nil
}

// loadTxFromDir loads the transactions from the given packages directory, recursively
func loadTxFromDir(path string) ([]std.Tx, error) {
	txs, err := gnoland.LoadPackagesFromDir(path, genesisDeployAddress, genesisDeployFee)
	if err != nil {
		return nil, fmt.Errorf("unable to load txs from directory, %w", err)
	}

	return txs, nil
}
