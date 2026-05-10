package fork

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	bftypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// dirSource reads chain state from a local directory.
//
// Expected directory layouts (tries both):
//
//	/path/to/dir/
//	  config/genesis.json   ← gnoland default node layout
//	  genesis.json          ← flat layout (e.g. manual export)
//	  txs.jsonl             ← optional: pre-exported txs with metadata
//
// If txs.jsonl is present it is used directly (no block store access needed).
// If txs.jsonl is absent, FetchTxs returns an empty slice with a warning —
// reading directly from the block store will be added in a future version.
type dirSource struct {
	dir         string
	genesisPath string // resolved path to genesis.json
	txsPath     string // resolved path to txs.jsonl (empty if not found)
}

func newDirSource(dir string) (*dirSource, error) {
	s := &dirSource{dir: dir}

	// Find genesis.json
	candidates := []string{
		filepath.Join(dir, "config", "genesis.json"),
		filepath.Join(dir, "genesis.json"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			s.genesisPath = c
			break
		}
	}
	if s.genesisPath == "" {
		return nil, fmt.Errorf("genesis.json not found in %s (tried config/genesis.json and genesis.json)", dir)
	}

	// Find txs.jsonl (optional)
	txsCandidates := []string{
		filepath.Join(dir, "txs.jsonl"),
		filepath.Join(dir, "historical-txs.jsonl"),
	}
	for _, c := range txsCandidates {
		if _, err := os.Stat(c); err == nil {
			s.txsPath = c
			break
		}
	}

	return s, nil
}

func (s *dirSource) Description() string { return "local directory" }
func (s *dirSource) Close() error        { return nil }

func (s *dirSource) FetchGenesis(ctx context.Context) (*bftypes.GenesisDoc, error) {
	data, err := os.ReadFile(s.genesisPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", s.genesisPath, err)
	}

	var genDoc bftypes.GenesisDoc
	if err := amino.UnmarshalJSON(data, &genDoc); err != nil {
		return nil, fmt.Errorf("parsing genesis: %w", err)
	}

	return &genDoc, nil
}

// LatestHeight returns the halt height from the genesis InitialHeight if set,
// otherwise falls back to -1 (user must specify --halt-height).
//
// For a proper auto-detect from a local node directory, reading the block store
// would be needed. That is tracked as a future enhancement.
func (s *dirSource) LatestHeight(_ context.Context) (int64, error) {
	data, err := os.ReadFile(s.genesisPath)
	if err != nil {
		return 0, fmt.Errorf("reading genesis: %w", err)
	}

	// Try to extract a height hint from the genesis file itself
	var raw struct {
		InitialHeight int64 `json:"initial_height"`
	}
	_ = amino.UnmarshalJSON(data, &raw)
	if raw.InitialHeight > 1 {
		// This is already a hardfork genesis — use InitialHeight-1 as the halt height
		return raw.InitialHeight - 1, nil
	}

	return 0, fmt.Errorf(
		"cannot auto-detect halt height from local directory %s; "+
			"please specify --halt-height explicitly, or point --source to a running node RPC",
		s.dir,
	)
}

// FetchTxs reads transactions from txs.jsonl if present.
// If no txs file is found, returns an empty slice with a warning.
// Full block-store reading will be added in a future version.
func (s *dirSource) FetchTxs(_ context.Context, fromHeight, toHeight int64, io commands.IO) ([]gnoland.TxWithMetadata, error) {
	if s.txsPath == "" {
		io.Println("  WARNING: no txs.jsonl found in source directory.")
		io.Println("  Historical tx replay will be empty — only genesis-mode txs will be included.")
		io.Println("  To include historical txs, provide a txs.jsonl file alongside genesis.json,")
		io.Println("  or use --source with an RPC URL instead.")
		return nil, nil
	}

	io.Printf("  Reading txs from: %s\n", s.txsPath)

	f, err := os.Open(s.txsPath)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", s.txsPath, err)
	}
	defer f.Close()

	var txs []gnoland.TxWithMetadata
	scanner := bufio.NewScanner(f)
	// Increase buffer for large tx lines.
	scanner.Buffer(make([]byte, 0, 4096), 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var tx gnoland.TxWithMetadata
		if err := amino.UnmarshalJSON(line, &tx); err != nil {
			return nil, fmt.Errorf("decoding tx: %w", err)
		}

		// Filter to requested height range
		if tx.Metadata != nil && tx.Metadata.BlockHeight > 0 {
			if tx.Metadata.BlockHeight < fromHeight || tx.Metadata.BlockHeight > toHeight {
				continue
			}
		}

		txs = append(txs, tx)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading txs: %w", err)
	}

	return txs, nil
}

// fileSource handles a single genesis.json file (no txs).
type fileSource struct {
	path string
}

func newFileSource(path string) (*fileSource, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("file %q: %w", path, err)
	}
	return &fileSource{path: path}, nil
}

func (s *fileSource) Description() string { return "genesis file" }
func (s *fileSource) Close() error        { return nil }

func (s *fileSource) FetchGenesis(ctx context.Context) (*bftypes.GenesisDoc, error) {
	d := &dirSource{dir: filepath.Dir(s.path), genesisPath: s.path}
	return d.FetchGenesis(ctx)
}

func (s *fileSource) LatestHeight(ctx context.Context) (int64, error) {
	d := &dirSource{dir: filepath.Dir(s.path), genesisPath: s.path}
	return d.LatestHeight(ctx)
}

func (s *fileSource) FetchTxs(_ context.Context, _, _ int64, io commands.IO) ([]gnoland.TxWithMetadata, error) {
	io.Println("  WARNING: single genesis.json source — no historical txs available.")
	io.Println("  Use --source with a directory (containing txs.jsonl) or an RPC URL.")
	return nil, nil
}
