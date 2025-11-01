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

func TestSettersAreIdempotent(t *testing.T) {
	// Calling setters multiple times with the same value should be safe
	// Note: We use the default values to avoid interfering with other tests
	// since sync.Once persists across tests in the same package
	defaultPrefix := "g"
	defaultPubKeyPrefix := "gpub"

	// Multiple calls should not panic or cause issues
	SetBech32AddrPrefix(defaultPrefix)
	SetBech32AddrPrefix(defaultPrefix)
	SetBech32AddrPrefix(defaultPrefix)

	SetBech32PubKeyPrefix(defaultPubKeyPrefix)
	SetBech32PubKeyPrefix(defaultPubKeyPrefix)
	SetBech32PubKeyPrefix(defaultPubKeyPrefix)

	// Values should remain as default (or whatever was set first)
	require.NotEmpty(t, GetBech32AddrPrefix())
	require.NotEmpty(t, GetBech32PubKeyPrefix())
}
