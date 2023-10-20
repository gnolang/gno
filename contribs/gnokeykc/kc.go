package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/zalando/go-keyring"
)

func newKcCmd(io *commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "kc",
			ShortUsage: "kc <command>",
			ShortHelp:  "Manage OS keychain",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)
	cmd.AddSubCommands(
		newKcSetCmd(io),
		newKcDeleteCmd(io),
	)
	return cmd
}

func newKcSetCmd(io *commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "set",
			ShortUsage: "set <name>",
			ShortHelp:  "set password for name in OS keychain",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execKcSet(args, io)
		},
	)
}

func execKcSet(args []string, io *commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	// XXX: check if name is an already existing key
	name := strings.TrimSpace(args[0])
	if name == "" {
		return flag.ErrHelp
	}

	insecurePasswordStdin := false // XXX: cfg.rootCfg.InsecurePasswordStdin
	password, err := io.GetPassword("Enter password.", insecurePasswordStdin)
	if err != nil {
		return fmt.Errorf("cannot read password: %w", err)
	}

	err = keyring.Set("gnokey", name, password)
	if err != nil {
		return fmt.Errorf("cannot set password is OS keychain")
	}

	io.Printfln("Successfully added password for key %q.", name)
	return nil
}

func newKcDeleteCmd(io *commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "delete",
			ShortUsage: "delete <name>",
			ShortHelp:  "delete password for name in OS keychain",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execKcDelete(args, io)
		},
	)
}

func execKcDelete(args []string, io *commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	name := strings.TrimSpace(args[0])
	if name == "" {
		return flag.ErrHelp
	}

	err := keyring.Delete("gnokey", name)
	if err != nil {
		return fmt.Errorf("cannot delete password from OS keychain")
	}

	io.Printfln("Successfully deleted password for key %q.", name)
	return nil
}
