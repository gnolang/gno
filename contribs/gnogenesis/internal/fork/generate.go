package fork

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	bftypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type generateCfg struct {
	source          string
	chainID         string
	originalChainID string
	haltHeight      int64
	output          string
	txsOutput       string
	patchRealms     patchRealmList
	migrationTxs    stringList
	skipTxs         bool
	noVerify        bool
}

// patchRealmList accepts repeated --patch-realm flags. Each value is
// "pkgpath=srcdir"; the tool rewrites the matching genesis-mode addpkg
// tx's Package.Files with the contents of srcdir.
type patchRealmList []string

func (p *patchRealmList) String() string { return strings.Join(*p, ",") }
func (p *patchRealmList) Set(v string) error {
	*p = append(*p, v)
	return nil
}

// stringList accepts repeated string flags.
type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func newGenerateCmd(io commands.IO) *commands.Command {
	cfg := &generateCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "generate",
			ShortUsage: "generate [flags]",
			ShortHelp:  "assemble a hardfork genesis from a source chain",
			LongHelp: `Generates a hardfork genesis.json by extracting state from a source chain
and assembling it with the hardfork parameters (original_chain_id, initial_height).

The source chain provides the base genesis (balances, validators, auth state)
and the historical transaction history. Both are embedded in the new genesis
so the new chain can replay all historical activity starting from the halt height.

Examples:

  # From a running or recently-halted node via RPC:
  gnogenesis fork generate --source http://rpc.gno.land:26657 --chain-id gnoland-1

  # From a local node data directory (offline, reads block store):
  gnogenesis fork generate --source /var/lib/gnoland --chain-id gnoland-1

  # From a pre-exported tarball (genesis.json + txs.jsonl):
  gnogenesis fork generate --source /tmp/gnoland1-export.tar.gz --chain-id gnoland-1

  # Preview only (skip tx export — fast summary of genesis structure):
  gnogenesis fork generate --source http://rpc.gno.land:26657 --skip-txs`,
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execGenerate(ctx, cfg, io)
		},
	)
}

func (c *generateCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.source, "source", "", "source: RPC URL, local data dir, or exported file (.json/.jsonl/.tar.gz)")
	fs.StringVar(&c.chainID, "chain-id", "gnoland-1", "new chain ID")
	fs.StringVar(&c.originalChainID, "original-chain-id", "", "source chain ID for signature verification (auto-detected from source genesis if empty)")
	fs.Int64Var(&c.haltHeight, "halt-height", 0, "block height at which source chain halted (auto-detected from source if 0)")
	fs.StringVar(&c.output, "output", "genesis.json", "output genesis file path")
	fs.StringVar(&c.txsOutput, "txs-output", "", "also write extracted txs to this .jsonl file (optional)")
	fs.Var(&c.migrationTxs, "migration-tx", "append migration txs at the END of appState.Txs "+
		"(after historical replay). Repeatable. FILE is a .jsonl where each "+
		"line is an amino-JSON gnoland.TxWithMetadata. These are genesis-mode "+
		"txs (BlockHeight==0) that run through the same --skip-genesis-sig-"+
		"verification code path as original genesis-mode txs, but are placed "+
		"after the historical stream so they can mutate replayed state "+
		"(e.g. govDAO prop to update r/sys/validators/v2 to the new valset).")
	fs.Var(&c.patchRealms, "patch-realm", "patch a genesis-mode addpkg tx in place: repeatable, PKGPATH=SRCDIR. "+
		"Replaces Package.Files with the *.gno + gnomod.toml files found in SRCDIR. "+
		"Source genesis on disk is NOT modified; the patch is applied in memory "+
		"before writing the hardfork genesis. Use this to deliver realm upgrades "+
		"as part of the fork (e.g. adding a new .gno file to an existing realm).")
	fs.BoolVar(&c.skipTxs, "skip-txs", false, "skip tx export (only copy genesis structure — useful for quick preview)")
	fs.BoolVar(&c.noVerify, "no-verify", false, "skip genesis verification after assembly")
}

func execGenerate(ctx context.Context, cfg *generateCfg, io commands.IO) error {
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

	// Apply --patch-realm rewrites on genesis-mode addpkg txs (in-memory only).
	for _, spec := range cfg.patchRealms {
		parts := strings.SplitN(spec, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("--patch-realm needs PKGPATH=SRCDIR, got %q", spec)
		}
		pkgPath, srcDir := parts[0], parts[1]
		n, err := patchGenesisModeAddPkg(appState, pkgPath, srcDir)
		if err != nil {
			return fmt.Errorf("patch %s: %w", pkgPath, err)
		}
		if n == 0 {
			io.Printf("  WARNING: --patch-realm %s did not match any genesis-mode addpkg tx\n", pkgPath)
		} else {
			io.Printf("  patched %s from %s (%d tx rewritten)\n", pkgPath, srcDir, n)
		}
	}
	// Append --migration-tx files at the END of appState.Txs (post-history).
	// Each file is a .jsonl of gnoland.TxWithMetadata. We force BlockHeight=0
	// so they go through the genesis-mode path (chain-id via PastChainIDs[0],
	// sig verify skipped under --skip-genesis-sig-verification).
	for _, path := range cfg.migrationTxs {
		migTxs, err := readMigrationTxs(path)
		if err != nil {
			return fmt.Errorf("migration-tx %s: %w", path, err)
		}
		appState.Txs = append(appState.Txs, migTxs...)
		io.Printf("  appended %d migration tx(s) from %s\n", len(migTxs), path)
	}
	newGenDoc.AppState = *appState

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

	// Default gas_replay_mode to "source" so historical txs don't get
	// re-gassed under the new VM's gas meter (which will reject most of
	// them as "insufficient funds error" if the gas model has changed
	// between source chain and hardfork). Operators who actively want
	// strict re-gassing for validation can set it explicitly post-hoc.
	if appState.GasReplayMode == "" {
		appState.GasReplayMode = "source"
	}

	// Source chains generated before the gas-storage refactor (PR #5415)
	// have no min_*/fixed_*_depth_100 or iter_next_cost_flat fields in
	// vm.params. When deserialized into the post-refactor Params struct
	// these default to 0, which fails Validate() (iter_next_cost_flat must
	// be > 0). Populate from code defaults when every field is unset, so
	// the resulting genesis boots on a post-refactor node without manual
	// patching. Do not overwrite if any value is already set — an operator
	// may have intentionally tuned these.
	if appState.VM.Params.IterNextCostFlat == 0 &&
		appState.VM.Params.MinGetReadDepth100 == 0 &&
		appState.VM.Params.MinSetReadDepth100 == 0 &&
		appState.VM.Params.MinWriteDepth100 == 0 &&
		appState.VM.Params.FixedGetReadDepth100 == 0 &&
		appState.VM.Params.FixedSetReadDepth100 == 0 &&
		appState.VM.Params.FixedWriteDepth100 == 0 {
		defaults := vm.DefaultParams()
		appState.VM.Params.MinGetReadDepth100 = defaults.MinGetReadDepth100
		appState.VM.Params.MinSetReadDepth100 = defaults.MinSetReadDepth100
		appState.VM.Params.MinWriteDepth100 = defaults.MinWriteDepth100
		appState.VM.Params.FixedGetReadDepth100 = defaults.FixedGetReadDepth100
		appState.VM.Params.FixedSetReadDepth100 = defaults.FixedSetReadDepth100
		appState.VM.Params.FixedWriteDepth100 = defaults.FixedWriteDepth100
		appState.VM.Params.IterNextCostFlat = defaults.IterNextCostFlat
	}

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

// readMigrationTxs reads a .jsonl file of gnoland.TxWithMetadata entries.
// BlockHeight is forced to 0 so each line is treated as a genesis-mode tx
// when replayed (uses PastChainIDs[0] for chain-id; sig verify skipped under
// --skip-genesis-sig-verification). Blank lines and # comments are ignored.
func readMigrationTxs(path string) ([]gnoland.TxWithMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	out := make([]gnoland.TxWithMetadata, 0, len(lines))
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		var tx gnoland.TxWithMetadata
		if err := amino.UnmarshalJSON([]byte(line), &tx); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		if tx.Metadata == nil {
			tx.Metadata = &gnoland.GnoTxMetadata{}
		}
		tx.Metadata.BlockHeight = 0 // always genesis-mode
		out = append(out, tx)
	}
	return out, nil
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

// patchGenesisModeAddPkg rewrites every genesis-mode addpkg tx whose package
// path matches `pkgPath` in-place — replacing its Package.Files slice with
// the *.gno + gnomod.toml files read from `srcDir`.
//
// This is how realm upgrades ride along in a hardfork: instead of adding a
// new tx (which would run with a different caller + account state and may
// collide with existing state), we rewrite the tx that originally deployed
// the realm so the forked chain initialises it with the new source.
//
// The source genesis on disk is NOT touched — this operates on the in-memory
// GnoGenesisState that we assembled for the output.
//
// Returns the number of txs rewritten.
func patchGenesisModeAddPkg(appState *gnoland.GnoGenesisState, pkgPath, srcDir string) (int, error) {
	files, err := loadGnoPackageFiles(srcDir)
	if err != nil {
		return 0, fmt.Errorf("load %s: %w", srcDir, err)
	}
	if len(files) == 0 {
		return 0, fmt.Errorf("no .gno/.toml files in %s", srcDir)
	}

	patched := 0
	for i := range appState.Txs {
		txm := &appState.Txs[i]
		if txm.Metadata != nil && txm.Metadata.BlockHeight > 0 {
			continue // historical tx, leave alone
		}
		for mi, msg := range txm.Tx.Msgs {
			addpkg, ok := msg.(vm.MsgAddPackage)
			if !ok {
				continue
			}
			if addpkg.Package == nil || addpkg.Package.Path != pkgPath {
				continue
			}
			addpkg.Package.Files = files
			// Refresh package name in case a .gno's `package ...` declaration
			// matters downstream.
			for _, f := range files {
				if strings.HasSuffix(f.Name, ".gno") {
					if name := gnoPackageNameFromFileBody(f.Name, f.Body); name != "" {
						addpkg.Package.Name = name
					}
					break
				}
			}
			txm.Tx.Msgs[mi] = addpkg
			patched++
		}
	}
	return patched, nil
}

// loadGnoPackageFiles reads *.gno and gnomod.toml files from srcDir
// (non-recursive) and returns them as ordered std.MemFile entries.
// Skips _test.gno, _filetest.gno, and hidden files.
func loadGnoPackageFiles(srcDir string) ([]*std.MemFile, error) {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		n := e.Name()
		if strings.HasSuffix(n, "_test.gno") || strings.HasSuffix(n, "_filetest.gno") {
			continue
		}
		if strings.HasSuffix(n, ".gno") || n == "gnomod.toml" {
			names = append(names, n)
		}
	}
	sort.Strings(names)

	files := make([]*std.MemFile, 0, len(names))
	for _, n := range names {
		body, err := os.ReadFile(filepath.Join(srcDir, n))
		if err != nil {
			return nil, err
		}
		files = append(files, &std.MemFile{Name: n, Body: string(body)})
	}
	return files, nil
}

// gnoPackageNameFromFileBody extracts `package NAME` from the top of a .gno
// file. Returns "" if not found. (Intentionally lightweight — avoids pulling
// in the gnovm parser.)
func gnoPackageNameFromFileBody(_ string, body string) string {
	for _, line := range strings.Split(body, "\n") {
		l := strings.TrimSpace(line)
		if strings.HasPrefix(l, "package ") {
			rest := strings.TrimPrefix(l, "package ")
			if i := strings.IndexAny(rest, " \t/"); i >= 0 {
				rest = rest[:i]
			}
			return rest
		}
	}
	return ""
}
