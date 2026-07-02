package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newConfigCmd(base *baseCfg, io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "config",
			ShortUsage: "gnopie config <subcommand>",
			ShortHelp:  "Manage gnopie configuration.",
			LongHelp: `Manage persistent gnopie settings.

  gnopie config set key=moul    Set default key
  gnopie config get key          Get a config value
  gnopie config list             Show all settings`,
		},
		nil,
		commands.HelpExec,
	)
	cmd.AddSubCommands(
		newConfigSetCmd(base, io),
		newConfigGetCmd(base, io),
		newConfigListCmd(base, io),
	)
	return cmd
}

func newConfigSetCmd(base *baseCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "set",
			ShortUsage: "gnopie config set <key>=<value>",
			ShortHelp:  "Set a config value.",
		},
		nil,
		func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("usage: gnopie config set <key>=<value>")
			}
			parts := strings.SplitN(args[0], "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("usage: gnopie config set <key>=<value>")
			}
			cfg, err := LoadConfig(base.home)
			if err != nil {
				return err
			}
			if err := ConfigSet(cfg, parts[0], parts[1]); err != nil {
				return err
			}
			if err := SaveConfig(base.home, cfg); err != nil {
				return err
			}
			io.Printfln("%s=%s", parts[0], parts[1])
			return nil
		},
	)
}

func newConfigGetCmd(base *baseCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "get",
			ShortUsage: "gnopie config get <key>",
			ShortHelp:  "Get a config value.",
		},
		nil,
		func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("usage: gnopie config get <key>")
			}
			cfg, err := LoadConfig(base.home)
			if err != nil {
				return err
			}
			val, err := ConfigGet(cfg, args[0])
			if err != nil {
				return err
			}
			io.Println(val)
			return nil
		},
	)
}

func newConfigListCmd(base *baseCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "list",
			ShortUsage: "gnopie config list",
			ShortHelp:  "Show all config settings.",
		},
		nil,
		func(_ context.Context, _ []string) error {
			cfg, err := LoadConfig(base.home)
			if err != nil {
				return err
			}
			io.Printfln("%s", ConfigList(cfg))
			return nil
		},
	)
}
