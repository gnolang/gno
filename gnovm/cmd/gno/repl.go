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
	imports   []string
	funcs     []string
	lastInput string
	// TODO: support setting global vars
	// TODO: switch to state machine, and support rollback of anything
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
		imports:   make([]string, 0),
		funcs:     make([]string, 0),
		lastInput: "// your code will be here", // initial value, to make it easier to identify with '/src'
	}

	for _, imp := range strings.Split(cfg.initialImports, ",") {
		if strings.TrimSpace(imp) == "" {
			continue
		}
		state.imports = append(state.imports, `import "`+imp+`"`)
	}

	if cfg.initialCommand != "" {
		// TODO: implement
		panic("not implemented")
	}

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

		if strings.TrimSpace(input) == "" {
			i--
			continue
		}

		funcName := fmt.Sprintf("repl_%d", i)
		// FIXME: support ";" as line separator?
		// FIXME: support multiline when unclosed parenthesis, etc

		imports := strings.Join(state.imports, "\n")
		funcs := strings.Join(state.funcs, "\n")
		src := "package test\n" + imports + "\n" + funcs + "\nfunc " + funcName + "() {\nINPUT\n}"

		fields := strings.Fields(input)
		command := fields[0]
		switch {
		case command == "/import":
			imp := fields[1]
			state.imports = append(state.imports, `import "`+imp+`"`)
			// TODO: check if valid, else rollback
			continue
		case command == "/func":
			state.funcs = append(state.funcs, input[1:])
			// TODO: check if valid, else rollback
			continue
		case command == "/src":
			// TODO: use go/format for pretty print
			src = strings.ReplaceAll(src, "INPUT", state.lastInput)
			println(src)
			continue
		case command == "/exit":
			break
		case strings.HasPrefix(command, "/"):
			println("unsupported command")
			continue
		default:
			// not a command, probably code to run
		}

		state.lastInput = input
		src = strings.ReplaceAll(src, "INPUT", input)
		n := gno.MustParseFile(funcName+".gno", src)
		// TODO: run fmt check + linter
		m.RunFiles(n)
		// TODO: smart recover system
		m.RunStatement(gno.S(gno.Call(gno.X(funcName))))
		// TODO: if output is empty, consider that it's a persisted variable?
	}

	return nil
}
