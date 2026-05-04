package fork

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/amino"
	tmcfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type testCfg struct {
	genesis                    string
	timeout                    time.Duration
	verbose                    bool
	keepRunning                bool
	skipFailingTxs             bool
	skipGenesisSigVerification bool
}

func newTestCmd(io commands.IO) *commands.Command {
	cfg := &testCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "test",
			ShortUsage: "test [flags]",
			ShortHelp:  "smoke-test a hardfork genesis by replaying it in-process",
			LongHelp: `Smoke-tests a hardfork genesis by loading it into an in-memory gnoland node
and replaying all transactions (genesis-mode and historical).

A fresh single-validator identity is generated for the test — it replaces the
real validators in the genesis so the node can produce blocks without requiring
the actual validator keys. The app state (txs, balances, packages) is kept
exactly as-is.

SkipGenesisSigVerification is enabled for genesis-mode txs. Historical txs
(those with block_height > 0) go through the normal ante handler using the
original_chain_id from the genesis to verify signatures.

Exit code: 0 on success (all txs replayed, first block produced), non-zero on failure.

Examples:

  # Smoke-test the default output of hardfork genesis:
  hardfork test --genesis genesis.json

  # With a longer timeout and verbose tx logging:
  hardfork test --genesis genesis.json --timeout 2h --verbose

  # Keep the node running after replay for manual inspection via RPC:
  hardfork test --genesis genesis.json --keep-running`,
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execTest(ctx, cfg, io)
		},
	)
}

func (c *testCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.genesis, "genesis", "genesis.json", "path to the hardfork genesis.json to test")
	fs.DurationVar(&c.timeout, "timeout", 30*time.Minute, "maximum time to wait for genesis replay to complete")
	fs.BoolVar(&c.verbose, "verbose", false, "print each tx result during replay")
	fs.BoolVar(&c.keepRunning, "keep-running", false, "keep the node running after genesis replay (for manual RPC inspection)")
	fs.BoolVar(&c.skipFailingTxs, "skip-failing-genesis-txs", false,
		"count failed genesis txs as informational (report count, still exit 0) instead of failing the test. "+
			"Match this to production node flags when the chain runs with -skip-failing-genesis-txs.")
	fs.BoolVar(&c.skipGenesisSigVerification, "skip-genesis-sig-verification", true,
		"bypass signature verification for genesis-mode txs (default true, matching production node behavior). "+
			"Set to false to exercise sig verification as a stricter consistency check.")
}

func execTest(ctx context.Context, cfg *testCfg, io commands.IO) error {
	// -------------------------------------------------------------------------
	// Step 1: Load and parse the genesis file
	// -------------------------------------------------------------------------
	io.Printf("Loading genesis: %s\n", cfg.genesis)

	data, err := os.ReadFile(cfg.genesis)
	if err != nil {
		return fmt.Errorf("reading genesis file: %w", err)
	}

	var genDoc bft.GenesisDoc
	if err := amino.UnmarshalJSON(data, &genDoc); err != nil {
		return fmt.Errorf("parsing genesis: %w", err)
	}

	if err := genDoc.ValidateAndComplete(); err != nil {
		return fmt.Errorf("genesis validation failed: %w", err)
	}

	// Extract app state for summary
	appState, ok := genDoc.AppState.(gnoland.GnoGenesisState)
	if !ok {
		raw, err := amino.MarshalJSON(genDoc.AppState)
		if err != nil {
			return fmt.Errorf("marshalling app state: %w", err)
		}
		if err := amino.UnmarshalJSON(raw, &appState); err != nil {
			return fmt.Errorf("unmarshalling app state: %w", err)
		}
	}

	genesisModeTxs := baseGenesisModeTxs(&appState)
	historicalTxs := len(appState.Txs) - len(genesisModeTxs)

	io.Printf("  Past chain IDs:    %v\n", appState.PastChainIDs)
	io.Printf("  New chain ID:      %s\n", genDoc.ChainID)
	io.Printf("  Initial height:    %d\n", genDoc.InitialHeight)
	io.Printf("  Genesis-mode txs:  %d\n", len(genesisModeTxs))
	io.Printf("  Historical txs:    %d\n", historicalTxs)
	io.Printf("  Total txs:         %d\n", len(appState.Txs))

	if len(appState.PastChainIDs) == 0 && historicalTxs > 0 {
		io.Println("  WARNING: past_chain_ids is empty — historical tx signatures cannot be verified.")
	}

	// -------------------------------------------------------------------------
	// Step 2: Replace validators with a local test identity
	// -------------------------------------------------------------------------
	pv := bft.NewMockPV()
	pk := pv.PubKey()
	genDoc.Validators = []bft.GenesisValidator{
		{
			Address: pk.Address(),
			PubKey:  pk,
			Power:   10,
			Name:    "hardfork-test-node",
		},
	}

	// -------------------------------------------------------------------------
	// Step 3: Find GNOROOT (needed for stdlibs)
	// -------------------------------------------------------------------------
	gnoroot, err := gnoenv.GuessRootDir()
	if err != nil {
		return fmt.Errorf("cannot locate GNOROOT (set the GNOROOT env var): %w", err)
	}

	stdlibDir := filepath.Join(gnoroot, "gnovm", "stdlibs")
	if _, err := os.Stat(stdlibDir); err != nil {
		return fmt.Errorf("stdlibs directory not found at %s (is GNOROOT correct?): %w", stdlibDir, err)
	}

	// -------------------------------------------------------------------------
	// Step 4: Set up tx result tracking
	// -------------------------------------------------------------------------
	var txFailures atomic.Int64
	var txProcessed atomic.Int64

	txResultHandler := func(ctx sdk.Context, tx std.Tx, res sdk.Result) {
		txProcessed.Add(1)
		if res.IsErr() {
			txFailures.Add(1)
			if cfg.verbose {
				io.Printf("  [FAIL] height=%d error=%s\n", ctx.BlockHeight(), res.Log)
			}
		} else if cfg.verbose {
			msgs := make([]string, len(tx.Msgs))
			for i, m := range tx.Msgs {
				msgs[i] = m.Type()
			}
			io.Printf("  [OK]   height=%d msgs=%v\n", ctx.BlockHeight(), msgs)
		}
	}

	// -------------------------------------------------------------------------
	// Step 5: Configure in-memory node
	// -------------------------------------------------------------------------
	tmConfig := tmcfg.TestConfig().SetRootDir(gnoroot)
	tmConfig.Consensus.WALDisabled = true
	tmConfig.Consensus.SkipTimeoutCommit = true
	tmConfig.Consensus.CreateEmptyBlocks = false
	tmConfig.RPC.ListenAddress = "tcp://127.0.0.1:0" // random port, avoids conflicts
	tmConfig.P2P.ListenAddress = "tcp://127.0.0.1:0"

	nodeCfg := &gnoland.InMemoryNodeConfig{
		PrivValidator:              pv,
		Genesis:                    &genDoc,
		TMConfig:                   tmConfig,
		DB:                         memdb.NewMemDB(),
		SkipGenesisSigVerification: cfg.skipGenesisSigVerification,
		InitChainerConfig: gnoland.InitChainerConfig{
			GenesisTxResultHandler: txResultHandler,
			StdlibDir:              stdlibDir,
			CacheStdlibLoad:        false,
			// fork test injects a fresh MockPV as the sole genesis
			// validator; its signing addr has no valoper profile, so
			// the hardfork-mode coverage assertion would fire spuriously.
			SkipValoperCoverageAssertion: true,
		},
	}

	// Choose logger: quiet by default, real output when verbose
	var nodeLogger *slog.Logger
	if cfg.verbose {
		nodeLogger = slog.Default()
	} else {
		nodeLogger = log.NewNoopLogger()
	}

	// -------------------------------------------------------------------------
	// Step 6: Start the node
	// -------------------------------------------------------------------------
	io.Println()
	io.Println("Starting in-memory node for genesis replay...")

	n, err := gnoland.NewInMemoryNode(nodeLogger, nodeCfg)
	if err != nil {
		return fmt.Errorf("creating in-memory node: %w", err)
	}

	start := time.Now()

	if err := n.Start(); err != nil {
		return fmt.Errorf("starting node: %w", err)
	}

	defer func() {
		if stopErr := n.Stop(); stopErr != nil {
			io.Printf("WARNING: error stopping node: %v\n", stopErr)
		}
	}()

	// -------------------------------------------------------------------------
	// Step 7: Wait for genesis replay to complete (first block produced)
	// -------------------------------------------------------------------------
	io.Printf("Replaying %d txs (timeout: %s)...\n", len(appState.Txs), cfg.timeout)

	// Progress ticker: print elapsed time every 30s so the user knows it's alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.timeout)
	defer cancel()

	for {
		select {
		case <-n.Ready():
			elapsed := time.Since(start)
			failures := txFailures.Load()
			processed := txProcessed.Load()

			io.Println()
			io.Println("=== Test Results ===")
			io.Printf("  Elapsed:           %s\n", elapsed.Round(time.Second))
			io.Printf("  Txs processed:     %d / %d\n", processed, len(appState.Txs))
			io.Printf("  Failures:          %d\n", failures)

			if failures > 0 {
				io.Println()
				if cfg.skipFailingTxs {
					// --skip-failing-genesis-txs matches production cluster
					// behavior: failed genesis txs are absorbed so the chain
					// still boots. Report count for visibility; don't fail.
					io.Printf("WARN: %d transaction(s) failed during genesis replay (absorbed by --skip-failing-genesis-txs).\n", failures)
					io.Println("Run with --verbose to see individual failures.")
					io.Println()
					io.Println("PASS: genesis replay completed (failures were suppressed).")
				} else {
					io.Printf("FAIL: %d transaction(s) failed during genesis replay.\n", failures)
					io.Println("Run with --verbose to see individual failures.")
					return fmt.Errorf("genesis replay completed with %d failures", failures)
				}
			} else {
				io.Println()
				io.Println("PASS: genesis replay completed successfully.")
			}

			if cfg.keepRunning {
				io.Println()
				io.Printf("Node is running at: %s\n", tmConfig.RPC.ListenAddress)
				io.Println("Press Ctrl+C to stop.")
				<-ctx.Done()
			}

			return nil

		case <-ticker.C:
			elapsed := time.Since(start)
			processed := txProcessed.Load()
			io.Printf("  ... still replaying: %d/%d txs, %s elapsed\n",
				processed, len(appState.Txs), elapsed.Round(time.Second))

		case <-timeoutCtx.Done():
			processed := txProcessed.Load()
			return fmt.Errorf("genesis replay timed out after %s (%d/%d txs processed)",
				cfg.timeout, processed, len(appState.Txs))
		}
	}
}
