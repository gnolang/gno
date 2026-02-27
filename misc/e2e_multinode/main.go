package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	cmd := newRootCmd(commands.NewDefaultIO())
	cmd.Execute(context.Background(), os.Args[1:])
}

func newRootCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "E2E Multi-Node Determinism Test",
			LongHelp:   "Standalone binary for testing multi-node consensus determinism",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newTestCmd(io),
	)

	return cmd
}
