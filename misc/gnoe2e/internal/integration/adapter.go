package integration

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"github.com/rogpeppe/go-internal/testscript"
)

// TestscriptT adapts slog.Logger to satisfy testscript.T.
// It uses runtime.Goexit() for FailNow/Fatal/Skip to match
// the semantics that testscript expects (same as testing.T).
type TestscriptT struct {
	Logger  *slog.Logger
	Failed  bool
	Skipped bool

	// verbose is unexported to avoid conflicting with the Verbose() interface method.
	verbose bool
}

var _ testscript.T = (*TestscriptT)(nil)

// NewTestscriptT creates a TestscriptT with the given logger and verbosity.
func NewTestscriptT(logger *slog.Logger, verbose bool) *TestscriptT {
	return &TestscriptT{Logger: logger, verbose: verbose}
}

func (t *TestscriptT) FailNow() {
	t.Failed = true
	runtime.Goexit()
}

func (t *TestscriptT) Fatal(args ...any) {
	t.Log(args...)
	t.FailNow()
}

func (t *TestscriptT) Skip(args ...any) {
	t.Log(args...)
	t.Skipped = true
	runtime.Goexit()
}

func (t *TestscriptT) Parallel() {} // no-op: run scripts sequentially

func (t *TestscriptT) Log(args ...any) {
	// testing.T.Log, which this stands in for, formats like Println: a space
	// between every operand, where fmt.Sprint adds one only between operands
	// that are not both strings.
	t.Logger.Info(strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func (t *TestscriptT) Verbose() bool { return t.verbose }

// Run executes f in an isolated goroutine so that runtime.Goexit()
// within f only terminates that goroutine, not the caller.
// This matches how testing.T.Run works internally.
func (t *TestscriptT) Run(name string, f func(testscript.T)) {
	t.Log("=== RUN", name)

	child := &TestscriptT{Logger: t.Logger, verbose: t.verbose}
	done := make(chan struct{})
	go func() {
		defer close(done)
		f(child)
	}()
	<-done

	if child.Failed {
		t.Log("--- FAIL", name)
		t.Failed = true
	} else if child.Skipped {
		t.Log("--- SKIP", name)
	} else {
		t.Log("--- PASS", name)
	}
}
