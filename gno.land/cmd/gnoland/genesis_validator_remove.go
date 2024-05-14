package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

var errValidatorNotPresent = errors.New("validator not present in genesis.json")

// newValidatorRemoveCmd creates the genesis validator remove subcommand
func newValidatorRemoveCmd(rootCfg *validatorCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "remove",
			ShortUsage: "validator remove [flags]",
			ShortHelp:  "removes a validator from the genesis.json",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, _ []string) error {
			return execValidatorRemove(rootCfg, io)
		},
	)
}

func execValidatorRemove(cfg *validatorCfg, io commands.IO) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.genesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Check the validator address
	address, err := crypto.AddressFromString(cfg.address)
	if err != nil {
		return fmt.Errorf("invalid validator address, %w", err)
	}

	index := -1

	for indx, validator := range genesis.Validators {
		if validator.Address == address {
			index = indx

			break
		}
	}

	if index < 0 {
		return errors.New("validator not present in genesis.json")
	}

	// Drop the validator
	genesis.Validators = append(genesis.Validators[:index], genesis.Validators[index+1:]...)

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"Validator with address %s removed from genesis file",
		cfg.address,
	)

	return nil
}
