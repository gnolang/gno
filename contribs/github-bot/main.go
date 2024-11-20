package main

import (
	"context"
	"os"

	p "github.com/gnolang/gno/contribs/github-bot/params"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	params := &p.Params{}

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "[flags]",
			ShortHelp:  "checks requirements for a PR to be merged",
			LongHelp:   "This tool checks if the requirements for a PR to be merged are satisfied (defined in config.go) and displays PR status checks accordingly.\nA valid GitHub Token must be provided by setting the GITHUB_TOKEN environment variable.",
		},
		params,
		func(_ context.Context, _ []string) error {
			params.ValidateFlags()
			return execBot(params)
		},
	)

	cmd.Execute(context.Background(), os.Args[1:])
}
