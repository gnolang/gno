package main

import (
	goerrors "errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
	gno "github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/tests"
)

type replOptions struct {
	Verbose bool   `flag:"verbose" help:"verbose"`
	RootDir string `flag:"root-dir" help:"clone location of github.com/gnolang/gno (gnodev tries to guess it)"`
	// Run string `flag:"run" help:"test name filtering pattern"`
	// Timeout time.Duration `flag:"timeout" help:"max execution time"`
	// VM Options
	// A flag about if we should download the production realms
	// UseNativeLibs bool // experimental, but could be useful for advanced developer needs
	// AutoImport bool
	// ImportPkgs...
}

var defaultReplOptions = replOptions{
	Verbose: false,
	RootDir: "",
}

func replApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(replOptions)

	if len(args) > 0 {
		cmd.ErrPrintfln("Usage: repl [flags]")

		return errors.New("invalid args")
	}

	if opts.RootDir == "" {
		opts.RootDir = guessRootDir()
	}

	return runRepl(opts.RootDir, opts.Verbose)
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
