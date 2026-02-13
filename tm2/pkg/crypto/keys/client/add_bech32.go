package client

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
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

	// Parse the public key early so the address is known for collision checks
	publicKey, err := crypto.PubKeyFromBech32(cfg.PublicKey)
	if err != nil {
		return fmt.Errorf("unable to parse public key from bech32, %w", err)
	}

	// Read the keybase from the home directory
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.Home)
	if err != nil {
		return fmt.Errorf("unable to read keybase, %w", err)
	}

	// Check if a key with signing capability already exists at this address.
	// Adding a public-key-only reference when a local or ledger key already
	// exists at the same address is redundant and should be skipped.
	newAddress := publicKey.Address()

	existingKey, err := kb.GetByAddress(newAddress)
	if err != nil && !keyerror.IsErrKeyNotFound(err) {
		return fmt.Errorf("unable to fetch key by address, %w", err)
	}

	if existingKey != nil {
		existingType := existingKey.GetType()
		if existingType == keys.TypeLocal || existingType == keys.TypeLedger {
			io.Println("A key with signing capability already exists at this address:")
			printNewInfo(existingKey, io)
			io.Println("Adding a public-key-only reference is redundant. Skipping.")

			return nil
		}
	}

	// Check for name collision
	confirmedKey, err := checkNameCollision(kb, name, io)
	if err != nil {
		return err
	}

	// Check for address collision
	if err := checkAddressCollision(kb, newAddress, confirmedKey, io); err != nil {
		return err
	}

	// Save it offline in the keybase
	_, err = kb.CreateOffline(name, publicKey)
	if err != nil {
		return fmt.Errorf("unable to save public key, %w", err)
	}

	io.Printfln("Key %q saved to disk.\n", name)

	return nil
}
