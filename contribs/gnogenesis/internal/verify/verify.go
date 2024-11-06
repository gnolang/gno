package verify

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var errInvalidGenesisState = errors.New("invalid genesis state type")

type verifyCfg struct {
	common.Cfg
}

// NewVerifyCmd creates the genesis verify subcommand
func NewVerifyCmd(io commands.IO) *commands.Command {
	cfg := &verifyCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "verify",
			ShortUsage: "[flags]",
			ShortHelp:  "verifies a genesis.json",
			LongHelp:   "Verifies a node's genesis.json",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execVerify(cfg, io)
		},
	)
}

func (c *verifyCfg) RegisterFlags(fs *flag.FlagSet) {
	c.Cfg.RegisterFlags(fs)
}

func execVerify(cfg *verifyCfg, io commands.IO) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.GenesisPath)
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
			if validateErr := tx.Tx.ValidateBasic(); validateErr != nil {
				return fmt.Errorf("invalid transacton, %w", validateErr)
			}
		}

		// Validate the initial balances
		for _, balance := range state.Balances {
			if err := balance.Verify(); err != nil {
				return fmt.Errorf("invalid balance: %w", err)
			}
		}
	}

	io.Printfln("Genesis at %s is valid", cfg.GenesisPath)

	return nil
}
