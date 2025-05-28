package client

import (
	"context"
	"errors"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

type DeleteCfg struct {
	RootCfg *BaseCfg

	Yes   bool
	Force bool
}

func NewDeleteCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &DeleteCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "delete",
			ShortUsage: "delete [flags] <key-name>",
			ShortHelp:  "deletes a key from the keybase",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execDelete(cfg, args, io)
		},
	)
}

func (c *DeleteCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.Yes,
		"yes",
		false,
		"skip confirmation prompt",
	)

	fs.BoolVar(
		&c.Force,
		"force",
		false,
		"remove key unconditionally",
	)
}

func execDelete(cfg *DeleteCfg, args []string, io commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	if cfg.RootCfg.Json {
		io.ErrPrintln("warning: -json flag has no effect on `delete` command")
	}

	nameOrBech32 := args[0]

	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.Home)
	if err != nil {
		return err
	}

	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}

	if info.GetType() == keys.TypeLedger || info.GetType() == keys.TypeOffline {
		if !cfg.Yes {
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
	skipPass := cfg.Force
	var oldpass string
	if !skipPass {
		msg := "DANGER - enter password to permanently delete key:"
		if oldpass, err = io.GetPassword(msg, cfg.RootCfg.InsecurePasswordStdin); err != nil {
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

func confirmDeletion(io commands.IO) error {
	answer, err := io.GetConfirmation("Key reference will be deleted. Continue?")
	if err != nil {
		return err
	}

	if !answer {
		return errors.New("aborted")
	}

	return nil
}
