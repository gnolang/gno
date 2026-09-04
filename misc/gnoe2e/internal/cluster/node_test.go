package cluster

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWaitForNodeReadyAbortsWhenTheProcessHasExited pins that readiness
// polling gives up as soon as the child is gone.
//
// Nothing else notices a node that dies during boot: the RPC it polls answers
// "connection refused" whether the node has not opened its port yet or will
// never open it again, so without a liveness check the poll burns its whole
// readiness budget and then blames a timeout. The process is the only thing
// that tells those two apart, and they need opposite fixes.
func TestWaitForNodeReadyAbortsWhenTheProcessHasExited(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process signals not supported on windows")
	}

	cmd := exec.Command("sleep", "30")
	require.NoError(t, cmd.Start())
	require.NoError(t, cmd.Process.Kill())

	// Nothing ever listens here, so the RPC poll can never succeed on its
	// own -- only the process check can end this call.
	node := &Node{Index: 2, DataDir: t.TempDir(), RPCAddr: "tcp://127.0.0.1:1", Process: cmd.Process}

	start := time.Now()
	err := WaitForNodeReady(context.Background(), node)
	elapsed := time.Since(start)

	require.Error(t, err)
	require.Contains(t, err.Error(), "exited",
		"the error has to name the process death, not a timeout")
	require.Less(t, elapsed, 30*time.Second,
		"must abort on process death rather than waiting out the readiness budget")
}

// TestCleanupNodesIsSilentWhenTheProcessAlreadyExited pins that tearing down a
// node whose process is already gone reports nothing.
//
// Signal and Kill both answer os.ErrProcessDone once a process has been
// reaped. Reading a failed signal as a reason to kill, and a failed kill as a
// fault worth logging, turns every already-stopped node into an ERROR line:
// a scenario that halts its own cluster then ends by printing failures it
// caused on purpose, and a reader learns to skip that line.
func TestCleanupNodesIsSilentWhenTheProcessAlreadyExited(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process signals not supported on windows")
	}

	cmd := exec.Command("true")
	require.NoError(t, cmd.Start())
	require.NoError(t, cmd.Wait())

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	CleanupNodes(logger, []*Node{{Index: 0, DataDir: t.TempDir(), Process: cmd.Process}})

	assert.Empty(t, logs.String(), "a node that has already stopped is not a failure")
}

// TestNodeSetup tests the basic node setup functionality
func TestNodeSetup(t *testing.T) {
	tempDir := t.TempDir()

	validator, err := SetupValidatorNode(tempDir, 0)
	require.NoError(t, err)
	defer validator.Cleanup()
	assert.NotNil(t, validator, "validator should not be nil")
	assert.Equal(t, 0, validator.Index, "validator index should be 0")
	assert.Greater(t, validator.P2PPort, 0, "validator should have valid P2P port")
	assert.NotEmpty(t, validator.NodeID, "validator should have NodeID")
	assert.NotEmpty(t, validator.DataDir, "validator should have DataDir")
}

// writerFunc adapts a function to io.Writer.
type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

// TestNodeExitedWaitsForTheOutputToBeWritten pins that a node counts as exited
// only once everything it wrote has been written out.
//
// os/exec drains a child's stdout and stderr on goroutines of its own, and only
// Cmd.Wait joins them. Reaping the process directly reports the exit while the
// end of the node's log is still in flight -- and the end is the part that says
// why it died, which is what every caller of Exited reads next.
func TestNodeExitedWaitsForTheOutputToBeWritten(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helpers not available on windows")
	}

	// Holding the copier inside a Write puts the node in the state the race
	// produces on a loaded machine -- process gone, output not yet delivered --
	// without having to lose the race to observe it.
	writing := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	held := writerFunc(func(p []byte) (int, error) {
		once.Do(func() { close(writing) })
		<-release
		return len(p), nil
	})

	node := &Node{Index: 0, DataDir: t.TempDir()}
	require.NoError(t, StartNode(t.Context(), "sh", node, []string{"-c", "echo done >&2"}, GnolandStartOpts{extraWriter: held}))

	<-writing
	exited := node.Exited()
	select {
	case <-exited:
		t.Fatal("reported the node gone while its output was still being written")
	case <-time.After(500 * time.Millisecond):
	}

	close(release)
	select {
	case <-exited:
	case <-time.After(30 * time.Second):
		t.Fatal("never reported the node gone once its output was written")
	}
}

// TestStartNodeKeepsWhatAPreviousRunLogged pins that restarting a validator
// does not erase the log of the run that stopped.
//
// A stopped validator keeps its data dir so RestartValidator can bring the same
// node back, and the reason it went down is in that dir. Truncating on restart
// throws away the post-mortem for the only failure a restart scenario is about.
func TestStartNodeKeepsWhatAPreviousRunLogged(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helpers not available on windows")
	}

	node := &Node{Index: 0, DataDir: t.TempDir()}

	require.NoError(t, StartNode(t.Context(), "sh", node, []string{"-c", "echo FIRST_RUN >&2"}, GnolandStartOpts{}))
	<-node.Exited()
	require.NoError(t, StartNode(t.Context(), "sh", node, []string{"-c", "echo SECOND_RUN >&2"}, GnolandStartOpts{}))
	<-node.Exited()

	logged := nodeStderrTail(node)
	assert.Contains(t, logged, "FIRST_RUN", "the stopped run's log must survive the restart")
	assert.Contains(t, logged, "SECOND_RUN")
}

// TestCleanupNodesKillsAProcessThatIgnoresTheInterrupt pins that teardown has a
// deadline.
//
// gnoland traps SIGINT to flush the consensus WAL, the mempool WAL and the app
// DB, so a node wedged in that flush never exits on its own. Waiting on it
// without a deadline stops the run: the cluster the scenario after it wanted is
// never built, and the temp directories of the one that hung are never removed.
func TestCleanupNodesKillsAProcessThatIgnoresTheInterrupt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process signals not supported on windows")
	}

	cmd := spawnUnstoppable(t)
	node := &Node{Index: 0, DataDir: t.TempDir(), Process: cmd.Process}

	done := make(chan struct{})
	go func() {
		defer close(done)
		cleanupNodes(slog.New(slog.DiscardHandler), []*Node{node}, 200*time.Millisecond)
	}()

	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("teardown never returned on a node that ignored the interrupt")
	}
}

// TestCleanupNodesSkipsSlotsThatWereNeverSetUp pins that teardown survives a
// half-built validator set.
//
// StartCluster registers this cleanup over the whole slice before it fills it,
// so a setup failure at any index leaves nil entries behind. Panicking here
// replaces the error naming the validator that failed with a nil-pointer stack
// trace raised from inside a deferred function.
func TestCleanupNodesSkipsSlotsThatWereNeverSetUp(t *testing.T) {
	assert.NotPanics(t, func() {
		CleanupNodes(slog.New(slog.DiscardHandler), []*Node{nil, {Index: 1, DataDir: t.TempDir()}, nil})
	})
}

// TestWaitForNodeReadyQuotesTheNodeLogWhenItTimesOut pins that the readiness
// deadline reports what the node said.
//
// A node that comes up too slowly and one that is wedged look identical from
// the RPC side, and the difference is in the node's own log. The temp directory
// holding that log is removed as soon as the boot fails, so an error that does
// not carry the log leaves nothing to read afterwards.
func TestWaitForNodeReadyQuotesTheNodeLogWhenItTimesOut(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process signals not supported on windows")
	}

	// Alive, so the process-death branch cannot fire, and pointed at a port
	// nothing listens on, so readiness can only ever time out.
	cmd := exec.Command("sleep", "30")
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	node := &Node{Index: 2, DataDir: t.TempDir(), RPCAddr: "tcp://127.0.0.1:1", Process: cmd.Process}
	require.NoError(t, os.WriteFile(
		filepath.Join(node.DataDir, "stderr.log"),
		[]byte("panic: TELLTALE reason the node never opened its port\n"),
		FilePermissions))

	err := waitForNodeReady(t.Context(), node, 50*time.Millisecond)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "TELLTALE", "the node's own log has to reach the error")
	assert.Contains(t, err.Error(), "127.0.0.1:1", "the error has to name where it polled")
}

// TestWaitForNodeReadySaysWhenTheRunWasCancelled pins that a cancelled run is
// not reported as a node too slow to boot. They need opposite responses, and
// the second reading sends whoever reads it looking for a node fault that
// never happened.
func TestWaitForNodeReadySaysWhenTheRunWasCancelled(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process signals not supported on windows")
	}

	cmd := exec.Command("sleep", "30")
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	node := &Node{Index: 0, DataDir: t.TempDir(), RPCAddr: "tcp://127.0.0.1:1", Process: cmd.Process}
	err := waitForNodeReady(ctx, node, 90*time.Second)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.NotContains(t, err.Error(), "90s", "a cancelled run did not wait out any deadline")
}
