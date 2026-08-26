package client

import (
	"context"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDialerInitializedFromDialTimeout verifies the internal dialer carries
// the configured (or default) dial timeout.
func TestDialerInitializedFromDialTimeout(t *testing.T) {
	t.Parallel()

	logger := log.NewNoopLogger()

	t.Run("default dial timeout", func(t *testing.T) {
		t.Parallel()

		// Set up a remote signer server so the client can connect.
		unixSocket := testUnixSocket(t)
		rss := newRemoteSignerServer(t, unixSocket, nil)
		require.NotNil(t, rss)
		require.NoError(t, rss.Start())
		defer rss.Stop()

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		rsc, err := NewRemoteSignerClient(ctx, unixSocket, logger)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		defer rsc.Close()

		assert.Equal(t, defaultDialTimeout, rsc.dialTimeout)
		assert.Equal(t, defaultDialTimeout, rsc.dialer.Timeout,
			"the dialer timeout should match the configured dial timeout")
	})

	t.Run("custom dial timeout", func(t *testing.T) {
		t.Parallel()

		// Set up a remote signer server so the client can connect.
		unixSocket := testUnixSocket(t)
		rss := newRemoteSignerServer(t, unixSocket, nil)
		require.NotNil(t, rss)
		require.NoError(t, rss.Start())
		defer rss.Stop()

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		rsc, err := NewRemoteSignerClient(ctx, unixSocket, logger, WithDialTimeout(time.Second))
		require.NotNil(t, rsc)
		require.NoError(t, err)
		defer rsc.Close()

		assert.Equal(t, time.Second, rsc.dialTimeout)
		assert.Equal(t, time.Second, rsc.dialer.Timeout,
			"the dialer timeout should match the configured dial timeout")
	})
}
