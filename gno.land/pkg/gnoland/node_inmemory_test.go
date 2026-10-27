package gnoland_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	rpccore "github.com/gnolang/gno/tm2/pkg/bft/rpc/core"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// TestInMemoryNode_AllowCodeExecutionFalse checks that an in-memory node with
// rpc.allow_code_execution = false refuses every query that runs code.
func TestInMemoryNode_AllowCodeExecutionFalse(t *testing.T) {
	cfg := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	cfg.TMConfig.RPC.AllowCodeExecution = false
	n, _ := integration.TestingInMemoryNode(t, log.NewTestingLogger(t), cfg)
	t.Cleanup(func() { require.NoError(t, n.Stop()) })

	for _, path := range []string{".app/simulate", "vm/qrender", "vm/qeval", "vm/qeval_json"} {
		_, err := n.RPCEnvironment().ABCIQuery(&rpctypes.Context{}, path, nil, 0, false)
		assert.Equal(t, rpccore.NewCodeExecutionRefusedError(path), err, "path %q", path)
	}
}
