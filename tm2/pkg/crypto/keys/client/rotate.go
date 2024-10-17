package client

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

type RotateCfg struct {
	RootCfg *BaseCfg

	Force bool
}

func NewRotateCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &RotateCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "rotate",
			ShortUsage: "rotate [flags] <key-name>",
			ShortHelp:  "rotate the password of a key in the keybase to a new password",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execRotate(cfg, args, io)
		},
	)
}

func (c *RotateCfg) RegisterFlags(fs *flag.FlagSet) {
}

func execRotate(cfg *RotateCfg, args []string, io commands.IO) error {
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
	err = kb.Rotate(nameOrBech32, oldpass, getNewpass)
	if err != nil {
		return err
	}
	io.ErrPrintln("Password rotated")

	return nil
}
