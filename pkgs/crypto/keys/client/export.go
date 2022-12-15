package client

import (
	"fmt"
	"os"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/errors"
)

var (
	errInvalidExportArgs = errors.New("invalid export arguments provided")
)

type ExportOptions struct {
	BaseOptions

	// The name or address of the private key to be exported
	NameOrBech32 string `flag:"key" help:"Name or Bech32 address of the private key"`

	// Output path for the encrypted private key armor
	OutputPath string `flag:"output-path" help:"The desired output path for the encrypted armor file"`
}

var DefaultExportOptions = ExportOptions{
	BaseOptions: DefaultBaseOptions,
}

// exportApp performs private key exports using the provided params
func exportApp(cmd *command.Command, _ []string, iopts interface{}) error {
	// Read the flag values
	opts, ok := iopts.(ExportOptions)
	if !ok {
		return errInvalidExportArgs
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

	// Get the key-base decrypt password
	decryptPassword, err := cmd.GetPassword(
		"Enter a passphrase to decrypt your private key from disk:",
		false,
	)
	if err != nil {
		return fmt.Errorf(
			"unable to retrieve decrypt password from user, %w",
			err,
		)
	}

	// Get the armor encrypt password
	encryptPassword, err := cmd.GetCheckPassword(
		"Enter a passphrase to encrypt your private key armor:",
		"Repeat the passphrase:")
	if err != nil {
		return fmt.Errorf(
			"unable to retrieve armor encrypt password from user, %w",
			err,
		)
	}

	// Generate the encrypted armor
	armor, err := kb.ExportPrivKey(
		opts.NameOrBech32,
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
		opts.OutputPath,
		[]byte(armor),
		0644,
	); err != nil {
		return fmt.Errorf(
			"unable to write encrypted armor to file, %w",
			err,
		)
	}

	cmd.Printfln("Encrypted private key armor successfully outputted to %s", opts.OutputPath)

	return nil
}
