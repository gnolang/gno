// go:build !js
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

const (
	indentSize = 4

	importExample = "import \"gno.land/p/demo/avl\""
	funcExample   = "func a() string { return \"a\" }"
	printExample  = "println(a())"
	srcCommand    = "/src"
	editorCommand = "/editor"
	resetCommand  = "/reset"
	exitCommand   = "/exit"
	clearCommand  = "/clear"
	helpCommand   = "/help"
	gnoREPL       = "gno> "
	inEditMode    = "...  "

	helpText = `// Usage:
//   gno> %-35s // import the p/demo/avl package
//   gno> %-35s // declare a new function named a
//   gno> %-35s // print current generated source
//   gno> %-35s // enter in multi-line mode, end with ';'
//   gno> %-35s // clear the terminal screen
//   gno> %-35s // remove all previously inserted code
//   gno> %-35s // print the result of calling a()
//   gno> %-35s // alternative to <Ctrl-D>

`
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

	var (
		inEdit      bool
		prev        string
		indentLevel int
	)

	liner := bufio.NewScanner(os.Stdin)

	for liner.Scan() {
		line := liner.Text()

		trimmedLine := strings.TrimSpace(line)

		indentLevel = updateIndentLevel(trimmedLine, indentLevel)
		line, inEdit = handleEditor(line)

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

		printPrompt(indentLevel, prev)
	}

	return nil
}

func handleEditor(line string) (string, bool) {
	if l := strings.TrimSpace(line); l == ";" {
		return "", false
	} else if l == editorCommand {
		fmt.Fprintln(os.Stdout, "// enter a single ';' to quit and commit")
		return "", true
	}

	return line, false
}

func updateIndentLevel(line string, indentLevel int) int {
	openCount, closeCount := countBrackets(line)
	increaseIndent := shouldIncreaseIndent(line)
	indentLevel += openCount - closeCount

	if indentLevel < 0 {
		indentLevel = 0
	}

	if increaseIndent {
		indentLevel++
	}

	return indentLevel
}

func countBrackets(line string) (int, int) {
	openCount, closeCount := 0, 0
	inString, inComment, inSingleLineComment := false, false, false
	var stringChar rune

	for i, char := range line {
		if !inString && !inComment && !inSingleLineComment {
			switch char {
			case '{', '(', '[':
				openCount++
			case '}', ')', ']':
				closeCount++
			case '"', '\'':
				inString = true
				stringChar = char
			case '/':
				if i < len(line)-1 {
					if line[i+1] == '/' {
						inSingleLineComment = true
					} else if line[i+1] == '*' {
						inComment = true
					}
				}
			}
		} else if inString && char == stringChar {
			inString = false
		} else if inComment && i < len(line)-1 && char == '*' && line[i+1] == '/' {
			inComment = false
		}
	}

	return openCount, closeCount
}

func shouldIncreaseIndent(line string) bool {
	openIndex := strings.IndexAny(line, "{([")
	if openIndex != -1 && openIndex < len(line)-1 && line[openIndex+1] == '\n' {
		return true
	}
	return false
}

func printPrompt(indentLevel int, prev string) {
	indent := strings.Repeat(" ", indentLevel*indentSize)
	if prev == "" {
		fmt.Fprintf(os.Stdout, "%s%s", gnoREPL, indent)
	} else {
		fmt.Fprintf(os.Stdout, "%s%s", inEditMode, indent)
	}
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
		clearScreen(&RealCommandExecutor{}, RealOSGetter{})
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

type CommandExecutor interface {
	Execute(cmd *exec.Cmd) error
}

type RealCommandExecutor struct{}

func (e *RealCommandExecutor) Execute(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

type OsGetter interface {
	Get() string
}

type RealOSGetter struct{}

func (r RealOSGetter) Get() string {
	return runtime.GOOS
}

func clearScreen(executor CommandExecutor, osGetter OsGetter) {
	var cmd *exec.Cmd

	if osGetter.Get() == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	executor.Execute(cmd)
}

func printHelp() {
	fmt.Printf(
		helpText, importExample, funcExample,
		srcCommand, editorCommand, clearCommand,
		resetCommand, exitCommand, printExample,
	)
}
