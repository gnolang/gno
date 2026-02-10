package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
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

	io := commands.NewDefaultIO()
	cmd, cfg := newGnocliCmd(io)

	args := os.Args[1:]

	if err := cmd.Parse(args); err != nil {
		handleError(err)
		return
	}

	// Apply directory change after parsing, before running
	if err := cfg.ApplyDirectory(); err != nil {
		handleError(err)
		return
	}

	ctx := context.Background()
	if err := cmd.Run(ctx); err != nil {
		handleError(err)
	}
}

func handleError(err error) {
	var ece commands.ExitCodeError
	switch {
	case errors.Is(err, flag.ErrHelp): // just exit with 1 (help already printed)
	case errors.As(err, &ece):
		os.Exit(int(ece))
	default:
		fmt.Fprintf(os.Stderr, "%+v\n", err)
	}
	os.Exit(1)
}

// rootConfig handles global flags
type rootConfig struct {
	ChangeDir string
}

// RegisterFlags registers the -C flag for changing directory
func (cfg *rootConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.ChangeDir, "C", "", "change to directory before running command")
}

// ApplyDirectory changes to the specified directory if set
func (cfg *rootConfig) ApplyDirectory() error {
	if cfg.ChangeDir != "" {
		if err := os.Chdir(cfg.ChangeDir); err != nil {
			return fmt.Errorf("failed to change directory to %q: %w", cfg.ChangeDir, err)
		}
	}
	return nil
}

func newGnocliCmd(io commands.IO) (*commands.Command, *rootConfig) {
	cfg := &rootConfig{}

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "gno <command> [arguments]",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newBugCmd(io),
		// build
		newCleanCmd(io),
		newDocCmd(io),
		newEnvCmd(io),
		// fix
    newFixCmd(io),
		newDoctestCmd(io),
		newFmtCmd(io),
		// generate
		// get
		// install
		newListCmd(io),
		newLintCmd(io),
		newModCmd(io),
		// work
		newReplCmd(),
		newRunCmd(io),
		// telemetry
		newTestCmd(io),
		newToolCmd(io),
		// version -- show cmd/gno, golang versions
		newGnoVersionCmd(io),
		// vet
	)

	return cmd, cfg
}
