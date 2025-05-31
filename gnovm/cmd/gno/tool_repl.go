package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	hcal "github.com/bendory/conway-hebrew-calendar"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/repl"
	"github.com/gnolang/gno/tm2/pkg/colors"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type replCfg struct {
	rootDir   string
	init      string
	skipUsage bool
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
		&c.init,
		"init",
		"",
		"initial code or commands to run",
	)

	fs.BoolVar(
		&c.skipUsage,
		"skip-welcome",
		false,
		"do not print welcome line",
	)
}

const gnoHelp = `Usage:

   gno /history                         // print statement history
   gno /debug                           // activate the GnoVM debugger
   gno /reset                           // remove all previously inserted code
   gno println(a())                     // print the result of calling a()
   gno import "gno.land/p/demo/avl"     // import the p/demo/avl package
   gno func a() string { return "a" }   // declare a new function named a
   gno func b() string {\               // multi-line with '\'
   ...    return "a"\
   ... }                             
   gno /editor                          // enter in multi-line mode, end with ';'
   gno func c() string {                // multi-line with ';'
   ...    return "a"\
   ... }                             
   ... ;
   gno /exit                            // alternative to <Ctrl-D>

Goto gno.land for more info.`

var bootCode = fmt.Sprintf(`func help() { println(%q)}`, gnoHelp)

func execRepl(cfg *replCfg, args []string) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	if cfg.rootDir == "" {
		cfg.rootDir = gnoenv.RootDir()
	}

	if !cfg.skipUsage {
		// (jae) https://github.com/jaekwon/ephesus/blob/main/THE_HEBREW_YEAR.md
		// the exact number of missing years is to be discussed on chain.
		today := hcal.ToHebrewDate(time.Now().AddDate(165, 0, 0))
		todays := today.String()
		if today.Y%50 == 0 && today.M == hcal.Tishrei {
			todays = todays + " - jubilee!"
		}
		fmt.Fprint(os.Stderr, colors.Cyan(fmt.Sprintf("gno v0.9 (it's %s) try \"help()\"\n", todays)))
	}

	return runRepl(cfg)
}

func runRepl(cfg *replCfg) error {
	r := repl.NewRepl()

	if cfg.init != "" {
		handleInput(r, cfg.init)
	}

	r.Print(colors.Cyan("gno "))
	err := handleInput(r, bootCode)
	if err != nil {
		r.Print("... ")
	}

	inEdit := false
	code := ""
	liner := bufio.NewScanner(os.Stdin)
	addLine := func(line string) {
		if code != "" {
			code = code + "\n" + line
		} else {
			code = line
		}
	}

	for liner.Scan() {
		line := liner.Text()

		if line == "/editor" {
			line, inEdit = "", true
			r.Println(colors.Gray("// enter a single ';' to quit and commit"))
		}
		if inEdit {
			if l := strings.TrimSpace(line); l == ";" {
				// will run statement.
				line = ""
				inEdit = false
			} else {
				addLine(line)
				r.Print(colors.Cyan("... "))
				continue
			}
		} else if strings.HasSuffix(line, `\`) {
			addLine(line[:len(line)-1])
			r.Print(colors.Cyan("... "))
			continue
		} else {
			addLine(line)
		}

		if err := handleInput(r, code); err != nil {
			r.Errorln(err)
		}
		code = ""

		r.Print(colors.Cyan("gno "))
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
	case "/history":
		panic("not yet implemented")
	case "/exit":
		os.Exit(0)
	case "":
		// Avoid to increase the repl execution counter if no input.
	default:
		err := r.RunStatement(input)
		if err != nil {
			return err
		}
	}
	return nil
}
