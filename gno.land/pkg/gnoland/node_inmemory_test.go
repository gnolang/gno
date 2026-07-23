package gnoland

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNewDefaultGenesisConfig_AppStateLoads pins the AppState type the helper
// emits. loadAppState type-switches on GnoGenesisState and *GenesisStateRef, so
// handing out a *GnoGenesisState makes InitChainer reject the genesis, which
// the handshake turns into a boot failure.
func TestNewDefaultGenesisConfig_AppStateLoads(t *testing.T) {
	t.Parallel()

	genesis := NewDefaultGenesisConfig("test-chain", "gno.land")

	_, ok := genesis.AppState.(GnoGenesisState)
	require.True(t, ok, "AppState must be a GnoGenesisState value, got %T", genesis.AppState)
}
