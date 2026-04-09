package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	cmd := newRootCmd(commands.NewDefaultIO())
	cmd.Execute(context.Background(), os.Args[1:])
}

func newRootCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "hardfork",
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "gnoland hardfork tooling",
			LongHelp: `Tools for preparing and testing gnoland chain hardforks.

A hardfork genesis has three components:
  1. SOURCE CHAIN  — provides historical state (genesis + tx history)
  2. NEW BINARY    — the updated gnoland built from this repo
  3. OVERLAY       — a directory of scripts applied to genesis before tx replay

Source modes (auto-detected from --source):
  http(s)://...    RPC of a running or recently-halted node
  /path/to/dir     local node data directory (must contain config/genesis.json)
  /path/to/file    exported file: genesis.json (no txs) or .jsonl (txs) or .tar.gz`,
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newGenesisCmd(io),
	)

	return cmd
}
