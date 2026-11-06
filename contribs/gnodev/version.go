package main

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/version"
)

// newVersionCmd creates a new version command
func newVersionCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "version",
			ShortUsage: "version",
			ShortHelp:  "display installed gnodev version",
		},
		nil,
		func(_ context.Context, _ []string) error {
			io.Println("gnodev version:", version.Version)
			return nil
		},
	)
}
