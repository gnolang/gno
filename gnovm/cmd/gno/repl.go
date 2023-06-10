package main

import (
	"context"
	goerrors "errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/tests"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"golang.org/x/term"
)

type replCfg struct {
	verbose        bool
	rootDir        string
	initialImports string
	initialCommand string
	skipUsage      bool
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

	fs.StringVar(
		&c.initialImports,
		"imports",
		"gno.land/p/demo/avl,gno.land/p/demo/ufmt",
		"initial imports, separated by a comma",
	)

	fs.StringVar(
		&c.initialCommand,
		"command",
		"",
		"initial command to run",
	)

	fs.BoolVar(
		&c.skipUsage,
		"skip-usage",
		false,
		"do not print usage",
	)
}

func execRepl(cfg *replCfg, args []string) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	if cfg.rootDir == "" {
		cfg.rootDir = guessRootDir()
	}

	if !cfg.skipUsage {
		fmt.Fprint(os.Stderr, `// Usage:
//   gno:1> /import "gno.land/p/demo/avl"     // import the p/demo/avl package
//   gno:2> /func a() string { return "a" }   // declare a new function named a
//   gno:3> /src                              // print current generated source
//   gno:4> println(a())                      // print the result of calling a()
//   gno:5> /exit
`)
	}

	return runRepl(cfg)
}

type state struct {
	imports []string
	funcs   []string
	// TODO: vars
}

func runRepl(cfg *replCfg) error {
	stdin := os.Stdin
	stdout := os.Stdout
	stderr := os.Stderr

	// init store and machine
	testStore := tests.TestStore(cfg.rootDir, "", stdin, stdout, stderr, tests.ImportModeStdlibsOnly)
	if cfg.verbose {
		testStore.SetLogStoreOps(true)
	}
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: "test",
		Output:  stdout,
		Store:   testStore,
	})

	defer m.Release()

	// init termui
	rw := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stderr}
	t := term.NewTerminal(rw, "")

	state := state{
		imports: make([]string, 0),
		funcs:   make([]string, 0),
	}

	// main loop
	for i := 1; ; /* continue until break */ i++ {
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

		if strings.TrimSpace(input) == "" {
			i--
			continue
		}

		funcName := fmt.Sprintf("repl_%d", i)
		// FIXME: support ";" as line separator?
		// FIXME: gofmt as linter + formatter
		// FIXME: support multiline when unclosed parenthesis, etc

		imports := strings.Join(state.imports, "\n")
		funcs := strings.Join(state.funcs, "\n")
		src := "package test\n" + imports + "\n" + funcs + "\nfunc " + funcName + "() {\nINPUT\n}"

		fields := strings.Fields(input)
		switch {
		case fields[0] == "/import":
			state.imports = append(state.imports, input[1:])
			continue
		case fields[0] == "/func":
			state.funcs = append(state.funcs, input[1:])
			continue
		case fields[0] == "/src":
			println(src)
		case fields[0] == "/exit":
			break
		}

		src = strings.Replace(src, "INPUT", input, 0)
		n := gno.MustParseFile(funcName+".gno", src)
		m.RunFiles(n)
		m.RunStatement(gno.S(gno.Call(gno.X(funcName))))

	}

	return nil
}
