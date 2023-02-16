package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/pkgs/commands"
	"github.com/gnolang/gno/pkgs/crypto/keys"
)

type exportCfg struct {
	rootCfg *baseCfg

	nameOrBech32 string
	outputPath   string
}

func newExportCmd(rootCfg *baseCfg) *commands.Command {
	cfg := &exportCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "export",
			ShortUsage: "export [flags]",
			ShortHelp:  "Exports encrypted private key armor",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execExport(cfg, args, bufio.NewReader(os.Stdin))
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
		"The desired output path for the encrypted armor file",
	)
}

func execExport(cfg *exportCfg, args []string, input *bufio.Reader) error {
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
	decryptPassword, err := commands.GetPassword(
		"Enter a passphrase to decrypt your private key from disk:",
		cfg.rootCfg.InsecurePasswordStdin,
		input,
	)
	if err != nil {
		return fmt.Errorf(
			"unable to retrieve decrypt password from user, %w",
			err,
		)
	}

	// Get the armor encrypt password
	encryptPassword, err := commands.GetCheckPassword(
		"Enter a passphrase to encrypt your private key armor:",
		"Repeat the passphrase:",
		cfg.rootCfg.InsecurePasswordStdin,
		input,
	)
	if err != nil {
		return fmt.Errorf(
			"unable to retrieve armor encrypt password from user, %w",
			err,
		)
	}

	// Generate the encrypted armor
	armor, err := kb.ExportPrivKey(
		cfg.nameOrBech32,
		decryptPassword,
		encryptPassword,
	)
	if err != nil {
		return fmt.Errorf(
			"unable to export the private key, %w",
			err,
		)
	}

	// Write the encrypted armor to disk
	if err := os.WriteFile(
		cfg.outputPath,
		[]byte(armor),
		0644,
	); err != nil {
		return fmt.Errorf(
			"unable to write encrypted armor to file, %w",
			err,
		)
	}

	fmt.Printf("Encrypted private key armor successfully outputted to %s\n", cfg.outputPath)

	return nil
}
