package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/pkgs/commands"
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
		"",
		commands.Metadata{
			Name:       "",
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Exports or imports transactions from the node",
		},
		cfg,
	)

	cmd.AddSubCommands(
		newImportCommand(cfg),
		newExportCommand(cfg),
	)

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)

		os.Exit(1)
	}
}

func (c *config) Exec(ctx context.Context, args []string) error {
	return commands.HelpExec(ctx, args)
}

func (c *config) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remote,
		"remote",
		"localhost:26657",
		"remote RPC address <addr:port>",
	)
}
