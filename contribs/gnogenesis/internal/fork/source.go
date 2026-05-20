package fork

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	bftypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// Source is a provider of chain state for hardfork genesis assembly.
type Source interface {
	// Description returns a human-readable source type label.
	Description() string

	// FetchGenesis returns the source chain's genesis document.
	FetchGenesis(ctx context.Context) (*bftypes.GenesisDoc, error)

	// LatestHeight returns the latest committed block height.
	// Used to auto-detect halt height when --halt-height is not specified.
	LatestHeight(ctx context.Context) (int64, error)

	// FetchTxs fetches all successful transactions in [fromHeight, toHeight]
	// with metadata (BlockHeight, Timestamp, ChainID populated).
	// Progress is reported via io.
	FetchTxs(ctx context.Context, fromHeight, toHeight int64, io commands.IO) ([]gnoland.TxWithMetadata, error)

	// Close releases any resources held by the source.
	Close() error
}

// openSource auto-detects the source type from the provided string and
// returns the appropriate Source implementation. rpcWorkersPerEndpoint is
// only consulted when the resolved source is an RPC pool; pass 0 to fall
// back to defaultWorkersPerEndpoint.
//
// Detection order:
//  1. comma-separated http(s) URLs → multi-endpoint RPC source
//  2. http:// or https:// prefix   → single-endpoint RPC source
//  3. directory path that exists   → local directory source
//  4. file ending in .json          → single genesis file source
//  5. file ending in .tar.gz/.tgz   → tarball source (future)
func openSource(s string, rpcWorkersPerEndpoint int) (Source, error) {
	// RPC source: any comma-bearing input, or anything starting with an
	// http(s) scheme (after trimming surrounding whitespace), is routed
	// through openRPCSource which parses one or more URLs.
	trimmed := strings.TrimSpace(s)
	if strings.Contains(s, ",") || strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return openRPCSource(s, rpcWorkersPerEndpoint)
	}

	// Local path
	fi, err := os.Stat(s)
	if err != nil {
		return nil, fmt.Errorf("source path %q: %w", s, err)
	}

	if fi.IsDir() {
		return newDirSource(s)
	}

	// Single genesis file
	if strings.HasSuffix(s, ".json") {
		return newFileSource(s)
	}

	// Tarball (not yet implemented)
	if strings.HasSuffix(s, ".tar.gz") || strings.HasSuffix(s, ".tgz") {
		return nil, fmt.Errorf("tarball source not yet implemented; extract first and use --source /path/to/dir")
	}

	return nil, fmt.Errorf("unrecognised source %q: expected http(s) URL, directory, .json file, or .tar.gz", s)
}

// openRPCSource parses s as one or more http(s) URLs separated by commas and
// returns an rpcSource whose client pool spans every URL. Whitespace is
// trimmed; empty segments are skipped; any non-http(s) segment is an error.
func openRPCSource(s string, workersPerEndpoint int) (Source, error) {
	parts := strings.Split(s, ",")
	urls := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "http://") && !strings.HasPrefix(p, "https://") {
			return nil, fmt.Errorf("--source segment %q must start with http:// or https://", p)
		}
		if _, err := url.Parse(p); err != nil {
			return nil, fmt.Errorf("invalid RPC URL %q: %w", p, err)
		}
		urls = append(urls, p)
	}
	if len(urls) == 0 {
		return nil, fmt.Errorf("--source contained no usable URLs")
	}
	return newRPCSource(workersPerEndpoint, urls...)
}
