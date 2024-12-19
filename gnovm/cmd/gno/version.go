package main

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

var buildVersion string

// newVersionCmd creates a new version command
func newVersionCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "version",
			ShortUsage: "version",
			ShortHelp:  "Display installed gno version",
		},
		nil,
		func(_ context.Context, args []string) error {
			version := getGnoVersion()
			io.Println("gno version:", version)
			return nil
		},
	)
}

func getGnoVersion() string {
	if buildVersion != "" {
		return buildVersion
	}
	return "unknown version"
}
