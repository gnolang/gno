package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	errAppStateNotSet    = errors.New("genesis app state not set")
	errNoTxHashSpecified = errors.New("no transaction hashes specified")
	errTxNotFound        = errors.New("transaction not present in genesis.json")
)

// newTxsRemoveCmd creates the genesis txs remove subcommand
func newTxsRemoveCmd(txsCfg *txsCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "remove",
			ShortUsage: "txs remove <tx-hash ...>",
			ShortHelp:  "removes the transactions from the genesis.json",
			LongHelp:   "Removes the transactions using the transaction hash",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execTxsRemove(txsCfg, io, args)
		},
	)
}

func execTxsRemove(cfg *txsCfg, io commands.IO, args []string) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.genesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Check if the genesis state is set at all
	if genesis.AppState == nil {
		return errAppStateNotSet
	}

	// Make sure the transaction hashes are set
	if len(args) == 0 {
		return errNoTxHashSpecified
	}

	state := genesis.AppState.(gnoland.GnoGenesisState)

	for _, inputHash := range args {
		index := -1

		for indx, tx := range state.Txs {
			// Find the hash of the transaction
			hash, err := getTxHash(tx)
			if err != nil {
				return fmt.Errorf("unable to generate tx hash, %w", err)
			}

			// Check if the hashes match
			if strings.ToLower(hash) == strings.ToLower(inputHash) {
				index = indx

				break
			}
		}

		if index < 0 {
			return errTxNotFound
		}

		state.Txs = append(state.Txs[:index], state.Txs[index+1:]...)

		io.Printfln(
			"Transaction %s removed from genesis.json",
			inputHash,
		)
	}

	genesis.AppState = state

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	return nil
}

// getTxHash returns the hex hash representation of
// the transaction (Amino encoded)
func getTxHash(tx std.Tx) (string, error) {
	encodedTx, err := amino.Marshal(tx)
	if err != nil {
		return "", fmt.Errorf("unable to marshal transaction, %w", err)
	}

	txHash := types.Tx(encodedTx).Hash()

	return fmt.Sprintf("%X", txHash), nil
}
