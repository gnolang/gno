package integration

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventuallySucceedsAfterTransientFailures(t *testing.T) {
	calls := 0
	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"flaky": func(ts *testscript.TestScript, neg bool, args []string) {
			calls++
			if calls < 3 {
				ts.Fatalf("not yet")
			}
		},
	}
	cmds["eventually"] = EventuallyCmd(cmds)

	testscript.RunT(NewTestscriptT(testLogger(t), false), testscript.Params{
		Files: []string{writeScript(t, "eventually 5s 10ms flaky\n")},
		Cmds:  cmds,
	})

	require.Equal(t, 3, calls)
}

func TestEventuallyGivesUpAtTheDeadline(t *testing.T) {
	adapter := NewTestscriptT(testLogger(t), false)
	calls := 0
	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"never": func(ts *testscript.TestScript, neg bool, args []string) {
			calls++
			ts.Fatalf("still failing")
		},
	}
	cmds["eventually"] = EventuallyCmd(cmds)

	timeout := 200 * time.Millisecond
	start := time.Now()
	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, "eventually 200ms 10ms never\n")},
		Cmds:  cmds,
	})
	elapsed := time.Since(start)

	require.True(t, adapter.Failed, "eventually must fail the script when the deadline passes")
	require.GreaterOrEqual(t, elapsed, timeout, "eventually must not give up before the configured timeout")
	require.Less(t, elapsed, 5*time.Second)
	require.Greater(t, calls, 1, "eventually must actually retry, not give up after the first attempt")
}

// Every assertion that follows an eventually must see the winning attempt alone,
// whichever way the earlier attempts failed.
//
// The two abort modes below reach eventually differently. ts.Fatalf flushes the
// builtin buffers on its way out (testscript.go:1202-1207), which happens to
// leave stdout clean; a panic never reaches that flush, so the attempt's output
// stays in the buffer for the next attempt to append to. Its guarantee must
// not depend on which one it got -- and stderr is unsound under both, because
// its own per-attempt diagnostic lands in the buffer the winning attempt
// shares.
func TestEventuallyOutputIsOnlyTheSucceedingAttempt(t *testing.T) {
	aborts := map[string]func(ts *testscript.TestScript){
		"aborting with ts.Fatalf": func(ts *testscript.TestScript) { ts.Fatalf("not yet") },
		"panicking":               func(ts *testscript.TestScript) { panic("not yet") },
	}
	scripts := map[string]string{
		"a positive match still sees the winning attempt": "stdout 'settled'\n",
		"a negation does not see a failed attempt":        "! stdout 'not ready'\n",
		"stderr carries only the winning attempt":         "stderr 'settled on stderr'\n! stderr 'complained'\n",
		"the per-attempt diagnostic stays out of stderr":  "! stderr 'did not succeed'\n",
		"a capture holds one attempt, not all of them": "cp stdout got.txt\ncmp got.txt want.txt\n" +
			"\n-- want.txt --\nsettled\n",
	}

	for abortName, abort := range aborts {
		for scriptName, script := range scripts {
			t.Run(abortName+", "+scriptName, func(t *testing.T) {
				logger, logBuf := bufferedTestLogger(t)
				// verbose: true, matching the -verbose run `make scenarios` does.
				// testscript rewinds (discards)
				// a passing phase's log detail when its T is not verbose
				// (testscript.go:541), which would erase the eventually diagnostic
				// below before it ever reached this buffer.
				adapter := NewTestscriptT(logger, true)
				calls := 0
				cmds := map[string]func(*testscript.TestScript, bool, []string){
					"noisy": func(ts *testscript.TestScript, neg bool, args []string) {
						calls++
						if calls < 3 {
							fmt.Fprintf(ts.Stdout(), "attempt %d: not ready\n", calls)
							fmt.Fprintf(ts.Stderr(), "attempt %d complained\n", calls)
							abort(ts)
						}
						fmt.Fprintln(ts.Stdout(), "settled")
						fmt.Fprintln(ts.Stderr(), "settled on stderr")
					},
				}
				cmds["eventually"] = EventuallyCmd(cmds)

				testscript.RunT(adapter, testscript.Params{
					Files: []string{writeScript(t, "eventually 5s 10ms noisy\n"+script)},
					Cmds:  cmds,
				})

				require.False(t, adapter.Failed, "script log:\n%s", logBuf)
				require.Equal(t, 3, calls, "the script must have exercised a real retry")
				// Two Contains rather than one on the fully formatted line:
				// slog's TextHandler backslash-escapes the %q-quoted command
				// name when it quotes this multi-line message as a whole, and
				// asserting on that escaping would pin the buffered handler's
				// choice of encoding rather than EventuallyCmd's own diagnostic text.
				logged := logBuf.String()
				require.Contains(t, logged, "eventually: attempt 1 for", "the per-attempt diagnostic must reach the test log, not just stay out of stderr")
				require.Contains(t, logged, "did not succeed", "the per-attempt diagnostic must reach the test log, not just stay out of stderr")
			})
		}
	}
}

func writeScript(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "script.txtar")
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	return path
}

func testLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// bufferedTestLogger is testLogger's counterpart for tests that need to
// assert on what actually reached the log, not merely observe that nothing
// panicked writing to it. testLogger's io.Discard sink is shared by other
// test files in this package (gnokey_test.go, gpao_test.go, httpget_test.go),
// so it keeps its original signature; this is a separate helper rather than
// a change to it.
func bufferedTestLogger(t *testing.T) (*slog.Logger, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	return slog.New(slog.NewTextHandler(&buf, nil)), &buf
}

// The scenarios rely on the resolved budget, not on the arguments they wrote:
// almost every eventually line names none, so the defaults below are a contract
// and the literals here pin them. The last four cases are the reason the two
// durations can be optional at all -- a command name, or a command's own
// duration argument, must never be eaten as a budget.
func TestParseEventuallyArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    eventuallyArgs
		wantErr string
	}{
		{
			name: "a line naming no budget gets the defaults",
			args: []string{"gnokey", "query", "vm/qinertpaths"},
			want: eventuallyArgs{
				timeout:  30 * time.Second,
				interval: time.Second,
				cmd:      "gnokey",
				cmdArgs:  []string{"query", "vm/qinertpaths"},
			},
		},
		{
			name: "a timeout alone keeps the default interval",
			args: []string{"90s", "gnokey", "query", "vm/qrender"},
			want: eventuallyArgs{
				timeout:  90 * time.Second,
				interval: time.Second,
				cmd:      "gnokey",
				cmdArgs:  []string{"query", "vm/qrender"},
			},
		},
		{
			name: "both durations override both defaults",
			args: []string{"5s", "10ms", "flaky"},
			want: eventuallyArgs{
				timeout:  5 * time.Second,
				interval: 10 * time.Millisecond,
				cmd:      "flaky",
				cmdArgs:  []string{},
			},
		},
		{
			name: "a command's own duration argument is not a budget",
			args: []string{"sleep", "5s"},
			want: eventuallyArgs{
				timeout:  30 * time.Second,
				interval: time.Second,
				cmd:      "sleep",
				cmdArgs:  []string{"5s"},
			},
		},
		{
			name: "a timeout followed by such a command takes only the timeout",
			args: []string{"60s", "sleep", "5s"},
			want: eventuallyArgs{
				timeout:  60 * time.Second,
				interval: time.Second,
				cmd:      "sleep",
				cmdArgs:  []string{"5s"},
			},
		},
		{
			name:    "no arguments at all",
			args:    []string{},
			wantErr: "usage: eventually [timeout [interval]] [-stdout regex] <cmd> [args...]",
		},
		{
			name:    "a timeout with no command to run",
			args:    []string{"30s"},
			wantErr: "usage: eventually [timeout [interval]] [-stdout regex] <cmd> [args...]",
		},
		{
			name:    "both durations with no command to run",
			args:    []string{"30s", "1s"},
			wantErr: "usage: eventually [timeout [interval]] [-stdout regex] <cmd> [args...]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEventuallyArgs(tt.args)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

// A wait that only checks the exit status cannot express "until this value
// appears": a query that answers an empty list exits 0, so the wait returns at
// once and the assertion after it reads state that has not arrived. The gate
// moves that check inside the attempt.
func TestParseEventuallyArgsStdoutGate(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    eventuallyArgs
		wantErr string
	}{
		{
			name: "a gate before the command",
			args: []string{"-stdout", "gno.land/r/x", "gnokey", "query", "vm/qinertpaths"},
			want: eventuallyArgs{
				timeout:  30 * time.Second,
				interval: time.Second,
				stdout:   "gno.land/r/x",
				cmd:      "gnokey",
				cmdArgs:  []string{"query", "vm/qinertpaths"},
			},
		},
		{
			name: "a gate after an explicit budget",
			args: []string{"60s", "2s", "-stdout", "board", "gnokey", "query", "vm/qrender"},
			want: eventuallyArgs{
				timeout:  60 * time.Second,
				interval: 2 * time.Second,
				stdout:   "board",
				cmd:      "gnokey",
				cmdArgs:  []string{"query", "vm/qrender"},
			},
		},
		{
			name: "no gate leaves it empty",
			args: []string{"gnokey", "query", "vm/qinertpaths"},
			want: eventuallyArgs{
				timeout:  30 * time.Second,
				interval: time.Second,
				cmd:      "gnokey",
				cmdArgs:  []string{"query", "vm/qinertpaths"},
			},
		},
		{
			// The gate is a regex, so a bad one is the author's mistake and has
			// to be reported as one rather than as a wait that never matches.
			name:    "an unparsable pattern is a usage error",
			args:    []string{"-stdout", "[", "gnokey"},
			wantErr: "invalid -stdout pattern",
		},
		{
			name:    "a gate with no pattern",
			args:    []string{"-stdout"},
			wantErr: eventuallyUsage,
		},
		{
			name:    "a gate with a pattern but no command",
			args:    []string{"-stdout", "board"},
			wantErr: eventuallyUsage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEventuallyArgs(tt.args)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// The gate has to be evaluated per attempt, or it is no better than an
// assertion on the line after the wait.
func TestEventuallyStdoutGatePollsUntilTheValueAppears(t *testing.T) {
	calls := 0
	cmds := map[string]func(*testscript.TestScript, bool, []string){
		// Exits 0 every time, as a list query does, and only reports the value
		// on its third answer.
		"listing": func(ts *testscript.TestScript, neg bool, args []string) {
			calls++
			if calls < 3 {
				fmt.Fprintln(ts.Stdout(), "other/paths")
				return
			}
			fmt.Fprintln(ts.Stdout(), "gno.land/r/out/blind")
		},
	}
	cmds["eventually"] = EventuallyCmd(cmds)

	logger, logBuf := bufferedTestLogger(t)
	adapter := NewTestscriptT(logger, true)
	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, "eventually 5s 10ms -stdout 'gno.land/r/out/blind' listing\nstdout 'blind'\n")},
		Cmds:  cmds,
	})

	require.False(t, adapter.Failed, "script log:\n%s", logBuf)
	assert.Equal(t, 3, calls, "a command exiting 0 with the wrong output must not end the wait")
}

// Giving up has to name the gate, or the failure reads as the command being
// broken when the command answered every time.
func TestEventuallyStdoutGateReportsTheUnmetPattern(t *testing.T) {
	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"listing": func(ts *testscript.TestScript, neg bool, args []string) {
			fmt.Fprintln(ts.Stdout(), "other/paths")
		},
	}
	cmds["eventually"] = EventuallyCmd(cmds)

	logger, logBuf := bufferedTestLogger(t)
	adapter := NewTestscriptT(logger, true)
	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, "eventually 100ms 10ms -stdout 'never-appears' listing\n")},
		Cmds:  cmds,
	})

	require.True(t, adapter.Failed, "an unmet gate must fail the script")
	assert.Contains(t, logBuf.String(), "never-appears",
		"the deadline message must name the pattern that never matched")
}

// repeat's own usage string documents "repeat [-all] N <cmd> [args...]", and a
// sub-command that takes no arguments is the shortest form of it. Refusing that
// form leaves the only repeatable commands the ones that happen to take an
// argument, and reports the refusal with the very usage line being obeyed.
func TestRepeatRunsASubCommandThatTakesNoArguments(t *testing.T) {
	calls := 0
	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"tick": func(ts *testscript.TestScript, neg bool, args []string) { calls++ },
	}
	cmds["repeat"] = RepeatCmd(cmds)

	adapter := NewTestscriptT(testLogger(t), false)
	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, "repeat 3 tick\n")},
		Cmds:  cmds,
	})

	require.False(t, adapter.Failed, "the shortest documented form has to run")
	assert.Equal(t, 3, calls)
}

// A wait that gives up has to name itself. Both verbs recover a failed attempt
// through one helper, and a diagnostic hard-coding the other verb's name sends
// the reader looking through their script for a line that is not there.
func TestEventuallyNamesItselfWhenAnAttemptPanics(t *testing.T) {
	logger, logBuf := bufferedTestLogger(t)
	adapter := NewTestscriptT(logger, true)
	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"exploding": func(ts *testscript.TestScript, neg bool, args []string) {
			panic("boom")
		},
	}
	cmds["eventually"] = EventuallyCmd(cmds)

	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, "eventually 50ms 10ms exploding\n")},
		Cmds:  cmds,
	})

	require.True(t, adapter.Failed)
	assert.Contains(t, logBuf.String(), "eventually: iteration error")
	assert.NotContains(t, logBuf.String(), "repeat:", "the diagnostic must not name a verb the script never used")
}
