package main

import (
	"context"
	"flag"

	"github.com/gnolang/gno/pkgs/commands"
	gno "github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/tests"
)

type runCfg struct {
	verboseStruct
	rootDirStruct
}

func newRunCmd(io *commands.IO) *commands.Command {
	cfg := &runCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "run",
			ShortUsage: "run [flags] <file> [<file>...]",
			ShortHelp:  "Runs the specified gno files",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execRun(cfg, args, io)
		},
	)
}

func (c *runCfg) RegisterFlags(fs *flag.FlagSet) {
	c.verboseStruct.RegisterFlags(fs)
	c.rootDirStruct.RegisterFlags(fs)
}

func execRun(cfg *runCfg, args []string, io *commands.IO) error {
	if len(args) == 0 {
		return flag.ErrHelp
	}

	if cfg.rootDir == "" {
		cfg.rootDir = guessRootDir()
	}

	stdin := io.In
	stdout := io.Out
	stderr := io.Err

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
