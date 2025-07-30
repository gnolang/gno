package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

// Node represents a gnoland node instance
type Node struct {
	Index      int
	NodeID     string
	DataDir    string
	P2PPort    int
	SocketAddr string
	Genesis    string
	Process    *os.Process
}

type testCfg struct {
	numValidators    int
	numNonValidators int
	numTransactions  int
	targetHeight     int64
	maxTestTime      time.Duration
	verbose          bool
}

func newTestCmd(io commands.IO) *commands.Command {
	cfg := &testCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "test",
			ShortUsage: "test [flags]",
			ShortHelp:  "runs the multi-node determinism test",
			LongHelp:   "Runs the E2E multi-node determinism test with configurable validators, non-validators, and target height",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execTest(ctx, cfg, io)
		},
	)
}

func (c *testCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.IntVar(
		&c.numValidators,
		"validators",
		2,
		"number of validator nodes",
	)

	fs.IntVar(
		&c.numNonValidators,
		"non-validators",
		3,
		"number of non-validator nodes",
	)

	fs.IntVar(
		&c.numTransactions,
		"transactions",
		5,
		"number of test transactions",
	)

	fs.Int64Var(
		&c.targetHeight,
		"height",
		205,
		"target blockchain height",
	)

	fs.DurationVar(
		&c.maxTestTime,
		"timeout",
		15*time.Minute,
		"maximum test duration",
	)

	fs.BoolVar(
		&c.verbose,
		"verbose",
		false,
		"enable verbose logging",
	)
}

func execTest(ctx context.Context, c *testCfg, io commands.IO) error {
	// Validate parameters
	if c.numValidators < 1 {
		return fmt.Errorf("at least 1 validator required")
	}
	if c.numNonValidators < 0 {
		return fmt.Errorf("non-validators must be >= 0")
	}
	if c.targetHeight < 10 {
		return fmt.Errorf("target height must be >= 10")
	}

	// Setup slog logger
	var logger *slog.Logger
	if c.verbose {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}
	slog.SetDefault(logger)

	slog.Info("🎯 E2E Multi-Node Determinism Test")
	slog.Info("📋 Configuration",
		"validators", c.numValidators,
		"non_validators", c.numNonValidators,
		"target_height", c.targetHeight)

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "e2e_multinode_*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	slog.Info("📁 Working directory", "path", tempDir)

	// Run the determinism test
	if err := runDeterminismTest(ctx, tempDir, c); err != nil {
		return fmt.Errorf("test failed: %w", err)
	}

	slog.Info("✅ Test completed successfully!")
	return nil
}

func runDeterminismTest(ctx context.Context, tempDir string, c *testCfg) error {
	testCtx, cancel := context.WithTimeout(ctx, c.maxTestTime)
	defer cancel()

	// Build gnoland binary
	binaryPath, err := buildGnolandBinary(tempDir)
	if err != nil {
		return fmt.Errorf("failed to build gnoland binary: %w", err)
	}

	var wg sync.WaitGroup
	totalNodes := c.numValidators + c.numNonValidators
	nodes := make([]*Node, 0, totalNodes)

	// Create validator nodes
	validators := make([]*Node, c.numValidators)
	for i := 0; i < c.numValidators; i++ {
		validators[i] = setupValidatorNode(tempDir, i)
		slog.Info("Created validator", "index", i+1, "node_id", validators[i].NodeID, "port", validators[i].P2PPort)
	}

	// Create non-validator nodes  
	nonValidators := make([]*Node, c.numNonValidators)
	for i := 0; i < c.numNonValidators; i++ {
		nonValidators[i] = setupNonValidatorNode(tempDir, c.numValidators+i)
		slog.Info("Created non-validator", "index", i+1, "node_id", nonValidators[i].NodeID, "port", nonValidators[i].P2PPort)
	}

	// Combine all nodes
	nodes = append(nodes, validators...)
	nodes = append(nodes, nonValidators...)

	// Cleanup processes at the end
	defer func() {
		cancel()
		wg.Wait()
		cleanupNodes(nodes)
	}()

	// Create shared genesis
	slog.Info("📋 Creating shared genesis file", "validators", c.numValidators)
	if err := createSharedGenesis(tempDir, validators); err != nil {
		return fmt.Errorf("failed to create genesis: %w", err)
	}

	// Copy genesis to all nodes
	for _, node := range nodes {
		if err := copySharedGenesis(tempDir, node); err != nil {
			return fmt.Errorf("failed to copy genesis to node %d: %w", node.Index, err)
		}
	}

	// Configure P2P topology
	slog.Info("📋 Configuring P2P topology")
	if err := configureP2PTopology(validators, nonValidators); err != nil {
		return fmt.Errorf("failed to configure P2P: %w", err)
	}

	// Configure consensus settings
	slog.Info("📋 Configuring consensus settings")
	for _, node := range nodes {
		if err := configureConsensusForSync(node); err != nil {
			return fmt.Errorf("failed to configure consensus for node %d: %w", node.Index, err)
		}
	}

	// Print configurations if verbose
	if c.verbose {
		printNodeConfigurations(nodes, c)
	}

	// Start nodes and run test
	slog.Info("📋 Starting nodes and running determinism test")
	return runMultiNodeTest(testCtx, &wg, binaryPath, validators, nonValidators, c)
}