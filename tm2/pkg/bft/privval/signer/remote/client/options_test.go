package client

import (
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRemoteSignerClient(t *testing.T) {
	t.Parallel()

	logger := log.NewNoopLogger()

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()

		rsc, err := NewRemoteSignerClient("", nil)
		require.Nil(t, rsc)
		assert.ErrorIs(t, err, ErrNilLogger)
	})

	t.Run("invalid protocol", func(t *testing.T) {
		t.Parallel()

		invalidAddressProtocol := "udp://127.0.0.1"
		rsc, err := NewRemoteSignerClient(invalidAddressProtocol, logger)
		require.Nil(t, rsc)
		assert.ErrorIs(t, err, ErrInvalidAddressProtocol)
	})

	t.Run("option dialMaxRetries", func(t *testing.T) {
		t.Parallel()

		// Set up a remote signer server for testing.
		unixSocket := testUnixSocket(t)
		rss := newRemoteSignerServer(t, unixSocket, types.NewMockSigner())
		require.NotNil(t, rss)
		require.NoError(t, rss.Start())
		defer rss.Stop()

		// Test default dialMaxRetries.
		rsc, err := NewRemoteSignerClient(unixSocket, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.Equal(t, defaultDialMaxRetries, rsc.dialMaxRetries)
		rsc.Close()

		// Test functional option.
		option := WithDialMaxRetries(3)
		rsc, err = NewRemoteSignerClient(unixSocket, logger, option)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.Equal(t, 3, rsc.dialMaxRetries)
	})

	t.Run("option dialRetryInterval", func(t *testing.T) {
		t.Parallel()

		// Set up a remote signer server for testing.
		unixSocket := testUnixSocket(t)
		rss := newRemoteSignerServer(t, unixSocket, types.NewMockSigner())
		require.NotNil(t, rss)
		require.NoError(t, rss.Start())
		defer rss.Stop()

		// Test default dialRetryInterval.
		rsc, err := NewRemoteSignerClient(unixSocket, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.Equal(t, defaultDialRetryInterval, rsc.dialRetryInterval)
		rsc.Close()

		// Test functional option.
		option := WithDialRetryInterval(42)
		rsc, err = NewRemoteSignerClient(unixSocket, logger, option)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.Equal(t, time.Duration(42), rsc.dialRetryInterval)
	})

	t.Run("option dialTimeout", func(t *testing.T) {
		t.Parallel()

		// Set up a remote signer server for testing.
		unixSocket := testUnixSocket(t)
		rss := newRemoteSignerServer(t, unixSocket, types.NewMockSigner())
		require.NotNil(t, rss)
		require.NoError(t, rss.Start())
		defer rss.Stop()

		// Test default dialTimeout.
		rsc, err := NewRemoteSignerClient(unixSocket, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.Equal(t, defaultDialTimeout, rsc.dialTimeout)
		rsc.Close()

		// Test functional option.
		option := WithDialTimeout(time.Millisecond)
		rsc, err = NewRemoteSignerClient(unixSocket, logger, option)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.Equal(t, time.Millisecond, rsc.dialTimeout)
	})

	t.Run("option keepAlivePeriod", func(t *testing.T) {
		t.Parallel()

		// Set up a remote signer server for testing.
		unixSocket := testUnixSocket(t)
		rss := newRemoteSignerServer(t, unixSocket, types.NewMockSigner())
		require.NotNil(t, rss)
		require.NoError(t, rss.Start())
		defer rss.Stop()

		// Test default keepAlivePeriod.
		rsc, err := NewRemoteSignerClient(unixSocket, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.Equal(t, defaultKeepAlivePeriod, rsc.keepAlivePeriod)
		rsc.Close()

		// Test functional option.
		option := WithKeepAlivePeriod(42)
		rsc, err = NewRemoteSignerClient(unixSocket, logger, option)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.Equal(t, time.Duration(42), rsc.keepAlivePeriod)
	})

	t.Run("option requestTimeout", func(t *testing.T) {
		t.Parallel()

		// Set up a remote signer server for testing.
		unixSocket := testUnixSocket(t)
		rss := newRemoteSignerServer(t, unixSocket, types.NewMockSigner())
		require.NotNil(t, rss)
		require.NoError(t, rss.Start())
		defer rss.Stop()

		// Test default requestTimeout.
		rsc, err := NewRemoteSignerClient(unixSocket, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.Equal(t, defaultRequestTimeout, rsc.requestTimeout)
		rsc.Close()

		// Test functional option.
		option := WithRequestTimeout(time.Millisecond)
		rsc, err = NewRemoteSignerClient(unixSocket, logger, option)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.Equal(t, time.Millisecond, rsc.requestTimeout)
	})

	t.Run("option clientPrivKey", func(t *testing.T) {
		t.Parallel()

		// Set up a remote signer server for testing.
		unixSocket := testUnixSocket(t)
		rss := newRemoteSignerServer(t, unixSocket, types.NewMockSigner())
		require.NotNil(t, rss)
		require.NoError(t, rss.Start())
		defer rss.Stop()

		// Test default clientPrivKey.
		rsc, err := NewRemoteSignerClient(unixSocket, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.NotNil(t, rsc.clientPrivKey)
		rsc.Close()

		// Test functional option.
		privKey := ed25519.GenPrivKey()
		option := WithClientPrivKey(privKey)
		rsc, err = NewRemoteSignerClient(unixSocket, logger, option)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.Equal(t, privKey, rsc.clientPrivKey)
	})

	t.Run("option authorizedKeys", func(t *testing.T) {
		t.Parallel()

		// Set up a remote signer server for testing.
		unixSocket := testUnixSocket(t)
		rss := newRemoteSignerServer(t, unixSocket, types.NewMockSigner())
		require.NotNil(t, rss)
		require.NoError(t, rss.Start())
		defer rss.Stop()

		// Test default authorizedKeys.
		rsc, err := NewRemoteSignerClient(unixSocket, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.Empty(t, rsc.authorizedKeys)
		rsc.Close()

		// Test functional option.
		keys := []ed25519.PubKeyEd25519{ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)}
		option := WithAuthorizedKeys(keys)
		rsc, err = NewRemoteSignerClient(unixSocket, logger, option)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.Equal(t, keys, rsc.authorizedKeys)
	})
}
