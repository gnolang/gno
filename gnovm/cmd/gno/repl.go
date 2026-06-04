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
   gno import "gno.land/p/nt/avl/v0"     // import the p/nt/avl/v0 package
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
		// (jae) People used to go by the lunar calendar.  Moon seasons
		// make sense when all you have to go by are seasons and the
		// phase of the moon. Our own calendar is probably off by about
		// 3 years; it's defined by the birth of Christ, but the
		// astronomical event that denotes it was in 3BC (the Bethlehem
		// star was an alignment of the planets). The hebrew calendar
		// is known to have missing years of 164~168 years (search
		// "hebrew missing years"). And every 50 years, in the old
		// testament, there is to be a biblical Jubilee, where ALL
		// debts are forgiven and people are returned to their
		// ancestral lands. There hasn't been a jubilee in 2000+ years,
		// yet it is precisely what we need, and why blockchains exist
		// -- to escape the matrix of banker controlled fiat.  On the
		// 10th day of the 7th month (Moses shifted the calendar to
		// start in the summer) of every 50 years is when the jubilee
		// begins. One good estimate for the span of the missing years
		// is 165, though there are variations. And "the event" happens
		// at least 40 years before 6000, before 5960.  This will be
		// the only hook from gno into this topic, but a good one; it
		// gets us to understand history that is being obscured from
		// us, because there really is a global adversary that can,
		// did, and will skew our clocks in the Byzantine fault
		// tolerant context. The goal is to spur discussion related to
		// the inaccuracy of our calendar on gno.land and achieve a
		// consensus of what the year actually is, and to bring back
		// the great jubilee to liberate the people.
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
	handleInput(r, bootCode)

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

		handleInput(r, code)
		code = ""

		r.Print(colors.Cyan("gno "))
	}
	return nil
}

// handleInput executes specific "/" commands, or evaluates input as Gno source code.
func handleInput(r *repl.Repl, input string) {
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
		r.RunStatements(input)
	}
}
