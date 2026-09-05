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
	Index     int
	NodeID    string
	DataDir   string
	P2PPort   int
	RPCAddr   string
	Genesis   string
	Process   *os.Process
	closeLogs func()

	// Guards the reaper started by Exited. A process can be waited for
	// exactly once, and both readiness polling and the stop paths need to know
	// when it
	// is gone, so the wait is shared rather than duplicated.
	mu sync.Mutex
	// cmd is the command StartNode ran, kept so the reaper can wait through it
	// rather than through Process. Nil on a node a caller assembled from a
	// process it started itself.
	cmd     *exec.Cmd
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
		p, cmd := n.Process, n.cmd
		if p == nil {
			close(n.exited)
			return n.exited
		}
		done := n.exited
		go func() {
			err := reap(p, cmd)
			n.mu.Lock()
			n.waitErr = err
			n.mu.Unlock()
			close(done)
		}()
	}
	return n.exited
}

// reap waits for a node's process to end and reports why.
//
// Through Cmd.Wait wherever the harness started the process itself: os/exec
// drains the child's stdout and stderr on goroutines of its own, and only
// Cmd.Wait joins them. Reaping the process directly returns as soon as the
// kernel does, with the end of the node's log still in the pipe -- and the end
// is the part that says why it died. Cmd.Wait also releases the pipes and the
// context watcher, which a run that restarts a node would otherwise leak once
// per cycle.
//
// A node assembled from a bare process is reaped through it.
func reap(p *os.Process, cmd *exec.Cmd) error {
	if cmd != nil {
		return cmd.Wait()
	}
	_, err := p.Wait()
	return err
}

// WaitErr reports why the process ended. Only meaningful once the channel from
// Exited is closed. A process the harness started reports the *exec.ExitError
// of a non-zero exit, the ones a signal causes included; a node assembled from
// a bare process reports an error only when the wait itself failed.
func (n *Node) WaitErr() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.waitErr
}

// adoptProcess takes ownership of a newly started process and of the closer
// for the log files it writes to, discarding the previous one's wait so a
// restarted node is not reported dead by the reaper that watched the process
// it replaced.
//
// The discard and the assignment happen together under the lock. Split apart,
// an Exited() landing between them starts a reaper on the OLD process, which
// has already exited -- so a node that just started successfully reports
// having exited before it was ready.
//
// The generation before this one has already exited, so its log files are
// closed here rather than at cleanup: a scenario that stops and restarts a
// validator would otherwise hold two more descriptors per cycle until the
// whole cluster goes.
func (n *Node) adoptProcess(cmd *exec.Cmd, closeLogs func()) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.closeLogs != nil {
		n.closeLogs()
	}
	n.closeLogs = closeLogs
	n.exited = nil
	n.waitErr = nil
	n.cmd = cmd
	n.Process = cmd.Process
}

// clearProcess forgets the process a node was running, so everything that
// reads liveness through Exited sees a stopped node as gone.
func (n *Node) clearProcess() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Process = nil
}

// Cleanup closes the node's log files. Idempotent, and safe on a node that
// was never started.
func (n *Node) Cleanup() {
	n.mu.Lock()
	closeLogs := n.closeLogs
	n.closeLogs = nil
	n.mu.Unlock()

	if closeLogs != nil {
		closeLogs()
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

// SetupValidatorNode creates and initializes a validator node under tempDir.
func SetupValidatorNode(tempDir string, index int) (*Node, error) {
	node := &Node{
		Index: index,
	}

	nodeDir := filepath.Join(tempDir, fmt.Sprintf("validator_%d", index))
	if err := os.MkdirAll(nodeDir, DirPermissions); err != nil {
		return nil, fmt.Errorf("create node directory: %w", err)
	}
	node.DataDir = nodeDir

	nodeID, err := createSecretsAndGenerateKeys(nodeDir)
	if err != nil {
		return nil, fmt.Errorf("initialize validator secrets: %w", err)
	}
	node.NodeID = nodeID

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

	if err := initializeNodeConfig(nodeDir, node.RPCAddr, node.P2PPort); err != nil {
		return nil, fmt.Errorf("initialize node config: %w", err)
	}

	slog.Debug("initialized node", "index", index, "dir", nodeDir, "p2p_port", node.P2PPort)
	return node, nil
}

func createSecretsAndGenerateKeys(dataDir string) (string, error) {
	secretsDir := filepath.Join(dataDir, config.DefaultSecretsDir)
	if err := os.MkdirAll(secretsDir, DirPermissions); err != nil {
		return "", fmt.Errorf("create secrets directory: %w", err)
	}

	validatorKeyPath := filepath.Join(secretsDir, DefaultValidatorKeyName)
	if _, err := signer.GeneratePersistedFileKey(validatorKeyPath); err != nil {
		return "", fmt.Errorf("generate validator key: %w", err)
	}

	validatorStatePath := filepath.Join(secretsDir, defaultValidatorStateName)
	if _, err := fstate.GeneratePersistedFileState(validatorStatePath); err != nil {
		return "", fmt.Errorf("generate validator state: %w", err)
	}

	nodeKeyPath := filepath.Join(secretsDir, DefaultNodeKeyName)
	nodeKey, err := types.GeneratePersistedNodeKey(nodeKeyPath)
	if err != nil {
		return "", fmt.Errorf("generate node key: %w", err)
	}

	return string(nodeKey.ID()), nil
}

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

// nodeReadyTimeout bounds the wait for one node to answer RPC.
const nodeReadyTimeout = 90 * time.Second

// exitReason renders why a wait ended. A node reaped through its process rather
// than its command reports nil whatever its exit status was, and "%v" on that
// prints "<nil>" where the reader is looking for a reason. Rendered rather than
// wrapped: this is one field of a diagnostic and no caller matches on it.
func exitReason(err error) string {
	if err == nil {
		return "no error reported"
	}
	return err.Error()
}

// nodeLogTails quotes the end of every node's stderr, labelled by index, and
// skips the nodes that wrote nothing.
//
// A cluster that never commits its first block usually says why on the nodes
// the poll never touched: the one being polled is healthy and waiting for a
// quorum the others never joined.
func nodeLogTails(nodes []*Node) string {
	var b strings.Builder
	for _, node := range nodes {
		if node == nil {
			continue
		}
		tail := nodeStderrTail(node)
		if tail == "" {
			continue
		}
		fmt.Fprintf(&b, "--- validator %d stderr ---\n%s\n", node.Index, tail)
	}
	return b.String()
}

// WaitForNodeReady waits for a node to be ready to accept RPC calls
func WaitForNodeReady(ctx context.Context, node *Node) error {
	return waitForNodeReady(ctx, node, nodeReadyTimeout)
}

// waitForNodeReady is WaitForNodeReady with the deadline exposed, so a test can
// reach the timeout branch without waiting out the real one.
func waitForNodeReady(ctx context.Context, node *Node, timeout time.Duration) error {
	rpcClient, err := client.NewHTTPClient(node.RPCAddr)
	if err != nil {
		return fmt.Errorf("create RPC client: %w", err)
	}

	run := ctx
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// A run that was cancelled and a node that will not come up need
			// opposite responses, and only the parent context tells them apart.
			if err := run.Err(); err != nil {
				return fmt.Errorf("waiting for node %d at %s: %w", node.Index, node.RPCAddr, err)
			}
			// The tail travels with the error because the directory holding it
			// is removed as soon as this boot is reported failed.
			return fmt.Errorf("node %d did not answer RPC at %s within %s\n%s",
				node.Index, node.RPCAddr, timeout, nodeStderrTail(node))
		case <-node.Exited():
			// The RPC answers "connection refused" both before a node opens
			// its port and after it dies, so polling alone cannot tell a slow
			// boot from a dead one. Saying which it was is the difference
			// between waiting longer and reading the node's stderr.
			return fmt.Errorf("node %d exited before it was ready (%s)\n%s",
				node.Index, exitReason(node.WaitErr()), nodeStderrTail(node))
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

// StartNode starts a gnoland node process. opts says where its output goes
// beyond the per-node log files; the zero value writes only those.
func StartNode(ctx context.Context, binaryPath string, node *Node, args []string, opts GnolandStartOpts) error {
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	cmd.Dir = node.DataDir

	stdoutPath := filepath.Join(node.DataDir, "stdout.log")
	stderrPath := filepath.Join(node.DataDir, "stderr.log")

	stdout, err := openNodeLog(stdoutPath)
	if err != nil {
		return fmt.Errorf("open stdout log: %w", err)
	}

	stderr, err := openNodeLog(stderrPath)
	if err != nil {
		stdout.Close()
		return fmt.Errorf("open stderr log: %w", err)
	}

	stdoutWriters := []io.Writer{stdout}
	stderrWriters := []io.Writer{stderr}
	if opts.TeeNodeLogs {
		stdoutWriters = append(stdoutWriters, os.Stderr)
		stderrWriters = append(stderrWriters, os.Stderr)
	}
	if opts.extraWriter != nil {
		stdoutWriters = append(stdoutWriters, opts.extraWriter)
		stderrWriters = append(stderrWriters, opts.extraWriter)
	}
	cmd.Stdout = io.MultiWriter(stdoutWriters...)
	cmd.Stderr = io.MultiWriter(stderrWriters...)

	slog.Debug("starting node", "index", node.Index, "args", strings.Join(args, " "))

	if err := cmd.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		return fmt.Errorf("failed to start node: %w", err)
	}

	node.adoptProcess(cmd, func() {
		stdout.Close()
		stderr.Close()
	})
	return nil
}

// openNodeLog opens one of a node's log files for appending.
//
// Appending rather than truncating because a stopped validator keeps its data
// dir so RestartValidator can bring the same node back, and what the stopped
// run logged on its way down is the post-mortem for the failure the restart is
// usually about.
func openNodeLog(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, FilePermissions)
}

// CleanupNodes terminates all node processes. A nil logger reports through
// slog.Default().
//
// Waits through Node.Exited rather than Process.Wait: since readiness polling
// began reaping, every started node already has a waiter, and a second
// concurrent wait on the same process is not something os.Process supports.
func CleanupNodes(logger *slog.Logger, nodes []*Node) {
	cleanupNodes(logger, nodes, haltTimeout)
}

// cleanupNodes is CleanupNodes with the grace period a wedged node is given
// before it is killed, so a test can watch the escalation without waiting out
// the real one.
func cleanupNodes(logger *slog.Logger, nodes []*Node, grace time.Duration) {
	if logger == nil {
		logger = slog.Default()
	}
	for _, node := range nodes {
		// StartCluster registers this over the whole validator slice before it
		// fills it, so a setup failure at any index leaves nil entries behind.
		if node == nil {
			continue
		}
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

			// gnoland traps the interrupt to flush its WALs and its app DB, so
			// a node wedged in that flush never answers it. Teardown cannot
			// wait it out: the run has the rest of its scenarios to boot, and
			// this node's directory is removed the moment this returns.
			select {
			case <-node.Exited():
			case <-time.After(grace):
				logger.Warn("node did not stop on the interrupt; killing it",
					"node_index", node.Index, "waited", grace)
				if err := node.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
					logger.Error("failed to kill node process", "node_index", node.Index, "error", err)
				}
				<-node.Exited()
			}
		}
		node.Cleanup()
	}
}

// GnolandStartOpts tunes how a node is started: what its argv carries and
// where its stdout/stderr are tee'd beyond the per-node log files.
type GnolandStartOpts struct {
	// TeeNodeLogs copies node stdout/stderr to os.Stderr in addition to
	// the per-node log files.
	TeeNodeLogs bool

	// extraWriter receives a copy of stdout+stderr alongside the per-node log
	// files. Nil captures nothing beyond those files.
	extraWriter io.Writer
}

// StartGnolandNode starts a gnoland node and returns once it answers RPC.
func StartGnolandNode(ctx context.Context, binaryPath string, node *Node, opts GnolandStartOpts) error {
	args := []string{
		"start",
		"--skip-genesis-sig-verification",
		"--genesis", node.Genesis,
		"--data-dir", node.DataDir,
	}

	if err := StartNode(ctx, binaryPath, node, args, opts); err != nil {
		return err
	}

	return WaitForNodeReady(ctx, node)
}
