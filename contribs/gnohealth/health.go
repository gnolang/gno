package main

import (
	"github.com/gnolang/gno/contribs/gnohealth/internal/timestamp"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newHealthCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "gno health check suite",
			LongHelp:   "Gno health check suite, to verify that different parts of Gno are working correctly",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		timestamp.NewTimestampCmd(io),
	)

	return cmd
}
