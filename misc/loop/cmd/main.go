package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type service struct {
	// TODO(albttx): put getter on it with RMutex
	portalLoop *snapshotter

	portalLoopURL string
}

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
		newServeCmd(cliIO),
		newBackupCmd(cliIO),
		newSwitchCmd(cliIO),
	)

	cmd.Execute(context.Background(), os.Args[1:])
}
