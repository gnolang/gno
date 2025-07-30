package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	fstate "github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

const (
	defaultValidatorKeyName   = "priv_validator_key.json"
	defaultNodeKeyName        = "node_key.json"
	defaultValidatorStateName = "priv_validator_state.json"
)

// Extended Node structure with all necessary fields
type ExtendedNode struct {
	*Node
	Client client.Client
}

// buildGnolandBinary builds a temporary gnoland binary for testing
func buildGnolandBinary(tempDir string) (string, error) {
	binaryPath := filepath.Join(tempDir, "gnoland-test")

	// Use gnoenv.RootDir() to get the root of the gno project
	gnoRoot := gnoenv.RootDir()
	gnolandDir := filepath.Join(gnoRoot, "gno.land", "cmd", "gnoland")

	// Verify the gnoland directory exists
	if _, err := os.Stat(gnolandDir); os.IsNotExist(err) {
		return "", fmt.Errorf("gnoland directory not found at: %s", gnolandDir)
	}

	// Build the gnoland binary
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = gnolandDir

	slog.Info("Building gnoland binary", "source_dir", gnolandDir)
	// For debugging, always show build output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to build gnoland binary: %w", err)
	}

	slog.Info("Built temporary gnoland binary", "path", binaryPath)
	return binaryPath, nil
}


// setupValidatorNode creates and initializes a validator node
func setupValidatorNode(tempDir string, index int) *Node {
	node := &Node{
		Index: index,
	}

	// Create node-specific directories
	nodeDir := filepath.Join(tempDir, fmt.Sprintf("validator_%d", index))
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		slog.Error("Failed to create validator directory", "error", err)
		os.Exit(1)
	}
	node.DataDir = nodeDir

	// Initialize secrets (validator key, node key, state)
	node.NodeID = initializeValidatorSecrets(nodeDir)

	// Initialize configuration
	initializeNodeConfig(nodeDir)

	// Set up network addresses
	node.P2PPort = 26656 + index
	node.SocketAddr = fmt.Sprintf("unix://%s", createSocketPath(fmt.Sprintf("validator_%d.sock", index)))
	node.Genesis = filepath.Join(nodeDir, "test_genesis.json")

	slog.Info("Initialized validator node", "index", index, "dir", nodeDir, "p2p_port", node.P2PPort)
	return node
}

// setupNonValidatorNode creates and initializes a non-validator node
func setupNonValidatorNode(tempDir string, index int) *Node {
	node := &Node{
		Index: index,
	}

	// Create node-specific directories
	nodeDir := filepath.Join(tempDir, fmt.Sprintf("nonvalidator_%d", index))
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		slog.Error("Failed to create non-validator directory", "error", err)
		os.Exit(1)
	}
	node.DataDir = nodeDir

	// Initialize only node key for non-validators
	node.NodeID = initializeNonValidatorSecrets(nodeDir)

	// Initialize configuration
	initializeNodeConfig(nodeDir)

	// Set up network addresses
	node.P2PPort = 26656 + index
	node.SocketAddr = fmt.Sprintf("unix://%s", createSocketPath(fmt.Sprintf("nonvalidator_%d.sock", index)))
	node.Genesis = filepath.Join(nodeDir, "test_genesis.json")

	slog.Info("Initialized non-validator node", "index", index, "dir", nodeDir, "p2p_port", node.P2PPort)
	return node
}

// initializeValidatorSecrets generates and saves validator key, node key, and validator state
func initializeValidatorSecrets(dataDir string) string {
	// Create secrets directory
	secretsDir := filepath.Join(dataDir, config.DefaultSecretsDir)
	if err := os.MkdirAll(secretsDir, 0755); err != nil {
		slog.Error("Failed to create secrets directory", "error", err)
		os.Exit(1)
	}

	// Generate and save validator key
	validatorKeyPath := filepath.Join(secretsDir, defaultValidatorKeyName)
	if _, err := signer.GeneratePersistedFileKey(validatorKeyPath); err != nil {
		slog.Error("Failed to generate validator key", "error", err)
		os.Exit(1)
	}

	// Generate and save node key
	nodeKeyPath := filepath.Join(secretsDir, defaultNodeKeyName)
	nodeKey, err := types.GeneratePersistedNodeKey(nodeKeyPath)
	if err != nil {
		slog.Error("Failed to generate node key", "error", err)
		os.Exit(1)
	}

	// Save validator state
	validatorStatePath := filepath.Join(secretsDir, defaultValidatorStateName)
	if _, err := fstate.GeneratePersistedFileState(validatorStatePath); err != nil {
		slog.Error("Failed to generate validator state", "error", err)
		os.Exit(1)
	}

	return string(nodeKey.ID())
}

// initializeNonValidatorSecrets generates only node key for non-validators
func initializeNonValidatorSecrets(dataDir string) string {
	// Create secrets directory
	secretsDir := filepath.Join(dataDir, config.DefaultSecretsDir)
	if err := os.MkdirAll(secretsDir, 0755); err != nil {
		slog.Error("Failed to create secrets directory", "error", err)
		os.Exit(1)
	}

	// Generate and save node key
	nodeKeyPath := filepath.Join(secretsDir, defaultNodeKeyName)
	nodeKey, err := types.GeneratePersistedNodeKey(nodeKeyPath)
	if err != nil {
		slog.Error("Failed to generate node key", "error", err)
		os.Exit(1)
	}

	return string(nodeKey.ID())
}

// initializeNodeConfig creates a default config.toml for the node
func initializeNodeConfig(dataDir string) {
	configPath := filepath.Join(dataDir, "config", "config.toml")

	// Create config directory
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		slog.Error("Failed to create config directory", "error", err)
		os.Exit(1)
	}

	// Initialize default config
	cfg := config.DefaultConfig()
	cfg.SetRootDir(dataDir)

	// Write config file
	if err := config.WriteConfigFile(configPath, cfg); err != nil {
		slog.Error("Failed to write config file", "error", err)
		os.Exit(1)
	}
}

// createSocketPath creates a unique socket path for the node
func createSocketPath(filename string) string {
	// Create a temporary directory for socket files
	tempDir, err := os.MkdirTemp("", "socktest-")
	if err != nil {
		slog.Error("Failed to create socket directory", "error", err)
		os.Exit(1)
	}
	return filepath.Join(tempDir, filename)
}

// waitForNodeReady waits for a node to be ready to accept RPC calls
func waitForNodeReady(ctx context.Context, node *ExtendedNode) error {
	// Create RPC client
	client, err := client.NewHTTPClient(node.SocketAddr)
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}
	node.Client = client

	// Wait for node to be ready using polling
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context timeout while waiting for node %d", node.Index)
		case <-ticker.C:
			info, err := node.Client.ABCIInfo()
			if err == nil && info.Response.Error == nil {
				// Node is ready
				slog.Debug("Node is ready", "index", node.Index)
				return nil
			}
		}
	}
}

// startNode starts a gnoland node process
func startNode(ctx context.Context, binaryPath string, node *Node, args []string) error {
	// Build command
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	cmd.Dir = node.DataDir

	// Create log files
	stdoutPath := filepath.Join(node.DataDir, "stdout.log")
	stderrPath := filepath.Join(node.DataDir, "stderr.log")

	stdout, err := os.Create(stdoutPath)
	if err != nil {
		return fmt.Errorf("failed to create stdout log: %w", err)
	}

	stderr, err := os.Create(stderrPath)
	if err != nil {
		return fmt.Errorf("failed to create stderr log: %w", err)
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	node.Process = cmd.Process
	slog.Info("Started node", "index", node.Index, "pid", node.Process.Pid)

	return nil
}

// cleanupNodes terminates all node processes
func cleanupNodes(nodes []*Node) {
	slog.Info("🧹 Cleaning up nodes", "count", len(nodes))

	for _, node := range nodes {
		if node.Process != nil {
			if err := node.Process.Signal(os.Interrupt); err != nil {
				// If interrupt fails, force kill
				if err := node.Process.Kill(); err != nil {
					slog.Warn("Failed to kill process", "node_index", node.Index, "error", err)
				}
			}

			// Wait for process to exit
			_, _ = node.Process.Wait()
		}

		// Clean up socket file if exists
		if node.SocketAddr != "" {
			socketPath := node.SocketAddr[7:] // Remove "unix://" prefix
			if dir := filepath.Dir(socketPath); strings.HasPrefix(dir, "/tmp/socktest-") {
				os.RemoveAll(dir)
			}
		}
	}
}
