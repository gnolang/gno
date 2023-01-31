package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
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

	fs := flag.NewFlagSet("", flag.ExitOnError)
	cfg.registerFlags(fs)

	cmd := &ffcli.Command{
		Name:       "",
		ShortUsage: "<subcommand> [flags] [<arg>...]",
		LongHelp:   "Exports or imports transactions from the node",
		FlagSet:    fs,
		Exec: func(_ context.Context, _ []string) error {
			return flag.ErrHelp
		},
	}

	cmd.Subcommands = []*ffcli.Command{
		newImportCommand(cfg),
		newExportCommand(cfg),
	}

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)

		os.Exit(1)
	}
}

func (c *config) registerFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remote,
		"remote",
		"localhost:26657",
		"remote RPC address <addr:port>",
	)
}
