package termlog

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The one output path a run has. Nothing else in the module reads a record, so
// what this handler drops is dropped for good.
func TestHandlerRendersARecord(t *testing.T) {
	var out bytes.Buffer
	logger := slog.New(NewHandler(&out, true)).With("component", "cluster")

	logger.Info("cluster ready", "validators", 3)

	line := out.String()
	assert.Contains(t, line, "cluster ready")
	assert.Contains(t, line, "cluster", "the component names which part of the run spoke")
	assert.Contains(t, line, "validators=", "an attribute reaches the line")
	assert.Contains(t, line, "3")
}

// Verbosity is the only level knob a run has: -verbose is what a scenario
// author turns on to see what the cluster and the oracle did.
func TestHandlerLevelFollowsVerbosity(t *testing.T) {
	for _, verbose := range []bool{false, true} {
		var out bytes.Buffer
		slog.New(NewHandler(&out, verbose)).Debug("a diagnostic")

		if verbose {
			assert.Contains(t, out.String(), "a diagnostic")
		} else {
			assert.Empty(t, out.String(), "debug records are for -verbose alone")
		}
	}
}

// A value that computes itself is rendered as what it computes, not as the
// thing that would have computed it.
type lazyValue struct{}

func (lazyValue) LogValue() slog.Value { return slog.StringValue("RESOLVED") }

func TestHandlerResolvesAComputedValue(t *testing.T) {
	var out bytes.Buffer
	slog.New(NewHandler(&out, true)).Info("deploying", "package", lazyValue{})

	assert.Contains(t, out.String(), "RESOLVED")
}

// slog requires a handler to treat WithGroup("") as a no-op, and a handler that
// invents an attribute for it labels every later record with an empty group.
// Called on the handler rather than through slog.Logger, which short-circuits
// the empty name before the handler ever sees it.
func TestHandlerIgnoresAnEmptyGroup(t *testing.T) {
	var out bytes.Buffer
	slog.New(NewHandler(&out, true).WithGroup("")).Info("still plain")

	assert.NotContains(t, out.String(), "group=")
}

// Every logger in a run derives from one handler and they all write to the same
// stderr: the cluster's, the oracle's, and the scenario driver's. Interleaved
// writes would corrupt the only record of what happened.
func TestHandlerSerializesConcurrentWrites(t *testing.T) {
	var out bytes.Buffer
	base := NewHandler(&out, true)

	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			logger := slog.New(base.WithAttrs([]slog.Attr{slog.Int("worker", i)}))
			logger.Info("working")
		})
	}
	wg.Wait()

	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	require.Len(t, lines, 20)
	for _, line := range lines {
		assert.Contains(t, line, "working")
	}
}

// Enabled is what slog asks before it builds a record at all.
func TestHandlerEnabled(t *testing.T) {
	quiet := NewHandler(&bytes.Buffer{}, false)
	assert.False(t, quiet.Enabled(context.Background(), slog.LevelDebug))
	assert.True(t, quiet.Enabled(context.Background(), slog.LevelInfo))
}
