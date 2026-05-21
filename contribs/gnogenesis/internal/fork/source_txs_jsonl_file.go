package fork

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// jsonlFileTxsSource reads pre-exported transactions from a single .jsonl file.
// Each line is an amino-JSON gnoland.TxWithMetadata; SignerInfo, BlockHeight,
// ChainID and Failed are already populated by whoever produced the file —
// no sequence brute-forcing is needed here.
type jsonlFileTxsSource struct {
	path string
}

func newJSONLFileTxsSource(path string) (*jsonlFileTxsSource, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("txs jsonl file %q: %w", path, err)
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("--source-txs-jsonl-file expects a file, got directory %q", path)
	}
	return &jsonlFileTxsSource{path: path}, nil
}

func (s *jsonlFileTxsSource) Description() string { return "JSONL txs file" }
func (s *jsonlFileTxsSource) Close() error        { return nil }

// LatestHeight scans the file once to find the maximum block_height.
// For multi-hundred-MB archives this is acceptable: the caller uses it
// only when --halt-height is not specified, and the full FetchTxs that
// follows reads the file anyway.
func (s *jsonlFileTxsSource) LatestHeight(_ context.Context) (int64, error) {
	f, err := os.Open(s.path)
	if err != nil {
		return 0, fmt.Errorf("opening %s: %w", s.path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 4096), 10*1024*1024)
	var maxHeight int64
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var tx gnoland.TxWithMetadata
		if err := amino.UnmarshalJSON(line, &tx); err != nil {
			return 0, fmt.Errorf("decoding tx: %w", err)
		}
		if tx.Metadata != nil && tx.Metadata.BlockHeight > maxHeight {
			maxHeight = tx.Metadata.BlockHeight
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("reading txs: %w", err)
	}
	if maxHeight == 0 {
		return 0, fmt.Errorf("no block_height metadata found in %s; specify --halt-height explicitly", s.path)
	}
	return maxHeight, nil
}

// FetchTxs reads every line and filters by [fromHeight, toHeight]. The
// chainID parameter is ignored — each line carries its own ChainID in the
// metadata, produced by an earlier export run.
func (s *jsonlFileTxsSource) FetchTxs(_ context.Context, _ string, fromHeight, toHeight int64, io commands.IO) ([]gnoland.TxWithMetadata, error) {
	io.Printf("  Reading txs from: %s\n", s.path)

	f, err := os.Open(s.path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", s.path, err)
	}
	defer f.Close()

	var txs []gnoland.TxWithMetadata
	scanner := bufio.NewScanner(f)
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
