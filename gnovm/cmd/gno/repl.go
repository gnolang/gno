package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"go/scanner"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/repl"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

const indentSize = 4

const (
	srcCommand    = "/src"
	editorCommand = "/editor"
	resetCommand  = "/reset"
	exitCommand   = "/exit"
	clearCommand  = "/clear"
	helpCommand   = "/help"
	gnoREPL       = "gno> "
	inEditMode    = "...  "
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
		"clone location of github.com/gnolang/gno (gno tries to guess it)",
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
		cfg.rootDir = gnoenv.RootDir()
	}

	if !cfg.skipUsage {
		printHelp()
	}

	return runRepl(cfg)
}

func runRepl(cfg *replCfg) error {
	r := repl.NewRepl()

	if cfg.initialCommand != "" {
		handleInput(r, cfg.initialCommand)
	}

	fmt.Fprint(os.Stdout, gnoREPL)

	inEdit := false
	prev := ""
	liner := bufio.NewScanner(os.Stdin)
	indentLevel := 0

	for liner.Scan() {
		line := liner.Text()

		trimmedLine := strings.TrimSpace(line)
		openCount, closeCount := 0, 0

		for _, char := range trimmedLine {
			switch char {
			case '{', '(', '[':
				openCount++
			case '}', ')', ']':
				closeCount++
			}
		}

		if closeCount > 0 {
			indentLevel -= closeCount
		}

		indentLevel += openCount

		if indentLevel < 0 {
			indentLevel = 0
		}

		if strings.HasSuffix(trimmedLine, ":") {
			indentLevel++
		}

		if l := strings.TrimSpace(line); l == ";" {
			line, inEdit = "", false
		} else if l == editorCommand {
			line, inEdit = "", true

			fmt.Fprintln(os.Stdout, "// enter a single ';' to quit and commit")
		}

		if prev != "" {
			line = prev + "\n" + line
			prev = ""
		}

		if inEdit {
			fmt.Fprint(os.Stdout, inEditMode)
			prev = line

			continue
		}

		if err := handleInput(r, line); err != nil {
			var goScanError scanner.ErrorList
			if errors.As(err, &goScanError) {
				// We assume that a Go scanner error indicates an incomplete Go statement.
				// Append next line and retry.
				prev = line
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
		}

		if prev == "" {
			fmt.Fprintf(os.Stdout, "gno> %s", strings.Repeat(" ", indentLevel*indentSize))
		} else {
			fmt.Fprintf(os.Stdout, "... %s", strings.Repeat(" ", indentLevel*indentSize))
		}
	}

	return nil
}

// handleInput executes specific "/" commands, or evaluates input as Gno source code.
func handleInput(r *repl.Repl, input string) error {
	input = strings.TrimSpace(input)
	switch input {
	case resetCommand:
		r.Reset()
	case srcCommand:
		fmt.Fprintln(os.Stdout, r.Src())
	case clearCommand:
		clearScreen()
	case exitCommand:
		os.Exit(0)
	case helpCommand:
		printHelp()
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

func clearScreen() {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}

func printHelp() {
	fmt.Fprint(os.Stderr, `Gno REPL Usage Instructions:
--------------------------------
- /src:      Display the current generated source code.
	Example:   Prints the current generated source code to the console.

- /editor:   Enter multi-line mode. End input with ';'.
	Example:   Allows writing code in multiple lines, finish by entering ';'.

- /clear:    Clear the screen.

- /reset:    Remove all previously inserted code and reset the session.
	Example:   Clears all code entered in the current REPL session.

- /exit:     Exit the REPL session (alternative to pressing <Ctrl-D>).
	Example:   Exits the REPL session.

- println:   Execute a function and print the result.
	Usage:     gno> println(a())
	Example:   Prints the result of calling the function 'a'.

Note: Prefix commands with 'gno>' to execute in the Gno REPL environment.

`)
}
