package client

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

type exportCfg struct {
	rootCfg *baseCfg

	nameOrBech32 string
	outputPath   string
	unsafe       bool
}

func newExportCmd(rootCfg *baseCfg) *commands.Command {
	cfg := &exportCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "export",
			ShortUsage: "export [flags]",
			ShortHelp:  "Exports private key armor",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execExport(cfg, commands.NewDefaultIO())
		},
	)
}

func (c *exportCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.nameOrBech32,
		"key",
		"",
		"Name or Bech32 address of the private key",
	)

	fs.StringVar(
		&c.outputPath,
		"output-path",
		"",
		"The desired output path for the armor file",
	)

	fs.BoolVar(
		&c.unsafe,
		"unsafe",
		false,
		"Export the private key armor as unencrypted",
	)
}

func execExport(cfg *exportCfg, io *commands.IO) error {
	// Create a new instance of the key-base
	kb, err := keys.NewKeyBaseFromDir(cfg.rootCfg.Home)
	if err != nil {
		return fmt.Errorf(
			"unable to create a key base from directory %s, %w",
			cfg.rootCfg.Home,
			err,
		)
	}

	// Get the key-base decrypt password
	decryptPassword, err := io.GetPassword(
		"Enter a passphrase to decrypt your private key from disk:",
		cfg.rootCfg.InsecurePasswordStdin,
	)
	if err != nil {
		return fmt.Errorf(
			"unable to retrieve decrypt password from user, %w",
			err,
		)
	}

	var (
		armor     string
		exportErr error
	)

	if cfg.unsafe {
		// Generate the unencrypted armor
		armor, exportErr = kb.ExportPrivKeyUnsafe(
			cfg.nameOrBech32,
			decryptPassword,
		)
	} else {
		// Get the armor encrypt password
		encryptPassword, err := io.GetCheckPassword(
			[2]string{
				"Enter a passphrase to encrypt your private key armor:",
				"Repeat the passphrase:",
			},
			cfg.rootCfg.InsecurePasswordStdin,
		)
		if err != nil {
			return fmt.Errorf(
				"unable to retrieve armor encrypt password from user, %w",
				err,
			)
		}

		// Generate the encrypted armor
		armor, exportErr = kb.ExportPrivKey(
			cfg.nameOrBech32,
			decryptPassword,
			encryptPassword,
		)
	}

	if exportErr != nil {
		return fmt.Errorf(
			"unable to export the private key, %w",
			exportErr,
		)
	}

	// Write the armor to disk
	if err := os.WriteFile(
		cfg.outputPath,
		[]byte(armor),
		0o644,
	); err != nil {
		return fmt.Errorf(
			"unable to write encrypted armor to file, %w",
			err,
		)
	}

	io.Printfln("Private key armor successfully outputted to %s", cfg.outputPath)

	return nil
}
