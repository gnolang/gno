package main

import (
	"context"
	"os"
	"runtime/pprof"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	cpup := os.Getenv("CPUPROFILE")
	if cpup != "" {
		f, err := os.Create(cpup)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		err = pprof.StartCPUProfile(f)
		if err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}

	cmd := newGnocliCmd(commands.NewDefaultIO())

	cmd.Execute(context.Background(), os.Args[1:])
}

func newGnocliCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "gno <command> [arguments]",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newBugCmd(io),
		// build
		newCleanCmd(io),
		newDocCmd(io),
		newEnvCmd(io),
		newFixCmd(io),
		newFmtCmd(io),
		// generate
		// get
		// install
		// list -- list packages
		newLintCmd(io),
		newModCmd(io),
		// work
		newPprofCmd(io),
		newReplCmd(),
		newRunCmd(io),
		// telemetry
		newTestCmd(io),
		newToolCmd(io),
		// version -- show cmd/gno, golang versions
		newGnoVersionCmd(io),
		// vet
	)

	return cmd
}
