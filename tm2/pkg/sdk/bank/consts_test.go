package bank

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSupplyStoreKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		denom    string
		expected []byte
	}{
		{
			name:     "standard denom",
			denom:    "ugnot",
			expected: []byte("/s/ugnot"),
		},
		{
			name:     "empty denom",
			denom:    "",
			expected: []byte("/s/"),
		},
		{
			name:     "denom with slash",
			denom:    "ibc/ABC123",
			expected: []byte("/s/ibc/ABC123"),
		},
		{
			name:     "single char denom",
			denom:    "a",
			expected: []byte("/s/a"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SupplyStoreKey(tt.denom)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestSupplyStoreKeyIsolation(t *testing.T) {
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
