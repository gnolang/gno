package client

import (
	"context"
	"errors"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
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
			return execDelete(cfg, args, commands.NewDefaultIO())
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

func execDelete(cfg *deleteCfg, args []string, io *commands.IO) error {
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
			if err := confirmDeletion(io); err != nil {
				return err
			}
		}

		if err := kb.Delete(nameOrBech32, "", true); err != nil {
			return err
		}
		io.ErrPrintln("Public key reference deleted")

		return nil
	}

	// skip passphrase check if run with --force
	skipPass := cfg.force
	var oldpass string
	if !skipPass {
		msg := "DANGER - enter password to permanently delete key:"
		if oldpass, err = io.GetPassword(msg, cfg.rootCfg.InsecurePasswordStdin); err != nil {
			return err
		}
	}

	err = kb.Delete(nameOrBech32, oldpass, skipPass)
	if err != nil {
		return err
	}
	io.ErrPrintln("Key deleted")

	return nil
}

func confirmDeletion(io *commands.IO) error {
	answer, err := io.GetConfirmation("Key reference will be deleted. Continue?")
	if err != nil {
		return err
	}

	if !answer {
		return errors.New("aborted")
	}

	return nil
}
