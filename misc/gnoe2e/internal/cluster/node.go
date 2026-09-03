package cluster

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	fstate "github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

// Node represents a gnoland node instance
type Node struct {
	Index    int
	NodeID   string
	DataDir  string
	P2PPort  int
	RPCAddr  string
	Genesis  string
	Process  *os.Process
	cleanups []func()

	// Guards the reaper started by Exited. A process can be waited for
	// exactly once, and both readiness polling and Halt need to know when it
	// is gone, so the wait is shared rather than duplicated.
	mu      sync.Mutex
	exited  chan struct{}
	waitErr error
}

// Exited returns a channel closed once the node's process has exited and been
// reaped, starting the reaper on first call. Safe to call from several
// goroutines and more than once; every caller observes the same wait.
//
// Callers need this rather than os.Process.Wait because that may only be
// called once, and because a child that has exited without being reaped is a
// zombie whose PID still answers signals -- so the usual liveness probes
// report a dead node as alive.
//
// A node with no process is treated as already gone: nothing will start it,
// so a caller waiting on it would wait forever.
func (n *Node) Exited() <-chan struct{} {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.exited == nil {
		n.exited = make(chan struct{})
		p := n.Process
		if p == nil {
			close(n.exited)
			return n.exited
		}
		done := n.exited
		go func() {
			_, err := p.Wait()
			n.mu.Lock()
			n.waitErr = err
			n.mu.Unlock()
			close(done)
		}()
	}
	return n.exited
}

// WaitErr reports why the process ended. Only meaningful once the channel from
// Exited is closed. A normal exit, including a non-zero one, reports nil:
// os.Process.Wait returns an error only when the wait itself fails.
func (n *Node) WaitErr() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.waitErr
}

// adoptProcess takes ownership of a newly started process, discarding the
// previous one's wait so a restarted node is not reported dead by the reaper
// that watched the process it replaced.
//
// The discard and the assignment happen together under the lock. Split apart,
// an Exited() landing between them starts a reaper on the OLD process, which
// has already exited -- so a node that just started successfully reports
// having exited before it was ready.
func (n *Node) adoptProcess(p *os.Process) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.exited = nil
	n.waitErr = nil
	n.Process = p
}

// Cleanup runs all registered cleanup functions in reverse order.
func (n *Node) Cleanup() {
	for i := len(n.cleanups) - 1; i >= 0; i-- {
		n.cleanups[i]()
	}
}

// Default file names for node secrets
const (
	DefaultValidatorKeyName   = "priv_validator_key.json"
	DefaultNodeKeyName        = "node_key.json"
	defaultValidatorStateName = "priv_validator_state.json"

	// File permissions
	DirPermissions  = 0o755
	FilePermissions = 0o644

	// maxStderrTail bounds how much of a failed node's stderr is quoted back.
	// A Go panic prints its reason first and then a long stack, so the tail has
	// to be wide enough to keep the reason above the frames a boot panic
	// unwinds, or the one line that says why the node died is the line that
	// gets dropped.
	maxStderrTail = 8192
)

// ExtendedNode holds a Node with an RPC client.
type ExtendedNode struct {
	*Node
	Client client.Client
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
func setupNode(tempDir string, index int, nodeType NodeType) (*Node, error) {
	node := &Node{
		Index: index,
	}

	// Create node directory
	nodeDir := filepath.Join(tempDir, fmt.Sprintf("%s_%d", nodeType, index))
	if err := os.MkdirAll(nodeDir, DirPermissions); err != nil {
		return nil, fmt.Errorf("create node directory: %w", err)
	}
	node.DataDir = nodeDir

	// Initialize secrets
	var nodeID string
	var err error
	switch nodeType {
	case ValidatorNode:
		nodeID, err = initializeValidatorSecrets(nodeDir)
	case NonValidatorNode:
		nodeID, err = initializeNodeSecrets(nodeDir)
	}
	if err != nil {
		return nil, fmt.Errorf("initialize %s secrets: %w", nodeType, err)
	}
	node.NodeID = nodeID

	// Set up network addresses with dynamic ports
	p2pPort, err := FindAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("find available port: %w", err)
	}
	node.P2PPort = p2pPort

	rpcPort, err := FindAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("find RPC port: %w", err)
	}
	node.RPCAddr = fmt.Sprintf("tcp://127.0.0.1:%d", rpcPort)
	node.Genesis = filepath.Join(nodeDir, "test_genesis.json")

	// Initialize configuration
	if err := initializeNodeConfig(nodeDir, node.RPCAddr, node.P2PPort); err != nil {
		return nil, fmt.Errorf("initialize node config: %w", err)
	}

	slog.Debug("initialized node", "type", nodeType.String(), "index", index, "dir", nodeDir, "p2p_port", node.P2PPort)
	return node, nil
}

// SetupValidatorNode creates a validator node
func SetupValidatorNode(tempDir string, index int) (*Node, error) {
	return setupNode(tempDir, index, ValidatorNode)
}

// SetupNonValidatorNode creates a non-validator node
func SetupNonValidatorNode(tempDir string, index int) (*Node, error) {
	return setupNode(tempDir, index, NonValidatorNode)
}

// initializeValidatorSecrets generates validator secrets
func initializeValidatorSecrets(dataDir string) (string, error) {
	return createSecretsAndGenerateKeys(dataDir, true)
}

// initializeNodeSecrets generates node key for non-validators
func initializeNodeSecrets(dataDir string) (string, error) {
	return createSecretsAndGenerateKeys(dataDir, false)
}

// createSecretsAndGenerateKeys generates cryptographic keys
func createSecretsAndGenerateKeys(dataDir string, isValidator bool) (string, error) {
	secretsDir := filepath.Join(dataDir, config.DefaultSecretsDir)
	if err := os.MkdirAll(secretsDir, DirPermissions); err != nil {
		return "", fmt.Errorf("create secrets directory: %w", err)
	}

	if isValidator {
		validatorKeyPath := filepath.Join(secretsDir, DefaultValidatorKeyName)
		_, err := signer.GeneratePersistedFileKey(validatorKeyPath)
		if err != nil {
			return "", fmt.Errorf("generate validator key: %w", err)
		}

		validatorStatePath := filepath.Join(secretsDir, defaultValidatorStateName)
		_, err = fstate.GeneratePersistedFileState(validatorStatePath)
		if err != nil {
			return "", fmt.Errorf("generate validator state: %w", err)
		}
	}

	nodeKeyPath := filepath.Join(secretsDir, DefaultNodeKeyName)
	nodeKey, err := types.GeneratePersistedNodeKey(nodeKeyPath)
	if err != nil {
		return "", fmt.Errorf("generate node key: %w", err)
	}

	return string(nodeKey.ID()), nil
}

// initializeNodeConfig creates node configuration file
func initializeNodeConfig(dataDir string, rpcAddr string, p2pPort int) error {
	configPath := filepath.Join(dataDir, "config", "config.toml")

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, DirPermissions); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Write initial config
	cfg := config.DefaultConfig()
	cfg.SetRootDir(dataDir)
	cfg.RPC.ListenAddress = rpcAddr
	cfg.P2P.ListenAddress = fmt.Sprintf("tcp://0.0.0.0:%d", p2pPort)

	if err := config.WriteConfigFile(configPath, cfg); err != nil {
		return fmt.Errorf("write initial config file: %w", err)
	}

	slog.Debug("configured node", "rpc_addr", rpcAddr, "p2p_port", p2pPort)
	return nil
}

// FindAvailablePort finds an available TCP port dynamically.
// Note: there is an inherent TOCTOU race between closing this listener
// and the node process binding to the port.
func FindAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("find available port: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	if err := listener.Close(); err != nil {
		return 0, fmt.Errorf("close port listener: %w", err)
	}

	slog.Debug("found available port", "port", port)
	return port, nil
}

// nodeStderrTail returns the end of a node's stderr log, or "" if there is
// none to read. Bounded because a node that failed to boot can have written a
// great deal before giving up, and the last of it is the part that says why.
func nodeStderrTail(node *Node) string {
	data, err := os.ReadFile(filepath.Join(node.DataDir, "stderr.log"))
	if err != nil || len(data) == 0 {
		return ""
	}
	if len(data) > maxStderrTail {
		data = data[len(data)-maxStderrTail:]
	}
	return string(data)
}

// WaitForNodeReady waits for a node to be ready to accept RPC calls
func WaitForNodeReady(ctx context.Context, node *Node) error {
	rpcClient, err := client.NewHTTPClient(node.RPCAddr)
	if err != nil {
		return fmt.Errorf("create RPC client: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Debug("node not ready, stderr tail", "node", node.Index, "stderr", nodeStderrTail(node))
			return fmt.Errorf("timeout waiting for node %d to be ready", node.Index)
		case <-node.Exited():
			// The RPC answers "connection refused" both before a node opens
			// its port and after it dies, so polling alone cannot tell a slow
			// boot from a dead one. Saying which it was is the difference
			// between waiting longer and reading the node's stderr.
			return fmt.Errorf("node %d exited before it was ready (%v)\n%s",
				node.Index, node.WaitErr(), nodeStderrTail(node)) //nolint:errorlint // WaitErr is nil on a normal exit; %w prints %!w(<nil>)
		case <-ticker.C:
			info, err := rpcClient.ABCIInfo(ctx)
			if err != nil {
				continue
			}
			if info.Response.Error == nil {
				return nil
			}
		}
	}
}

// StartNode starts a gnoland node process.
// If teeNodeLogs is true, node stdout/stderr is also copied to os.Stderr.
// If extraWriter is non-nil, stdout/stderr are also copied to it (used
// by Cluster.bootLog capture for hardfork-replay summary scraping).
func StartNode(ctx context.Context, binaryPath string, node *Node, args []string, teeNodeLogs bool, extraWriter io.Writer) error {
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	cmd.Dir = node.DataDir

	stdoutPath := filepath.Join(node.DataDir, "stdout.log")
	stderrPath := filepath.Join(node.DataDir, "stderr.log")

	stdout, err := os.Create(stdoutPath)
	if err != nil {
		return fmt.Errorf("failed to create stdout log: %w", err)
	}

	stderr, err := os.Create(stderrPath)
	if err != nil {
		stdout.Close()
		return fmt.Errorf("failed to create stderr log: %w", err)
	}

	stdoutWriters := []io.Writer{stdout}
	stderrWriters := []io.Writer{stderr}
	if teeNodeLogs {
		stdoutWriters = append(stdoutWriters, os.Stderr)
		stderrWriters = append(stderrWriters, os.Stderr)
	}
	if extraWriter != nil {
		stdoutWriters = append(stdoutWriters, extraWriter)
		stderrWriters = append(stderrWriters, extraWriter)
	}
	cmd.Stdout = io.MultiWriter(stdoutWriters...)
	cmd.Stderr = io.MultiWriter(stderrWriters...)

	slog.Debug("starting node", "index", node.Index, "args", strings.Join(args, " "))

	if err := cmd.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		return fmt.Errorf("failed to start node: %w", err)
	}

	node.cleanups = append(node.cleanups, func() {
		stdout.Close()
		stderr.Close()
	})

	node.adoptProcess(cmd.Process)
	return nil
}

// CleanupNodes terminates all node processes. A nil logger reports through
// slog.Default().
//
// Waits through Node.Exited rather than Process.Wait: since readiness polling
// began reaping, every started node already has a waiter, and a second
// concurrent wait on the same process is not something os.Process supports.
func CleanupNodes(logger *slog.Logger, nodes []*Node) {
	if logger == nil {
		logger = slog.Default()
	}
	for _, node := range nodes {
		if node.Process != nil {
			// A node that has already been reaped -- one the scenario halted,
			// or one the run's context took down -- answers os.ErrProcessDone
			// to both the signal and the kill. Killing it again and reporting
			// the second refusal would make every deliberate stop look like a
			// teardown failure.
			err := node.Process.Signal(os.Interrupt)
			if err != nil && !errors.Is(err, os.ErrProcessDone) {
				if err := node.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
					logger.Error("failed to kill node process", "node_index", node.Index, "error", err)
				}
			}
			<-node.Exited()
		}
		node.Cleanup()
	}
}

// GnolandStartOpts tunes how StartGnolandNodeWithOpts builds the
// gnoland start command and where its stdout/stderr are tee'd.
type GnolandStartOpts struct {
	// TeeNodeLogs copies node stdout/stderr to os.Stderr in addition to
	// the per-node log files.
	TeeNodeLogs bool

	// ExtraWriter receives a copy of stdout+stderr alongside the per-node
	// log files. Cluster.BootFromGenesis sets this to the cluster-level
	// bootLog buffer so hardfork-replay can scrape the structured
	// "Genesis replay report" line via ParseReplayReport. Nil disables
	// capture (existing behavior for restart and StartCluster paths).
	ExtraWriter io.Writer

	// LogFormatJSON appends `--log-format json` to gnoland start. Required
	// for hardfork-replay (so the replay summary lands as one JSON object
	// per line); false for other paths to preserve human-readable console
	// output in stdout.log.
	LogFormatJSON bool
}

// StartGnolandNode starts a gnoland node (validator or non-validator)
// with default options (no JSON logs, no extra writer). Equivalent to
// StartGnolandNodeWithOpts with TeeNodeLogs set from the parameter.
func StartGnolandNode(ctx context.Context, binaryPath string, node *Node, teeNodeLogs bool) error {
	return StartGnolandNodeWithOpts(ctx, binaryPath, node, GnolandStartOpts{TeeNodeLogs: teeNodeLogs})
}

// StartGnolandNodeWithOpts starts a gnoland node honoring the supplied
// options. Used by BootFromGenesis to enable JSON logs and bootLog
// capture for the hardfork-replay summary scraper.
func StartGnolandNodeWithOpts(ctx context.Context, binaryPath string, node *Node, opts GnolandStartOpts) error {
	args := []string{
		"start",
		"--skip-genesis-sig-verification",
		"--genesis", node.Genesis,
		"--data-dir", node.DataDir,
	}
	if opts.LogFormatJSON {
		args = append(args, "--log-format", "json")
	}

	if err := StartNode(ctx, binaryPath, node, args, opts.TeeNodeLogs, opts.ExtraWriter); err != nil {
		return err
	}

	return WaitForNodeReady(ctx, node)
}
