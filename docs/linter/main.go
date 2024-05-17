package main

import (
	"context"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"os"
)

func main() {
	io := commands.NewDefaultIO()
	cmd := newRootCmd(io)

	cmd.Execute(context.Background(), os.Args[1:])
}

func newRootCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "The CLI for easy use of the r/blog realm",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	// Make Post command
	return commands.NewCommand(
		commands.Metadata{
			Name:       "post",
			ShortUsage: "post <FILE OR FILES_DIR> [flags]",
			LongHelp:   `Post one or more files. Passing in a file will post that single file, while passing in a directory will search for all README.md files and batch post them.`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execPost(io, args, cfg)
		},
	)

	cmd.Execute()

	cmd.AddSubCommands(
		newPostCommand(io),
	)

	return cmd
}
