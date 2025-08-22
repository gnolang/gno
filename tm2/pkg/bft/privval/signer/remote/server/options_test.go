package server

import (
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRemoteSignerServer(t *testing.T) {
	t.Parallel()

	var (
		signer        = types.NewMockSigner()
		listenAddress = "tcp://127.0.0.1"
		logger        = log.NewNoopLogger()
	)

	t.Run("nil signer", func(t *testing.T) {
		t.Parallel()

		rss, err := NewRemoteSignerServer(nil, "", nil)
		require.Nil(t, rss)
		assert.ErrorIs(t, err, ErrNilSigner)
	})

	t.Run("invalid listenAddress", func(t *testing.T) {
		t.Parallel()

		// Test empty listenAddress.
		invalidListenAddress := "udp://127.0.0.1"
		rss, err := NewRemoteSignerServer(signer, invalidListenAddress, nil)
		require.Nil(t, rss)
		assert.ErrorIs(t, err, ErrInvalidAddressProtocol)
	})

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()

		rss, err := NewRemoteSignerServer(signer, listenAddress, nil)
		require.Nil(t, rss)
		assert.ErrorIs(t, err, ErrNilLogger)
	})

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		rss, err := NewRemoteSignerServer(signer, listenAddress, logger)
		require.NotNil(t, rss)
		assert.NoError(t, err)
	})

	t.Run("option keepAlivePeriod", func(t *testing.T) {
		t.Parallel()

		// Test default keepAlivePeriod.
		rss, err := NewRemoteSignerServer(signer, listenAddress, logger)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.Equal(t, DefaultKeepAlivePeriod, rss.keepAlivePeriod)

		// Test functional option.
		option := WithKeepAlivePeriod(42)
		rss, err = NewRemoteSignerServer(signer, listenAddress, logger, option)
		require.NotNil(t, rss)
		require.NoError(t, err)
		assert.Equal(t, time.Duration(42), rss.keepAlivePeriod)
	})

	t.Run("option responseTimeout", func(t *testing.T) {
		t.Parallel()

		// Test default responseTimeout.
		rss, err := NewRemoteSignerServer(signer, listenAddress, logger)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.Equal(t, DefaultResponseTimeout, rss.responseTimeout)

		// Test functional option.
		option := WithResponseTimeout(42)
		rss, err = NewRemoteSignerServer(signer, listenAddress, logger, option)
		require.NotNil(t, rss)
		require.NoError(t, err)
		assert.Equal(t, time.Duration(42), rss.responseTimeout)
	})

	t.Run("option serverPrivKey", func(t *testing.T) {
		t.Parallel()

		// Test default serverPrivKey.
		rss, err := NewRemoteSignerServer(signer, listenAddress, logger)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NotNil(t, rss.serverPrivKey)

		// Test functional option.
		privKey := ed25519.GenPrivKey()
		option := WithServerPrivKey(privKey)
		rss, err = NewRemoteSignerServer(signer, listenAddress, logger, option)
		require.NotNil(t, rss)
		require.NoError(t, err)
		assert.Equal(t, privKey, rss.serverPrivKey)
	})

	t.Run("option authorizedKeys", func(t *testing.T) {
		t.Parallel()

		// Test default authorizedKeys.
		rss, err := NewRemoteSignerServer(signer, listenAddress, logger)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.Empty(t, rss.authorizedKeys)

		// Test functional option.
		keys := []ed25519.PubKeyEd25519{ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)}
		option := WithAuthorizedKeys(keys)
		rss, err = NewRemoteSignerServer(signer, listenAddress, logger, option)
		require.NotNil(t, rss)
		require.NoError(t, err)
		assert.Equal(t, keys, rss.authorizedKeys)
	})
}
