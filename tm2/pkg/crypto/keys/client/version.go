package client

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/version"
)

func NewVersionCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "version",
			ShortUsage: "version",
			ShortHelp:  "display the gnokey binary version",
			LongHelp:   "Displays detailed version information for the gnokey binary, including the build number, commit hash, and build date. Useful for verifying the exact version of gnokey you are running.",
		},
		nil,
		func(_ context.Context, _ []string) error {
			io.Println("gnokey version:", version.Version)
			return nil
		},
	)
}
