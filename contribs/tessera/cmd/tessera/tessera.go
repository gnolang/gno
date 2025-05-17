package main

import (
	"github.com/gnolang/gno/contribs/tessera/cmd/tessera/run"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newTesseraCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "gno.land testing harness, minus the bullshit",
			LongHelp:   "gno.land testing harness, for executing live cluster tests",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		run.NewRunCmd(io),
		// list.NewListCmd(io),
		// validate.NewValidateCmd(io),
		// create.NewCreateCmd(io),
	)

	return cmd
}
