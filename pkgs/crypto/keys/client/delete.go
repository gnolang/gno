package client

import (
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/errors"
)

type DeleteOptions struct {
	BaseOptions
	Yes   bool `flag:"yes" help:"skip confirmation prompt"`
	Force bool `flag:"force" help:"remove key unconditionally"`
}

var DefaultDeleteOptions = DeleteOptions{
	BaseOptions: DefaultBaseOptions,
}

func deleteApp(cmd *command.Command, args []string, iopts interface{}) error {
	var opts DeleteOptions = iopts.(DeleteOptions)

	if len(args) != 1 {
		cmd.ErrPrintfln("Usage: delete <keyname or address>")
		return errors.New("invalid args")
	}

	nameOrBech32 := args[0]

	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}

	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}

	if info.GetType() == keys.TypeLedger || info.GetType() == keys.TypeOffline {
		if !opts.Yes {
			if err := confirmDeletion(cmd); err != nil {
				return err
			}
		}
		if err := kb.Delete(nameOrBech32, "", true); err != nil {
			return err
		}
		cmd.ErrPrintln("Public key reference deleted")
		return nil
	}

	// skip passphrase check if run with --force
	skipPass := opts.Force
	var oldpass string
	if !skipPass {
		msg := "DANGER - enter password to permanently delete key:"
		if oldpass, err = cmd.GetPassword(msg, false); err != nil {
			return err
		}
	}

	err = kb.Delete(nameOrBech32, oldpass, skipPass)
	if err != nil {
		return err
	}
	cmd.ErrPrintln("Key deleted")
	return nil
}

func confirmDeletion(cmd *command.Command) error {
	answer, err := cmd.GetConfirmation("Key reference will be deleted. Continue?")
	if err != nil {
		return err
	}
	if !answer {
		return errors.New("aborted")
	}
	return nil
}
