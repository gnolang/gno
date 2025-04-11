package client

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
)

type GenerateCfg struct {
	RootCfg *BaseCfg

	CustomEntropy bool
}

func NewGenerateCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &GenerateCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "generate",
			ShortUsage: "generate [flags]",
			ShortHelp:  "generates a bip39 mnemonic",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execGenerate(cfg, args, io)
		},
	)
}

func (c *GenerateCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.CustomEntropy,
		"entropy",
		false,
		"supply custom entropy",
	)
}

func execGenerate(cfg *GenerateCfg, args []string, io commands.IO) error {
	customEntropy := cfg.CustomEntropy

	if len(args) != 0 {
		return flag.ErrHelp
	}

	if cfg.RootCfg.Json {
		io.ErrPrintln("warning: -json flag has no effect on `generate` command")
	}

	var entropySeed []byte

	if customEntropy {
		// prompt the user to enter some entropy
		inputEntropy, err := io.GetString(
			"WARNING: Generate at least 256-bits of entropy and enter the results here:",
		)
		if err != nil {
			return err
		}
		if len(inputEntropy) < 43 {
			return fmt.Errorf("256-bits is 43 characters in Base-64, and 100 in Base-6. You entered %v, and probably want more", len(inputEntropy))
		}
		conf, err := io.GetConfirmation(
			fmt.Sprintf("Input length: %d", len(inputEntropy)),
		)
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
		entropySeed, err = bip39.NewEntropy(mnemonicEntropySize)
		if err != nil {
			return err
		}
	}

	mnemonic, err := bip39.NewMnemonic(entropySeed[:])
	if err != nil {
		return err
	}

	io.Println(mnemonic)

	return nil
}
