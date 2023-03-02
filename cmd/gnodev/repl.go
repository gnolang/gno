package main

import (
	"context"
	goerrors "errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/gnolang/gno/pkgs/commands"
	"golang.org/x/term"

	gno "github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/tests"
)

type replCfg struct {
	verbose bool
	rootDir string
}

func newReplCmd() *commands.Command {
	cfg := &replCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "repl",
			ShortUsage: "repl [flags]",
			ShortHelp:  "Starts a GnoVM REPL",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execRepl(cfg, args)
		},
	)
}

func (c *replCfg) RegisterFlags(fs *flag.FlagSet) {
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

func execRepl(cfg *replCfg, args []string) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	if cfg.rootDir == "" {
		cfg.rootDir = guessRootDir()
	}

	return runRepl(cfg.rootDir, cfg.verbose)
}

func runRepl(rootDir string, verbose bool) error {
	stdin := os.Stdin
	stdout := os.Stdout
	stderr := os.Stderr

	// init store and machine
	testStore := tests.TestStore(rootDir, "", stdin, stdout, stderr, tests.ImportModeStdlibsOnly)
	if verbose {
		testStore.SetLogStoreOps(true)
	}
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: "test",
		Output:  stdout,
		Store:   testStore,
	})

	// init termui
	rw := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stderr}
	t := term.NewTerminal(rw, "")

	// main loop
	for i := 1; ; i++ {
		// parse line and execute
		t.SetPrompt(fmt.Sprintf("gno:%d> ", i))
		oldState, err := term.MakeRaw(0)
		input, err := t.ReadLine()
		if err != nil {
			term.Restore(0, oldState)
			if goerrors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("term error: %w", err)
		}
		term.Restore(0, oldState)

		funcName := fmt.Sprintf("repl_%d", i)
		src := "package test\nfunc " + funcName + "() {\n" + input + "\n}"
		// FIXME: support ";" as line separator?
		// FIXME: gofmt as linter + formatter
		// FIXME: support multiline when unclosed parenthesis, etc

		n := gno.MustParseFile(funcName+".gno", src)
		m.RunFiles(n)
		m.RunStatement(gno.S(gno.Call(gno.X(funcName))))
	}

	return nil
}
