package client

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

type UpdateCfg struct {
	RootCfg *BaseCfg

	Force bool
}

func NewUpdateCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &UpdateCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "update",
			ShortUsage: "update [flags] <key-name>",
			ShortHelp:  "update the password of a key in the keybase",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execUpdate(cfg, args, io)
		},
	)
}

func (c *UpdateCfg) RegisterFlags(fs *flag.FlagSet) {
}

func execUpdate(cfg *UpdateCfg, args []string, io commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	nameOrBech32 := args[0]

	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.Home)
	if err != nil {
		return err
	}

	oldpass, err := io.GetPassword("Enter the current password:", cfg.RootCfg.InsecurePasswordStdin)
	if err != nil {
		return err
	}

	newpass, err := io.GetCheckPassword(
		[2]string{
			"Enter the new password to encrypt your key to disk:",
			"Repeat the password:",
		},
		cfg.RootCfg.InsecurePasswordStdin,
	)
	if err != nil {
		return fmt.Errorf("unable to parse provided password, %w", err)
	}

	getNewpass := func() (string, error) { return newpass, nil }
	err = kb.Update(nameOrBech32, oldpass, getNewpass)
	if err != nil {
		return err
	}
	io.ErrPrintln("Password updated")

	return nil
}
