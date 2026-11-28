package integration

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

// The adapter stands in for testing.T, whose Log formats like Println: a space
// between every operand, including two strings.
func TestAdapterLogSeparatesOperands(t *testing.T) {
	var buf bytes.Buffer
	adapter := NewTestscriptT(slog.New(slog.NewTextHandler(&buf, nil)), false)

	adapter.Run("tour", func(testscript.T) {})

	require.Contains(t, buf.String(), "=== RUN tour")
	require.Contains(t, buf.String(), "--- PASS tour")
}

// TestRunTurnsAPanicIntoAFailedScript pins that a fault in the harness's own Go
// code fails the script rather than the process.
//
// testscript re-panics anything that is not its own failNow sentinel, and this
// goroutine is the last frame before the runtime. A panic that escapes it takes
// the process down without unwinding the run: the cluster's validators and the
// oracle stay alive, and the temp directories they were given stay behind.
func TestRunTurnsAPanicIntoAFailedScript(t *testing.T) {
	adapter := NewTestscriptT(testLogger(t), false)
	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"exploding": func(ts *testscript.TestScript, neg bool, args []string) {
			panic("a bug in a verb")
		},
	}

	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, "exploding\n")},
		Cmds:  cmds,
	})

	require.True(t, adapter.Failed, "the script must be reported failed")
}
