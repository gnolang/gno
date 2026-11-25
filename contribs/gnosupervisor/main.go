// Command gnosupervisor runs a gnoland node across the coordinated upgrades
// recorded in a chain's upgrade ledger. See README.md.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	io := commands.NewDefaultIO()
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags]",
			LongHelp:   "Runs a gnoland node across the upgrades recorded in a chain's upgrade ledger (upgrades.json).",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)
	cmd.AddSubCommands(newRunCmd(io))
	cmd.Execute(ctx, os.Args[1:])
}
