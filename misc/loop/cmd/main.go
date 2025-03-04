package main

import (
	"context"
	"os"

	cmd_ "loop/cmd/cmd"

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
		cmd_.NewBackupCmd(cliIO),
		cmd_.NewServeCmd(cliIO),
		cmd_.NewSwitchCmd(cliIO),
	)

	cmd.Execute(context.Background(), os.Args[1:])
}
