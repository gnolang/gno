package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
}
