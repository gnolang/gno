package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
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
			ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
			defer cancel()

			return execTest(ctx, cfg, io)
		},
	)
}

func (c *testCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.IntVar(&c.numValidators, "validators", 2, "number of validator nodes")
	fs.IntVar(&c.numNonValidators, "non-validators", 1, "number of non-validator nodes")
	fs.IntVar(&c.numTransactions, "transactions", 5, "number of test transactions")
	fs.Int64Var(&c.targetHeight, "height", 50, "target blockchain height")
	fs.DurationVar(&c.maxTestTime, "timeout", 15*time.Minute, "maximum test duration")
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

	// Setup slog logger with info level
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create SlogTestingT for structured test logging
	st, cleanup := NewSlogTestingT(logger)
	defer cleanup()

	st.Log("🎯 E2E Multi-Node Determinism Test")
	st.Logf("📋 Configuration - Validators: %d, Non-validators: %d, Target height: %d",
		c.numValidators, c.numNonValidators, c.targetHeight)

	// Run the determinism test with structured logging
	defer func() {
		if r := recover(); r != nil {
			// Test failed and called FailNow() which panics
			st.Log("❌ Test failed with panic - check logs for details")
			panic(r)
		}
	}()

	runDeterminismTest(st, ctx, c)

	st.Log("✅ Test completed successfully!")
	return nil
}

func runDeterminismTest(t TestingT, ctx context.Context, c *testCfg) {
	testCtx, cancel := context.WithTimeout(ctx, c.maxTestTime)
	defer cancel()

	tempDir := t.TempDir()
	t.Logf("📁 Working directory: %s", tempDir)

	// Build gnoland binary
	t.Log("🔨 Building gnoland binary...")
	binaryPath, err := buildGnolandBinary(t, tempDir)
	require.NoError(t, err, "failed to build gnoland binary")
	t.Logf("✅ Built binary: %s", binaryPath)

	var wg sync.WaitGroup
	totalNodes := c.numValidators + c.numNonValidators
	nodes := make([]*Node, 0, totalNodes)

	// Create validator nodes
	t.Logf("📋 Creating %d validator nodes...", c.numValidators)
	validators := make([]*Node, c.numValidators)
	for i := 0; i < c.numValidators; i++ {
		validators[i] = setupValidatorNode(t, tempDir, i)
		t.Logf("Created validator %d - ID: %s, Port: %d", i+1, validators[i].NodeID, validators[i].P2PPort)
	}

	// Create non-validator nodes
	t.Logf("📋 Creating %d non-validator nodes...", c.numNonValidators)
	nonValidators := make([]*Node, c.numNonValidators)
	for i := 0; i < c.numNonValidators; i++ {
		nonValidators[i] = setupNonValidatorNode(t, tempDir, c.numValidators+i)
		t.Logf("Created non-validator %d - ID: %s, Port: %d", i+1, nonValidators[i].NodeID, nonValidators[i].P2PPort)
	}

	// Combine all nodes
	nodes = append(nodes, validators...)
	nodes = append(nodes, nonValidators...)

	// Create shared genesis
	t.Log("📋 Creating shared genesis file")
	t.Logf("Genesis will include %d validators", c.numValidators)
	createSharedGenesis(t, tempDir, validators)

	// Copy genesis to all nodes
	t.Log("📋 Copying genesis to all nodes")
	for _, node := range nodes {
		copySharedGenesis(t, tempDir, node)
	}

	// Configure P2P topology
	t.Log("📋 Configuring P2P topology")
	configureP2PTopology(t, validators, nonValidators)

	// Configure consensus settings
	t.Log("📋 Configuring consensus settings")
	for _, node := range nodes {
		configureConsensusForSync(t, node)
	}

	// Print node configurations
	printNodeConfigurations(t, nodes, c)

	// Cleanup processes at the end
	defer func() {
		cancel()
		wg.Wait()
		cleanupNodes(t, nodes)
	}()

	// Start nodes and run test
	t.Log("📋 Starting nodes and running determinism test")
	runMultiNodeTest(t, testCtx, &wg, binaryPath, validators, nonValidators, c)
}
