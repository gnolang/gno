package cluster

import (
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/stretchr/testify/require"
)

// Every node in a local cluster answers on the same loopback address, so
// tm2's one-peer-per-IP guard refuses every peer past the first, and the
// refused side redials with no backoff. Under it, three validators hold ~4800
// sockets in TIME_WAIT and unrelated RPC connects stall for seconds; with the
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
