package integration

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
)

// ---- sleep command

func SleepCmd() func(ts *testscript.TestScript, neg bool, args []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		if neg {
			ts.Fatalf("sleep does not support negation")
		}
		if len(args) != 1 {
			ts.Fatalf("usage: sleep <duration>")
		}

		d, err := time.ParseDuration(args[0])
		if err != nil {
			ts.Fatalf("invalid duration %q: %v", args[0], err)
		}
		time.Sleep(d)
	}
}

// ---- repeat command

// RepeatCmd returns a testscript command that runs a subcommand N times.
// It takes a reference to the full commands map for dispatching.
func RepeatCmd(cmds map[string]func(*testscript.TestScript, bool, []string)) func(ts *testscript.TestScript, neg bool, args []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		// The shortest form the usage describes is "repeat N cmd", so two.
		// The bounds checks below reject a count with no command after it.
		if len(args) < 2 {
			ts.Fatalf("usage: repeat [-all] N <cmd> [args...]")
		}

		runAll := false
		idx := 0
		if args[0] == "-all" {
			runAll = true
			idx++
		}

		if idx >= len(args) {
			ts.Fatalf("usage: repeat [-all] N <cmd> [args...]")
		}

		count, err := strconv.Atoi(args[idx])
		if err != nil || count < 1 {
			ts.Fatalf("invalid repeat count %q", args[idx])
		}
		idx++

		if idx >= len(args) {
			ts.Fatalf("usage: repeat [-all] N <cmd> [args...]")
		}

		subCmd := args[idx]
		subArgs := args[idx+1:]

		cmdFn, ok := cmds[subCmd]
		if !ok {
			ts.Fatalf("unknown command %q for repeat", subCmd)
		}

		passed, failed := 0, 0
		firstFailAt := -1

		for i := range count {
			iterFailed := runCmdIteration(ts, "repeat", cmdFn, false, subArgs)
			if iterFailed {
				failed++
				if firstFailAt < 0 {
					firstFailAt = i + 1
				}
				if !runAll {
					fmt.Fprintf(ts.Stderr(), "repeat: failed at iteration %d/%d\n", i+1, count)
					TSValidateError(ts, "repeat", neg, fmt.Errorf("iteration %d failed", i+1))
					return
				}
			} else {
				passed++
			}
		}

		// Write summary for -all mode
		if runAll {
			if failed > 0 {
				fmt.Fprintf(ts.Stderr(), "repeat: %d/%d passed, %d/%d failed (first at iteration %d)\n",
					passed, count, failed, count, firstFailAt)
			} else {
				fmt.Fprintf(ts.Stderr(), "repeat: %d/%d passed\n", passed, count)
			}
		}

		if failed > 0 {
			TSValidateError(ts, "repeat", neg, fmt.Errorf("%d/%d iterations failed", failed, count))
		} else {
			TSValidateError(ts, "repeat", neg, nil)
		}
	}
}

// runCmdIteration runs a testscript command function and catches any
// fatal/panic from the command. Returns true if the iteration failed.
//
// verb names the command doing the repeating, because both repeat and
// eventually run their attempts through here and a script that used one must
// not be told about the other.
func runCmdIteration(ts *testscript.TestScript, verb string, fn func(*testscript.TestScript, bool, []string), neg bool, args []string) (failed bool) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(ts.Stderr(), "%s: iteration error: %v\n", verb, r)
			failed = true
		}
	}()
	fn(ts, neg, args)
	return false
}

// ---- eventually command

const (
	// defaultEventuallyTimeout bounds a wait whose line names no budget.
	defaultEventuallyTimeout = 30 * time.Second
	// defaultEventuallyInterval matches gpao's own -poll-interval default
	// (contribs/gpao/main.go:40). Every wait the scenarios keep is waiting on
	// the oracle, which looks for new blocks once a second, so polling faster
	// cannot see its progress any sooner.
	defaultEventuallyInterval = time.Second

	// stdoutGateFlag turns an attempt's output into part of what counts as
	// success, for a command that answers without erring when the value the
	// scenario waits for has not arrived.
	stdoutGateFlag = "-stdout"

	eventuallyUsage = "usage: eventually [timeout [interval]] [-stdout regex] <cmd> [args...]"
)

// eventuallyArgs is one eventually line resolved: what to run, the budget to
// run it under, and what counts as an answer.
type eventuallyArgs struct {
	timeout  time.Duration
	interval time.Duration
	// stdout is a regex an attempt's output must match to count as success.
	// Empty means the exit status alone decides.
	stdout  string
	cmd     string
	cmdArgs []string
}

// parseEventuallyArgs reads the optional leading timeout and interval, filling
// in defaultEventuallyTimeout and defaultEventuallyInterval for whichever the
// line omits.
//
// Positional optional durations stay unambiguous because no command name
// parses as a time.Duration. A duration needs a unit, the sole exception being
// a bare "0" (time/format.go:1635-1637), which is not a command name either.
// The cost is that a unitless typo ("30" for "30s") is read as the command to
// run and reported as an unknown command.
func parseEventuallyArgs(args []string) (eventuallyArgs, error) {
	p := eventuallyArgs{timeout: defaultEventuallyTimeout, interval: defaultEventuallyInterval}
	for _, budget := range []*time.Duration{&p.timeout, &p.interval} {
		if len(args) == 0 {
			break
		}
		d, err := time.ParseDuration(args[0])
		if err != nil {
			break
		}
		*budget, args = d, args[1:]
	}
	if len(args) > 0 && args[0] == stdoutGateFlag {
		if len(args) < 2 {
			return eventuallyArgs{}, errors.New(eventuallyUsage)
		}
		// Compiled here so a malformed pattern is a usage error rather than a
		// wait that quietly never matches.
		if _, err := regexp.Compile(args[1]); err != nil {
			return eventuallyArgs{}, fmt.Errorf("invalid %s pattern %q: %w", stdoutGateFlag, args[1], err)
		}
		p.stdout, args = args[1], args[2:]
	}
	if len(args) == 0 {
		return eventuallyArgs{}, errors.New(eventuallyUsage)
	}
	p.cmd, p.cmdArgs = args[0], args[1:]
	return p, nil
}

// EventuallyCmd returns a testscript command that reruns a subcommand until it
// succeeds or the timeout passes.
//
// timeout bounds when a new attempt starts, not the wall-clock time of the
// command itself: the deadline is only checked between attempts, so a slow
// or hanging sub-command can overrun it. A hard per-attempt cutoff would
// need to abandon a still-running attempt from another goroutine, leaving
// ts in a half-mutated state, which is worse than overrunning. The outer
// test run timeout is the real backstop for a hung command.
//
// It dispatches only to the custom command map, not to testscript builtins,
// because the builtin table is not reachable from a user command.
//
// There is no chain event announcing that a package went live, so every
// assertion about the oracle is a poll. This is what keeps those polls from
// being fixed sleeps.
//
// Every attempt starts with the builtin output buffers emptied, so a following
// "stdout", "! stdout", "stderr", "! stderr" or "cp stdout" sees the attempt
// that succeeded and nothing else. That is what lets a read needing a negation
// or a capture still be retried instead of running once and flaking on a
// transient RPC blip.
//
// It does not poll a following assertion: without a gate it returns as soon as
// the sub-command exits 0, so a "stdout <pattern>" line after it runs exactly
// once, against whatever that first success produced. That is unsound for a
// query answering an empty result rather than an error -- vm/qinertpaths lists
// nothing and exits 0 -- because the wait ends before the value arrives.
//
// "-stdout <regex>" is the answer to that: the pattern is checked inside the
// attempt, so an exit 0 whose output does not match is not yet an answer and
// the wait runs the command again. HTTPGetCmd's second argument gates its own
// body the same way (httpget.go:86).
func EventuallyCmd(cmds map[string]func(*testscript.TestScript, bool, []string)) func(*testscript.TestScript, bool, []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		if neg {
			ts.Fatalf("eventually does not support negation")
		}
		p, err := parseEventuallyArgs(args)
		if err != nil {
			ts.Fatalf("%v", err)
		}
		cmdFn, ok := cmds[p.cmd]
		if !ok {
			ts.Fatalf("unknown command %q for eventually", p.cmd)
		}
		// Compiled once rather than per attempt; parseEventuallyArgs already
		// rejected a pattern that does not compile.
		var gate *regexp.Regexp
		if p.stdout != "" {
			gate = regexp.MustCompile(p.stdout)
		}

		deadline := time.Now().Add(p.timeout)
		attempts := 0
		for {
			attempts++
			resetBuiltinOutput(ts)
			failed := runCmdIteration(ts, "eventually", cmdFn, false, p.cmdArgs)
			if !failed && gate == nil {
				return
			}
			// Read before the next attempt's reset empties it. A command that
			// exits 0 with the wrong output has not answered yet, which is the
			// case the exit status alone cannot express.
			if !failed && gate.MatchString(builtinStdout(ts)) {
				return
			}

			// runCmdIteration's own diagnostic recovers testscript's fixed
			// failNow sentinel, not the formatted Fatalf message, so it
			// carries no useful detail here. Report the attempt ourselves --
			// to the test log rather than to stderr, both because the next
			// attempt's reset would wipe it and because it is a note about
			// the wait, not output of the command being retried.
			ts.Logf("eventually: attempt %d for %q did not succeed", attempts, p.cmd)
			if time.Now().After(deadline) {
				if gate != nil {
					ts.Fatalf("eventually: %q ran %d times in %s without stdout matching %q",
						p.cmd, attempts, p.timeout, p.stdout)
				}
				ts.Fatalf("eventually: %q did not succeed within %s (%d attempts)", p.cmd, p.timeout, attempts)
			}
			time.Sleep(p.interval)
		}
	}
}

// builtinStdout reads what the current attempt wrote, before the next
// attempt's reset empties it. Checked the way resetBuiltinOutput checks:
// testscript returning something other than the builder it accumulates into
// must fail here rather than silently gate on an empty string.
func builtinStdout(ts *testscript.TestScript) string {
	b, ok := ts.Stdout().(*strings.Builder)
	if !ok {
		ts.Fatalf("eventually: testscript builtin output is %T, not *strings.Builder: cannot read it for -stdout", ts.Stdout())
	}
	return b.String()
}

// resetBuiltinOutput empties the buffers testscript hands a builtin, so what one
// attempt writes cannot reach an assertion that follows a later one.
//
// testscript flushes those buffers into ts.stdout/ts.stderr once, when the
// builtin returns (callBuiltinCmd's defer, testscript.go:758-763), and every
// attempt inside a single wait shares them. ts.Fatalf happens to flush on its
// way out (testscript.go:1202-1207), which hides the problem for a sub-command
// that aborts that way, but an attempt recovered from a panic by
// runCmdIteration leaves its output behind, and EventuallyCmd's and
// runCmdIteration's own stderr diagnostics accumulate either way.
//
// clearBuiltinStd is unexported, but ts.Stdout() and ts.Stderr() return the
// *strings.Builder they accumulate into (testscript.go:931-949), which can be
// reset directly. A testscript that ever returns something else must fail here
// rather than silently go back to accumulating.
func resetBuiltinOutput(ts *testscript.TestScript) {
	for _, w := range []io.Writer{ts.Stdout(), ts.Stderr()} {
		b, ok := w.(*strings.Builder)
		if !ok {
			ts.Fatalf("eventually: testscript builtin output is %T, not *strings.Builder: cannot clear it between attempts", w)
		}
		b.Reset()
	}
}
