package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/gnovm/cmd/gno/internal/pkgdownload"
	"github.com/gnolang/gno/gnovm/cmd/gno/internal/pkgdownload/gnopkgfetcher"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	cmd := newGnocliCmd(commands.NewDefaultIO(), gnopkgfetcher.New())

	cmd.Execute(context.Background(), os.Args[1:])
}

func newGnocliCmd(io commands.IO, packageFetcher pkgdownload.PackageFetcher) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Runs the gno development toolkit",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newModCmd(io, packageFetcher),
		newTestCmd(io),
		newLintCmd(io),
		newRunCmd(io),
		newTranspileCmd(io),
		newCleanCmd(io),
		newReplCmd(),
		newDocCmd(io),
		newEnvCmd(io),
		newBugCmd(io),
		newFmtCmd(io),
		// graph
		// vendor -- download deps from the chain in vendor/
		// list -- list packages
		// render -- call render()?
		// publish/release
		// generate
		// "vm" -- starts an in-memory chain that can be interacted with?
		// version -- show cmd/gno, golang versions
	)

	return cmd
}
