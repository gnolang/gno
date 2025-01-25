package main

import (
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newToolCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "tool",
			ShortUsage: "gno tool command [args...]",
			ShortHelp:  "run specified gno tool",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		// go equivalent commands:
		//
		// compile
		// transpile
		// pprof
		// trace
		// vet

		// gno specific commands:
		//
		// ast
		newLintCmd(io),
		// publish/release
		// render -- call render()?
		newReplCmd(),
		newTranspileCmd(io),
		// "vm" -- starts an in-memory chain that can be interacted with?
	)

	return cmd
}
