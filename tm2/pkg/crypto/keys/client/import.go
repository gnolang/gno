package client

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

type importCfg struct {
	rootCfg *baseCfg

	keyName   string
	armorPath string
	unsafe    bool
}

func newImportCmd(rootCfg *baseCfg, io commands.IO) *commands.Command {
	cfg := &importCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "import",
			ShortUsage: "import [flags]",
			ShortHelp:  "Imports encrypted private key armor",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execImport(cfg, io)
		},
	)
}

func (c *importCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.keyName,
		"name",
		"",
		"The name of the private key",
	)

	fs.StringVar(
		&c.armorPath,
		"armor-path",
		"",
		"The path to the encrypted armor file",
	)

	fs.BoolVar(
		&c.unsafe,
		"unsafe",
		false,
		"Import the private key armor as unencrypted",
	)
}

func execImport(cfg *importCfg, io commands.IO) error {
	// Create a new instance of the key-base
	kb, err := keys.NewKeyBaseFromDir(cfg.rootCfg.Home)
	if err != nil {
		return fmt.Errorf(
			"unable to create a key base from directory %s, %w",
			cfg.rootCfg.Home,
			err,
		)
	}

	// Read the raw encrypted armor
	armor, err := os.ReadFile(cfg.armorPath)
	if err != nil {
		return fmt.Errorf(
			"unable to read armor from path %s, %w",
			cfg.armorPath,
			err,
		)
	}

	var (
		decryptPassword string
		encryptPassword string
	)

	if !cfg.unsafe {
		// Get the armor decrypt password
		decryptPassword, err = io.GetPassword(
			"Enter a passphrase to decrypt your private key armor:",
			cfg.rootCfg.InsecurePasswordStdin,
		)
		if err != nil {
			return fmt.Errorf(
				"unable to retrieve armor decrypt password from user, %w",
				err,
			)
		}
	}

	// Get the key-base encrypt password
	encryptPassword, err = io.GetCheckPassword(
		[2]string{
			"Enter a passphrase to encrypt your private key:",
			"Repeat the passphrase:",
		},
		cfg.rootCfg.InsecurePasswordStdin,
	)
	if err != nil {
		return fmt.Errorf(
			"unable to retrieve key encrypt password from user, %w",
			err,
		)
	}

	if cfg.unsafe {
		// Import the unencrypted private key
		if err := kb.ImportPrivKeyUnsafe(
			cfg.keyName,
			string(armor),
			encryptPassword,
		); err != nil {
			return fmt.Errorf(
				"unable to import the unencrypted private key, %w",
				err,
			)
		}
	} else {
		// Import the encrypted private key
		if err := kb.ImportPrivKey(
			cfg.keyName,
			string(armor),
			decryptPassword,
			encryptPassword,
		); err != nil {
			return fmt.Errorf(
				"unable to import the encrypted private key, %w",
				err,
			)
		}
	}

	io.Printfln("Successfully imported private key %s", cfg.keyName)

	return nil
}
