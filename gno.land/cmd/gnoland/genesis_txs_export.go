package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var errNoOutputFile = errors.New("no output file path specified")

// newTxsExportCmd creates the genesis txs export subcommand
func newTxsExportCmd(txsCfg *txsCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "export",
			ShortUsage: "txs export [flags] <output-path>",
			ShortHelp:  "exports the transactions from the genesis.json",
			LongHelp:   "Exports the transactions from the genesis.json to an output file",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execTxsExport(txsCfg, io, args)
		},
	)
}

func execTxsExport(cfg *txsCfg, io commands.IO, args []string) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.genesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Load the genesis state
	if genesis.AppState == nil {
		return errAppStateNotSet
	}

	state := genesis.AppState.(gnoland.GnoGenesisState)
	if len(state.Txs) == 0 {
		io.Println("No genesis transactions to export")

		return nil
	}

	// Make sure the output file path is specified
	if len(args) == 0 {
		return errNoOutputFile
	}

	// Open output file
	outputFile, err := os.OpenFile(
		args[0],
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0o755,
	)
	if err != nil {
		return fmt.Errorf("unable to create output file, %w", err)
	}

	// Save the transactions
	for _, tx := range state.Txs {
		// Marshal tx individual tx into JSON
		jsonData, err := amino.MarshalJSON(tx)
		if err != nil {
			return fmt.Errorf("unable to marshal JSON data, %w", err)
		}

		// Write the JSON data as a line to the file
		if _, err = outputFile.Write(jsonData); err != nil {
			return fmt.Errorf("unable to write to output, %w", err)
		}

		// Write a newline character to separate JSON objects
		if _, err = outputFile.WriteString("\n"); err != nil {
			return fmt.Errorf("unable to write newline output, %w", err)
		}
	}

	io.Printfln(
		"Exported %d transactions",
		len(state.Txs),
	)

	return nil
}
