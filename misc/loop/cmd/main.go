package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	cliIO := commands.NewDefaultIO()

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "portalloopd",
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Portalloop commands interactions",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		NewBackupCmd(cliIO),
		NewServeCmd(cliIO),
		NewSwitchCmd(cliIO),
	)

	cmd.Execute(context.Background(), os.Args[1:])
}
