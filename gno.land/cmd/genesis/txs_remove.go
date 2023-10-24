package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var (
	errAppStateNotSet = errors.New("genesis app state not set")
	errTxNotFound     = errors.New("transaction not present in genesis.json")
)

type txsRemoveCfg struct {
	rootCfg *txsCfg

	hash string
}

// newTxsRemoveCmd creates the genesis txs remove subcommand
func newTxsRemoveCmd(txsCfg *txsCfg, io *commands.IO) *commands.Command {
	cfg := &txsRemoveCfg{
		rootCfg: txsCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "remove",
			ShortUsage: "txs remove [flags]",
			ShortHelp:  "Removes the transaction from the genesis.json",
			LongHelp:   "Removes the transaction using the transaction hash",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execTxsRemove(cfg, io)
		},
	)
}

func (c *txsRemoveCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.hash,
		"hash",
		"",
		"the transaction hash (hex format)",
	)
}

func execTxsRemove(cfg *txsRemoveCfg, io *commands.IO) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.rootCfg.genesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Check if the genesis state is set at all
	if genesis.AppState == nil {
		return errAppStateNotSet
	}

	var (
		state = genesis.AppState.(gnoland.GnoGenesisState)
		index = -1
	)

	for indx, tx := range state.Txs {
		// Find the hash of the transaction
		hash, err := getTxHash(tx)
		if err != nil {
			return fmt.Errorf("unable to generate tx hash, %w", err)
		}

		// Check if the hashes match
		if strings.ToLower(hash) == strings.ToLower(cfg.hash) {
			index = indx

			break
		}
	}

	if index < 0 {
		return errTxNotFound
	}

	state.Txs = append(state.Txs[:index], state.Txs[index+1:]...)
	genesis.AppState = state

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.rootCfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"Transaction %s removed from genesis.json",
		cfg.hash,
	)

	return nil
}
