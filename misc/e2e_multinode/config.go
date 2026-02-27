package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/stretchr/testify/require"
)

// configureP2PTopology sets up the P2P connections between nodes
func configureP2PTopology(t TestingT, validators, nonValidators []*Node) {
	t.Logf("📋 Configuring P2P topology, validators: %d, non_validators: %d", len(validators), len(nonValidators))

	// Configure validator mesh topology (all validators connect to each other)
	for i, validator := range validators {
		var peerAddrs []string
		for j, otherValidator := range validators {
			if i != j {
				peerAddr := fmt.Sprintf("%s@localhost:%d", otherValidator.NodeID, otherValidator.P2PPort)
				peerAddrs = append(peerAddrs, peerAddr)
			}
		}
		configurePersistentPeers(t, validator, peerAddrs)
		t.Logf("Validator %d configured with %d peers", i+1, len(peerAddrs))
	}

	// Configure non-validator chain topology
	if len(nonValidators) > 0 {
		// First non-validator connects to first validator
		if len(validators) > 0 {
			peerAddr := fmt.Sprintf("%s@localhost:%d", validators[0].NodeID, validators[0].P2PPort)
			configurePersistentPeers(t, nonValidators[0], []string{peerAddr})
			t.Logf("Non-validator 1 connects to validator 1")
		}

		// Each subsequent non-validator connects to the previous one (chain topology)
		for i := 1; i < len(nonValidators); i++ {
			peerAddr := fmt.Sprintf("%s@localhost:%d", nonValidators[i-1].NodeID, nonValidators[i-1].P2PPort)
			configurePersistentPeers(t, nonValidators[i], []string{peerAddr})
			t.Logf("Non-validator %d connects to node %d", i+1, i)
		}
	}

	t.Log("✅ P2P topology configuration completed")
}

// configurePersistentPeers configures a node to use persistent_peers
func configurePersistentPeers(t TestingT, node *Node, peerAddrs []string) {
	configPath := filepath.Join(node.DataDir, "config", "config.toml")

	// Load current config
	cfg, err := config.LoadConfigFile(configPath)
	require.NoError(t, err, "failed to load config")

	// Set persistent peers
	peerList := ""
	for i, addr := range peerAddrs {
		if i > 0 {
			peerList += ","
		}
		peerList += addr
	}
	cfg.P2P.PersistentPeers = peerList

	// Set P2P listen address
	cfg.P2P.ListenAddress = fmt.Sprintf("tcp://0.0.0.0:%d", node.P2PPort)

	// Write config back
	err = config.WriteConfigFile(configPath, cfg)
	require.NoError(t, err, "failed to write config")

	t.Logf("Node %d peers: %s", node.Index, peerList)
}

// configureConsensusForSync configures consensus parameters for fast synchronization
func configureConsensusForSync(t TestingT, node *Node) {
	configPath := filepath.Join(node.DataDir, "config", "config.toml")

	// Load current config
	cfg, err := config.LoadConfigFile(configPath)
	require.NoError(t, err, "failed to load config")

	// Set extremely fast consensus timeouts to reach target height quickly
	// Note: Block times are limited by BlockTimeIotaMS (100ms) in tm2/pkg/bft/types/params.go
	cfg.Consensus.TimeoutCommit = 10 * time.Millisecond             // Ultra fast commits
	cfg.Consensus.SkipTimeoutCommit = true                          // Skip timeout for faster sync
	cfg.Consensus.CreateEmptyBlocks = true                          // Keep creating blocks
	cfg.Consensus.CreateEmptyBlocksInterval = 10 * time.Millisecond // Create empty blocks very frequently
	cfg.Consensus.TimeoutPropose = 10 * time.Millisecond            // Ultra fast proposals
	cfg.Consensus.TimeoutPrevote = 10 * time.Millisecond            // Ultra fast prevotes
	cfg.Consensus.TimeoutPrecommit = 10 * time.Millisecond          // Ultra fast precommits

	// Configure P2P for faster message propagation
	cfg.P2P.FlushThrottleTimeout = 10 * time.Millisecond // Reduce P2P message batching delay

	// Configure RPC to use unix socket
	cfg.RPC.ListenAddress = node.SocketAddr

	// Write updated config
	err = config.WriteConfigFile(configPath, cfg)
	require.NoError(t, err, "failed to write config")

	t.Logf("Node %d consensus configured", node.Index)
}

// printNodeConfigurations prints node configurations
func printNodeConfigurations(t TestingT, nodes []*Node, cfg *testCfg) {
	t.Log("📋 Node Configurations")

	for i, node := range nodes {
		nodeType := "non-validator"
		if i < cfg.numValidators { // First nodes are validators based on our setup
			nodeType = "validator"
		}

		t.Logf("Node configuration, index: %d, type: %s, node_id: %s, p2p_port: %d, socket: %s, data_dir: %s",
			node.Index, nodeType, node.NodeID, node.P2PPort, node.SocketAddr, filepath.Base(node.DataDir))

		// Read and print key config settings
		configPath := filepath.Join(node.DataDir, "config", "config.toml")
		if nodeCfg, err := config.LoadConfigFile(configPath); err == nil {
			t.Logf("Node %d P2P: %s, peers: %s", node.Index, nodeCfg.P2P.ListenAddress, nodeCfg.P2P.PersistentPeers)
			t.Logf("Node %d RPC: %s", node.Index, nodeCfg.RPC.ListenAddress)
		}
	}
}
