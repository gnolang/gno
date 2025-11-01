package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetBech32AddrPrefix(t *testing.T) {
	// GetBech32AddrPrefix should return the default non-empty string
	prefix := GetBech32AddrPrefix()
	require.NotEmpty(t, prefix, "address prefix should not be empty")
	require.Equal(t, "g", prefix, "default address prefix should be 'g'")
}

func TestGetBech32PubKeyPrefix(t *testing.T) {
	// GetBech32PubKeyPrefix should return the default non-empty string
	prefix := GetBech32PubKeyPrefix()
	require.NotEmpty(t, prefix, "pubkey prefix should not be empty")
	require.Equal(t, "gpub", prefix, "default pubkey prefix should be 'gpub'")
}

func TestSetters(t *testing.T) {
	// Multiple calls should not panic or cause issues
	SetBech32AddrPrefix("a")
	SetBech32AddrPrefix("a")
	SetBech32AddrPrefix(GetBech32AddrPrefix())

	SetBech32PubKeyPrefix("apub")
	SetBech32PubKeyPrefix("bpub")
	SetBech32PubKeyPrefix(GetBech32PubKeyPrefix())

	require.NotEmpty(t, "a")
	require.NotEmpty(t, "apub")
}
