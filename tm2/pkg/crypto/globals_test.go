package crypto

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBech32AddrPrefix(t *testing.T) {
	// Bech32AddrPrefix should return the default non-empty string
	prefix := Bech32AddrPrefix()
	require.NotEmpty(t, prefix, "address prefix should not be empty")
	require.Equal(t, "g", prefix, "default address prefix should be 'g'")
}

func TestBech32PubKeyPrefix(t *testing.T) {
	// Bech32PubKeyPrefix should return the default non-empty string
	prefix := Bech32PubKeyPrefix()
	require.NotEmpty(t, prefix, "pubkey prefix should not be empty")
	require.Equal(t, "gpub", prefix, "default pubkey prefix should be 'gpub'")
}

func TestSetters(t *testing.T) {
	t.Cleanup(func() { onceBech32Prefixes = sync.Once{} })

	// check default values
	require.Equal(t, Bech32AddrPrefix(), "g")
	require.Equal(t, Bech32PubKeyPrefix(), "gpub")

	// set custom values
	require.Panics(t, func() {
		SetBech32AddrPrefix("", "bpub")
	})
	SetBech32AddrPrefix("b", "bpub")

	// verify custom values
	require.Equal(t, Bech32AddrPrefix(), "b")
	require.Equal(t, Bech32PubKeyPrefix(), "bpub")

	// cannot be set again
	require.Panics(t, func() {
		SetBech32AddrPrefix("b", "bpub")
	})

	t.Cleanup(func() { onceBech32Prefixes = sync.Once{} })
}
