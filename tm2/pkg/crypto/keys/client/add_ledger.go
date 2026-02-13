package client

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

// NewAddLedgerCmd creates a gnokey add ledger command
func NewAddLedgerCmd(cfg *AddCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "ledger",
			ShortUsage: "add ledger [flags] <key-name>",
			ShortHelp:  "adds a Ledger key reference to the keybase",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execAddLedger(cfg, args, io)
		},
	)
}

func execAddLedger(cfg *AddCfg, args []string, io commands.IO) error {
	// Validate a key name was provided
	if len(args) != 1 {
		return flag.ErrHelp
	}

	name := args[0]

	// Read the keybase from the home directory
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.Home)
	if err != nil {
		return fmt.Errorf("unable to read keybase, %w", err)
	}

	// If not forcing, check for collisions with existing keys
	if !cfg.Force {
		// Handle name collision if any
		handled, err := handleCollision(kb, name, crypto.Address{}, keys.TypeLedger, io)
		if err != nil {
			return err
		}
		// If a collision was found and handled, we can skip saving the new key
		if handled {
			return nil
		}
	}

	// Create the ledger reference
	info, err := kb.CreateLedger(
		name,
		keys.Secp256k1,
		crypto.Bech32AddrPrefix,
		uint32(cfg.Account),
		uint32(cfg.Index),
	)
	if err != nil {
		return fmt.Errorf("unable to create Ledger reference in keybase, %w", err)
	}

	// Print the information
	printCreate(info, false, "", io)

	return nil
}
