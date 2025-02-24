package client

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("default config", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, DefaultRemoteSignerClientConfig().ValidateBasic())
	})

	t.Run("test config", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, TestRemoteSignerClientConfig().ValidateBasic())
	})

	t.Run("default config with invalid keys", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultRemoteSignerClientConfig()
		cfg.AuthorizedKeys = []string{"invalid_key"}

		require.ErrorIs(t, cfg.ValidateBasic(), errInvalidAuthorizedKey)
	})
}

func TestAuthorizedKeys(t *testing.T) {
	t.Parallel()

	t.Run("invalid key bech32", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultRemoteSignerClientConfig()
		cfg.AuthorizedKeys = []string{"invalid_key"}

		keys, err := cfg.authorizedKeys()
		require.Nil(t, keys)
		require.ErrorIs(t, err, errInvalidAuthorizedKey)
	})

	t.Run("invalid key type", func(t *testing.T) {
		t.Parallel()

		invalidPubKey := secp256k1.GenPrivKey().PubKey().String() // Not an ed25519 key

		cfg := DefaultRemoteSignerClientConfig()
		cfg.AuthorizedKeys = []string{invalidPubKey}

		keys, err := cfg.authorizedKeys()
		require.Nil(t, keys)
		require.ErrorIs(t, err, errInvalidAuthorizedKey)
	})

	t.Run("valid authorized keys", func(t *testing.T) {
		t.Parallel()

		validKeys := make([]string, 3)
		for i := range validKeys {
			validKeys[i] = ed25519.GenPrivKey().PubKey().String()
		}

		cfg := DefaultRemoteSignerClientConfig()
		cfg.AuthorizedKeys = validKeys

		keys, err := cfg.authorizedKeys()
		require.Equal(t, len(keys), 3)
		require.NoError(t, err)
	})
}

func TestNewRemoteSignerClientFromConfig(t *testing.T) {
	t.Parallel()

	var (
		privKey = ed25519.GenPrivKey()
		logger  = log.NewNoopLogger()
	)

	t.Run("invalid key type", func(t *testing.T) {
		t.Parallel()

		invalidPubKey := secp256k1.GenPrivKey().PubKey().String() // Not an ed25519 key

		cfg := DefaultRemoteSignerClientConfig()
		cfg.AuthorizedKeys = []string{invalidPubKey}

		client, err := NewRemoteSignerClientFromConfig(cfg, privKey, logger)
		require.Nil(t, client)
		require.ErrorIs(t, err, errInvalidAuthorizedKey)
	})

	t.Run("valid authorized keys", func(t *testing.T) {
		t.Parallel()

		validKeys := make([]string, 3)
		for i := range validKeys {
			validKeys[i] = ed25519.GenPrivKey().PubKey().String()
		}

		cfg := DefaultRemoteSignerClientConfig()
		cfg.AuthorizedKeys = validKeys

		client, err := NewRemoteSignerClientFromConfig(cfg, privKey, logger)
		require.NotNil(t, client)
		require.NoError(t, err)

		keys, err := cfg.authorizedKeys()
		require.NotNil(t, keys)
		require.NoError(t, err)
		require.Equal(t, client.authorizedKeys, keys)
		require.Equal(t, client.clientPrivKey, privKey)
		client.Close()
	})
}
