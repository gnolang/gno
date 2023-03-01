package client

import (
	"errors"
	"fmt"
	"os"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
)

var errInvalidImportArgs = errors.New("invalid import arguments provided")

type ImportOptions struct {
	BaseOptions

	// Name of the private key in the key-base
	KeyName string `flag:"name" help:"The name of the private key"`

	// Path to the encrypted private key armor
	ArmorPath string `flag:"armor-path" help:"The path to the encrypted armor file"`

	// Unsafe flag for specifying the input as unencrypted
	Unsafe bool `flag:"unsafe" help:"Import the private key armor as unencrypted"`
}

var DefaultImportOptions = ImportOptions{
	BaseOptions: DefaultBaseOptions,
}

// importApp performs private key imports using the provided params
func importApp(cmd *command.Command, _ []string, iopts interface{}) error {
	// Read the flag values
	opts, ok := iopts.(ImportOptions)
	if !ok {
		return errInvalidImportArgs
	}

	// Create a new instance of the key-base
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return fmt.Errorf(
			"unable to create a key base from directory %s, %w",
			opts.Home,
			err,
		)
	}

	// Read the raw encrypted armor
	armor, err := os.ReadFile(opts.ArmorPath)
	if err != nil {
		return fmt.Errorf(
			"unable to read armor from path %s, %w",
			opts.ArmorPath,
			err,
		)
	}

	var (
		decryptPassword string
		encryptPassword string
	)

	if !opts.Unsafe {
		// Get the armor decrypt password
		decryptPassword, err = cmd.GetPassword(
			"Enter a passphrase to decrypt your private key armor:",
			false,
		)
		if err != nil {
			return fmt.Errorf(
				"unable to retrieve armor decrypt password from user, %w",
				err,
			)
		}
	}

	// Get the key-base encrypt password
	encryptPassword, err = cmd.GetCheckPassword(
		"Enter a passphrase to encrypt your private key:",
		"Repeat the passphrase:")
	if err != nil {
		return fmt.Errorf(
			"unable to retrieve key encrypt password from user, %w",
			err,
		)
	}

	if opts.Unsafe {
		// Import the unencrypted private key
		if err := kb.ImportPrivKeyUnsafe(
			opts.KeyName,
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
			opts.KeyName,
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

	cmd.Printfln("Successfully imported private key %s", opts.KeyName)

	return nil
}
