package types

import (
	"math"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/stretchr/testify/require"
)

func TestGasMeter(t *testing.T) {
	t.Parallel()

	cases := []struct {
		limit Gas
		usage []Gas
	}{
		{10, []Gas{1, 2, 3, 4}},
		{1000, []Gas{40, 30, 20, 10, 900}},
		{100000, []Gas{99999, 1}},
		{100000000, []Gas{50000000, 40000000, 10000000}},
		{65535, []Gas{32768, 32767}},
		{65536, []Gas{32768, 32767, 1}},
	}

	for tcnum, tc := range cases {
		meter := NewGasMeter(tc.limit)
		used := int64(0)

		for unum, usage := range tc.usage {
			used += usage
			require.NotPanics(t, func() { meter.ConsumeGas(usage, "") }, "Not exceeded limit but panicked. tc #%d, usage #%d", tcnum, unum)
			require.Equal(t, used, meter.GasConsumed(), "Gas consumption not match. tc #%d, usage #%d", tcnum, unum)
			require.Equal(t, used, meter.GasConsumedToLimit(), "Gas consumption (to limit) not match. tc #%d, usage #%d", tcnum, unum)
			require.False(t, meter.IsPastLimit(), "Not exceeded limit but got IsPastLimit() true")
			if unum < len(tc.usage)-1 {
				require.False(t, meter.IsOutOfGas(), "Not yet at limit but got IsOutOfGas() true")
			} else {
				require.True(t, meter.IsOutOfGas(), "At limit but got IsOutOfGas() false")
			}
		}

		require.Panics(t, func() { meter.ConsumeGas(1, "") }, "Exceeded but not panicked. tc #%d", tcnum)
		require.Equal(t, meter.GasConsumedToLimit(), meter.Limit(), "Gas consumption (to limit) not match limit")
		require.Equal(t, meter.GasConsumed(), meter.Limit()+1, "Gas consumption not match limit+1")
	}
}

func TestAddUint64Overflow(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		a, b     int64
		result   int64
		overflow bool
	}{
		{0, 0, 0, false},
		{100, 100, 200, false},
		{math.MaxInt64 / 2, math.MaxInt64/2 + 1, math.MaxInt64, false},
		{math.MaxInt64 / 2, math.MaxInt64/2 + 2, math.MinInt64, true},
	}

	for i, tc := range testCases {
		res, ok := overflow.Add(tc.a, tc.b)
		overflow := !ok
		require.Equal(
			t, tc.overflow, overflow,
			"invalid overflow result; tc: #%d, a: %d, b: %d", i, tc.a, tc.b,
		)
		require.Equal(
			t, tc.result, res,
			"invalid int64 result; tc: #%d, a: %d, b: %d", i, tc.a, tc.b,
		)
	}
}

// TestWillIterator asserts that WillIterator charges ReadCostFlat
// (modelling the seek as one Get-equivalent tree walk).
func TestWillIterator(t *testing.T) {
	t.Parallel()
	meter := NewGasMeter(1 << 62)
	cfg := DefaultGasConfig()
	gctx := &GasContext{Meter: meter, Config: cfg}

	gctx.WillIterator()
	require.Equal(t, cfg.ReadCostFlat, meter.GasConsumed())

	// nil gctx is safe and a no-op.
	var nilCtx *GasContext
	require.NotPanics(t, func() { nilCtx.WillIterator() })
}

// TestWillIterNext asserts the per-step formula:
// IterNextCostFlat + ReadCostPerByte * len(value).
func TestWillIterNext(t *testing.T) {
	t.Parallel()
	meter := NewGasMeter(1 << 62)
	cfg := DefaultGasConfig()
	gctx := &GasContext{Meter: meter, Config: cfg}

	val := []byte("hello world") // 11 bytes
	gctx.WillIterNext(val)

	want := cfg.IterNextCostFlat + 11*cfg.ReadCostPerByte
	require.Equal(t, want, meter.GasConsumed())

	// Empty value charges only the flat.
	meter2 := NewGasMeter(1 << 62)
	gctx2 := &GasContext{Meter: meter2, Config: cfg}
	gctx2.WillIterNext(nil)
	require.Equal(t, cfg.IterNextCostFlat, meter2.GasConsumed())

	// nil gctx is safe.
	var nilCtx *GasContext
	require.NotPanics(t, func() { nilCtx.WillIterNext(val) })
}

// Both meters implement one interface, so they must agree on what an input may
// be. A negative consumption lowers the total instead of raising it, and the
// infinite meter reports that total as the transaction's gas used.
//
// Not reachable today: every caller derives gas from a length or a validated
// parameter, and the vm parameters that scale it are checked positive. This
// keeps the two implementations from disagreeing about the contract anyway.
func TestBothGasMetersRefuseANegativeConsumption(t *testing.T) {
	t.Parallel()

	for name, meter := range map[string]GasMeter{
		"metered":  NewGasMeter(1000),
		"infinite": NewInfiniteGasMeter(),
	} {
		require.Panics(t, func() { meter.ConsumeGas(-1, "negative") },
			"%s meter must refuse a negative consumption", name)
		require.Panics(t, func() { meter.RefundGas(-1, "negative") },
			"%s meter must refuse a negative refund", name)
	}

	// The control: a real amount is still counted, and a refund still lands.
	m := NewInfiniteGasMeter()
	m.ConsumeGas(100, "work")
	require.Equal(t, Gas(100), m.GasConsumed())
	m.RefundGas(40, "dedup")
	require.Equal(t, Gas(60), m.GasConsumed())
	// A refund larger than the total floors at zero rather than going negative.
	m.RefundGas(1000, "over")
	require.Equal(t, Gas(0), m.GasConsumed())
}
