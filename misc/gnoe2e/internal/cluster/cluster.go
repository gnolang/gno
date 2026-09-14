package cluster

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
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
	ChainID      string
	MaxGas       int64
	MaxTxBytes   int64
	MaxDataBytes int64
	TimeIotaMS   int64
	LoadExamples bool
	Balances     map[string]int64 // extra funded accounts (addr -> ugnot)
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

// DefaultGenesisConfig returns the genesis parameters a cluster starts from
// when a scenario declares none of its own.
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
	return nil
}

// ResolveGenesisState settles what the chain will really boot with: the named
// fields first, then the generic `genesis.` family, which is applied last so a
// path spelling wins over the named key covering the same field.
//
// Anything that needs to know what a scenario asked for reads this rather than
// the named fields, which a path can supersede without ever touching them.
func ResolveGenesisState(cfg GenesisConfig) (gnoland.GnoGenesisState, error) {
	state := gnoland.DefaultGenState()
	if cfg.CodeSubmissionPolicy != "" {
		state.VM.Params.CodeSubmissionPolicy = cfg.CodeSubmissionPolicy
	}
	if len(cfg.PkgApprovers) > 0 {
		state.VM.Params.PkgApprovers = cfg.PkgApprovers
	}
	if err := applyGenesisParams(&state, cfg.Params); err != nil {
		return gnoland.GnoGenesisState{}, err
	}
	return state, nil
}

// ValidateGenesisState reports a genesis no scenario can be written against.
//
// Separate from GenesisConfig.Validate, and checked on the resolved state
// rather than on the fields, for two reasons: the `genesis.` family reaches the
// same settings without touching those fields, and `run` fills the approver set
// per scenario, so its flag template legitimately holds an inert policy with
// nobody in it.
func ValidateGenesisState(state gnoland.GnoGenesisState) error {
	if state.VM.Params.CodeSubmissionPolicy == vm.CodeSubmissionPolicyInert &&
		len(state.VM.Params.PkgApprovers) == 0 {
		return errors.New("an inert chain needs at least one package approver: " +
			"with code_submission_policy=inert and an empty pkg_approvers, every package " +
			"submitted after genesis parks with nobody able to enable it")
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
}

// Cleanup stops all nodes and cleans up resources.
// Node logs remain in TempDir/validator_N/{stdout,stderr}.log until TempDir is removed.
func (c *Cluster) Cleanup() {
	CleanupNodes(c.logger, c.Validators)
	if c.TempDir == "" {
		return
	}

	if err := os.RemoveAll(c.TempDir); err != nil {
		// Hundreds of megabytes per validator, so a directory that stays has
		// to be reported or nobody goes looking for it.
		logger := c.logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Warn("cluster temp dir not removed", "dir", c.TempDir, "error", err)
	}
}

// haltTimeout bounds the wait for a node to exit after it has been signalled.
// gnoland's signal handler flushes consensus WAL, mempool WAL, and app DB
// before exit; 30s is comfortably above the longest observed flush.
const haltTimeout = 30 * time.Second

// firstBlockTimeout bounds the wait for the chain to commit its first block.
// A validator set that never reaches quorum never commits one, so this is the
// deadline that reports a cluster which came up but cannot make progress.
const firstBlockTimeout = 60 * time.Second

// The ways a caller can name something the cluster cannot act on. Separate from
// the errors a stop or a restart produces by trying, because these say the
// action was never attempted: a verb that lets a script negate one of these
// would pass a scenario whose cluster never lost a node at all.
var (
	ErrUnknownValidator = errors.New("no validator")
	ErrValidatorRunning = errors.New("validator is already running")
)

// validatorAt resolves a validator by position in Validators, which is the
// order scenarios see as RPC_ADDR_0 upward.
func (c *Cluster) validatorAt(index int) (*Node, error) {
	if index < 0 || index >= len(c.Validators) {
		return nil, fmt.Errorf("%w %d: cluster has %d", ErrUnknownValidator, index, len(c.Validators))
	}
	n := c.Validators[index]
	if n == nil {
		return nil, fmt.Errorf("%w %d: it was never set up", ErrUnknownValidator, index)
	}
	return n, nil
}

// StopValidator terminates one validator and leaves the rest of the cluster
// running.
//
// Cleanup takes every validator down at once, which is the whole cluster going
// away rather than a node failing, so fault injection needs this instead: losing
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

	// A validator that died on its own keeps its *os.Process, and its PID
	// answers signals until it is reaped, so neither tells this apart from a
	// node that is about to be stopped. Reporting success would tell a scenario
	// it caused an outage that had in fact already started, and the reason the
	// node went down would never be reported at all.
	alreadyExited := func() error {
		<-n.Exited()
		n.clearProcess()
		return fmt.Errorf("validator %d had already exited before it was stopped (%s)",
			index, exitReason(n.WaitErr()))
	}
	select {
	case <-n.Exited():
		return alreadyExited()
	default:
	}

	if err := n.Process.Signal(syscall.SIGTERM); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return alreadyExited()
		}
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

	if err := n.WaitErr(); err != nil && !isExpectedStopExit(err) {
		return fmt.Errorf("validator %d wait: %w", index, err)
	}
	n.clearProcess()
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
		return fmt.Errorf("%w: stop validator %d first", ErrValidatorRunning, index)
	}
	if c.BinaryPath == "" {
		return fmt.Errorf("cluster has no binary path; it was never started")
	}

	if err := StartGnolandNode(ctx, c.BinaryPath, n, GnolandStartOpts{}); err != nil {
		return fmt.Errorf("restart validator %d: %w", index, err)
	}
	return nil
}

// isExpectedStopExit reports whether err from Node.WaitErr represents a normal
// exit triggered by the signal the caller sent. gnoland's signal handler
// returns nil from cmd.Execute on a clean shutdown, so nil is one case; a node
// that exits non-zero, or dies to the signal, surfaces as *exec.ExitError, and
// that is expected too because the caller asked for the stop.
func isExpectedStopExit(err error) bool {
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
// The binary is the caller's to produce: a run boots a cluster per scenario and
// the binary is the same for all of them, so building here would repeat
// identical work per boot.
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

	if err := BuildGenesis(tempDir, cfg.Genesis, validators); err != nil {
		return nil, fmt.Errorf("build genesis: %w", err)
	}

	for _, node := range validators {
		if err := CopySharedGenesis(tempDir, node); err != nil {
			return nil, fmt.Errorf("copy genesis to node %d: %w", node.Index, err)
		}
	}

	if err := ConfigureP2PTopology(validators); err != nil {
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

	for i, val := range validators {
		logger.Info("starting validator", "index", i+1)
		if err := StartGnolandNode(ctx, binaryPath, val, GnolandStartOpts{TeeNodeLogs: cfg.TeeNodeLogs}); err != nil {
			return nil, fmt.Errorf("start validator %d: %w", i+1, err)
		}
		logger.Info("validator ready", "index", i+1)
	}

	if err := WaitForFirstBlock(ctx, validators[0]); err != nil {
		// The deferred cleanup above removes the temp dir, and every
		// validator_N/stderr.log with it, so what the nodes said has to be
		// carried out in the error itself.
		if tails := nodeLogTails(validators); tails != "" {
			return nil, fmt.Errorf("%w\n%s", err, tails)
		}
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
	return waitForFirstBlock(ctx, node, firstBlockTimeout)
}

// waitForFirstBlock is WaitForFirstBlock with the deadline exposed, so a test
// can reach the timeout branch without waiting out the real one.
func waitForFirstBlock(ctx context.Context, node *Node, timeout time.Duration) error {
	rpcClient, err := client.NewHTTPClient(node.RPCAddr)
	if err != nil {
		return fmt.Errorf("create RPC client: %w", err)
	}

	run := ctx
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	// What the last poll saw. A validator set that never reaches quorum answers
	// RPC perfectly and stays at height zero, so the height and the last RPC
	// error are the whole difference between "nobody is proposing" and "this
	// node cannot be reached".
	var (
		lastHeight int64
		lastErr    error
	)

	for {
		select {
		case <-ctx.Done():
			if err := run.Err(); err != nil {
				return fmt.Errorf("waiting for the chain to commit its first block: %w", err)
			}
			return fmt.Errorf(
				"the chain did not commit a block within %s (polled node %d at %s, last height %d, last RPC error: %s)",
				timeout, node.Index, node.RPCAddr, lastHeight, exitReason(lastErr))
		case <-node.Exited():
			return fmt.Errorf("node %d exited before the chain produced a block (%s)\n%s",
				node.Index, exitReason(node.WaitErr()), nodeStderrTail(node))
		case <-ticker.C:
			info, err := rpcClient.ABCIInfo(ctx)
			if err != nil {
				lastErr = err
				continue
			}
			if info.Response.Error != nil {
				lastErr = info.Response.Error
				continue
			}
			lastHeight = info.Response.LastBlockHeight
			if lastHeight >= 1 {
				return nil
			}
		}
	}
}

// ---- Genesis building

// BuildGenesis creates a shared genesis file in tempDir using the given config.
func BuildGenesis(tempDir string, cfg GenesisConfig, validators []*Node) error {
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

	gen.Validators = make([]bft.GenesisValidator, len(validators))
	for i, key := range validatorKeys {
		gen.Validators[i] = bft.GenesisValidator{
			Address: key.Address,
			PubKey:  key.PubKey,
			Power:   1,
			Name:    fmt.Sprintf("testval%d", i+1),
		}
	}

	balances, err := clusterBalances(validatorKeys, cfg.Balances)
	if err != nil {
		return err
	}

	defaultGenState, err := ResolveGenesisState(cfg)
	if err != nil {
		return err
	}
	// Checked here rather than on the config, because here is where every
	// spelling of every setting has landed on the params the node will read.
	if err := ValidateGenesisState(defaultGenState); err != nil {
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

		fundDeployer(balances, txSender, len(pkgsTxs))

		allTxs = append(allTxs, pkgsTxs...)
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

func findExamplesDir() (string, error) {
	gnoRoot := gnoenv.RootDir()
	examplesPath := filepath.Join(gnoRoot, "examples")

	if info, err := os.Stat(examplesPath); err == nil && info.IsDir() {
		slog.Debug("found examples directory", "path", examplesPath)
		return examplesPath, nil
	}

	return "", fmt.Errorf("examples directory not found: %s", examplesPath)
}
