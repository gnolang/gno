package client

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/armor"
)

type ExportCfg struct {
	RootCfg *BaseCfg

	NameOrBech32 string
	OutputPath   string
}

func NewExportCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &ExportCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "export",
			ShortUsage: "export [flags]",
			ShortHelp:  "exports private key armor",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execExport(cfg, io)
		},
	)
}

func (c *ExportCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.NameOrBech32,
		"key",
		"",
		"name or bech32 address of the private key",
	)

	fs.StringVar(
		&c.OutputPath,
		"output-path",
		"",
		"the desired output path for the armor file",
	)
}

func execExport(cfg *ExportCfg, io commands.IO) error {
	// check keyname
	if cfg.NameOrBech32 == "" {
		return errors.New("key to be exported shouldn't be empty")
	}

	// Create a new instance of the key-base
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.Home)
	if err != nil {
		return fmt.Errorf(
			"unable to create a key base from directory %s, %w",
			cfg.RootCfg.Home,
			err,
		)
	}

	// Get the key-base decrypt password
	decryptPassword, err := io.GetPassword(
		"Enter a passphrase to decrypt your private key from disk:",
		cfg.RootCfg.InsecurePasswordStdin,
	)
	if err != nil {
		return fmt.Errorf(
			"unable to retrieve decrypt password from user, %w",
			err,
		)
	}

	var keyArmor string

	// Export the private key from the keybase
	privateKey, err := kb.ExportPrivKey(cfg.NameOrBech32, decryptPassword)
	if err != nil {
		return fmt.Errorf("unable to export private key, %w", err)
	}

	// Get the armor encrypt password
	pw, err := promptPassphrase(io, cfg.RootCfg.InsecurePasswordStdin)
	if err != nil {
		return err
	}

	keyArmor = armor.EncryptArmorPrivKey(privateKey, pw)

	// Write the armor to disk
	if err := os.WriteFile(
		cfg.OutputPath,
		[]byte(keyArmor),
		0o644,
	); err != nil {
		return fmt.Errorf(
			"unable to write encrypted armor to file, %w",
			err,
		)
	}

	io.Printfln("Key armor successfully saved to %s", cfg.OutputPath)

	return nil
}
