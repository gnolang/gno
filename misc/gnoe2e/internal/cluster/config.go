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

// clusterTimings sets the consensus and gossip timings a local cluster commits
// at. Block times are still floored by BlockTimeIotaMS, 100ms in
// tm2/pkg/bft/types/params.go.
func clusterTimings(cfg *config.Config) {
	cfg.Consensus.TimeoutCommit = 10 * time.Millisecond
	cfg.Consensus.SkipTimeoutCommit = true
	cfg.Consensus.CreateEmptyBlocks = true
	cfg.Consensus.CreateEmptyBlocksInterval = 10 * time.Millisecond
	cfg.Consensus.TimeoutPropose = 10 * time.Millisecond
	cfg.Consensus.TimeoutPrevote = 10 * time.Millisecond
	cfg.Consensus.TimeoutPrecommit = 10 * time.Millisecond

	cfg.P2P.FlushThrottleTimeout = 10 * time.Millisecond
}

// clusterPeering sets what a mesh of nodes sharing one loopback address needs.
//
// tm2 refuses a second inbound connection from an IP it already holds a peer on
// (p2p/switch.go:841). Past two nodes that refusal is permanent, and the
// refused side reads EOF then redials with no backoff, because StopPeerForError
// goes straight to the dial queue (p2p/switch.go:235) instead of the redial loop
// that owns the backoff. Without this, three validators hold thousands of
// sockets in TIME_WAIT and unrelated RPC connects stall for seconds.
func clusterPeering(cfg *config.Config) {
	cfg.P2P.AllowDuplicateIP = true
}

// DefaultNodeConfig returns the config a validator boots with when a scenario
// sets nothing of its own: tm2's defaults with the harness's passes applied.
// The listen addresses are left as they come, since the harness assigns each
// node its own at boot.
func DefaultNodeConfig() *config.Config {
	cfg := config.DefaultConfig()
	clusterTimings(cfg)
	clusterPeering(cfg)
	return cfg
}

// ConfigDefaults lists every "config." key a scenario can set in its
// "-- cluster --" section, with the value a validator boots with.
func ConfigDefaults() []Override {
	return fieldDefaults(reflect.ValueOf(DefaultNodeConfig()).Elem(), nodeConfigSelector)
}

// ConfigureP2PTopology gives every validator a peer list naming the others, so
// the set is a full mesh.
func ConfigureP2PTopology(validators []*Node) error {
	slog.Debug("configuring P2P topology", "validators", len(validators))

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

	clusterPeering(cfg)

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

	clusterTimings(cfg)

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

	if err := applyConfigOverrides(cfg, overrides); err != nil {
		return err
	}

	if err := config.WriteConfigFile(configPath, cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	slog.Debug("node config overridden", "node_index", node.Index, "overrides", len(overrides))
	return nil
}

// applyConfigOverrides sets a scenario's overrides on a node config in memory.
func applyConfigOverrides(cfg *config.Config, overrides []Override) error {
	root := reflect.ValueOf(cfg).Elem()
	for _, o := range overrides {
		if err := applyOverride(root, nodeConfigSelector, o); err != nil {
			// The key is named the way the scenario wrote it, prefix included,
			// rather than as the path the config was traversed by.
			return fmt.Errorf("config.%w", err)
		}
	}
	return nil
}

// ValidateNodeConfig reports whether overrides resolve and parse against the
// config a validator boots with, writing nothing.
//
// This is what lets a misspelled key fail when the scenario is read rather than
// when its own cluster is built, which in a run is after every scenario ahead
// of it has already booted one.
func ValidateNodeConfig(overrides []Override) error {
	return applyConfigOverrides(DefaultNodeConfig(), overrides)
}
