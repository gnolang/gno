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
// returns the appropriate Source implementation.
//
// Detection order:
//  1. http:// or https:// prefix → RPC source
//  2. directory path that exists → local directory source
//  3. file ending in .json        → single genesis file source
//  4. file ending in .tar.gz/.tgz → tarball source (future)
func openSource(s string) (Source, error) {
	// RPC source
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		u, err := url.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("invalid RPC URL: %w", err)
		}
		return newRPCSource(u.String())
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
