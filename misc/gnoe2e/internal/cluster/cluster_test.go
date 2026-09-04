package cluster

import (
	"context"
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

// TestMain dispatches to the signal-ignoring helper subprocess when invoked
// with the GNOE2E_TEST_IGNORE_SIGNALS env var, so the tests that need a process
// which refuses to stop have one.
func TestMain(m *testing.M) {
	if os.Getenv("GNOE2E_TEST_IGNORE_SIGNALS") == "1" {
		runUnstoppableHelper()
		return
	}
	os.Exit(m.Run())
}

// runUnstoppableHelper is the child half of spawnUnstoppable: it ignores both
// signals the harness stops a node with, announces that the handlers are
// installed by creating the file named in GNOE2E_TEST_READY_FILE, and then
// blocks until it is killed.
func runUnstoppableHelper() {
	signal.Ignore(syscall.SIGTERM, syscall.SIGINT)
	if path := os.Getenv("GNOE2E_TEST_READY_FILE"); path != "" {
		if err := os.WriteFile(path, []byte("ready"), FilePermissions); err != nil {
			os.Exit(1)
		}
	}
	time.Sleep(30 * time.Second)
	os.Exit(0)
}

// spawnUnstoppable starts a helper subprocess that ignores SIGINT and SIGTERM,
// and returns only once its handlers are installed.
//
// Waiting for the handshake rather than sleeping is what keeps these tests
// honest: a helper still in runtime startup takes the default disposition and
// dies to the very signal the test needs it to survive, which turns the test
// green for the opposite of the reason it was written.
func spawnUnstoppable(t *testing.T) *exec.Cmd {
	t.Helper()

	ready := filepath.Join(t.TempDir(), "ready")
	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(),
		"GNOE2E_TEST_IGNORE_SIGNALS=1",
		"GNOE2E_TEST_READY_FILE="+ready,
	)
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	require.Eventually(t, func() bool {
		_, err := os.Stat(ready)
		return err == nil
	}, 30*time.Second, 5*time.Millisecond, "helper never installed its signal handlers")

	return cmd
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

// TestStopValidatorReportsAValidatorThatHadAlreadyDied pins that stop does not
// claim a node it never stopped.
//
// Nothing watches liveness between boot and a "validator stop" line, and a
// validator that died on its own still holds its *os.Process. Reporting success
// tells a scenario it caused an outage that had in fact started earlier, and
// the reason the node went down is never reported at all.
func TestStopValidatorReportsAValidatorThatHadAlreadyDied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process signals not supported on windows")
	}
	c := spawnFakeValidators(t, 2)

	// Killed behind the cluster's back, which is what a crash looks like from
	// here: no bookkeeping is updated and the node keeps its process.
	require.NoError(t, c.Validators[1].Process.Kill())
	<-c.Validators[1].Exited()

	err := c.StopValidator(context.Background(), 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exited")
	assert.Nil(t, c.Validators[1].Process, "a validator that is gone holds no process")

	// The stop failed, but the cluster still describes itself correctly.
	require.Len(t, c.Validators, 2)
	assert.NotNil(t, c.Validators[0].Process, "the other validator is untouched")
}

// TestWaitForFirstBlockNamesWhatItPolled pins that the one deadline a cluster
// which cannot reach quorum actually hits says something usable.
//
// Every node can be up and answering while the chain commits nothing, so this
// timeout is the normal report of a broken validator set. Without the node, the
// address and the height it got to, it names no suspect at all.
func TestWaitForFirstBlockNamesWhatItPolled(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process signals not supported on windows")
	}

	cmd := exec.Command("sleep", "30")
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	node := &Node{Index: 1, DataDir: t.TempDir(), RPCAddr: "tcp://127.0.0.1:1", Process: cmd.Process}
	err := waitForFirstBlock(t.Context(), node, 50*time.Millisecond)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "node 1", "the error has to name the node it polled")
	assert.Contains(t, err.Error(), "127.0.0.1:1", "and the address it polled")
}

// TestNodeLogTailsQuotesEveryNodeThatWroteSomething pins the aggregate the
// first-block failure reports.
//
// The polled node is the one node that is usually fine: it is waiting for a
// quorum the others never joined, so their logs hold the reason and the
// cluster's directory is removed as soon as the boot is reported failed.
func TestNodeLogTailsQuotesEveryNodeThatWroteSomething(t *testing.T) {
	withStderr := func(content string) string {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "stderr.log"), []byte(content), FilePermissions))
		return dir
	}

	got := nodeLogTails([]*Node{
		{Index: 0, DataDir: withStderr("WHY_ZERO_DIED\n")},
		{Index: 1, DataDir: t.TempDir()}, // never wrote anything
		nil,                              // never set up
		{Index: 3, DataDir: withStderr("WHY_THREE_DIED\n")},
	})

	assert.Contains(t, got, "WHY_ZERO_DIED")
	assert.Contains(t, got, "WHY_THREE_DIED")
	assert.Contains(t, got, "validator 0")
	assert.Contains(t, got, "validator 3")
	assert.NotContains(t, got, "validator 1", "a node with an empty log adds nothing but noise")
}
