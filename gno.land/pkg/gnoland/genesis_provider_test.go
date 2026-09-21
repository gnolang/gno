package gnoland

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestStreamingGenesisProvider verifies that the helper used by gnoland's
// node entry point returns a GenesisDocProvider that, when invoked, produces
// a doc whose AppState is the streaming on-disk-backed handle (not the
// in-memory GnoGenesisState). This is what makes the type-switch in
// InitChainerConfig.loadAppState route to the streaming path on real
// gnoland start.
func TestStreamingGenesisProvider(t *testing.T) {
	src := copySlimFixture(t)
	cacheRoot := t.TempDir()

	provider := StreamingGenesisProvider(src, cacheRoot, nil)
	require.NotNil(t, provider)

	doc, err := provider()
	require.NoError(t, err)
	require.NotNil(t, doc)

	ref, ok := doc.AppState.(*GenesisStateRef)
	require.True(t, ok, "AppState must be *GenesisStateRef, got %T", doc.AppState)
	require.Equal(t, slimFixtureBalanceCount, ref.BalanceCount())
	require.Equal(t, slimFixtureTxCount, ref.TxCount())
}
