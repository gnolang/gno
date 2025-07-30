package main

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
)

// configureP2PTopology sets up the P2P connections between nodes
func configureP2PTopology(validators, nonValidators []*Node) error {
	slog.Info("📋 Configuring P2P topology", "validators", len(validators), "non_validators", len(nonValidators))

	// Configure validator mesh topology (all validators connect to each other)
	for i, validator := range validators {
		var peerAddrs []string
		for j, otherValidator := range validators {
			if i != j {
				peerAddr := fmt.Sprintf("%s@localhost:%d", otherValidator.NodeID, otherValidator.P2PPort)
				peerAddrs = append(peerAddrs, peerAddr)
			}
		}
		if err := configurePersistentPeers(validator, peerAddrs); err != nil {
			return fmt.Errorf("failed to configure peers for validator %d: %w", i, err)
		}
		// Log peer connections for debugging
		slog.Debug("Validator peer connections", "validator", i+1, "peer_count", len(peerAddrs))
	}

	// Configure non-validator chain topology
	if len(nonValidators) > 0 {
		// First non-validator connects to first validator
		if len(validators) > 0 {
			peerAddr := fmt.Sprintf("%s@localhost:%d", validators[0].NodeID, validators[0].P2PPort)
			if err := configurePersistentPeers(nonValidators[0], []string{peerAddr}); err != nil {
				return fmt.Errorf("failed to configure peers for non-validator 0: %w", err)
			}
			slog.Debug("Non-validator peer connection", "non_validator", 1, "connects_to", "validator 1")
		}

		// Each subsequent non-validator connects to the previous one (chain topology)
		for i := 1; i < len(nonValidators); i++ {
			peerAddr := fmt.Sprintf("%s@localhost:%d", nonValidators[i-1].NodeID, nonValidators[i-1].P2PPort)
			if err := configurePersistentPeers(nonValidators[i], []string{peerAddr}); err != nil {
				return fmt.Errorf("failed to configure peers for non-validator %d: %w", i, err)
			}
			slog.Debug("Non-validator peer connection", "non_validator", i+1, "connects_to", fmt.Sprintf("non-validator %d", i))
		}
	}

	slog.Info("✅ P2P topology configuration completed")
	return nil
}

// configurePersistentPeers configures a node to use persistent_peers
func configurePersistentPeers(node *Node, peerAddrs []string) error {
	configPath := filepath.Join(node.DataDir, "config", "config.toml")

	// Load current config
	cfg, err := config.LoadConfigFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

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
	if err := config.WriteConfigFile(configPath, cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	slog.Debug("Configured node persistent peers", "node_index", node.Index, "peers", peerList)

	return nil
}

// configureConsensusForSync configures consensus parameters for fast synchronization
func configureConsensusForSync(node *Node) error {
	configPath := filepath.Join(node.DataDir, "config", "config.toml")

	// Load current config
	cfg, err := config.LoadConfigFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Set extremely fast consensus timeouts to reach target height quickly
	cfg.Consensus.TimeoutCommit = 50 * time.Millisecond             // Extremely fast commits
	cfg.Consensus.SkipTimeoutCommit = true                          // Skip timeout for faster sync
	cfg.Consensus.CreateEmptyBlocks = true                          // Keep creating blocks
	cfg.Consensus.CreateEmptyBlocksInterval = 50 * time.Millisecond // Create empty blocks very frequently
	cfg.Consensus.TimeoutPropose = 100 * time.Millisecond           // Very fast proposals
	cfg.Consensus.TimeoutPrevote = 50 * time.Millisecond            // Extremely fast prevotes
	cfg.Consensus.TimeoutPrecommit = 50 * time.Millisecond          // Extremely fast precommits

	// Configure RPC to use unix socket
	cfg.RPC.ListenAddress = node.SocketAddr

	// Write updated config
	if err := config.WriteConfigFile(configPath, cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	slog.Debug("Configured node consensus settings", "node_index", node.Index)

	return nil
}

// printNodeConfigurations prints detailed node configurations for debugging
func printNodeConfigurations(nodes []*Node, cfg *testCfg) {
	slog.Info("📋 Node Configurations")

	for i, node := range nodes {
		nodeType := "non-validator"
		if i < cfg.numValidators { // First nodes are validators based on our setup
			nodeType = "validator"
		}

		slog.Info("Node configuration",
			"index", node.Index,
			"type", nodeType,
			"node_id", node.NodeID,
			"p2p_port", node.P2PPort,
			"socket", node.SocketAddr,
			"data_dir", filepath.Base(node.DataDir))

		// Read and print key config settings
		configPath := filepath.Join(node.DataDir, "config", "config.toml")
		if nodeCfg, err := config.LoadConfigFile(configPath); err == nil {
			slog.Debug("P2P configuration",
				"node_index", node.Index,
				"listen_address", nodeCfg.P2P.ListenAddress,
				"seeds", nodeCfg.P2P.Seeds,
				"persistent_peers", nodeCfg.P2P.PersistentPeers,
				"max_inbound_peers", nodeCfg.P2P.MaxNumInboundPeers,
				"max_outbound_peers", nodeCfg.P2P.MaxNumOutboundPeers)

			slog.Debug("Consensus configuration",
				"node_index", node.Index,
				"timeout_commit", nodeCfg.Consensus.TimeoutCommit,
				"skip_timeout_commit", nodeCfg.Consensus.SkipTimeoutCommit,
				"create_empty_blocks", nodeCfg.Consensus.CreateEmptyBlocks,
				"create_empty_blocks_interval", nodeCfg.Consensus.CreateEmptyBlocksInterval,
				"timeout_propose", nodeCfg.Consensus.TimeoutPropose,
				"timeout_prevote", nodeCfg.Consensus.TimeoutPrevote,
				"timeout_precommit", nodeCfg.Consensus.TimeoutPrecommit)

			slog.Debug("RPC configuration",
				"node_index", node.Index,
				"listen_address", nodeCfg.RPC.ListenAddress,
				"grpc_listen_address", nodeCfg.RPC.GRPCListenAddress)
		}
	}
}
