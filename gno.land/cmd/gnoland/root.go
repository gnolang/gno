package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	cmd := newRootCmd(commands.NewDefaultIO())

	// Setup wait context to ensure correct cleanup on [interrupt] signal
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	cmd.Execute(ctx, os.Args[1:])
}

func newRootCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "manages the gnoland blockchain node",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newStartCmd(io),
		newSecretsCmd(io),
		newConfigCmd(io),
	)

	return cmd
}
