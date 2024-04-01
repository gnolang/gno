package client

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"sort"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
)

var (
	errOverwriteAborted       = errors.New("overwrite aborted")
	errUnableToVerifyMultisig = errors.New("unable to verify multisig threshold")
)

type AddMultisigCfg struct {
	RootCfg *AddCfg

	NoSort            bool
	Multisig          commands.StringArr
	MultisigThreshold int
}

// NewAddMultisigCmd creates a gnokey add multisig command
func NewAddMultisigCmd(rootCfg *AddCfg, io commands.IO) *commands.Command {
	cfg := &AddMultisigCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "multisig",
			ShortUsage: "add multisig [flags] <key-name>",
			ShortHelp:  "adds a multisig key reference to the keybase",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execAddMultisig(cfg, args, io)
		},
	)
}

func (c *AddMultisigCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.NoSort,
		"nosort",
		false,
		"keys passed to --multisig are taken in the order they're supplied",
	)

	fs.Var(
		&c.Multisig,
		"multisig",
		"construct and store a multisig public key",
	)

	fs.IntVar(
		&c.MultisigThreshold,
		"threshold",
		1,
		"K out of N required signatures",
	)
}

func execAddMultisig(cfg *AddMultisigCfg, args []string, io commands.IO) error {
	// Validate a key name was provided
	if len(args) != 1 {
		return flag.ErrHelp
	}

	// Validate the multisig threshold
	if err := keys.ValidateMultisigThreshold(
		cfg.MultisigThreshold,
		len(cfg.Multisig),
	); err != nil {
		return errUnableToVerifyMultisig
	}

	name := args[0]

	// Read the keybase from the home directory
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.Home)
	if err != nil {
		return fmt.Errorf("unable to read keybase, %w", err)
	}

	// Check if the key exists
	exists, err := kb.HasByName(name)
	if err != nil {
		return fmt.Errorf("unable to fetch key, %w", err)
	}

	// Get overwrite confirmation, if any
	if exists {
		overwrite, err := io.GetConfirmation(fmt.Sprintf("Override the existing name %s", name))
		if err != nil {
			return fmt.Errorf("unable to get confirmation, %w", err)
		}

		if !overwrite {
			return errOverwriteAborted
		}
	}

	publicKeys := make([]crypto.PubKey, 0)
	for _, keyName := range cfg.Multisig {
		k, err := kb.GetByName(keyName)
		if err != nil {
			return fmt.Errorf("unable to fetch key, %w", err)
		}

		publicKeys = append(publicKeys, k.GetPubKey())
	}

	// Check if the keys should be sorted
	if !cfg.NoSort {
		sort.Slice(publicKeys, func(i, j int) bool {
			return publicKeys[i].Address().Compare(publicKeys[j].Address()) < 0
		})
	}

	// Create a new public key with the multisig threshold
	if _, err := kb.CreateMulti(
		name,
		multisig.NewPubKeyMultisigThreshold(cfg.MultisigThreshold, publicKeys),
	); err != nil {
		return fmt.Errorf("unable to create multisig key reference, %w", err)
	}

	io.Printfln("Key %q saved to disk.\n", name)

	return nil
}
