package main

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

const version = "0.1.0-dev"

func newVersionCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "version",
			ShortUsage: "gnosh version",
			ShortHelp:  "Print version information.",
		},
		nil,
		func(_ context.Context, _ []string) error {
			io.Printfln("gnosh %s", version)
			return nil
		},
	)
}
