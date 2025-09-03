package main

import (
	"context"
	"fmt"
	"net"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Default file names for node secrets
const (
	defaultValidatorKeyName   = "priv_validator_key.json"
	defaultNodeKeyName        = "node_key.json"
	defaultValidatorStateName = "priv_validator_state.json"

	// File permissions
	dirPermissions  = 0o755
	filePermissions = 0o644
)

// Extended Node structure with all necessary fields
type ExtendedNode struct {
	*Node
	Client client.Client
}

// buildGnolandBinary builds a temporary gnoland binary
func buildGnolandBinary(t TestingT, tempDir string) (string, error) {
	binaryPath := filepath.Join(tempDir, "gnoland-test")

	gnoRoot := gnoenv.RootDir()
	gnolandDir := filepath.Join(gnoRoot, "gno.land", "cmd", "gnoland")

	// Verify gnoland directory exists
	if _, err := os.Stat(gnolandDir); os.IsNotExist(err) {
		return "", fmt.Errorf("gnoland directory not found at: %s", gnolandDir)
	}

	// Build binary
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = gnolandDir

	t.Logf("Building gnoland binary, source_dir: %s", gnolandDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to build gnoland binary: %w", err)
	}

	t.Logf("Built temporary gnoland binary, path: %s", binaryPath)
	return binaryPath, nil
}

// NodeType represents the type of node being set up
type NodeType int

const (
	ValidatorNode NodeType = iota
	NonValidatorNode
)

// String returns the string representation of NodeType
func (nt NodeType) String() string {
	switch nt {
	case ValidatorNode:
		return "validator"
	case NonValidatorNode:
		return "non-validator"
	default:
		return "unknown"
	}
}

// setupNode creates and initializes a node
func setupNode(t TestingT, tempDir string, index int, nodeType NodeType) *Node {
	node := &Node{
		Index: index,
	}

	// Create node directory
	nodeDir := filepath.Join(tempDir, fmt.Sprintf("%s_%d", nodeType, index))
	require.NoError(t, os.MkdirAll(nodeDir, dirPermissions), "Failed to create node directory")
	node.DataDir = nodeDir

	// Initialize secrets
	switch nodeType {
	case ValidatorNode:
		node.NodeID = initializeValidatorSecrets(t, nodeDir)
	case NonValidatorNode:
		node.NodeID = initializeNodeSecrets(t, nodeDir)
	}

	// Set up network addresses with dynamic ports
	node.P2PPort = findAvailablePort(t)
	node.SocketAddr = fmt.Sprintf("unix://%s", createSocketPath(t, fmt.Sprintf("%s_%d.sock", nodeType, index)))
	node.Genesis = filepath.Join(nodeDir, "test_genesis.json")

	// Initialize configuration
	initializeNodeConfig(t, nodeDir, node.SocketAddr, node.P2PPort)

	t.Logf("Initialized node, type: %s, index: %d, dir: %s, p2p_port: %d", nodeType, index, nodeDir, node.P2PPort)
	return node
}

// setupValidatorNode creates a validator node
func setupValidatorNode(t TestingT, tempDir string, index int) *Node {
	return setupNode(t, tempDir, index, ValidatorNode)
}

// setupNonValidatorNode creates a non-validator node
func setupNonValidatorNode(t TestingT, tempDir string, index int) *Node {
	return setupNode(t, tempDir, index, NonValidatorNode)
}

// initializeValidatorSecrets generates validator secrets
func initializeValidatorSecrets(t TestingT, dataDir string) string {
	return createSecretsAndGenerateKeys(t, dataDir, true)
}

// initializeNodeSecrets generates node key for non-validators
func initializeNodeSecrets(t TestingT, dataDir string) string {
	return createSecretsAndGenerateKeys(t, dataDir, false)
}

// createSecretsAndGenerateKeys generates cryptographic keys
func createSecretsAndGenerateKeys(t TestingT, dataDir string, isValidator bool) string {
	secretsDir := filepath.Join(dataDir, config.DefaultSecretsDir)
	require.NoError(t, os.MkdirAll(secretsDir, dirPermissions), "Failed to create secrets directory")

	if isValidator {
		validatorKeyPath := filepath.Join(secretsDir, defaultValidatorKeyName)
		_, err := signer.GeneratePersistedFileKey(validatorKeyPath)
		require.NoError(t, err, "Failed to generate validator key")

		validatorStatePath := filepath.Join(secretsDir, defaultValidatorStateName)
		_, err = fstate.GeneratePersistedFileState(validatorStatePath)
		require.NoError(t, err, "Failed to generate validator state")
	}

	nodeKeyPath := filepath.Join(secretsDir, defaultNodeKeyName)
	nodeKey, err := types.GeneratePersistedNodeKey(nodeKeyPath)
	require.NoError(t, err, "Failed to generate node key")

	return string(nodeKey.ID())
}

// initializeNodeConfig creates node configuration file
func initializeNodeConfig(t TestingT, dataDir string, socketAddr string, p2pPort int) {
	configPath := filepath.Join(dataDir, "config", "config.toml")

	configDir := filepath.Dir(configPath)
	require.NoError(t, os.MkdirAll(configDir, dirPermissions), "Failed to create config directory")

	// Write initial config
	cfg := config.DefaultConfig()
	cfg.SetRootDir(dataDir)
	cfg.RPC.ListenAddress = socketAddr
	cfg.P2P.ListenAddress = fmt.Sprintf("tcp://0.0.0.0:%d", p2pPort)

	require.NoError(t, config.WriteConfigFile(configPath, cfg), "Failed to write initial config file")

	t.Logf("Configured node with RPC socket: %s, P2P port: %d", socketAddr, p2pPort)
}

// createSocketPath creates a unique socket path avoiding length limits
func createSocketPath(t TestingT, filename string) string {
	randSuffix := fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
	tempDir := filepath.Join("/tmp", "gno-"+randSuffix)

	err := os.MkdirAll(tempDir, 0755)
	require.NoError(t, err, "Failed to create socket directory")

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	socketPath := filepath.Join(tempDir, filename)

	if len(socketPath) > 100 { // Unix socket path limit
		t.Fatalf("Socket path too long (%d chars): %s", len(socketPath), socketPath)
	}

	return socketPath
}

// findAvailablePort finds an available TCP port dynamically
func findAvailablePort(t TestingT) int {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	// Close listener
	require.NoError(t, listener.Close())

	t.Logf("Found available port: %d", port)
	return port
}

// waitForNodeReady waits for a node to be ready to accept RPC calls
func waitForNodeReady(t TestingT, ctx context.Context, node *ExtendedNode) error {
	// Create RPC client
	client, err := client.NewHTTPClient(node.SocketAddr)
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}
	node.Client = client

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	deadline, _ := ctx.Deadline()

	success := assert.EventuallyWithT(t, func(c *assert.CollectT) {
		info, err := node.Client.ABCIInfo(ctx)
		require.NoError(c, err)
		require.NoError(c, info.Response.Error)
	}, time.Until(deadline), 500*time.Millisecond, "node %d failed to become ready", node.Index)

	if !success {
		return fmt.Errorf("timeout waiting for node %d to be ready", node.Index)
	}

	return nil
}

// startNode starts a gnoland node process
func startNode(t TestingT, ctx context.Context, binaryPath string, node *Node, args []string) error {
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

	cmd.Stdout, cmd.Stderr = stdout, stderr

	// Log the command being executed
	t.Logf("Starting node %d with command: %s %s", node.Index, binaryPath, strings.Join(args, " "))

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	node.Process = cmd.Process
	t.Logf("Started node, index: %d, pid: %d", node.Index, node.Process.Pid)

	return nil
}

// cleanupNodes terminates all node processes
func cleanupNodes(t TestingT, nodes []*Node) {
	t.Logf("🧹 Cleaning up nodes, count: %d", len(nodes))

	for _, node := range nodes {
		if node.Process == nil {
			continue
		}

		if err := node.Process.Signal(os.Interrupt); err != nil {
			// If interrupt fails, force kill
			if err := node.Process.Kill(); err != nil {
				t.Errorf("WARNING: Failed to kill process, node_index: %d, error: %v", node.Index, err)
			}
		}

		// Wait for process to exit
		_, _ = node.Process.Wait()
	}
}
