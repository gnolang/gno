package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gnolang/gno/pkgs/commands"
)

func main() {
	cmd := newGnodevCmd(commands.DefaultIO())

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v", err)

		os.Exit(1)
	}
}

func newGnodevCmd(io *commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Runs the gno development toolkit",
		},
		nil,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newRunCmd(io),
		newBuildCmd(io),
		newPrecompileCmd(io),
		newTestCmd(io),
		newReplCmd(),
		newModCmd(),
		// fmt -- gofmt
		// clean
		// graph
		// vendor -- download deps from the chain in vendor/
		// list -- list packages
		// render -- call render()?
		// publish/release
		// generate
		// doc -- godoc
		// "vm" -- starts an in-memory chain that can be interacted with?
		// bug -- start a bug report
		// version -- show gnodev, golang versions
	)

	return cmd
}
