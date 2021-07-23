package client

import (
	"errors"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
)

type DeleteOptions struct {
	BaseOptions      // home, ...
	Yes         bool // skip confirmation prompt
	Force       bool // remove key unconditionally
}

var DefaultDeleteOptions = DeleteOptions{}

func runDeleteCmd(cmd *command.Command) error {
	var opts DeleteOptions = cmd.Options.(DeleteOptions)
	var args = cmd.Args

	name := args[0]

	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}

	info, err := kb.Get(name)
	if err != nil {
		return err
	}

	if info.GetType() == keys.TypeLedger || info.GetType() == keys.TypeOffline {
		if !opts.Yes {
			if err := confirmDeletion(cmd); err != nil {
				return err
			}
		}
		if err := kb.Delete(name, "", true); err != nil {
			return err
		}
		cmd.ErrPrintln("Public key reference deleted")
		return nil
	}

	// skip passphrase check if run with --force
	skipPass := opts.Force
	var oldpass string
	if !skipPass {
		if oldpass, err = cmd.GetPassword(
			"DANGER - enter password to permanently delete key:"); err != nil {
			return err
		}
	}

	err = kb.Delete(name, oldpass, skipPass)
	if err != nil {
		return err
	}
	cmd.ErrPrintln("Key deleted forever (uh oh!)")
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
