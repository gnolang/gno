package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	stdio := commands.NewDefaultIO()

	cfg := LocalAppConfig{}
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "gnodev",
			ShortUsage: "gnodev <cmd> [flags] ",
			ShortHelp:  "Runs an in-memory node and gno.land web server for development purposes.",
			LongHelp:   `The gnodev command starts an in-memory node and a gno.land web interface primarily for realm package development. It automatically loads the 'examples' directory and any additional specified paths.`,
		},
		&cfg,
		func(_ context.Context, args []string) error {
			return execLocalApp(&cfg, args, stdio)
		},
	)

	// cmd.AddSubCommands(NewLocalCmd(stdio))
	cmd.AddSubCommands(NewStagingCmd(stdio))

	cmd.Execute(context.Background(), os.Args[1:])
}
