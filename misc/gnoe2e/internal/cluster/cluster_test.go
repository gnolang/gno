package cluster

import (
	"context"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain dispatches to the SIGTERM-ignoring helper subprocess when
// invoked with the GNOE2E_TEST_IGNORE_SIGTERM env var. Used by
// TestCluster_Halt_CtxCancelled to obtain a process that reliably
// refuses to exit on SIGTERM.
func TestMain(m *testing.M) {
	if os.Getenv("GNOE2E_TEST_IGNORE_SIGTERM") == "1" {
		signal.Ignore(syscall.SIGTERM)
		time.Sleep(30 * time.Second)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// ---- Halt tests

// TestCluster_Halt_NoValidators verifies Halt errors when the cluster has
// no running processes.
func TestCluster_Halt_NoValidators(t *testing.T) {
	c := &Cluster{}
	err := c.Halt(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

// TestCluster_Halt_SignalsAndWaits spawns a real subprocess that traps
// SIGTERM, registers it as a Node.Process, and verifies Halt sends
// SIGTERM and waits for the process to exit cleanly.
func TestCluster_Halt_SignalsAndWaits(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SIGTERM not supported on windows")
	}

	// `sleep 30` exits cleanly on SIGTERM (default disposition: terminate).
	// Halt sends SIGTERM and waits for Process.Wait, which is exactly the
	// behavior we want to verify without depending on a real gnoland binary.
	cmd := exec.Command("sleep", "30")
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		// Belt-and-braces in case the test fails before Halt runs.
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	node := &Node{Index: 0, Process: cmd.Process}
	c := &Cluster{Validators: []*Node{node}}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	require.NoError(t, c.Halt(ctx))
	elapsed := time.Since(start)

	// Halt should return promptly once the child exits, well below the
	// 30-second sleep duration.
	assert.Less(t, elapsed, 5*time.Second, "Halt should return quickly after SIGTERM")
}

// TestCluster_BootFromExistingDataDir_RejectsAlreadyRunning verifies
// the precondition check: a cluster with running validators must not be
// re-booted. Reusing a cluster value while its existing process is
// still alive risks leaking the old process and double-binding ports.
func TestCluster_BootFromExistingDataDir_RejectsAlreadyRunning(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("subprocess semantics differ on windows")
	}

	cmd := exec.Command("sleep", "30")
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	c := &Cluster{Validators: []*Node{{Index: 0, Process: cmd.Process}}}
	err := c.BootFromExistingDataDir(context.Background(), t.TempDir(), "/bin/echo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

// TestCluster_BootFromExistingDataDir_RejectsMissingDataDir verifies
// the precondition check on the data dir path.
func TestCluster_BootFromExistingDataDir_RejectsMissingDataDir(t *testing.T) {
	c := &Cluster{}
	err := c.BootFromExistingDataDir(context.Background(), "/nonexistent/data/dir", "/bin/echo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "data dir")
}

// TestCluster_BootFromExistingDataDir_RejectsMissingBinary verifies
// the precondition check on the binary path.
func TestCluster_BootFromExistingDataDir_RejectsMissingBinary(t *testing.T) {
	c := &Cluster{}
	err := c.BootFromExistingDataDir(context.Background(), t.TempDir(), "/nonexistent/binary")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "binary")
}

// ---- BootFromGenesis tests

// TestCluster_BootFromGenesis_RejectsAlreadyRunning verifies the
// precondition check: a cluster with running validators must not be
// re-booted, even with a fresh genesis. Reusing a cluster value while
// its existing process is still alive risks leaking the old process.
func TestCluster_BootFromGenesis_RejectsAlreadyRunning(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("subprocess semantics differ on windows")
	}

	cmd := exec.Command("sleep", "30")
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	tmp := t.TempDir()
	genesisPath := filepath.Join(tmp, "genesis.json")
	require.NoError(t, os.WriteFile(genesisPath, []byte("{}"), 0o644))

	c := &Cluster{Validators: []*Node{{Index: 0, Process: cmd.Process}}}
	err := c.BootFromGenesis(context.Background(), genesisPath, "/bin/echo", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

// TestCluster_BootFromGenesis_RejectsMissingGenesis verifies the
// precondition check on the supplied genesis file path.
func TestCluster_BootFromGenesis_RejectsMissingGenesis(t *testing.T) {
	c := &Cluster{}
	tmp := t.TempDir()
	missingPath := filepath.Join(tmp, "missing.json")

	err := c.BootFromGenesis(context.Background(), missingPath, "/bin/echo", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "genesis")
}

// TestCluster_BootFromGenesis_RejectsMissingBinary verifies the
// precondition check on the binary path.
func TestCluster_BootFromGenesis_RejectsMissingBinary(t *testing.T) {
	c := &Cluster{}

	tmp := t.TempDir()
	genesisPath := filepath.Join(tmp, "genesis.json")
	require.NoError(t, os.WriteFile(genesisPath, []byte("{}"), 0o644))

	err := c.BootFromGenesis(context.Background(), genesisPath, "/nonexistent/binary", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "binary")
}

// ---- BootLogReader tests

// TestCluster_BootLogReader_NilBeforeBoot verifies the accessor returns
// nil when no boot has occurred. HardforkReplayMode.AssertPostTransition
// distinguishes "boot did not capture logs" from "logs captured but no
// replay summary" by checking for nil first.
func TestCluster_BootLogReader_NilBeforeBoot(t *testing.T) {
	c := &Cluster{}
	assert.Nil(t, c.BootLogReader())
}

// TestCluster_BootLogReader_ReturnsSnapshot verifies the reader exposes
// captured bytes after manual writes to the buffer. Snapshots are
// independent of subsequent writes, so reading an old reader after a
// later write must not yield the new bytes.
func TestCluster_BootLogReader_ReturnsSnapshot(t *testing.T) {
	c := &Cluster{bootLog: newBootLogBuffer()}
	_, err := c.bootLog.Write([]byte("hello\n"))
	require.NoError(t, err)

	r := c.BootLogReader()
	require.NotNil(t, r)

	// First read sees the captured byte.
	got, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(got))

	// Subsequent writes don't bleed into already-handed-out readers.
	_, err = c.bootLog.Write([]byte("world\n"))
	require.NoError(t, err)
	got2, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Empty(t, got2, "previously-handed-out reader must be a snapshot")

	// A fresh reader sees the cumulative bytes.
	r2 := c.BootLogReader()
	got3, err := io.ReadAll(r2)
	require.NoError(t, err)
	assert.Equal(t, "hello\nworld\n", string(got3))
}

// TestCluster_Halt_CtxCancelled verifies Halt returns an error when ctx
// expires before the process exits. The helper subprocess (see
// TestMain) ignores SIGTERM via signal.Ignore so the wait blocks until
// ctx fires.
func TestCluster_Halt_CtxCancelled(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("signal semantics differ on windows")
	}

	// Re-invoke the test binary as a helper subprocess that ignores
	// SIGTERM via signal.Ignore (see TestMain). This guarantees the
	// child stays alive after Halt sends SIGTERM, forcing Halt to hit
	// the ctx.Done branch.
	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), "GNOE2E_TEST_IGNORE_SIGTERM=1")
	require.NoError(t, cmd.Start())
	// Wait for the helper to install signal.Ignore before sending SIGTERM,
	// otherwise the early SIGTERM uses the default disposition (terminate).
	time.Sleep(200 * time.Millisecond)
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	node := &Node{Index: 0, Process: cmd.Process}
	c := &Cluster{Validators: []*Node{node}}

	// Tight ctx so the test doesn't take 30s.
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	err := c.Halt(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "halt cancelled")
}

// ---- Per-validator lifecycle

// spawnFakeValidators returns a cluster of n nodes backed by real but inert
// child processes, so lifecycle bookkeeping can be exercised without booting
// a chain. `sleep` exits on SIGTERM, which is what StopValidator sends.
func spawnFakeValidators(t *testing.T, n int) *Cluster {
	t.Helper()
	c := &Cluster{TempDir: t.TempDir()}
	for i := range n {
		cmd := exec.Command("sleep", "30")
		require.NoError(t, cmd.Start())
		node := &Node{
			Index:   i,
			DataDir: t.TempDir(),
			RPCAddr: "tcp://127.0.0.1:1",
			Process: cmd.Process,
		}
		t.Cleanup(func() {
			_ = cmd.Process.Kill()
			<-node.Exited()
		})
		c.Validators = append(c.Validators, node)
	}
	return c
}

// TestStopValidatorLeavesTheRestOfTheClusterAlone is the property the outage
// scenarios depend on: Halt takes the whole cluster down at once, so without
// a per-node stop there is no way to express losing one validator while the
// others keep producing blocks.
func TestStopValidatorLeavesTheRestOfTheClusterAlone(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process signals not supported on windows")
	}
	c := spawnFakeValidators(t, 3)

	require.NoError(t, c.StopValidator(context.Background(), 1))

	require.Nil(t, c.Validators[1].Process, "a stopped validator holds no process")

	// The others are untouched, and the slice still describes the cluster the
	// caller built. Collapsing it is what the boot-from-data-dir paths do, and
	// it would leave a scenario unable to name the nodes it did not stop.
	require.Len(t, c.Validators, 3, "StopValidator must not collapse the validator set")
	for _, i := range []int{0, 2} {
		require.NotNil(t, c.Validators[i].Process, "validator %d should still be running", i)
		require.NoError(t, c.Validators[i].Process.Signal(syscall.Signal(0)),
			"validator %d should still be alive", i)
	}
}

// TestStopValidatorRejectsWhatItCannotStop covers the ways a caller can name
// something unstoppable. Each has to be an error rather than a silent no-op:
// a scenario that believes it stopped a node, and did not, goes on to assert
// against a cluster that never lost anything.
func TestStopValidatorRejectsWhatItCannotStop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process signals not supported on windows")
	}
	c := spawnFakeValidators(t, 2)

	require.Error(t, c.StopValidator(context.Background(), 5), "index past the end")
	require.Error(t, c.StopValidator(context.Background(), -1), "negative index")

	require.NoError(t, c.StopValidator(context.Background(), 0))
	require.Error(t, c.StopValidator(context.Background(), 0), "already stopped")
}

// TestRestartValidatorRejectsARunningNode pins that restart is not a silent
// no-op on a live validator. It would otherwise report success having done
// nothing, and a scenario asserting recovery would pass without the node ever
// having gone away.
func TestRestartValidatorRejectsARunningNode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process signals not supported on windows")
	}
	c := spawnFakeValidators(t, 2)

	err := c.RestartValidator(context.Background(), 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "running")

	require.Error(t, c.RestartValidator(context.Background(), 9), "index past the end")
}
