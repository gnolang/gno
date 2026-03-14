package bank

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSupplyStoreKey(t *testing.T) {
	t.Parallel()

	// Ensure successive calls don't share backing arrays.
	k1 := SupplyStoreKey("aaa")
	k2 := SupplyStoreKey("bbb")

	require.Equal(t, []byte("/s/aaa"), k1)
	require.Equal(t, []byte("/s/bbb"), k2)

	// Mutating k1 must not affect k2.
	k1[3] = 'x'
	require.Equal(t, []byte("/s/bbb"), k2)
}
