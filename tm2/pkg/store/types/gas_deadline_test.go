package types

import (
	"math"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeadlineGasMeter(t *testing.T) {
	t.Parallel()

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()
		_, ok := NewDeadlineGasMeter(10, nil).(*basicGasMeter)
		require.True(t, ok, "nil flag should give a plain meter")
	})

	t.Run("meters", func(t *testing.T) {
		t.Parallel()
		g := NewDeadlineGasMeter(100, new(atomic.Bool))
		g.ConsumeGas(60, "a")
		require.False(t, g.IsOutOfGas())
		// A real out-of-gas is still reported as such.
		require.PanicsWithValue(t, OutOfGasError{"c"}, func() { g.ConsumeGas(41, "c") })
		require.True(t, g.IsPastLimit())
		require.True(t, g.IsOutOfGas())
	})

	t.Run("fires on the shared flag", func(t *testing.T) {
		t.Parallel()
		expired := new(atomic.Bool)
		g := NewDeadlineGasMeter(math.MaxInt64, expired)
		g.ConsumeGas(1, "op")
		expired.Store(true)
		require.PanicsWithValue(t, DeadlineExceededError{"op"}, func() { g.ConsumeGas(1, "op") })
		// The charge that saw the flag is still counted, and the meter is
		// not out of gas: a caller can tell the two stops apart.
		require.Equal(t, Gas(2), g.GasConsumed())
		require.False(t, g.IsPastLimit())
		// Sticky: every later charge fails at once.
		require.PanicsWithValue(t, DeadlineExceededError{"next"}, func() { g.ConsumeGas(0, "next") })
	})

	t.Run("real out of gas wins after expiry", func(t *testing.T) {
		t.Parallel()
		expired := new(atomic.Bool)
		expired.Store(true)
		g := NewDeadlineGasMeter(10, expired)
		require.PanicsWithValue(t, OutOfGasError{"op"}, func() { g.ConsumeGas(11, "op") })
	})
}

// BenchmarkDeadlineGasMeter measures the per-ConsumeGas cost of the deadline
// meter against the plain one.
func BenchmarkDeadlineGasMeter(b *testing.B) {
	b.Run("base", func(b *testing.B) {
		g := GasMeter(NewGasMeter(math.MaxInt64))
		for b.Loop() {
			g.ConsumeGas(1, "op")
		}
	})
	b.Run("deadline", func(b *testing.B) {
		g := NewDeadlineGasMeter(math.MaxInt64, new(atomic.Bool))
		for b.Loop() {
			g.ConsumeGas(1, "op")
		}
	})
}
