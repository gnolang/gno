package replay

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	gnoLog "github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/events"
)

type replayCfg struct {
	dataDir    string
	genesis    string
	haltHeight int64
	timeout    time.Duration
	skipSigVer bool
	skipFail   bool
}

func NewReplayCmd(io commands.IO) *commands.Command {
	cfg := &replayCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "replay",
			ShortUsage: "replay [flags]",
			ShortHelp:  "replay chain state with current binary to verify upgrade compatibility",
			LongHelp: `Starts a gnoland node against an existing data directory to verify that
the current binary can successfully replay all committed blocks. This is the
primary smoke test for chain upgrades.

The node starts in single-validator mode and replays all blocks from the
existing database. A successful replay (no panic, no crash) indicates that
the new binary is compatible with the existing chain state.

Exit codes:
  0  replay completed successfully
  1  replay failed (panic, crash, or timeout)`,
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execReplay(ctx, cfg, io)
		},
	)
}

func (c *replayCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.dataDir,
		"data-dir",
		"",
		"path to the chain data directory (contains config/ and data/ subdirs)",
	)

	fs.StringVar(
		&c.genesis,
		"genesis",
		"",
		"path to genesis.json (default: <data-dir>/config/genesis.json)",
	)

	fs.Int64Var(
		&c.haltHeight,
		"halt-height",
		0,
		"stop replay after this block height (0 = replay all available blocks)",
	)

	fs.DurationVar(
		&c.timeout,
		"timeout",
		30*time.Minute,
		"maximum time to wait for replay to complete",
	)

	fs.BoolVar(
		&c.skipSigVer,
		"skip-sig-verification",
		false,
		"skip genesis transaction signature verification",
	)

	fs.BoolVar(
		&c.skipFail,
		"skip-failing-txs",
		false,
		"skip genesis transactions that fail instead of panicking",
	)
}

func execReplay(ctx context.Context, cfg *replayCfg, io commands.IO) error {
	if cfg.dataDir == "" {
		return errors.New("--data-dir is required")
	}

	// Resolve paths
	dataDir, err := filepath.Abs(cfg.dataDir)
	if err != nil {
		return fmt.Errorf("invalid data-dir: %w", err)
	}

	genesisPath := cfg.genesis
	if genesisPath == "" {
		genesisPath = filepath.Join(dataDir, "config", "genesis.json")
	}

	if _, err := os.Stat(genesisPath); os.IsNotExist(err) {
		return fmt.Errorf("genesis.json not found at %s", genesisPath)
	}

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return fmt.Errorf("data directory not found at %s", dataDir)
	}

	io.Printfln("=== Chain Upgrade Replay Smoke Test ===")
	io.Printfln("Data directory: %s", dataDir)
	io.Printfln("Genesis:        %s", genesisPath)
	io.Printfln("Timeout:        %s", cfg.timeout)
	if cfg.haltHeight > 0 {
		io.Printfln("Halt height:    %d", cfg.haltHeight)
	}
	io.Printfln("")

	// Set up logging
	zapLogger, err := gnoLog.InitializeZapLogger(os.Stderr, "debug", "console")
	if err != nil {
		return fmt.Errorf("unable to initialize logger: %w", err)
	}
	defer zapLogger.Sync()
	logger := gnoLog.ZapLoggerToSlog(zapLogger)

	// Load node configuration
	nodeCfg, err := config.LoadConfig(dataDir)
	if err != nil {
		return fmt.Errorf("unable to load config from %s: %w", dataDir, err)
	}

	// Create event switch
	evsw := events.NewEventSwitch()

	// Set up the genesis app config
	genesisCfg := gnoland.GenesisAppConfig{
		SkipFailingTxs:      cfg.skipFail,
		SkipSigVerification: cfg.skipSigVer,
	}

	// Create the application
	gnoApp, err := gnoland.NewApp(
		dataDir,
		genesisCfg,
		nodeCfg.Application,
		evsw,
		logger,
	)
	if err != nil {
		return fmt.Errorf("unable to create gnoland app: %w", err)
	}

	nodeCfg.LocalApp = gnoApp

	// Create the node
	// TODO: add node.WithHaltHeight(cfg.haltHeight) once gnolang/gno#5334 is merged
	gnoNode, err := node.DefaultNewNode(nodeCfg, genesisPath, evsw, logger)
	if err != nil {
		return fmt.Errorf("unable to create node: %w", err)
	}

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(ctx, cfg.timeout)
	defer cancel()

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start the node
	io.Printfln("Starting replay...")
	startTime := time.Now()

	if err := gnoNode.Start(); err != nil {
		return fmt.Errorf("node failed to start (replay failed): %w", err)
	}

	// Wait for the node to be ready (replay complete) or failure
	select {
	case <-gnoNode.Ready():
		elapsed := time.Since(startTime)
		io.Printfln("")
		io.Printfln("=== REPLAY SUCCEEDED ===")
		io.Printfln("Duration: %s", elapsed.Round(time.Millisecond))
		printNodeInfo(io, gnoNode, logger)
	case <-ctx.Done():
		io.Printfln("")
		io.Printfln("=== REPLAY TIMED OUT ===")
		_ = gnoNode.Stop()
		return fmt.Errorf("replay timed out after %s", cfg.timeout)
	case sig := <-sigCh:
		io.Printfln("")
		io.Printfln("=== REPLAY INTERRUPTED (%s) ===", sig)
		_ = gnoNode.Stop()
		return fmt.Errorf("interrupted by signal %s", sig)
	}

	// Clean shutdown
	_ = gnoNode.Stop()

	io.Printfln("")
	io.Printfln("Smoke test passed. The new binary successfully replayed the chain state.")
	return nil
}

func printNodeInfo(io commands.IO, n *node.Node, logger *slog.Logger) {
	// Try to get the latest block height from consensus state
	cs := n.ConsensusState()
	if cs != nil {
		rs := cs.GetRoundState()
		if rs != nil {
			io.Printfln("Latest height: %d", rs.Height)
		}
	}

	// Get validators
	_, vals := n.ConsensusState().GetValidators()
	io.Printfln("Validators:    %d", len(vals))

	// Check genesis doc for chain ID
	genesis := n.GenesisDoc()
	if genesis != nil {
		io.Printfln("Chain ID:      %s", genesis.ChainID)
	}
}

// getGnoRoot returns the gno root directory, checking GNOROOT env var first.
func getGnoRoot() string {
	if root := os.Getenv("GNOROOT"); root != "" {
		return root
	}
	return gnoenv.RootDir()
}

// loadGenesis loads and returns the genesis document from the given path.
func loadGenesis(path string) (*bft.GenesisDoc, error) {
	doc, err := bft.GenesisDocFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load genesis from %s: %w", path, err)
	}
	return doc, nil
}
