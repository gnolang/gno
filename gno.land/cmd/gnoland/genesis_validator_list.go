package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type validatorListCfg struct {
	rootCfg *validatorCfg
}

func newValidatorListCmd(validatorCfg *validatorCfg, io commands.IO) *commands.Command {
	cfg := &validatorListCfg{
		rootCfg: validatorCfg,
	}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "list",
			ShortUsage: "validator list [flags]",
			ShortHelp:  "lists current validator set in the genesis.json",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execValidatorList(cfg, io)
		},
	)
}

func (c *validatorListCfg) RegisterFlags(fs *flag.FlagSet) {
}

func execValidatorList(cfg *validatorListCfg, io commands.IO) error {
	// Load genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.rootCfg.genesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Print validator set
	io.Printf("Validator set in %s has %d validator(s):\n\n", cfg.rootCfg.genesisPath, len(genesis.Validators))
	for _, validator := range genesis.Validators {
		io.Printf("%s power=%d %s %s\n", 
			validator.Address.String(), 
			validator.Power,
			validator.Name, 
			validator.PubKey.String(),
		)
	}
	return nil
}
