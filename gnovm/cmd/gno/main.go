package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	cmd := newGnocliCmd(commands.NewDefaultIO())

	cmd.Execute(context.Background(), os.Args[1:])
}

func newGnocliCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "gno <command> [arguments]",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newBugCmd(io),
		// build
		newCleanCmd(io),
		newDocCmd(io),
		newEnvCmd(io),
		// fix
		newFmtCmd(io),
		// generate
		// get
		// install
		// list -- list packages
		newModCmd(io),
		// work
		newRunCmd(io),
		// telemetry
		newTestCmd(io),
		newToolCmd(io),
		// version -- show cmd/gno, golang versions
		newGnoVersionCmd(io),
		// vet
	)

	return cmd
}
