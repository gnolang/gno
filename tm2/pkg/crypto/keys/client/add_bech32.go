package client

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

type AddBech32Cfg struct {
	RootCfg *AddCfg

	PublicKey string
}

// NewAddBech32Cmd creates a gnokey add bech32 command
func NewAddBech32Cmd(rootCfg *AddCfg, io commands.IO) *commands.Command {
	cfg := &AddBech32Cfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "bech32",
			ShortUsage: "add bech32 [flags] <key-name>",
			ShortHelp:  "adds a public key to the keybase, using the bech32 representation",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execAddBech32(cfg, args, io)
		},
	)
}

func (c *AddBech32Cfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.PublicKey,
		"pubkey",
		"",
		"parse a public key in bech32 format and save it to disk",
	)
}

func execAddBech32(cfg *AddBech32Cfg, args []string, io commands.IO) error {
	// Validate a key name was provided
	if len(args) != 1 {
		return flag.ErrHelp
	}

	name := args[0]

	// Read the keybase from the home directory
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.Home)
	if err != nil {
		return fmt.Errorf("unable to read keybase, %w", err)
	}

	// Parse the public key
	publicKey, err := crypto.PubKeyFromBech32(cfg.PublicKey)
	if err != nil {
		return fmt.Errorf("unable to parse public key from bech32, %w", err)
	}

	// If not forcing, check for collisions with existing keys
	if !cfg.RootCfg.Force {
		// Derive the address to check for collision
		newAddress := publicKey.Address()

		// Handle address / name collision if any
		handled, err := handleCollision(kb, name, newAddress, keys.TypeOffline, io)
		if err != nil {
			return err
		}
		// If a collision was found and handled, we can skip saving the new key
		if handled {
			return nil
		}
	}

	// Save it offline in the keybase
	_, err = kb.CreateOffline(name, publicKey)
	if err != nil {
		return fmt.Errorf("unable to save public key, %w", err)
	}

	io.Printfln("Key %q saved to disk.\n", name)

	return nil
}
