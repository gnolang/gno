package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/gnolang/gno/gnovm/pkg/repl"
	"github.com/gnolang/gno/tm2/pkg/commands"
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
//   gno:1> import "gno.land/p/demo/avl"     // import the p/demo/avl package
//   gno:2> func a() string { return "a" }   // declare a new function named a
//   gno:3> /src                             // print current generated source
//   gno:3> /reset                           // remove all previously inserted code
//   gno:4> println(a())                     // print the result of calling a()
//   gno:5> /exit
`)
	}

	return runRepl(cfg)
}

type replCmd struct {
	stdout io.Writer

	repl *repl.Repl
}

func (r *replCmd) handleInput(input string) error {
	switch strings.TrimSpace(input) {
	case "/reset":
		r.repl.Reset()
	case "/src":
		fmt.Fprintln(r.stdout, r.repl.Src())
	case "/exit":
		os.Exit(0) // return special err?
	default:
		out, err := r.repl.Process(input)
		if err != nil {
			return err
		}
		fmt.Fprintln(r.stdout, out)
	}

	return nil
}

func runRepl(cfg *replCfg) error {
	stdout := os.Stdout

	// init repl state
	r := replCmd{
		repl:   repl.NewRepl(),
		stdout: stdout,
	}

	if cfg.initialCommand != "" {
		r.handleInput(cfg.initialCommand)
	}

	stdin := os.Stdin

	// main loop
	isTerm := term.IsTerminal(int(stdin.Fd()))

	if isTerm {
		rw := struct {
			io.Reader
			io.Writer
		}{os.Stdin, os.Stderr}
		t := term.NewTerminal(rw, "")
		for {
			// prompt and parse
			t.SetPrompt("gno:> ")
			oldState, err := term.MakeRaw(0)
			if err != nil {
				return fmt.Errorf("make term raw: %w", err)
			}

			input, err := t.ReadLine()
			if err != nil {
				term.Restore(0, oldState)
				if errors.Is(err, io.EOF) {
					return nil
				}
				return fmt.Errorf("term error: %w", err)
			}
			term.Restore(0, oldState)

			err = r.handleInput(input)
			if err != nil {
				return fmt.Errorf("handle repl input: %w", err)
			}
		}
	} else { // !isTerm
		scanner := bufio.NewScanner(stdin)
		for scanner.Scan() {
			input := scanner.Text()
			err := r.handleInput(input)
			if err != nil {
				return fmt.Errorf("handle repl input: %w", err)
			}
		}
		err := scanner.Err()
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
	}
	return nil
}
