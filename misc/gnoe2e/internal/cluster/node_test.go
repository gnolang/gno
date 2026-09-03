package cluster

import (
	"context"
	"os/exec"
	"runtime"
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

	nonValidator, err := SetupNonValidatorNode(tempDir, 1)
	require.NoError(t, err)
	defer nonValidator.Cleanup()
	assert.NotNil(t, nonValidator, "non-validator should not be nil")
	assert.Equal(t, 1, nonValidator.Index, "non-validator index should be 1")
	assert.Greater(t, nonValidator.P2PPort, 0, "non-validator should have valid P2P port")
	assert.NotEmpty(t, nonValidator.NodeID, "non-validator should have NodeID")
	assert.NotEmpty(t, nonValidator.DataDir, "non-validator should have DataDir")
}
