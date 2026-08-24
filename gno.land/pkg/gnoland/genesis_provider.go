package gnoland

import (
	"log/slog"

	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

// StreamingGenesisProvider returns a node.GenesisDocProvider that loads the
// genesis document from disk via LoadStreamingGenesisDoc, attaching a
// *GenesisStateRef as AppState rather than the typed in-memory
// GnoGenesisState. Use this in place of node.DefaultGenesisDocProviderFunc
// when starting a gnoland node from a real genesis file — the streaming
// loader keeps peak memory bounded regardless of source-file size.
//
// logger is forwarded to LoadStreamingGenesisDoc so unknown-field warnings
// surface in the node's normal logging path. A nil logger falls back to
// slog.Default.
//
// The provider is closure-only — each call re-runs the loader, which is
// cheap on a warm cache (hash check + open files) and does the full
// preprocessing pass on a cold cache.
func StreamingGenesisProvider(genesisFile, cacheRoot string, logger *slog.Logger) node.GenesisDocProvider {
	return func() (*types.GenesisDoc, error) {
		return LoadStreamingGenesisDoc(genesisFile, cacheRoot, logger)
	}
}
