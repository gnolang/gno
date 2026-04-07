package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newRemotesCmd(base *baseCfg, io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "remotes",
			ShortUsage: "gnopie remotes <subcommand>",
			ShortHelp:  "Manage network configurations.",
		},
		nil,
		commands.HelpExec,
	)
	cmd.AddSubCommands(
		newRemotesListCmd(base, io),
		newRemotesAddCmd(base, io),
		newRemotesRmCmd(base, io),
		newRemotesUpdateCmd(base, io),
		newRemotesDefaultCmd(base, io),
	)
	return cmd
}

func newRemotesListCmd(base *baseCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{Name: "list", ShortUsage: "gnopie remotes list", ShortHelp: "List configured remotes."},
		nil,
		func(_ context.Context, _ []string) error {
			cfg, err := LoadRemotes(base.home)
			if err != nil {
				return err
			}
			if base.jsonOut {
				return outputJSON(io, cfg.Remotes)
			}
			io.Printfln("%s", cfg.FormatTable())
			return nil
		},
	)
}

type remotesAddCfg struct {
	rpc, chainID, indexer string
}

func (c *remotesAddCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.rpc, "rpc", "", "RPC endpoint URL (required)")
	fs.StringVar(&c.chainID, "chain-id", "", "chain ID (required)")
	fs.StringVar(&c.indexer, "indexer", "", "indexer GraphQL URL")
}

func newRemotesAddCmd(base *baseCfg, io commands.IO) *commands.Command {
	cfg := &remotesAddCfg{}
	return commands.NewCommand(
		commands.Metadata{Name: "add", ShortUsage: "gnopie remotes add --rpc <url> --chain-id <id> <name>", ShortHelp: "Add a new remote."},
		cfg,
		func(_ context.Context, args []string) error {
			if len(args) != 1 || cfg.rpc == "" || cfg.chainID == "" {
				return fmt.Errorf("usage: gnopie remotes add --rpc <url> --chain-id <id> <name>")
			}
			rcfg, err := LoadRemotes(base.home)
			if err != nil {
				return err
			}
			if err := rcfg.Add(args[0], cfg.rpc, cfg.chainID, cfg.indexer); err != nil {
				return err
			}
			if err := rcfg.Save(); err != nil {
				return err
			}
			io.Printfln("Added remote %q", args[0])
			return nil
		},
	)
}

func newRemotesRmCmd(base *baseCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{Name: "rm", ShortUsage: "gnopie remotes rm <name>", ShortHelp: "Remove a remote."},
		nil,
		func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("usage: gnopie remotes rm <name>")
			}
			cfg, err := LoadRemotes(base.home)
			if err != nil {
				return err
			}
			if err := cfg.Remove(args[0]); err != nil {
				return err
			}
			return cfg.Save()
		},
	)
}

type remotesUpdateCfg struct {
	rpc, chainID, indexer string
}

func (c *remotesUpdateCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.rpc, "rpc", "", "new RPC URL")
	fs.StringVar(&c.chainID, "chain-id", "", "new chain ID")
	fs.StringVar(&c.indexer, "indexer", "", "new indexer URL")
}

func newRemotesUpdateCmd(base *baseCfg, io commands.IO) *commands.Command {
	cfg := &remotesUpdateCfg{}
	return commands.NewCommand(
		commands.Metadata{Name: "update", ShortUsage: "gnopie remotes update [flags] <name>", ShortHelp: "Update a remote."},
		cfg,
		func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("usage: gnopie remotes update <name> [--rpc <url>] [--chain-id <id>]")
			}
			rcfg, err := LoadRemotes(base.home)
			if err != nil {
				return err
			}
			if err := rcfg.Update(args[0], cfg.rpc, cfg.chainID, cfg.indexer); err != nil {
				return err
			}
			return rcfg.Save()
		},
	)
}

func newRemotesDefaultCmd(base *baseCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{Name: "default", ShortUsage: "gnopie remotes default <name>", ShortHelp: "Set the default remote."},
		nil,
		func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("usage: gnopie remotes default <name>")
			}
			cfg, err := LoadRemotes(base.home)
			if err != nil {
				return err
			}
			if err := cfg.SetDefault(args[0]); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			io.Printfln("Default set to %q", args[0])
			return nil
		},
	)
}
