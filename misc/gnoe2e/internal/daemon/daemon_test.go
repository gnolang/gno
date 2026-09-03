package daemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// The test binary impersonates a daemon when GNOE2E_FAKE_DAEMON is set, so the
// supervisor is exercised against a real child process without building one.
//
// The parked modes sleep rather than blocking on an empty select: a select
// with no cases is the sole goroutine here, so the runtime's deadlock detector
// kills the process instead of leaving it parked. A pending timer keeps the
// detector quiet, and SIGTERM still terminates the process, which is what Stop
// relies on.
func TestMain(m *testing.M) {
	switch os.Getenv("GNOE2E_FAKE_DAEMON") {
	case "":
		os.Exit(m.Run())
	case "serve":
		if path := os.Getenv("GNOE2E_FAKE_READY_FILE"); path != "" {
			_ = os.WriteFile(path, []byte("ready"), 0o644)
		}
		time.Sleep(time.Hour)
	case "die":
		fmt.Fprintln(os.Stderr, "fake: cannot reach the node")
		os.Exit(3)
	case "die_after_ready":
		if path := os.Getenv("GNOE2E_FAKE_READY_FILE"); path != "" {
			_ = os.WriteFile(path, []byte("ready"), 0o644)
		}
		// Long enough that the probe has almost certainly already polled the
		// ready file (readyPollInterval is 50ms) before this process dies, so
		// the race lands on the settle check rather than the pre-probe exit
		// path already covered by TestStartFailsFastWhenProcessExits.
		time.Sleep(100 * time.Millisecond)
		fmt.Fprintln(os.Stderr, "fake: ready, then the node vanished")
		os.Exit(3)
	case "exit_clean":
		// Exits 0, so cmd.Wait reports nil and the supervisor has no exit
		// status to quote.
		fmt.Fprintln(os.Stderr, "fake: nothing to serve")
		os.Exit(0)
	case "silent":
		time.Sleep(time.Hour)
	default:
		os.Exit(9)
	}
}

// fileProbe reports ready once path exists.
func fileProbe(path string) Probe {
	return func(context.Context) error {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("not ready: %w", err)
		}
		return nil
	}
}

func fakeConfig(t *testing.T, mode string, ready Probe) Config {
	t.Helper()
	return Config{
		Name:       "fake",
		BinaryPath: os.Args[0],
		Env:        []string{"GNOE2E_FAKE_DAEMON=" + mode},
		Ready:      ready,
		ReadyWait:  10 * time.Second,
	}
}

func TestStartWaitsForReady(t *testing.T) {
	readyFile := filepath.Join(t.TempDir(), "ready")
	cfg := fakeConfig(t, "serve", fileProbe(readyFile))
	cfg.Env = append(cfg.Env, "GNOE2E_FAKE_READY_FILE="+readyFile)

	d, err := Start(context.Background(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Stop() })

	require.FileExists(t, readyFile)
}

func TestStartFailsFastWhenProcessExits(t *testing.T) {
	readyFile := filepath.Join(t.TempDir(), "never")

	_, err := Start(context.Background(), fakeConfig(t, "die", fileProbe(readyFile)))

	require.Error(t, err)
	require.Contains(t, err.Error(), "exited before it was ready")
	require.Contains(t, err.Error(), "cannot reach the node",
		"the error must carry the process output, or a dead daemon is undiagnosable")
}

func TestStartFailsWhenProcessExitsAfterReady(t *testing.T) {
	readyFile := filepath.Join(t.TempDir(), "ready")
	cfg := fakeConfig(t, "die_after_ready", fileProbe(readyFile))
	cfg.Env = append(cfg.Env, "GNOE2E_FAKE_READY_FILE="+readyFile)

	_, err := Start(context.Background(), cfg)

	require.Error(t, err)
	require.Contains(t, err.Error(), "became ready, then exited")
	require.NotContains(t, err.Error(), "exited before it was ready",
		"the process WAS ready before it died; that message would send a reader looking for the wrong fault")
	require.Contains(t, err.Error(), "fake: ready, then the node vanished",
		"the error must carry the process output, or a dead daemon is undiagnosable")
}

// A process that exits 0 gives cmd.Wait a nil error, and %w on nil renders
// %!w(<nil>) -- in the middle of the one message an author reads to find out
// why the daemon never came up.
func TestStartErrorSurvivesACleanExit(t *testing.T) {
	readyFile := filepath.Join(t.TempDir(), "never")

	_, err := Start(context.Background(), fakeConfig(t, "exit_clean", fileProbe(readyFile)))

	require.Error(t, err)
	require.Contains(t, err.Error(), "exited before it was ready")
	require.NotContains(t, err.Error(), "%!w",
		"a nil wait error must not leave a broken format verb where the exit status goes")
	require.Contains(t, err.Error(), "fake: nothing to serve",
		"the error must carry the process output, or a dead daemon is undiagnosable")
}

func TestStartTimesOutWhenNeverReady(t *testing.T) {
	readyFile := filepath.Join(t.TempDir(), "never")
	cfg := fakeConfig(t, "silent", fileProbe(readyFile))
	cfg.ReadyWait = 300 * time.Millisecond

	_, err := Start(context.Background(), cfg)

	require.Error(t, err)
	require.Contains(t, err.Error(), "not ready after")
}

func TestStopIsIdempotent(t *testing.T) {
	readyFile := filepath.Join(t.TempDir(), "ready")
	cfg := fakeConfig(t, "serve", fileProbe(readyFile))
	cfg.Env = append(cfg.Env, "GNOE2E_FAKE_READY_FILE="+readyFile)

	d, err := Start(context.Background(), cfg)
	require.NoError(t, err)

	require.NoError(t, d.Stop())
	require.NoError(t, d.Stop())
}
