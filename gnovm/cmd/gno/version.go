package main

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/version"
)

// newVersionCmd creates a new version command
func newGnoVersionCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "version",
			ShortUsage: "version",
			ShortHelp:  "display installed gno version",
		},
		nil,
		func(_ context.Context, args []string) error {
			io.Println("gno version:", version.Version)
			return nil
		},
	)
}
