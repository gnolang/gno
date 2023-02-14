package main

import (
	"context"
	"flag"
	"os"

	"github.com/gnolang/gno/pkgs/commands"
	"github.com/gnolang/gno/pkgs/errors"
	gno "github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/tests"
)

type runCfg struct {
	verbose bool
	rootDir string
}

func newRunCmd() *commands.Command {
	cfg := &runCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "run",
			ShortUsage: "run [flags] <file> [<file>...]",
			ShortHelp:  "Runs the specified gno files",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execRun(cfg, args)
		},
	)
}

func (c *runCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.verbose,
		"verbose",
		false,
		"verbose output when running",
	)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"clone location of github.com/gnolang/gno (gnodev tries to guess it)",
	)
}

func execRun(cfg *runCfg, args []string) error {
	if len(args) == 0 {
		return errors.New("invalid args")
	}

	if cfg.rootDir == "" {
		cfg.rootDir = guessRootDir()
	}

	stdin := os.Stdin
	stdout := os.Stdout
	stderr := os.Stderr

	// init store and machine
	testStore := tests.TestStore(cfg.rootDir,
		"", stdin, stdout, stderr,
		tests.ImportModeStdlibsPreferred)
	if cfg.verbose {
		testStore.SetLogStoreOps(true)
	}

	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: "main",
		Output:  stdout,
		Store:   testStore,
	})

	// read files
	files := make([]*gno.FileNode, len(args))
	for i, fname := range args {
		files[i] = gno.MustReadFile(fname)
	}

	// run files
	m.RunFiles(files...)
	m.RunMain()

	return nil
}
