package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"go/scanner"
	"os"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/repl"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type replCfg struct {
	rootDir        string
	initialCommand string
	skipUsage      bool
}

func newReplCmd() *commands.Command {
	cfg := &replCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "repl",
			ShortUsage: "repl [flags]",
			ShortHelp:  "starts a GnoVM REPL",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execRepl(cfg, args)
		},
	)
}

func (c *replCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"clone location of github.com/gnolang/gno (gno tries to guess it)",
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
		cfg.rootDir = gnoenv.RootDir()
	}

	if !cfg.skipUsage {
		fmt.Fprint(os.Stderr, `// Usage:
//   gno> import "gno.land/p/demo/avl"     // import the p/demo/avl package
//   gno> func a() string { return "a" }   // declare a new function named a
//   gno> /src                             // print current generated source
//   gno> /debug                           // activate the GnoVM debugger
//   gno> /editor                          // enter in multi-line mode, end with ';'
//   gno> /reset                           // remove all previously inserted code
//   gno> println(a())                     // print the result of calling a()
//   gno> /exit                            // alternative to <Ctrl-D>
`)
	}

	return runRepl(cfg)
}

func runRepl(cfg *replCfg) error {
	r := repl.NewRepl()

	if cfg.initialCommand != "" {
		handleInput(r, cfg.initialCommand)
	}

	fmt.Fprint(os.Stdout, "gno> ")

	inEdit := false
	prev := ""
	liner := bufio.NewScanner(os.Stdin)

	for liner.Scan() {
		line := liner.Text()

		if l := strings.TrimSpace(line); l == ";" {
			line, inEdit = "", false
		} else if l == "/editor" {
			line, inEdit = "", true
			fmt.Fprintln(os.Stdout, "// enter a single ';' to quit and commit")
		}
		if prev != "" {
			line = prev + "\n" + line
			prev = ""
		}
		if inEdit {
			fmt.Fprint(os.Stdout, "...  ")
			prev = line
			continue
		}

		if err := handleInput(r, line); err != nil {
			var goScanError scanner.ErrorList
			if errors.As(err, &goScanError) {
				// We assune that a Go scanner error indicates an incomplete Go statement.
				// Append next line and retry.
				prev = line
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
		}

		if prev == "" {
			fmt.Fprint(os.Stdout, "gno> ")
		} else {
			fmt.Fprint(os.Stdout, "...  ")
		}
	}
	return nil
}

// handleInput executes specific "/" commands, or evaluates input as Gno source code.
func handleInput(r *repl.Repl, input string) error {
	switch strings.TrimSpace(input) {
	case "/reset":
		r.Reset()
	case "/debug":
		r.Debug()
	case "/src":
		fmt.Fprintln(os.Stdout, r.Src())
	case "/exit":
		os.Exit(0)
	case "":
		// Avoid to increase the repl execution counter if no input.
	default:
		out, err := r.Process(input)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, out)
	}
	return nil
}
