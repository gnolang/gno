package main

import (
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newInitCmd creates the new secrets init command
func newInitCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "init [subcommand] [flags]",
			ShortHelp:  "Initializes the Gno node secrets",
			LongHelp:   "Initializes the Gno node secrets locally, including the validator key, validator state and node key",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newInitAllCmd(io),
		// newInitSingleCmd(io),
	)

	return cmd
}
