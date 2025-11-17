package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/profiler"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"golang.org/x/term"
)

func maybeStartProfileShell(cmdIO commands.IO, opts *test.TestOptions) {
	if opts == nil || opts.Profile == nil || !opts.Profile.Interactive {
		return
	}
	profile := opts.Profile.LastProfile()
	if profile == nil {
		return
	}
	runProfileShell(cmdIO, profile, opts.Profile, test.NewStoreAdapter(opts.TestStore))
}

func runProfileShell(cmdIO commands.IO, profile *profiler.Profile, cfg *test.ProfileConfig, store profiler.Store) {
	if !isTerminalReader(cmdIO.In()) {
		fmt.Fprintln(cmdIO.Err(), "Profiler shell requires an interactive terminal on stdin; skipping.")
		return
	}

	reader := bufio.NewReader(cmdIO.In())
	fmt.Fprintln(cmdIO.Err(), "\nProfiler shell ready. Type 'help' for available commands, 'exit' to leave.")

	for {
		fmt.Fprint(cmdIO.Err(), "profile> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Fprintln(cmdIO.Err())
			} else {
				fmt.Fprintf(cmdIO.Err(), "error: %v\n", err)
			}
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if exit := executeProfileCommand(line, cmdIO, profile, cfg, store); exit {
			return
		}
	}
}

func executeProfileCommand(input string, cmdIO commands.IO, profile *profiler.Profile, cfg *test.ProfileConfig, store profiler.Store) bool {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return false
	}

	cmd := strings.ToLower(fields[0])
	args := strings.TrimSpace(input[len(fields[0]):])

	switch cmd {
	case "exit", "quit":
		fmt.Fprintln(cmdIO.Err(), "Exiting profiler shell.")
		return true
	case "help":
		fmt.Fprintln(cmdIO.Err(), "Commands:")
		fmt.Fprintln(cmdIO.Err(), "  help           - show this help")
		fmt.Fprintln(cmdIO.Err(), "  text           - print textual summary of the profile")
		fmt.Fprintln(cmdIO.Err(), "  top            - show top functions by cumulative cost")
		fmt.Fprintln(cmdIO.Err(), "  calltree       - display the call tree view")
		fmt.Fprintln(cmdIO.Err(), "  json           - dump the raw JSON profile")
		fmt.Fprintln(cmdIO.Err(), "  list <func>    - show line-by-line profile for matching functions")
		fmt.Fprintln(cmdIO.Err(), "  exit/quit      - leave the shell")
	case "text", "summary":
		if _, err := profile.WriteTo(cmdIO.Out()); err != nil {
			fmt.Fprintf(cmdIO.Err(), "error: %v\n", err)
		}
	case "top", "toplist":
		limit := 0
		if args != "" {
			val, err := strconv.Atoi(args)
			if err != nil || val <= 0 {
				fmt.Fprintln(cmdIO.Err(), "usage: top <positive-number>? e.g. 'top 5'")
				break
			}
			limit = val
		}
		if err := profile.WriteTopListLimit(cmdIO.Out(), limit); err != nil {
			fmt.Fprintf(cmdIO.Err(), "error: %v\n", err)
		}
	case "calltree":
		if err := profile.WriteCallTree(cmdIO.Out()); err != nil {
			fmt.Fprintf(cmdIO.Err(), "error: %v\n", err)
		}
	case "json":
		if err := profile.WriteJSON(cmdIO.Out()); err != nil {
			fmt.Fprintf(cmdIO.Err(), "error: %v\n", err)
		}
	case "list":
		target := strings.TrimSpace(args)
		if target == "" {
			target = cfg.FunctionList
			if target == "" {
				fmt.Fprintln(cmdIO.Err(), "usage: list <function-pattern>")
				break
			}
			fmt.Fprintf(cmdIO.Err(), "using default function pattern %q\n", target)
		}
		if err := profile.WriteFunctionList(cmdIO.Out(), target, store); err != nil {
			fmt.Fprintf(cmdIO.Err(), "error: %v\n", err)
		}
	default:
		fmt.Fprintf(cmdIO.Err(), "unknown command %q (type 'help' for a list of commands)\n", cmd)
	}

	return false
}

type fdReader interface {
	Fd() uintptr
}

func isTerminalReader(r io.Reader) bool {
	if v, ok := r.(fdReader); ok {
		return term.IsTerminal(int(v.Fd()))
	}
	return false
}
