package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	bftypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type genesisCfg struct {
	source         string
	chainID        string
	originalChainID string
	haltHeight     int64
	output         string
	txsOutput      string
	overlayDir     string
	skipTxs        bool
	noVerify       bool
}

func newGenesisCmd(io commands.IO) *commands.Command {
	cfg := &genesisCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "genesis",
			ShortUsage: "genesis [flags]",
			ShortHelp:  "generate a hardfork genesis from a source chain",
			LongHelp: `Generates a hardfork genesis.json by extracting state from a source chain
and assembling it with the hardfork parameters (original_chain_id, initial_height).

The source chain provides the base genesis (balances, validators, auth state)
and the historical transaction history. Both are embedded in the new genesis
so the new chain can replay all historical activity starting from the halt height.

Examples:

  # From a running or recently-halted node via RPC:
  hardfork genesis --source http://rpc.gno.land:26657 --chain-id gnoland-1

  # From a local node data directory (offline, reads block store):
  hardfork genesis --source /var/lib/gnoland --chain-id gnoland-1

  # From a pre-exported tarball (genesis.json + txs.jsonl):
  hardfork genesis --source /tmp/gnoland1-export.tar.gz --chain-id gnoland-1

  # Preview only (skip tx export — fast summary of genesis structure):
  hardfork genesis --source http://rpc.gno.land:26657 --skip-txs`,
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execGenesis(ctx, cfg, io)
		},
	)
}

func (c *genesisCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.source, "source", "", "source: RPC URL, local data dir, or exported file (.json/.jsonl/.tar.gz)")
	fs.StringVar(&c.chainID, "chain-id", "gnoland-1", "new chain ID")
	fs.StringVar(&c.originalChainID, "original-chain-id", "", "source chain ID for signature verification (auto-detected from source genesis if empty)")
	fs.Int64Var(&c.haltHeight, "halt-height", 0, "block height at which source chain halted (auto-detected from source if 0)")
	fs.StringVar(&c.output, "output", "genesis.json", "output genesis file path")
	fs.StringVar(&c.txsOutput, "txs-output", "", "also write extracted txs to this .jsonl file (optional)")
	fs.StringVar(&c.overlayDir, "overlay-dir", "", "directory of overlay scripts to apply before tx replay (optional)")
	fs.BoolVar(&c.skipTxs, "skip-txs", false, "skip tx export (only copy genesis structure — useful for quick preview)")
	fs.BoolVar(&c.noVerify, "no-verify", false, "skip genesis verification after assembly")
}

func execGenesis(ctx context.Context, cfg *genesisCfg, io commands.IO) error {
	if cfg.source == "" {
		return errors.New("--source is required (RPC URL, local data dir, or exported file)")
	}

	src, err := openSource(cfg.source)
	if err != nil {
		return fmt.Errorf("opening source %q: %w", cfg.source, err)
	}
	defer src.Close()

	io.Printf("Source: %s (%s)\n", src.Description(), cfg.source)

	// -------------------------------------------------------------------------
	// Step 1: Fetch base genesis from source
	// -------------------------------------------------------------------------
	io.Println("Step 1/4: Fetching base genesis...")

	baseGenDoc, err := src.FetchGenesis(ctx)
	if err != nil {
		return fmt.Errorf("fetching genesis: %w", err)
	}

	sourceChainID := baseGenDoc.ChainID
	io.Printf("  Source chain ID: %s\n", sourceChainID)
	io.Printf("  Source genesis time: %s\n", baseGenDoc.GenesisTime)

	// Use auto-detected chain ID if not explicitly provided
	if cfg.originalChainID == "" {
		cfg.originalChainID = sourceChainID
		io.Printf("  Original chain ID (auto-detected): %s\n", cfg.originalChainID)
	}

	// Auto-detect halt height from source
	if cfg.haltHeight == 0 {
		h, err := src.LatestHeight(ctx)
		if err != nil {
			return fmt.Errorf("detecting halt height: %w", err)
		}
		cfg.haltHeight = h
		io.Printf("  Halt height (auto-detected): %d\n", cfg.haltHeight)
	} else {
		io.Printf("  Halt height: %d\n", cfg.haltHeight)
	}

	// -------------------------------------------------------------------------
	// Step 2: Fetch historical transactions
	// -------------------------------------------------------------------------
	var txs []gnoland.TxWithMetadata

	if !cfg.skipTxs {
		io.Printf("Step 2/4: Fetching historical transactions (height 1..%d)...\n", cfg.haltHeight)

		txs, err = src.FetchTxs(ctx, 1, cfg.haltHeight, io)
		if err != nil {
			return fmt.Errorf("fetching transactions: %w", err)
		}

		io.Printf("  Fetched %d successful transactions\n", len(txs))

		// Write txs to separate file if requested
		if cfg.txsOutput != "" {
			if err := writeTxsJSONL(cfg.txsOutput, txs); err != nil {
				return fmt.Errorf("writing txs output: %w", err)
			}
			io.Printf("  Txs written to: %s\n", cfg.txsOutput)
		}
	} else {
		io.Println("Step 2/4: Skipping tx export (--skip-txs)")
	}

	// -------------------------------------------------------------------------
	// Step 3: Assemble hardfork genesis
	// -------------------------------------------------------------------------
	io.Println("Step 3/4: Assembling hardfork genesis...")

	initialHeight := cfg.haltHeight + 1

	newGenDoc, appState, err := buildHardforkGenesis(baseGenDoc, txs, cfg.chainID, cfg.originalChainID, initialHeight)
	if err != nil {
		return fmt.Errorf("building hardfork genesis: %w", err)
	}

	// Apply overlay if provided
	if cfg.overlayDir != "" {
		io.Printf("  Applying overlay from: %s\n", cfg.overlayDir)
		if err := applyOverlay(cfg.overlayDir, cfg.output, io); err != nil {
			return fmt.Errorf("applying overlay: %w", err)
		}
	}

	// -------------------------------------------------------------------------
	// Step 4: Write and verify output
	// -------------------------------------------------------------------------
	io.Println("Step 4/4: Writing genesis...")

	if err := writeGenesis(cfg.output, newGenDoc, appState); err != nil {
		return fmt.Errorf("writing genesis: %w", err)
	}

	stat, _ := os.Stat(cfg.output)
	io.Printf("  Written: %s", cfg.output)
	if stat != nil {
		io.Printf(" (%.1f MB)", float64(stat.Size())/(1024*1024))
	}
	io.Println()

	if !cfg.noVerify {
		if err := verifyGenesisFile(cfg.output); err != nil {
			return fmt.Errorf("genesis verification failed: %w (use --no-verify to skip)", err)
		}
		io.Println("  Verification: OK")
	}

	// Summary
	io.Println()
	io.Println("=== Hardfork Genesis Summary ===")
	io.Printf("  New chain ID:       %s\n", cfg.chainID)
	io.Printf("  Original chain ID:  %s\n", cfg.originalChainID)
	io.Printf("  Initial height:     %d\n", initialHeight)
	io.Printf("  Halt height:        %d\n", cfg.haltHeight)
	io.Printf("  Genesis-mode txs:   %d (from source genesis, no metadata)\n", len(baseGenesisModeTxs(appState)))
	io.Printf("  Historical txs:     %d (with block_height metadata)\n", len(txs))
	io.Printf("  Total txs:          %d\n", len(appState.Txs))
	io.Printf("  Output:             %s\n", cfg.output)
	io.Println()
	io.Println("Next steps:")
	io.Printf("  1. Test locally (in-process replay):\n")
	io.Printf("     hardfork test --genesis %s\n", cfg.output)
	io.Printf("  2. Verify with other validators (share SHA-256):\n")
	io.Printf("     sha256: $(sha256sum %s | cut -d' ' -f1)\n", cfg.output)

	_ = appState // suppress unused warning (used in summary above)
	return nil
}

// buildHardforkGenesis constructs the new genesis document.
// It takes the source chain's genesis as the base, injects the hardfork
// parameters, and appends historical txs (with block_height metadata).
func buildHardforkGenesis(
	srcGenDoc *bftypes.GenesisDoc,
	historicalTxs []gnoland.TxWithMetadata,
	newChainID string,
	originalChainID string,
	initialHeight int64,
) (*bftypes.GenesisDoc, *gnoland.GnoGenesisState, error) {
	// Extract app state from source genesis
	appState, ok := srcGenDoc.AppState.(gnoland.GnoGenesisState)
	if !ok {
		// Try amino JSON round-trip if the app state is a raw json.RawMessage
		raw, err := amino.MarshalJSON(srcGenDoc.AppState)
		if err != nil {
			return nil, nil, fmt.Errorf("marshalling source app state: %w", err)
		}
		if err := amino.UnmarshalJSON(raw, &appState); err != nil {
			return nil, nil, fmt.Errorf("unmarshalling source app state as GnoGenesisState: %w", err)
		}
	}

	// Inject hardfork fields
	appState.PastChainIDs = []string{originalChainID}
	appState.InitialHeight = initialHeight

	// Append historical txs after existing genesis-mode txs
	// Genesis-mode txs (no metadata or BlockHeight==0): package deploys, setup
	// Historical txs (BlockHeight > 0): replayed with original chain ID context
	appState.Txs = append(appState.Txs, historicalTxs...)

	// Build the new genesis doc
	newGenDoc := *srcGenDoc // copy
	newGenDoc.ChainID = newChainID
	newGenDoc.InitialHeight = initialHeight
	newGenDoc.AppState = appState

	return &newGenDoc, &appState, nil
}

// writeGenesis serializes and writes the genesis to a file.
func writeGenesis(path string, genDoc *bftypes.GenesisDoc, _ *gnoland.GnoGenesisState) error {
	data, err := amino.MarshalJSONIndent(genDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling genesis: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// writeTxsJSONL writes transactions to a file, one amino JSON per line.
// Uses amino.MarshalJSON to preserve interface type information (e.g. std.Msg).
func writeTxsJSONL(path string, txs []gnoland.TxWithMetadata) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, tx := range txs {
		data, err := amino.MarshalJSON(tx)
		if err != nil {
			return err
		}
		data = append(data, '\n')
		if _, err := f.Write(data); err != nil {
			return err
		}
	}
	return nil
}

// verifyGenesisFile runs basic validation on the written genesis file.
func verifyGenesisFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var genDoc bftypes.GenesisDoc
	if err := amino.UnmarshalJSON(data, &genDoc); err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	return genDoc.ValidateAndComplete()
}

// baseGenesisModeTxs returns only the genesis-mode txs (BlockHeight == 0) from app state.
func baseGenesisModeTxs(appState *gnoland.GnoGenesisState) []gnoland.TxWithMetadata {
	var out []gnoland.TxWithMetadata
	for _, tx := range appState.Txs {
		if tx.Metadata == nil || tx.Metadata.BlockHeight == 0 {
			out = append(out, tx)
		}
	}
	return out
}

// applyOverlay runs overlay scripts from a directory against the genesis file.
// This is a simple wrapper around the shell scripts in the overlay directory.
func applyOverlay(overlayDir, genesisPath string, io commands.IO) error {
	entries, err := os.ReadDir(overlayDir)
	if err != nil {
		return fmt.Errorf("reading overlay dir: %w", err)
	}

	var scripts []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sh") {
			scripts = append(scripts, e.Name())
		}
	}

	if len(scripts) == 0 {
		io.Printf("  No overlay scripts found in %s\n", overlayDir)
		return nil
	}

	// Overlay script execution is not yet implemented in the Go binary.
	// Return an error so the caller knows the genesis is incomplete.
	// Use generate-genesis.sh for full overlay support.
	return fmt.Errorf("overlay not yet implemented: found %d scripts in %s; use generate-genesis.sh instead", len(scripts), overlayDir)
}
