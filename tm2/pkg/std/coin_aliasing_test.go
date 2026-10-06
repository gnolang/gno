package std

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// NewCoins and Coins.Add must not write through to slices they were handed.
//
// Both used to. removeZeroCoins deleted in place, and NewCoins sorted the
// variadic array, which is the caller's own slice when spread with
// `NewCoins(myVar...)`. Deleting shifts the set down over the gap, so a caller
// came back holding a reordered set ending in a zero-amount coin it never had
// -- one that Validate then rejects.
//
// The Gno mirror in gnovm/stdlibs/chain/coins.gno had the same fault and is
// covered by its own tests. Keep the two in step.

func TestNewCoinsLeavesTheInputSliceAlone(t *testing.T) {
	t.Parallel()

	// A zero entry is what made removeZeroCoins shift the caller's array.
	mine := Coins{{Denom: "aaa", Amount: 0}, {Denom: "bbb", Amount: 5}, {Denom: "ccc", Amount: 7}}
	before := append(Coins(nil), mine...)

	got := NewCoins(mine...)

	require.Equal(t, before, mine, "NewCoins must not modify the slice it was given")
	require.Equal(t, Coins{{Denom: "bbb", Amount: 5}, {Denom: "ccc", Amount: 7}}, got)
	// Spelled out, because this is the shift the old code performed: the zero
	// stayed put instead of the tail moving down over it.
	require.Zero(t, mine[0].Amount, "the caller's zero entry must still be first")
}

func TestNewCoinsDoesNotReorderTheInputSlice(t *testing.T) {
	t.Parallel()

	// No zero entry here, so only the in-place Sort could reach the caller.
	mine := Coins{{Denom: "zzz", Amount: 1}, {Denom: "aaa", Amount: 2}}
	before := append(Coins(nil), mine...)

	got := NewCoins(mine...)

	require.Equal(t, before, mine, "NewCoins must not sort the slice it was given")
	require.Equal(t, Coins{{Denom: "aaa", Amount: 2}, {Denom: "zzz", Amount: 1}}, got,
		"the returned set is still sorted")
}

func TestCoinsAddLeavesBothOperandsAlone(t *testing.T) {
	t.Parallel()

	// AddUnsafe passes a view of whichever side it has exhausted to
	// removeZeroCoins, so check the receiver and the argument separately.
	t.Run("argument", func(t *testing.T) {
		t.Parallel()
		arg := Coins{{Denom: "aaa", Amount: 0}, {Denom: "bbb", Amount: 5}}
		before := append(Coins(nil), arg...)

		got := Coins{}.AddUnsafe(arg)

		require.Equal(t, before, arg, "Add must not modify its argument")
		require.Equal(t, Coins{{Denom: "bbb", Amount: 5}}, got)
	})

	t.Run("receiver", func(t *testing.T) {
		t.Parallel()
		recv := Coins{{Denom: "aaa", Amount: 0}, {Denom: "bbb", Amount: 5}}
		before := append(Coins(nil), recv...)

		got := recv.AddUnsafe(Coins{})

		require.Equal(t, before, recv, "Add must not modify its receiver")
		require.Equal(t, Coins{{Denom: "bbb", Amount: 5}}, got)
	})
}
