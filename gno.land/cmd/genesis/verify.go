package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var errInvalidGenesisState = errors.New("invalid genesis state type")

type verifyCfg struct {
	genesisPath string
}

// newVerifyCmd creates the genesis verify subcommand
func newVerifyCmd(io *commands.IO) *commands.Command {
	cfg := &verifyCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "verify",
			ShortUsage: "verify [flags]",
			LongHelp:   "Verifies a node's genesis.json",
			ShortHelp:  "Verifies a genesis.json",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execVerify(cfg, io)
		},
	)
}

func (c *verifyCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.genesisPath,
		"genesis-path",
		"./genesis.json",
		"the path to the genesis.json",
	)
}

func execVerify(cfg *verifyCfg, io *commands.IO) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.genesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Verify it
	if validateErr := genesis.Validate(); validateErr != nil {
		return fmt.Errorf("unable to verify genesis, %w", validateErr)
	}

	// Validate the genesis state
	if genesis.AppState != nil {
		state, ok := genesis.AppState.(gnoland.GnoGenesisState)
		if !ok {
			return errInvalidGenesisState
		}

		// Validate the initial transactions
		for _, tx := range state.Txs {
			if validateErr := tx.ValidateBasic(); validateErr != nil {
				return fmt.Errorf("invalid transacton, %w", validateErr)
			}
		}

		// Validate the initial balances
		for _, balance := range state.Balances {
			if _, parseErr := std.ParseCoins(balance); parseErr != nil {
				return fmt.Errorf("invalid balance %s, %w", balance, parseErr)
			}
		}
	}

	io.Printfln("Genesis at %s is valid", cfg.genesisPath)

	return nil
}
