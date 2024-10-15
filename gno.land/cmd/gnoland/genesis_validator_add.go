package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	_ "github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

var (
	errInvalidPower             = errors.New("invalid validator power")
	errInvalidName              = errors.New("invalid validator name")
	errPublicKeyAddressMismatch = errors.New("provided public key and address do not match")
	errAddressPresent           = errors.New("validator with same address already present in genesis.json")
)

type validatorAddCfg struct {
	rootCfg *validatorCfg

	pubKey string
	name   string
	power  int64
}

// newValidatorAddCmd creates the genesis validator add subcommand
func newValidatorAddCmd(validatorCfg *validatorCfg, io commands.IO) *commands.Command {
	cfg := &validatorAddCfg{
		rootCfg: validatorCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "add",
			ShortUsage: "validator add [flags]",
			ShortHelp:  "adds a new validator to the genesis.json",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execValidatorAdd(cfg, io)
		},
	)
}

func (c *validatorAddCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.pubKey,
		"pub-key",
		"",
		"the bech32 string representation of the validator's public key",
	)

	fs.StringVar(
		&c.name,
		"name",
		"",
		"the name of the validator (must be unique)",
	)

	fs.Int64Var(
		&c.power,
		"power",
		1,
		"the voting power of the validator (must be > 0)",
	)
}

func execValidatorAdd(cfg *validatorAddCfg, io commands.IO) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.rootCfg.genesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Check the validator address
	address, err := crypto.AddressFromString(cfg.rootCfg.address)
	if err != nil {
		return fmt.Errorf("invalid validator address, %w", err)
	}

	// Check the voting power
	if cfg.power < 1 {
		return errInvalidPower
	}

	// Check the name
	if cfg.name == "" {
		return errInvalidName
	}

	// Check the public key
	pubKey, err := crypto.PubKeyFromBech32(cfg.pubKey)
	if err != nil {
		return fmt.Errorf("invalid validator public key, %w", err)
	}

	// Check the public key matches the address
	if pubKey.Address() != address {
		return errPublicKeyAddressMismatch
	}

	validator := types.GenesisValidator{
		Address: address,
		PubKey:  pubKey,
		Power:   cfg.power,
		Name:    cfg.name,
	}

	// Check if the validator exists
	for _, genesisValidator := range genesis.Validators {
		// There is no need to check if the public keys match
		// since the address is derived from it, and the derivation
		// is checked already
		if validator.Address == genesisValidator.Address {
			return errAddressPresent
		}
	}

	// Add the validator
	genesis.Validators = append(genesis.Validators, validator)

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.rootCfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"Validator with address %s added to genesis file",
		cfg.rootCfg.address,
	)

	return nil
}
