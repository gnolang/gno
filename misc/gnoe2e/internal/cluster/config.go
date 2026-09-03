package cluster

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
)

// ConfigureP2PTopology sets up the P2P connections between nodes
func ConfigureP2PTopology(validators, nonValidators []*Node) error {
	slog.Debug("configuring P2P topology", "validators", len(validators), "non_validators", len(nonValidators))

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
			return fmt.Errorf("configure validator %d peers: %w", i, err)
		}
		slog.Debug("validator peers configured", "index", i+1, "peer_count", len(peerAddrs))
	}

	// Configure non-validator chain topology
	if len(nonValidators) > 0 {
		// First non-validator connects to first validator
		if len(validators) > 0 {
			peerAddr := fmt.Sprintf("%s@localhost:%d", validators[0].NodeID, validators[0].P2PPort)
			if err := configurePersistentPeers(nonValidators[0], []string{peerAddr}); err != nil {
				return fmt.Errorf("configure non-validator 0 peers: %w", err)
			}
			slog.Debug("non-validator 1 connects to validator 1")
		}

		// Each subsequent non-validator connects to the previous one (chain topology)
		for i := 1; i < len(nonValidators); i++ {
			peerAddr := fmt.Sprintf("%s@localhost:%d", nonValidators[i-1].NodeID, nonValidators[i-1].P2PPort)
			if err := configurePersistentPeers(nonValidators[i], []string{peerAddr}); err != nil {
				return fmt.Errorf("configure non-validator %d peers: %w", i, err)
			}
			slog.Debug("non-validator connected", "index", i+1, "connects_to", i)
		}
	}

	slog.Debug("P2P topology configuration completed")
	return nil
}

// configurePersistentPeers configures a node to use persistent_peers
func configurePersistentPeers(node *Node, peerAddrs []string) error {
	configPath := filepath.Join(node.DataDir, "config", "config.toml")

	// Load current config
	cfg, err := config.LoadConfigFile(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	cfg.P2P.PersistentPeers = strings.Join(peerAddrs, ",")

	// Every node in a local cluster answers on one loopback address, and tm2
	// refuses a second inbound connection from an IP it already holds a peer
	// on (p2p/switch.go:841). Past two nodes that refusal is permanent, and
	// the refused side reads EOF then redials with no backoff, because
	// StopPeerForError goes straight to the dial queue (p2p/switch.go:235)
	// instead of the redial loop that owns the backoff. Without this, three
	// validators hold thousands of sockets in TIME_WAIT and unrelated RPC
	// connects stall for seconds.
	cfg.P2P.AllowDuplicateIP = true

	// Set P2P listen address
	cfg.P2P.ListenAddress = fmt.Sprintf("tcp://0.0.0.0:%d", node.P2PPort)

	// Write config back
	if err := config.WriteConfigFile(configPath, cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	slog.Debug("node peers configured", "node_index", node.Index, "peers", cfg.P2P.PersistentPeers)
	return nil
}

// ConfigureConsensusForSync configures consensus parameters for fast synchronization
func ConfigureConsensusForSync(node *Node) error {
	configPath := filepath.Join(node.DataDir, "config", "config.toml")

	// Load current config
	cfg, err := config.LoadConfigFile(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

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

	// Configure RPC listen address
	cfg.RPC.ListenAddress = node.RPCAddr

	// Write updated config
	if err := config.WriteConfigFile(configPath, cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	slog.Debug("consensus configured for sync", "node_index", node.Index)
	return nil
}

// applyNodeConfig applies a scenario's node config overrides to a node's
// config.toml.
//
// Last of the configuration passes, so a scenario can have the timing and
// limits it asks for rather than the ones the harness picked for its own
// convenience. The listen addresses are not reachable this way: the harness
// assigns each node a free port and hands the addresses to the scripts, so the
// section refuses those keys before they arrive here.
func applyNodeConfig(node *Node, overrides []Override) error {
	if len(overrides) == 0 {
		return nil
	}

	configPath := filepath.Join(node.DataDir, "config", "config.toml")
	cfg, err := config.LoadConfigFile(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	root := reflect.ValueOf(cfg).Elem()
	for _, o := range overrides {
		// "toml", the selector `gnoland config set` resolves with: it is the
		// only one that names the top-level section's keys the way an operator
		// spells them, because config.BaseConfig carries no json tags.
		if err := applyOverride(root, "toml", o); err != nil {
			// The key is named the way the scenario wrote it, prefix included,
			// rather than as the path the config was traversed by.
			return fmt.Errorf("config.%w", err)
		}
	}

	if err := config.WriteConfigFile(configPath, cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	slog.Debug("node config overridden", "node_index", node.Index, "overrides", len(overrides))
	return nil
}
