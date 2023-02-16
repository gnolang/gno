package client

import (
	"crypto/sha256"
	"fmt"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/bip39"
	"github.com/gnolang/gno/pkgs/errors"
)

type GenerateOptions struct {
	CustomEntropy bool `flag:"entropy" help:"custom entropy"`
}

var DefaultGenerateOptions = GenerateOptions{
	// BaseOptions: DefaultBaseOptions,
}

func generateApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(GenerateOptions)
	customEntropy := opts.CustomEntropy

	if len(args) != 0 {
		cmd.ErrPrintfln("Usage: generate (no args)")
		return errors.New("invalid args")
	}

	var entropySeed []byte

	if customEntropy {
		// prompt the user to enter some entropy
		inputEntropy, err := cmd.GetString("WARNING: Generate at least 256-bits of entropy and enter the results here:")
		if err != nil {
			return err
		}
		if len(inputEntropy) < 43 {
			return fmt.Errorf("256-bits is 43 characters in Base-64, and 100 in Base-6. You entered %v, and probably want more", len(inputEntropy))
		}
		conf, err := cmd.GetConfirmation(fmt.Sprintf("Input length: %d", len(inputEntropy)))
		if err != nil {
			return err
		}
		if !conf {
			return nil
		}

		// hash input entropy to get entropy seed
		hashedEntropy := sha256.Sum256([]byte(inputEntropy))
		entropySeed = hashedEntropy[:]
	} else {
		// read entropy seed straight from crypto.Rand
		var err error
		entropySeed, err = bip39.NewEntropy(MnemonicEntropySize)
		if err != nil {
			return err
		}
	}

	mnemonic, err := bip39.NewMnemonic(entropySeed[:])
	if err != nil {
		return err
	}
	cmd.Println(mnemonic)

	return nil
}
