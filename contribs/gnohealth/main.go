package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/contribs/gnohealth/internal/timestamp"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags]",
			LongHelp:   "Gno health check suite, to verify that different parts of Gno are working correctly",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	io := commands.NewDefaultIO()
	cmd.AddSubCommands(
		timestamp.NewTimestampCmd(io),
	)

	cmd.Execute(context.Background(), os.Args[1:])
}
