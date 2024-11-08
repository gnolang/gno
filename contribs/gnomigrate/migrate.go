package main

import (
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gnomigrate/internal/txs"
)

func newMigrateCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "gno migration suite",
			LongHelp:   "Gno state migration suite, for managing legacy headaches",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		txs.NewTxsCmd(io),
	)

	return cmd
}
