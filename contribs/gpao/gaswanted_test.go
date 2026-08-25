package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
)

// TestGasWantedFor covers the sizing of an enable's GasWanted.
//
// The number matters in both directions. Too low and the approval fails on
// exactly the packages whose cost the oracle could not predict -- which is the
// state the shipped default was in, at 20,000,000 against a worst case near
// 40,000,000 for a 1 MB parked blob plus the namespace and CLA realm calls. Too
// high and the transaction is refused at CheckTx, because the ante caps
// GasWanted at the block limit.
func TestGasWantedFor(t *testing.T) {
	const (
		fallback = int64(20_000_000)
		ceiling  = int64(3_000_000_000)
	)

	for _, tc := range []struct {
		name      string
		estimated int64
		want      int64
	}{
		{
			name:      "a measured estimate gets 20% headroom",
			estimated: 10_000_000,
			want:      12_000_000,
		},
		{
			name: "headroom lifts a realistic worst case above the old default",
			// A 1 MB parked blob plus the two realm calls. The point of the
			// estimate is that this no longer has to be guessed at.
			estimated: 40_000_000,
			want:      48_000_000,
		},
		{
			name:      "no estimate falls back to the configured value",
			estimated: 0,
			want:      fallback,
		},
		{
			name:      "a negative estimate falls back too, rather than asking for negative gas",
			estimated: -1,
			want:      fallback,
		},
		{
			name: "headroom is capped at the block ceiling",
			// Just under the ceiling: +20% would clear it, so the cap binds.
			estimated: ceiling - 1,
			want:      ceiling,
		},
		{
			name:      "an estimate at the ceiling stays at the ceiling",
			estimated: ceiling,
			want:      ceiling,
		},
		{
			name: "an absurd estimate is capped, not overflowed",
			// Guarded before the multiply: 2^62 * 12 wraps negative, and a
			// non-positive GasWanted is refused by the ante -- so an overflow
			// here would read as "the oracle asked for negative gas".
			estimated: 1 << 62,
			want:      ceiling,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := gasWantedFor(tc.estimated, fallback, ceiling)
			assert.Equal(t, tc.want, got)
			assert.Positive(t, got, "GasWanted must always be positive")
			assert.LessOrEqual(t, got, ceiling,
				"and must never exceed the ceiling the ante enforces")
		})
	}

	t.Run("the ceiling is the chain's, not a constant", func(t *testing.T) {
		// The ante REFUSES a gas-wanted above Block.MaxGas rather than clamping
		// it, so a chain configured below the default would reject everything
		// sized against the default. This is why the ceiling is queried.
		const small = int64(500_000)
		assert.Equal(t, small, gasWantedFor(10_000_000, fallback, small),
			"an estimate above the chain's own limit must be cut to it")
		assert.Equal(t, int64(120_000), gasWantedFor(100_000, fallback, small),
			"and one below it still gets its headroom")
	})

	t.Run("a fallback above the chain's ceiling is cut to it", func(t *testing.T) {
		// The fallback is -gas-wanted, which is operator input and is not
		// checked against the chain. It is also what a failed estimate degrades
		// to -- so leaving it unbounded would refuse every enable on exactly
		// the chains the fallback exists to keep working on.
		const small = int64(500_000)
		assert.Greater(t, fallback, small, "premise: the fallback must exceed the ceiling")
		assert.Equal(t, small, gasWantedFor(0, fallback, small))
	})
}

// TestBlockMaxGasFrom covers how the gas ceiling is chosen from what the node
// reports.
//
// The ceiling has to be the chain's own Block.MaxGas, because the ante refuses
// a transaction above it rather than clamping. Everything unusable falls back to
// the tm2 default rather than to zero or to no bound: zero would make every
// approval fail instantly, and no bound would let one absurd estimate ask for
// unbounded gas.
func TestBlockMaxGasFrom(t *testing.T) {
	withMaxGas := func(v int64) *ctypes.ResultConsensusParams {
		return &ctypes.ResultConsensusParams{
			ConsensusParams: abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: v}},
		}
	}

	t.Run("the chain's own value is used", func(t *testing.T) {
		assert.Equal(t, int64(500_000), blockMaxGasFrom(withMaxGas(500_000), nil))
	})

	t.Run("a query error falls back", func(t *testing.T) {
		assert.Equal(t, defaultBlockMaxGas,
			blockMaxGasFrom(withMaxGas(500_000), errors.New("node unreachable")),
			"a reported value must not be trusted when the query itself failed")
	})

	t.Run("no bound falls back rather than being taken literally", func(t *testing.T) {
		// -1 is legal and means "no limit". Honouring it would remove the only
		// thing stopping an absurd estimate.
		assert.Equal(t, defaultBlockMaxGas, blockMaxGasFrom(withMaxGas(-1), nil))
		assert.Equal(t, defaultBlockMaxGas, blockMaxGasFrom(withMaxGas(0), nil))
	})

	t.Run("a malformed response falls back", func(t *testing.T) {
		assert.Equal(t, defaultBlockMaxGas, blockMaxGasFrom(nil, nil))
		assert.Equal(t, defaultBlockMaxGas,
			blockMaxGasFrom(&ctypes.ResultConsensusParams{}, nil),
			"a response with no Block section must not nil-deref")
	})
}
