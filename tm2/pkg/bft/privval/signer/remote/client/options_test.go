package client

import (
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRemoteSignerClient(t *testing.T) {
	t.Parallel()

	var (
		validTCP  = "tcp://127.0.0.1"
		validUnix = "unix:///tmp/remote_signer.sock"
		logger    = log.NewNoopLogger()
	)

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

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		// Test TCP connection.
		rsc, err := NewRemoteSignerClient(validTCP, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)

		// Test Unix socket connection.
		rsc, err = NewRemoteSignerClient(validUnix, logger)
		require.NotNil(t, rsc)
		assert.NoError(t, err)
	})

	t.Run("option dialMaxRetries", func(t *testing.T) {
		t.Parallel()

		// Test default dialMaxRetries.
		rsc, err := NewRemoteSignerClient(validTCP, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.Equal(t, defaultDialMaxRetries, rsc.dialMaxRetries)

		// Test functional option.
		option := WithDialMaxRetries(3)
		rsc, err = NewRemoteSignerClient(validTCP, logger, option)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.Equal(t, 3, rsc.dialMaxRetries)
	})

	t.Run("option dialRetryInterval", func(t *testing.T) {
		t.Parallel()

		// Test default dialRetryInterval.
		rsc, err := NewRemoteSignerClient(validTCP, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.Equal(t, defaultDialRetryInterval, rsc.dialRetryInterval)

		// Test functional option.
		option := WithDialRetryInterval(42)
		rsc, err = NewRemoteSignerClient(validTCP, logger, option)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.Equal(t, time.Duration(42), rsc.dialRetryInterval)
	})

	t.Run("option dialTimeout", func(t *testing.T) {
		t.Parallel()

		// Test default dialTimeout.
		rsc, err := NewRemoteSignerClient(validTCP, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.Equal(t, defaultDialTimeout, rsc.dialTimeout)

		// Test functional option.
		option := WithDialTimeout(time.Microsecond)
		rsc, err = NewRemoteSignerClient(validTCP, logger, option)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.Equal(t, time.Microsecond, rsc.dialTimeout)
	})

	t.Run("option keepAlivePeriod", func(t *testing.T) {
		t.Parallel()

		// Test default keepAlivePeriod.
		rsc, err := NewRemoteSignerClient(validTCP, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.Equal(t, defaultKeepAlivePeriod, rsc.keepAlivePeriod)

		// Test functional option.
		option := WithKeepAlivePeriod(42)
		rsc, err = NewRemoteSignerClient(validTCP, logger, option)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.Equal(t, time.Duration(42), rsc.keepAlivePeriod)
	})

	t.Run("option requestTimeout", func(t *testing.T) {
		t.Parallel()

		// Test default requestTimeout.
		rsc, err := NewRemoteSignerClient(validTCP, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.Equal(t, defaultRequestTimeout, rsc.requestTimeout)

		// Test functional option.
		option := WithRequestTimeout(42)
		rsc, err = NewRemoteSignerClient(validTCP, logger, option)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.Equal(t, time.Duration(42), rsc.requestTimeout)
	})

	t.Run("option clientPrivKey", func(t *testing.T) {
		t.Parallel()

		// Test default clientPrivKey.
		rsc, err := NewRemoteSignerClient(validTCP, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.NotNil(t, rsc.clientPrivKey)

		// Test functional option.
		privKey := ed25519.GenPrivKey()
		option := WithClientPrivKey(privKey)
		rsc, err = NewRemoteSignerClient(validTCP, logger, option)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.Equal(t, privKey, rsc.clientPrivKey)
	})

	t.Run("option authorizedKeys", func(t *testing.T) {
		t.Parallel()

		// Test default authorizedKeys.
		rsc, err := NewRemoteSignerClient(validTCP, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.Empty(t, rsc.authorizedKeys)

		// Test functional option.
		keys := []ed25519.PubKeyEd25519{ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)}
		option := WithAuthorizedKeys(keys)
		rsc, err = NewRemoteSignerClient(validTCP, logger, option)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.Equal(t, keys, rsc.authorizedKeys)
	})
}
