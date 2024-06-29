package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

var errValidatorNotPresent = errors.New("validator not present in genesis.json")

type validatorRemoveCfg struct {
	rootCfg *validatorCfg

	address string
}

// newValidatorRemoveCmd creates the genesis validator remove subcommand
func newValidatorRemoveCmd(validatorCfg *validatorCfg, io commands.IO) *commands.Command {
	cfg := &validatorRemoveCfg{
		rootCfg: validatorCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "remove",
			ShortUsage: "validator remove [flags]",
			ShortHelp:  "removes a validator from the genesis.json",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execValidatorRemove(cfg, io)
		},
	)
}

func execValidatorRemove(cfg *validatorRemoveCfg, io commands.IO) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.rootCfg.genesisPath)
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
	if err := genesis.SaveAs(cfg.rootCfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"Validator with address %s removed from genesis file",
		cfg.address,
	)

	return nil
}

func (c *validatorRemoveCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.address,
		"address",
		"",
		"the gno bech32 address of the validator",
	)
}
