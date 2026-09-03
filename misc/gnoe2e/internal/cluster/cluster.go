package cluster

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const DefaultChainID = "test-e2e"

// ---- Genesis Configuration

// GenesisConfig controls the genesis document for a local cluster.
type GenesisConfig struct {
	ChainID       string
	MaxGas        int64
	MaxTxBytes    int64
	MaxDataBytes  int64
	TimeIotaMS    int64
	LoadExamples  bool
	Balances      map[string]int64         // extra funded accounts (addr -> ugnot)
	ExtraTxs      []gnoland.TxWithMetadata // unsigned deploy TXs; BuildGenesis signs with validator key
	ExtraPackages []*std.MemPackage        // packages to deploy in genesis; BuildGenesis creates and signs TXs
	ExtraCallTxs  []gnoland.TxWithMetadata // genesis call TXs (Bootstrap, RegisterUser, etc.); signed by BuildGenesis
	// CodeSubmissionPolicy governs what MsgAddPackage does at heights above
	// genesis. Empty leaves the chain default. "inert" parks submissions until
	// an address in PkgApprovers enables them.
	CodeSubmissionPolicy vm.CodeSubmissionPolicy
	// PkgApprovers may send MsgEnablePackage. Required under "inert".
	PkgApprovers []crypto.Address
	// Params sets auth, vm and bank genesis params by `gnogenesis params set`
	// key. Applied after the policy and approver fields above, so an explicit
	// path wins over them.
	Params []Override
}

// DefaultGenesisConfig returns genesis defaults matching mainnet parameters.
func DefaultGenesisConfig() GenesisConfig {
	return GenesisConfig{
		ChainID:      DefaultChainID,
		MaxGas:       3_000_000_000,
		MaxTxBytes:   1_000_000,
		MaxDataBytes: 2_000_000,
		TimeIotaMS:   100,
		LoadExamples: true,
		Balances:     make(map[string]int64),
	}
}

func (g *GenesisConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&g.ChainID, "chain-id", g.ChainID, "chain ID")
	fs.Int64Var(&g.MaxGas, "block-max-gas", g.MaxGas, "per-block gas limit")
	fs.Int64Var(&g.MaxTxBytes, "max-tx-bytes", g.MaxTxBytes, "max transaction size in bytes")
	fs.Int64Var(&g.MaxDataBytes, "max-data-bytes", g.MaxDataBytes, "max block data size in bytes")
	fs.Int64Var(&g.TimeIotaMS, "block-time-iota", g.TimeIotaMS, "minimum block time in ms")
	fs.BoolVar(&g.LoadExamples, "load-examples", g.LoadExamples, "load example packages in genesis")
	fs.Func("code-submission-policy", "permissionless, permissioned, or inert", func(v string) error {
		g.CodeSubmissionPolicy = vm.CodeSubmissionPolicy(v)
		return nil
	})
	fs.Func("pkg-approver", "address allowed to enable inert packages (repeatable; run keeps only the last)", func(v string) error {
		addr, err := crypto.AddressFromBech32(v)
		if err != nil {
			return fmt.Errorf("invalid -pkg-approver %q: %w", v, err)
		}
		g.PkgApprovers = append(g.PkgApprovers, addr)
		return nil
	})
}

func (g *GenesisConfig) Validate() error {
	if g.ChainID == "" {
		return errors.New("--chain-id is required")
	}
	if g.MaxGas < 1 {
		return errors.New("--block-max-gas must be positive")
	}
	switch g.CodeSubmissionPolicy {
	case "",
		vm.CodeSubmissionPolicyPermissionless,
		vm.CodeSubmissionPolicyPermissioned,
		vm.CodeSubmissionPolicyInert:
	default:
		return fmt.Errorf("unknown --code-submission-policy %q", g.CodeSubmissionPolicy)
	}
	// An inert chain with nobody able to enable is a chain where no package
	// submitted after genesis can ever become live.
	if g.CodeSubmissionPolicy == vm.CodeSubmissionPolicyInert && len(g.PkgApprovers) == 0 {
		return errors.New("--code-submission-policy=inert requires at least one --pkg-approver")
	}
	return nil
}

// ---- Cluster Configuration

// ClusterConfig controls how the local validator cluster is set up.
type ClusterConfig struct {
	NumValidators int
	TeeNodeLogs   bool // tee node stdout/stderr to os.Stderr
	Genesis       GenesisConfig
	Logger        *slog.Logger // if nil, uses slog.Default()
	// NodeConfig sets node config options by `gnoland config set` key. Applied
	// to every validator after the harness's own configuration, so it can
	// override consensus timing -- but never the listen addresses, which the
	// harness assigns and hands to the scripts.
	NodeConfig []Override
}

// DefaultClusterConfig returns cluster defaults.
func DefaultClusterConfig() ClusterConfig {
	return ClusterConfig{
		NumValidators: 2,
		Genesis:       DefaultGenesisConfig(),
	}
}

func (c *ClusterConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.IntVar(&c.NumValidators, "validators", c.NumValidators, "number of validator nodes")
	c.Genesis.RegisterFlags(fs)
}

func (c *ClusterConfig) Validate() error {
	if c.NumValidators < 1 {
		return errors.New("at least 1 validator required")
	}
	return c.Genesis.Validate()
}

// ---- Cluster type

// Cluster holds the running state of a local validator cluster.
type Cluster struct {
	Validators []*Node
	RPCAddr    string // RPC address of the first validator
	BinaryPath string
	TempDir    string

	// logger is the one ClusterConfig supplied, kept so teardown reports
	// through the run's own handler rather than slog.Default(). nil on a
	// Cluster that was not built by StartCluster.
	logger *slog.Logger

	// bootLog captures stdout+stderr of the gnoland process started by
	// BootFromExistingDataDir or BootFromGenesis. nil before any boot. See
	// BootLogReader for read access semantics.
	bootLog *bootLogBuffer
}

// bootLogBuffer is a mutex-protected byte buffer used to capture
// stdout+stderr from gnoland subprocesses. Writes happen on whatever
// goroutine the os/exec pipe drains on; reads happen on the caller's,
// through BootLogReader. Snapshot returns a slices.Clone so callers can
// read independently of further writes.
type bootLogBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func newBootLogBuffer() *bootLogBuffer { return &bootLogBuffer{} }

func (b *bootLogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *bootLogBuffer) Snapshot() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return slices.Clone(b.buf.Bytes())
}

// BootLogReader exposes the node's boot log so a caller can scrape what
// gnoland said as it came up: a snapshot of the stdout+stderr captured by
// the most recent BootFromExistingDataDir or BootFromGenesis call. Only
// BootFromGenesis starts gnoland with --log-format json, so a caller after
// machine-readable lines has to boot through that one.
//
// The reader holds a copy, so it stays valid after Halt or Cleanup and is
// safe to read while gnoland is still running.
//
// Returns nil when no boot has captured anything yet. Callers must
// nil-check.
func (c *Cluster) BootLogReader() io.Reader {
	if c.bootLog == nil {
		return nil
	}
	return bytes.NewReader(c.bootLog.Snapshot())
}

// Cleanup stops all nodes and cleans up resources.
// Node logs remain in TempDir/validator_N/{stdout,stderr}.log until TempDir is removed.
func (c *Cluster) Cleanup() {
	CleanupNodes(c.logger, c.Validators)
	if c.TempDir != "" {
		os.RemoveAll(c.TempDir)
	}
}

// haltTimeout bounds the per-cluster wait for graceful exit after SIGTERM.
// gnoland's signal handler flushes consensus WAL, mempool WAL, and app DB
// before exit; 30s is comfortably above the longest observed flush.
const haltTimeout = 30 * time.Second

// firstBlockTimeout bounds the wait for the chain to commit its first block.
// A validator set that never reaches quorum never commits one, so this is the
// deadline that reports a cluster which came up but cannot make progress.
const firstBlockTimeout = 60 * time.Second

// Halt cleanly stops the cluster with state-flush guarantees. Sends
// SIGTERM to every running validator process and waits for clean exit.
// gnoland's signal handler (gnoNode.Stop -> LocalApp.Close) flushes
// consensus WAL, mempool WAL, and app DB before exit, so each node's
// data dir is consistent and resumable post-Halt. Halt does NOT remove
// the TempDir; callers that intend to resume from the same data dir
// must avoid Cleanup.
//
// Returns an error if no processes are running, if SIGTERM cannot be
// delivered, if any process fails to exit within haltTimeout, or if
// ctx is cancelled before all processes exit.
//
// For multi-validator clusters this signals every validator at once and
// waits for all of them. Nothing coordinates the height they stop at, so
// their data dirs agree with each other only if they happened to halt on
// the same block.
func (c *Cluster) Halt(ctx context.Context) error {
	running := make([]*Node, 0, len(c.Validators))
	for _, n := range c.Validators {
		if n != nil && n.Process != nil {
			running = append(running, n)
		}
	}
	if len(running) == 0 {
		return fmt.Errorf("cluster not running")
	}

	for _, n := range running {
		if err := n.Process.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("send SIGTERM to validator %d: %w", n.Index, err)
		}
	}

	// Wait for each process in its own goroutine; collect results on a
	// channel so we can apply a single bounded deadline for the whole
	// cluster.
	type waitResult struct {
		index int
		err   error
	}
	results := make(chan waitResult, len(running))
	for _, n := range running {
		go func(n *Node) {
			// Node.Exited rather than Process.Wait: a process can be waited
			// for once, and readiness polling may already be waiting on this
			// one. Both observe the same reap instead of racing for it.
			<-n.Exited()
			results <- waitResult{index: n.Index, err: n.WaitErr()}
		}(n)
	}

	deadline := time.NewTimer(haltTimeout)
	defer deadline.Stop()

	var errs []error
	for i := 0; i < len(running); i++ {
		select {
		case r := <-results:
			// Process.Wait returns *exec.ExitError for non-zero exits,
			// which is normal for SIGTERM-triggered shutdowns. Only
			// treat unexpected errors as failures.
			if r.err != nil && !isExpectedHaltExit(r.err) {
				errs = append(errs, fmt.Errorf("validator %d wait: %w", r.index, r.err))
			}
		case <-ctx.Done():
			return fmt.Errorf("halt cancelled: %w", ctx.Err())
		case <-deadline.C:
			return fmt.Errorf("halt timeout after %v: not all validators exited cleanly", haltTimeout)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// validatorAt resolves a validator by position in Validators, which is the
// order scenarios see as RPC_ADDR_0 upward.
func (c *Cluster) validatorAt(index int) (*Node, error) {
	if index < 0 || index >= len(c.Validators) {
		return nil, fmt.Errorf("no validator %d: cluster has %d", index, len(c.Validators))
	}
	n := c.Validators[index]
	if n == nil {
		return nil, fmt.Errorf("validator %d is not set up", index)
	}
	return n, nil
}

// StopValidator terminates one validator and leaves the rest of the cluster
// running.
//
// Halt stops every validator at once, which is the whole cluster going away
// rather than a node failing, so fault injection needs this instead: losing
// one validator of three is survivable and the chain should keep producing
// blocks, and that is only observable if the others stay up.
//
// The node keeps its data dir and its identity, so RestartValidator can bring
// the same validator back. Only the process is released, which is what marks
// it stopped.
func (c *Cluster) StopValidator(ctx context.Context, index int) error {
	n, err := c.validatorAt(index)
	if err != nil {
		return err
	}
	if n.Process == nil {
		return fmt.Errorf("validator %d is not running", index)
	}

	if err := n.Process.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return fmt.Errorf("send SIGTERM to validator %d: %w", index, err)
	}

	select {
	case <-n.Exited():
	case <-ctx.Done():
		return fmt.Errorf("stop validator %d cancelled: %w", index, ctx.Err())
	case <-time.After(haltTimeout):
		// The caller asked for this node to be gone. Leaving a process that
		// ignored SIGTERM still bound to its ports would make the restart
		// below fail for a reason unrelated to what the scenario is testing.
		_ = n.Process.Kill()
		<-n.Exited()
	}

	if err := n.WaitErr(); err != nil && !isExpectedHaltExit(err) {
		return fmt.Errorf("validator %d wait: %w", index, err)
	}
	n.Process = nil
	return nil
}

// RestartValidator starts a stopped validator again from its existing data
// dir and waits for it to serve RPC.
//
// The node keeps the identity, keys, genesis and ports it had before, so it
// rejoins the same chain rather than forming a new one -- which is what makes
// "did it catch up, and does its state match?" answerable.
//
// ctx governs the restarted process, not just this call: the node is killed
// when it is cancelled. Callers that want the node to outlive the request that
// restarted it -- a test script command, say -- pass a context that outlives
// it too.
func (c *Cluster) RestartValidator(ctx context.Context, index int) error {
	n, err := c.validatorAt(index)
	if err != nil {
		return err
	}
	if n.Process != nil {
		return fmt.Errorf("validator %d is already running; stop it first", index)
	}
	if c.BinaryPath == "" {
		return fmt.Errorf("cluster has no binary path; it was never started")
	}

	if err := StartGnolandNode(ctx, c.BinaryPath, n, false); err != nil {
		return fmt.Errorf("restart validator %d: %w", index, err)
	}
	return nil
}

// BootFromExistingDataDir starts gnoland against an existing data dir,
// typically one a halted cluster left behind, skipping genesis
// construction. Writing a data dir with one binary and booting it with
// another is what makes a cross-version data-format check possible.
//
// Preconditions: the cluster must not already be running, the data dir
// must exist, and the binary path must exist. Validator keys, genesis,
// and config.toml are loaded as-is from the data dir; this method does
// not derive or rewrite them.
//
// On success the cluster's Validators, RPCAddr, and BinaryPath are
// populated and the node is RPC-ready. The caller owns subsequent
// Halt/Cleanup. One node only: a data dir belongs to one validator.
func (c *Cluster) BootFromExistingDataDir(ctx context.Context, dataDir, binary string) error {
	for _, n := range c.Validators {
		if n != nil && n.Process != nil {
			return fmt.Errorf("cluster already running")
		}
	}
	if _, err := os.Stat(dataDir); err != nil {
		return fmt.Errorf("data dir %q: %w", dataDir, err)
	}
	if _, err := os.Stat(binary); err != nil {
		return fmt.Errorf("binary %q: %w", binary, err)
	}

	// Load config.toml from the existing data dir to recover the RPC
	// address; this is what tests/orchestration query against.
	configPath := filepath.Join(dataDir, "config", "config.toml")
	cfg, err := config.LoadConfigFile(configPath)
	if err != nil {
		return fmt.Errorf("load config from %q: %w", configPath, err)
	}

	genesisPath := filepath.Join(dataDir, "test_genesis.json")
	if _, err := os.Stat(genesisPath); err != nil {
		return fmt.Errorf("genesis in data dir %q: %w", dataDir, err)
	}

	node := &Node{
		Index:   0,
		DataDir: dataDir,
		RPCAddr: cfg.RPC.ListenAddress,
		Genesis: genesisPath,
	}

	if err := StartGnolandNode(ctx, binary, node, false); err != nil {
		// StartGnolandNode may have spawned the gnoland process before
		// failing (e.g. WaitForNodeReady timed out). Tear down the node
		// so the process and open log files don't leak past this failed
		// boot. CleanupNodes is safe on a node whose Process is nil.
		CleanupNodes(c.logger, []*Node{node})
		return fmt.Errorf("start gnoland: %w", err)
	}

	c.Validators = []*Node{node}
	c.RPCAddr = node.RPCAddr
	c.BinaryPath = binary
	return nil
}

// BootFromGenesis starts gnoland against a fresh data dir using the
// caller-supplied genesis file. A genesis carrying a halted chain's state
// is how a caller replays that history onto a new binary: the historical
// txs are genesis-mode txs, so gnoland runs them during InitChain.
//
// Preconditions: the cluster must not already be running, the genesis
// path must exist, and the binary path must exist. A fresh validator
// data dir is allocated under c.TempDir (or system tmp if c.TempDir is
// empty); secrets and config are bootstrapped via the in-process
// SetupValidatorNode helper, then the supplied genesis is copied into
// the node's genesis path with fsync before gnoland start reads it.
//
// validatorSecretsSrc, when non-empty, points to an existing validator
// data dir whose secrets/ subdirectory will be cloned into the target
// node's secrets/ (overwriting the freshly-generated keys). Required
// when the supplied genesis preserves the source chain's validator set
// (the default behaviour of `gnogenesis fork generate`); otherwise the
// fresh validator key is not in the genesis validator list and the
// target produces no blocks. Pass "" when the genesis substitutes a
// fresh validator set (e.g. the `gnogenesis fork test` flow).
//
// gnoland is started with `--log-format json`, so every line it logs
// while booting lands in the captured boot log as one JSON object a
// caller can parse. The capture covers stdout and stderr both; read it
// back through BootLogReader.
//
// On success the cluster's Validators, RPCAddr, and BinaryPath are
// populated and the node is RPC-ready. The caller owns subsequent
// Halt/Cleanup. One node only: the genesis is installed for a single
// validator.
func (c *Cluster) BootFromGenesis(ctx context.Context, genesisPath, binary, validatorSecretsSrc string) (retErr error) {
	for _, n := range c.Validators {
		if n != nil && n.Process != nil {
			return fmt.Errorf("cluster already running")
		}
	}
	if _, err := os.Stat(genesisPath); err != nil {
		return fmt.Errorf("genesis file %q: %w", genesisPath, err)
	}
	if _, err := os.Stat(binary); err != nil {
		return fmt.Errorf("binary %q: %w", binary, err)
	}

	// Allocate a fresh data dir. If the cluster has no TempDir set
	// (caller did not pre-provision one), fall back to system tmp; this
	// keeps the failure-path tests using bare &Cluster{} working without
	// special-casing. Track whether we created it so a later failure can
	// remove it without nuking a caller-owned dir.
	parentDir := c.TempDir
	tempDirOwned := false
	if parentDir == "" {
		tmp, err := os.MkdirTemp("", "e2e-cluster-*")
		if err != nil {
			return fmt.Errorf("create cluster temp dir: %w", err)
		}
		parentDir = tmp
		c.TempDir = tmp
		tempDirOwned = true
	}
	defer func() {
		if retErr != nil && tempDirOwned {
			os.RemoveAll(c.TempDir)
			c.TempDir = ""
		}
	}()

	node, err := SetupValidatorNode(parentDir, 0)
	if err != nil {
		return fmt.Errorf("setup validator node: %w", err)
	}

	// Place the supplied genesis at the path StartGnolandNode passes to
	// `gnoland start --genesis`. Sync before close so gnoland reads a
	// fully-flushed file.
	if err := copyFileSynced(genesisPath, node.Genesis); err != nil {
		return fmt.Errorf("install genesis at %q: %w", node.Genesis, err)
	}

	// When the genesis preserves the source's validator set, clone the
	// source's validator identity (priv_validator_key.json, node_key.json)
	// so the target node's identity matches a genesis validator. The
	// freshly-generated priv_validator_state.json is intentionally NOT
	// overwritten: cloning it would carry the source chain's last signed
	// height/round/step, tripping the consensus state-machine's
	// step-regression check on the post-fork chain (whose initial_height
	// equals or exceeds the source's halt height).
	if validatorSecretsSrc != "" {
		dstSecrets := filepath.Join(node.DataDir, config.DefaultSecretsDir)
		srcSecrets := filepath.Join(validatorSecretsSrc, config.DefaultSecretsDir)
		for _, name := range []string{DefaultValidatorKeyName, DefaultNodeKeyName} {
			src := filepath.Join(srcSecrets, name)
			dst := filepath.Join(dstSecrets, name)
			if err := copyFileSynced(src, dst); err != nil {
				return fmt.Errorf("clone %s from %q: %w", name, srcSecrets, err)
			}
		}
	}

	c.bootLog = newBootLogBuffer()
	if err := StartGnolandNodeWithOpts(ctx, binary, node, GnolandStartOpts{
		ExtraWriter:   c.bootLog,
		LogFormatJSON: true,
	}); err != nil {
		// StartGnolandNodeWithOpts may have spawned the gnoland process
		// before failing (e.g. WaitForNodeReady timed out). Tear down the
		// node so the process and open log files don't leak past this
		// failed boot. CleanupNodes is safe on a node whose Process is nil.
		CleanupNodes(c.logger, []*Node{node})
		return fmt.Errorf("start gnoland: %w", err)
	}

	c.Validators = []*Node{node}
	c.RPCAddr = node.RPCAddr
	c.BinaryPath = binary
	return nil
}

// copyFileSynced copies src to dst and fsyncs the destination before
// returning. The flush ensures gnoland reads a fully-written genesis
// when a write-then-read sequence happens within the same process.
// Surfaces close errors via named return so flush failures aren't
// silently dropped.
func copyFileSynced(src, dst string) (retErr error) {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && retErr == nil {
			retErr = fmt.Errorf("close %s: %w", dst, cerr)
		}
	}()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %s -> %s: %w", src, dst, err)
	}
	if err := out.Sync(); err != nil {
		return fmt.Errorf("sync %s: %w", dst, err)
	}
	return nil
}

// isExpectedHaltExit reports whether err from Process.Wait represents a
// normal exit triggered by SIGTERM. gnoland's signal handler returns
// nil from cmd.Execute on a clean shutdown, so a nil err is the common
// case. Some platforms surface the signal as *exec.ExitError; treat any
// ExitError as expected here, because Halt sent the signal itself.
func isExpectedHaltExit(err error) bool {
	if err == nil {
		return true
	}
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr)
}

// ---- Cluster startup

// StartCluster sets up validators from genesis and starts the cluster on the
// gnoland at binaryPath.
//
// The binary is the caller's to produce, the way BootFromExistingDataDir takes
// one: a run boots a cluster per scenario and the binary is the same for all of
// them, so building here would repeat identical work per boot.
func StartCluster(ctx context.Context, cfg ClusterConfig, binaryPath string) (_ *Cluster, retErr error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	tempDir, err := os.MkdirTemp("", "e2e-cluster-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		if retErr != nil {
			os.RemoveAll(tempDir)
		}
	}()

	// Setup validators
	validators := make([]*Node, cfg.NumValidators)
	defer func() {
		if retErr != nil {
			CleanupNodes(logger, validators)
		}
	}()

	for i := 0; i < cfg.NumValidators; i++ {
		node, err := SetupValidatorNode(tempDir, i)
		if err != nil {
			return nil, fmt.Errorf("setup validator %d: %w", i, err)
		}
		validators[i] = node
	}

	// Create genesis
	if err := BuildGenesis(tempDir, cfg.Genesis, validators); err != nil {
		return nil, fmt.Errorf("build genesis: %w", err)
	}

	for _, node := range validators {
		if err := CopySharedGenesis(tempDir, node); err != nil {
			return nil, fmt.Errorf("copy genesis to node %d: %w", node.Index, err)
		}
	}

	// Configure P2P and consensus
	if err := ConfigureP2PTopology(validators, nil); err != nil {
		return nil, fmt.Errorf("configure P2P topology: %w", err)
	}
	for _, node := range validators {
		if err := ConfigureConsensusForSync(node); err != nil {
			return nil, fmt.Errorf("configure consensus for node %d: %w", node.Index, err)
		}
	}

	// After the harness's own passes and before any node starts, so what a
	// scenario asked for is what every node boots with.
	for _, node := range validators {
		if err := applyNodeConfig(node, cfg.NodeConfig); err != nil {
			return nil, fmt.Errorf("configure node %d: %w", node.Index, err)
		}
	}

	// Start validators
	for i, val := range validators {
		logger.Info("starting validator", "index", i+1)
		if err := StartGnolandNode(ctx, binaryPath, val, cfg.TeeNodeLogs); err != nil {
			return nil, fmt.Errorf("start validator %d: %w", i+1, err)
		}
		logger.Info("validator ready", "index", i+1)
	}

	if err := WaitForFirstBlock(ctx, validators[0]); err != nil {
		return nil, err
	}

	return &Cluster{
		Validators: validators,
		RPCAddr:    validators[0].RPCAddr,
		BinaryPath: binaryPath,
		TempDir:    tempDir,
		logger:     logger,
	}, nil
}

// WaitForFirstBlock waits until the chain has committed a block.
//
// Per-node readiness cannot cover this. A node reports ready once its ABCI app
// answers, which is before InitChain has written genesis, and it cannot report
// a committed block either: nodes start one at a time, and a validator set
// needing a quorum has none until the last of them is up. So the first
// transaction of a run can reach a node whose store holds no accounts yet, and
// is refused for a signature that is in fact correct.
func WaitForFirstBlock(ctx context.Context, node *Node) error {
	rpcClient, err := client.NewHTTPClient(node.RPCAddr)
	if err != nil {
		return fmt.Errorf("create RPC client: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, firstBlockTimeout)
	defer cancel()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for the chain to commit its first block")
		case <-node.Exited():
			return fmt.Errorf("node %d exited before the chain produced a block (%v)\n%s",
				node.Index, node.WaitErr(), nodeStderrTail(node)) //nolint:errorlint // WaitErr is nil on a normal exit; %w prints %!w(<nil>)
		case <-ticker.C:
			info, err := rpcClient.ABCIInfo(ctx)
			if err != nil || info.Response.Error != nil {
				continue
			}
			if info.Response.LastBlockHeight >= 1 {
				return nil
			}
		}
	}
}

// ---- Genesis building

// BuildGenesis creates a shared genesis file in tempDir using the given config.
func BuildGenesis(tempDir string, cfg GenesisConfig, validators []*Node) error {
	// Read validator keys
	validatorKeys := make([]*signer.FileKey, len(validators))
	for i, validator := range validators {
		validatorKeyPath := filepath.Join(validator.DataDir, "secrets", DefaultValidatorKeyName)
		validatorFileKey, err := signer.LoadFileKey(validatorKeyPath)
		if err != nil {
			return fmt.Errorf("load validator %d key: %w", i, err)
		}
		validatorKeys[i] = validatorFileKey
	}

	gen := &bft.GenesisDoc{}
	gen.GenesisTime = time.Date(2025, 7, 25, 7, 0, 0, 0, time.UTC)
	gen.ChainID = cfg.ChainID
	gen.ConsensusParams = abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxTxBytes:   cfg.MaxTxBytes,
			MaxDataBytes: cfg.MaxDataBytes,
			MaxGas:       cfg.MaxGas,
			TimeIotaMS:   cfg.TimeIotaMS,
		},
	}

	// Set up validators
	gen.Validators = make([]bft.GenesisValidator, len(validators))
	for i, key := range validatorKeys {
		gen.Validators[i] = bft.GenesisValidator{
			Address: key.Address,
			PubKey:  key.PubKey,
			Power:   1,
			Name:    fmt.Sprintf("testval%d", i+1),
		}
	}

	// Build balances: validators + extra accounts
	balances, err := buildBalances(tempDir, validatorKeys, cfg.Balances)
	if err != nil {
		return fmt.Errorf("build balances: %w", err)
	}

	// Optionally load example packages
	defaultGenState := gnoland.DefaultGenState()
	if cfg.CodeSubmissionPolicy != "" {
		defaultGenState.VM.Params.CodeSubmissionPolicy = cfg.CodeSubmissionPolicy
	}
	if len(cfg.PkgApprovers) > 0 {
		defaultGenState.VM.Params.PkgApprovers = cfg.PkgApprovers
	}
	// Last, so a scenario that spells out a path overrides the named setting
	// covering the same field rather than losing to it.
	if err := applyGenesisParams(&defaultGenState, cfg.Params); err != nil {
		return err
	}
	var allTxs []gnoland.TxWithMetadata

	if cfg.LoadExamples {
		examplesDir, err := findExamplesDir()
		if err != nil {
			return fmt.Errorf("find examples dir: %w", err)
		}

		txSender := validatorKeys[0].Address
		deployFee := std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
		pkgsTxs, err := gnoland.LoadPackagesFromDir(examplesDir, txSender, deployFee)
		if err != nil {
			return fmt.Errorf("load packages from examples: %w", err)
		}
		slog.Debug("loaded package transactions from examples", "count", len(pkgsTxs))

		if err := gnoland.SignGenesisTxs(pkgsTxs, validatorKeys[0].PrivKey, gen.ChainID); err != nil {
			return fmt.Errorf("sign genesis transactions: %w", err)
		}

		// Ensure deployer has sufficient balance
		deployerBalance := int64(len(pkgsTxs)) * 50_000_000
		balances.Set(txSender, std.NewCoins(std.NewCoin("ugnot", deployerBalance)))

		allTxs = append(allTxs, pkgsTxs...)
	}

	// Sign and append ExtraTxs with the first validator's private key
	if len(cfg.ExtraTxs) > 0 {
		if err := gnoland.SignGenesisTxs(cfg.ExtraTxs, validatorKeys[0].PrivKey, gen.ChainID); err != nil {
			return fmt.Errorf("sign extra genesis transactions: %w", err)
		}
		allTxs = append(allTxs, cfg.ExtraTxs...)
	}

	// Convert ExtraPackages to signed TXs using the first validator's key
	if len(cfg.ExtraPackages) > 0 {
		txSender := validatorKeys[0].Address
		deployFee := std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
		pkgTxs := make([]gnoland.TxWithMetadata, 0, len(cfg.ExtraPackages))
		for _, mpkg := range cfg.ExtraPackages {
			tx, err := gnoland.LoadPackage(mpkg, txSender, deployFee, nil)
			if err != nil {
				return fmt.Errorf("load extra package %s: %w", mpkg.Path, err)
			}
			pkgTxs = append(pkgTxs, gnoland.TxWithMetadata{Tx: tx})
		}
		if err := gnoland.SignGenesisTxs(pkgTxs, validatorKeys[0].PrivKey, gen.ChainID); err != nil {
			return fmt.Errorf("sign extra package transactions: %w", err)
		}

		// Add deployer balance for extra packages (additive, not replacing)
		extraNeeded := int64(len(pkgTxs)) * 50_000_000
		existing, _ := balances.Get(txSender)
		total := existing.Amount.Add(std.NewCoins(std.NewCoin("ugnot", extraNeeded)))
		balances.Set(txSender, total)

		slog.Debug("loaded extra package transactions", "count", len(pkgTxs))
		allTxs = append(allTxs, pkgTxs...)
	}

	// Sign and append ExtraCallTxs (genesis function calls like Bootstrap, RegisterUser).
	// These must come AFTER package deploy TXs so the packages exist when calls execute.
	if len(cfg.ExtraCallTxs) > 0 {
		// Set the caller address on each MsgCall to the first validator.
		for i := range cfg.ExtraCallTxs {
			for j := range cfg.ExtraCallTxs[i].Tx.Msgs {
				if mc, ok := cfg.ExtraCallTxs[i].Tx.Msgs[j].(vm.MsgCall); ok {
					mc.Caller = validatorKeys[0].Address
					cfg.ExtraCallTxs[i].Tx.Msgs[j] = mc
				}
			}
		}
		if err := gnoland.SignGenesisTxs(cfg.ExtraCallTxs, validatorKeys[0].PrivKey, gen.ChainID); err != nil {
			return fmt.Errorf("sign extra call transactions: %w", err)
		}
		slog.Debug("loaded extra call transactions", "count", len(cfg.ExtraCallTxs))
		allTxs = append(allTxs, cfg.ExtraCallTxs...)
	}

	defaultGenState.Txs = allTxs
	defaultGenState.Balances = balances.List()
	gen.AppState = defaultGenState

	sharedGenesisPath := filepath.Join(tempDir, "shared_genesis.json")
	if err := gen.SaveAs(sharedGenesisPath); err != nil {
		return fmt.Errorf("save genesis: %w", err)
	}
	slog.Debug("created genesis", "path", sharedGenesisPath)

	PrintGenesisConfig(gen)
	return nil
}

// buildBalances creates a GenesisBalances object with validator + extra balances.
func buildBalances(tempDir string, validatorKeys []*signer.FileKey, extraBalances map[string]int64) (gnoland.Balances, error) {
	balanceFile, err := CreateEnhancedBalanceFile(tempDir, validatorKeys, extraBalances)
	if err != nil {
		return nil, fmt.Errorf("create balance file: %w", err)
	}
	balances, err := gnoland.LoadGenesisBalancesFile(balanceFile)
	if err != nil {
		return nil, fmt.Errorf("load genesis balances: %w", err)
	}
	return balances, nil
}

// findExamplesDir locates the examples directory using gnoenv.RootDir()
func findExamplesDir() (string, error) {
	gnoRoot := gnoenv.RootDir()
	examplesPath := filepath.Join(gnoRoot, "examples")

	if info, err := os.Stat(examplesPath); err == nil && info.IsDir() {
		slog.Debug("found examples directory", "path", examplesPath)
		return examplesPath, nil
	}

	return "", fmt.Errorf("examples directory not found: %s", examplesPath)
}
