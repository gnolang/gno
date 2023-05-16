package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

// config is the shared config for gnotxport, and its subcommands
type config struct {
	remote string `default:"localhost:26657"`
}

const (
	defaultFilePath = "txexport.log"
)

func main() {
	cfg := &config{}

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Exports or imports transactions from the node",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newImportCommand(cfg),
		newExportCommand(cfg),
	)

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v", err)

		os.Exit(1)
	}
}

func (c *config) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remote,
		"remote",
		"localhost:26657",
		"remote RPC address <addr:port>",
	)
}
