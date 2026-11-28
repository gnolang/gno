package cluster

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/stretchr/testify/require"
)

// A scenario's node config is applied after the harness's own passes, so the
// timing it asks for is the timing the node boots with. Read back with the
// loader the node itself uses rather than by matching the written toml.
func TestApplyNodeConfigOutlastsTheConsensusPass(t *testing.T) {
	node := &Node{Index: 0, NodeID: "node0", DataDir: t.TempDir(), RPCAddr: "tcp://127.0.0.1:30001", P2PPort: 30000}
	require.NoError(t, initializeNodeConfig(node.DataDir, node.RPCAddr, node.P2PPort))
	require.NoError(t, ConfigureConsensusForSync(node))

	require.NoError(t, applyNodeConfig(node, []Override{
		{Key: "consensus.timeout_commit", Value: "2s"},
		{Key: "consensus.skip_timeout_commit", Value: "false"},
		// A key of the top-level section, which only the toml tags name the
		// way an operator spells it.
		{Key: "moniker", Value: "scenario-node"},
	}))

	written, err := config.LoadConfigFile(filepath.Join(node.DataDir, "config", "config.toml"))
	require.NoError(t, err)
	require.Equal(t, 2*time.Second, written.Consensus.TimeoutCommit,
		"the consensus pass sets 10ms, so this is only 2s if the override ran last")
	require.False(t, written.Consensus.SkipTimeoutCommit)
	require.Equal(t, "scenario-node", written.Moniker)
	require.Equal(t, node.RPCAddr, written.RPC.ListenAddress, "the pass must leave the harness's own settings alone")
}

// A key the config has no field for is the scenario author's typo, and the
// cluster must not boot on a setting nobody applied.
func TestApplyNodeConfigRefusesAnUnknownKey(t *testing.T) {
	node := &Node{Index: 0, NodeID: "node0", DataDir: t.TempDir(), P2PPort: 30000}
	require.NoError(t, initializeNodeConfig(node.DataDir, "tcp://127.0.0.1:30001", node.P2PPort))

	err := applyNodeConfig(node, []Override{{Key: "p2p.nope", Value: "1"}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "config.p2p.nope")
}

// Every node in a local cluster answers on the same loopback address, so
// tm2's one-peer-per-IP guard refuses every peer past the first, and the
// refused side redials with no backoff. Under it, three validators hold
// thousands of sockets in TIME_WAIT and unrelated RPC connects stall for
// seconds; with the
// guard off, no validator reports a peer error at all.
func TestConfigurePersistentPeersAllowsPeersSharingAnIP(t *testing.T) {
	node := &Node{Index: 0, NodeID: "node0", DataDir: t.TempDir(), P2PPort: 30000}
	require.NoError(t, initializeNodeConfig(node.DataDir, "tcp://127.0.0.1:30001", node.P2PPort))

	configPath := filepath.Join(node.DataDir, "config", "config.toml")
	initial, err := config.LoadConfigFile(configPath)
	require.NoError(t, err)
	require.False(t, initial.P2P.AllowDuplicateIP, "tm2 defaults the guard on, which is what has to be overridden")

	require.NoError(t, configurePersistentPeers(node, []string{"node1@localhost:30002"}))

	written, err := config.LoadConfigFile(configPath)
	require.NoError(t, err)
	require.True(t, written.P2P.AllowDuplicateIP)
	require.Equal(t, "node1@localhost:30002", written.P2P.PersistentPeers)
}
