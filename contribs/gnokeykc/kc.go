package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/zalando/go-keyring"
)

const (
	kcService = "gnokey"
	kcName    = "encryption"
)

func newKcCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "kc",
			ShortUsage: "kc <command>",
			ShortHelp:  "manage OS keychain",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)
	cmd.AddSubCommands(
		newKcSetCmd(io),
		newKcUnsetCmd(io),
	)
	return cmd
}

func newKcSetCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "set",
			ShortUsage: "set",
			ShortHelp:  "set encryption password in OS keychain",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execKcSet(args, io)
		},
	)
}

func execKcSet(args []string, io commands.IO) error {
	if len(args) != 0 {
		return flag.ErrHelp
	}

	insecurePasswordStdin := false // XXX: cfg.rootCfg.InsecurePasswordStdin
	password, err := io.GetPassword("Enter password.", insecurePasswordStdin)
	if err != nil {
		return fmt.Errorf("cannot read password: %w", err)
	}

	err = keyring.Set(kcService, kcName, password)
	if err != nil {
		return fmt.Errorf("cannot set password is OS keychain")
	}

	io.Printfln("Successfully added password for key.")
	return nil
}

func newKcUnsetCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "unset",
			ShortUsage: "unset",
			ShortHelp:  "unset password in OS keychain",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execKcUnset(args, io)
		},
	)
}

func execKcUnset(args []string, io commands.IO) error {
	if len(args) != 0 {
		return flag.ErrHelp
	}

	err := keyring.Delete(kcService, kcName)
	if err != nil {
		return fmt.Errorf("cannot unset password from OS keychain")
	}

	io.Printfln("Successfully unset password")
	return nil
}
