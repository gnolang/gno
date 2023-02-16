package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/pkgs/commands"
	"github.com/gnolang/gno/pkgs/crypto/keys"
)

type deleteCfg struct {
	rootCfg *baseCfg

	yes   bool
	force bool
}

func newDeleteCmd(rootCfg *baseCfg) *commands.Command {
	cfg := &deleteCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "delete",
			ShortUsage: "delete [flags] <key-name>",
			ShortHelp:  "Deletes a key from the keybase",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execDelete(cfg, args, bufio.NewReader(os.Stdin))
		},
	)
}

func (c *deleteCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.yes,
		"yes",
		false,
		"skip confirmation prompt",
	)

	fs.BoolVar(
		&c.force,
		"force",
		false,
		"remove key unconditionally",
	)
}

func execDelete(cfg *deleteCfg, args []string, input *bufio.Reader) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	nameOrBech32 := args[0]

	kb, err := keys.NewKeyBaseFromDir(cfg.rootCfg.Home)
	if err != nil {
		return err
	}

	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}

	if info.GetType() == keys.TypeLedger || info.GetType() == keys.TypeOffline {
		if !cfg.yes {
			if err := confirmDeletion(input); err != nil {
				return err
			}
		}

		if err := kb.Delete(nameOrBech32, "", true); err != nil {
			return err
		}
		fmt.Println("Public key reference deleted")

		return nil
	}

	// skip passphrase check if run with --force
	skipPass := cfg.force
	var oldpass string
	if !skipPass {
		msg := "DANGER - enter password to permanently delete key:"
		if oldpass, err = commands.GetPassword(msg, cfg.rootCfg.InsecurePasswordStdin, input); err != nil {
			return err
		}
	}

	err = kb.Delete(nameOrBech32, oldpass, skipPass)
	if err != nil {
		return err
	}
	fmt.Println("Key deleted")

	return nil
}

func confirmDeletion(input *bufio.Reader) error {
	answer, err := commands.GetConfirmation(
		"Key reference will be deleted. Continue?",
		input,
	)

	if err != nil {
		return err
	}

	if !answer {
		return errors.New("aborted")
	}

	return nil
}
