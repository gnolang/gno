package abci

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRequestInitChain_InitialHeight_RoundTrip is a regression test that
// asserts amino preserves RequestInitChain.InitialHeight on the wire. A
// silent registration regression (wrong field tag, dropped field, name
// rename without rebuild) would otherwise only surface during an actual
// hardfork attempt, where the chain would boot at height 1 instead of
// the operator-supplied InitialHeight.
func TestRequestInitChain_InitialHeight_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		height int64
	}{
		{"zero (standard genesis)", 0},
		{"one (standard genesis)", 1},
		{"hardfork", 1234567},
		{"large hardfork height", 1_000_000_000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			original := RequestInitChain{
				ChainID:       "test-chain",
				InitialHeight: tc.height,
			}
			bz := amino.MustMarshal(original)
			var decoded RequestInitChain
			require.NoError(t, amino.Unmarshal(bz, &decoded))
			assert.Equal(t, tc.height, decoded.InitialHeight,
				"InitialHeight should round-trip; got %d, want %d",
				decoded.InitialHeight, tc.height)
		})
	}
}
