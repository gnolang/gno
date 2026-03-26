package main

import (
	"github.com/gnolang/contribs/gnoupgrade/internal/healthcheck"
	"github.com/gnolang/contribs/gnoupgrade/internal/replay"
	"github.com/gnolang/contribs/gnoupgrade/internal/statediff"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newUpgradeCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "chain upgrade testing and verification toolkit",
			LongHelp:   "Tools for testing and verifying gno.land chain upgrades: replay smoke tests, state diffs, and health checks.",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		replay.NewReplayCmd(io),
		statediff.NewStateDiffCmd(io),
		healthcheck.NewHealthCheckCmd(io),
	)

	return cmd
}
