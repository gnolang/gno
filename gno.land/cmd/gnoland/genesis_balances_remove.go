package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

var (
	errUnableToLoadGenesis = errors.New("unable to load genesis")
	errBalanceNotFound     = errors.New("genesis balances entry does not exist")
)

type balancesRemoveCfg struct {
	rootCfg *balancesCfg

	address string
}

// newBalancesRemoveCmd creates the genesis balances remove subcommand
func newBalancesRemoveCmd(rootCfg *balancesCfg, io commands.IO) *commands.Command {
	cfg := &balancesRemoveCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "remove",
			ShortUsage: "balances remove [flags]",
			ShortHelp:  "removes the balance information of a specific account",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execBalancesRemove(cfg, io)
		},
	)
}

func (c *balancesRemoveCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.address,
		"address",
		"",
		"the address of the account whose balance information should be removed from genesis.json",
	)
}

func execBalancesRemove(cfg *balancesRemoveCfg, io commands.IO) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.rootCfg.genesisPath)
	if loadErr != nil {
		return fmt.Errorf("%w, %w", errUnableToLoadGenesis, loadErr)
	}

	// Validate the address
	address, err := crypto.AddressFromString(cfg.address)
	if err != nil {
		return fmt.Errorf("%w, %w", errInvalidAddress, err)
	}

	// Check if the genesis state is set at all
	if genesis.AppState == nil {
		return errAppStateNotSet
	}

	// Construct the initial genesis balance sheet
	state := genesis.AppState.(gnoland.GnoGenesisState)
	genesisBalances, err := mapGenesisBalancesFromState(state)
	if err != nil {
		return err
	}

	// Check if the genesis balance for the account is present
	_, exists := genesisBalances[address]
	if !exists {
		return errBalanceNotFound
	}

	// Drop the account pre-mine
	delete(genesisBalances, address)

	// Save the balances
	state.Balances = genesisBalances.List()
	genesis.AppState = state

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.rootCfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"Pre-mine information for address %s removed",
		address.String(),
	)

	return nil
}
