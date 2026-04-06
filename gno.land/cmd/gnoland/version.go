package main

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/version"
)

func newVersionCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "version",
			ShortUsage: "version",
			ShortHelp:  "display the gnoland binary version",
		},
		nil,
		func(_ context.Context, _ []string) error {
			io.Println("gnoland version:", version.Version)
			return nil
		},
	)
}
