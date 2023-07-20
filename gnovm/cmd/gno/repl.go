package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

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
//   gno> import "gno.land/p/demo/avl"     // import the p/demo/avl package
//   gno> func a() string { return "a" }   // declare a new function named a
//   gno> /src                             // print current generated source
//   gno> /editor                          // enter in editor mode to add several lines
//   gno> /reset                           // remove all previously inserted code
//   gno> println(a())                     // print the result of calling a()
//   gno> /exit
`)
	}

	return runRepl(cfg)
}

func runRepl(cfg *replCfg) error {
	// init repl state
	r := repl.NewRepl()

	if cfg.initialCommand != "" {
		handleInput(r, cfg.initialCommand)
	}

	var multiline bool
	for {
		fmt.Fprint(os.Stdout, "gno> ")

		input, err := getInput(multiline)
		if err != nil {
			return err
		}

		multiline = handleInput(r, input)
	}
}

func handleInput(r *repl.Repl, input string) bool {
	switch strings.TrimSpace(input) {
	case "/reset":
		r.Reset()
	case "/src":
		fmt.Fprintln(os.Stdout, r.Src())
	case "/exit":
		os.Exit(0)
	case "/editor":
		fmt.Fprintln(os.Stdout, "// Entering editor mode (^D to finish)")
		return true
	case "":
		// avoid to increase the repl execution counter if sending empty content
		return false
	default:
		out, err := r.Process(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)

		}
		fmt.Fprintln(os.Stdout, out)
	}

	return false
}

func getInput(ml bool) (string, error) {
	s := bufio.NewScanner(os.Stdin)
	var mlOut bytes.Buffer
	for s.Scan() {
		line := s.Text()
		if !ml {
			return line, nil
		}

		if line == "^D" {
			break
		}

		mlOut.WriteString(line)
		mlOut.WriteString("\n")
	}

	if err := s.Err(); err != nil {
		return "", err
	}

	return mlOut.String(), nil
}
