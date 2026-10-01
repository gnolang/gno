package crypto

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func resetBech32Prefixes() {
	onceBech32Prefixes = sync.Once{}
	bech32AddrPrefix = ""
	bech32PubKeyPrefix = ""
}

func TestBech32AddrPrefix(t *testing.T) {
	resetBech32Prefixes()
	t.Cleanup(resetBech32Prefixes)

	// Bech32AddrPrefix should return the default non-empty string
	prefix := Bech32AddrPrefix()
	require.NotEmpty(t, prefix, "address prefix should not be empty")
	require.Equal(t, "g", prefix, "default address prefix should be 'g'")
}

func TestBech32PubKeyPrefix(t *testing.T) {
	resetBech32Prefixes()
	t.Cleanup(resetBech32Prefixes)

	// Bech32PubKeyPrefix should return the default non-empty string
	prefix := Bech32PubKeyPrefix()
	require.NotEmpty(t, prefix, "pubkey prefix should not be empty")
	require.Equal(t, "gpub", prefix, "default pubkey prefix should be 'gpub'")
}

func TestSetBech32Prefixes(t *testing.T) {
	resetBech32Prefixes()
	t.Cleanup(resetBech32Prefixes)

	// check default values
	require.Equal(t, "g", Bech32AddrPrefix())
	require.Equal(t, "gpub", Bech32PubKeyPrefix())
}

func TestSetBech32Prefixes_Custom(t *testing.T) {
	resetBech32Prefixes()
	t.Cleanup(resetBech32Prefixes)

	// set custom values
	require.Panics(t, func() {
		SetBech32Prefixes("", "bpub")
	})
	SetBech32Prefixes("b", "bpub")

	// verify custom values
	require.Equal(t, "b", Bech32AddrPrefix())
	require.Equal(t, "bpub", Bech32PubKeyPrefix())

	// cannot be set again
	require.Panics(t, func() {
		SetBech32Prefixes("b", "bpub")
	})
}
