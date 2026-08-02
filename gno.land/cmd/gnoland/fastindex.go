package main

import (
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newFastindexCmd creates the fastindex root command: offline maintenance and
// auditing of the bptree store's fast index (see tm2/pkg/bptree).
func newFastindexCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "fastindex",
			ShortUsage: "fastindex <subcommand> [flags]",
			ShortHelp:  "bptree fast-index maintenance and auditing",
			LongHelp: "Offline tools for the bptree store's fast index (an advisory read " +
				"accelerator, not part of the app hash). Run against a STOPPED node's " +
				"data directory or a captured copy.",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newFastindexVerifyCmd(io),
	)

	return cmd
}
