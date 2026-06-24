// Package fork provides the `gnogenesis fork` subcommands for building and
// smoke-testing hardfork genesis files.
//
// A hardfork genesis is built from:
//  1. SOURCE CHAIN  — provides historical state (genesis + tx history)
//  2. NEW BINARY    — the updated gnoland built from this repo
//
// Source modes (auto-detected from --source):
//
//	http(s)://...    RPC of a running or recently-halted node
//	/path/to/dir     local node data directory (must contain config/genesis.json)
//	/path/to/file    exported file: genesis.json (no txs) or .jsonl (txs) or .tar.gz
package fork

import (
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// NewForkCmd returns the `gnogenesis fork` parent command with its
// subcommands (`generate`, `test`) attached.
func NewForkCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "fork",
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "build and smoke-test hardfork genesis files",
			LongHelp: `Build a hardfork genesis from a source chain and smoke-test it locally.

Subcommands:
  generate     Assemble a new-chain genesis.json from a source chain's state + tx history.
  test         Run an in-memory InitChain replay against a genesis.json (fast smoke-test).
  valoper-seed Build a deterministic .jsonl of valopers.Register migration txs from a CSV.
  addpkg       Build a .jsonl of MsgAddPackage migration txs from local package dirs.

Source modes (auto-detected from --source):
  http(s)://...    RPC of a running or recently-halted node
  /path/to/dir     local node data directory (must contain config/genesis.json)
  /path/to/file    exported file: genesis.json (no txs) or .jsonl (txs) or .tar.gz`,
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newGenerateCmd(io),
		newTestCmd(io),
		newValoperSeedCmd(io),
		newAddpkgCmd(io),
	)

	return cmd
}
