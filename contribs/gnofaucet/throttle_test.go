package main

import (
	"context"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPThrottler_RegisterNewRequest(t *testing.T) {
	t.Parallel()

	t.Run("first request allowed", func(t *testing.T) {
		t.Parallel()

		addr := netip.MustParseAddr("127.0.0.1")

		th := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)

		assert.NoError(t, th.registerNewRequest(addr))
	})

	t.Run("second request rejected", func(t *testing.T) {
		t.Parallel()

		addr := netip.MustParseAddr("127.0.0.1")

		th := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)

		require.NoError(t, th.registerNewRequest(addr))

		assert.ErrorIs(t, th.registerNewRequest(addr), errInvalidNumberOfRequests)
	})
}

func TestIPThrottler_SecondRequestRejected(t *testing.T) {
	t.Parallel()

	addr := netip.MustParseAddr("192.168.1.1")

	// Use a long interval so no tokens regenerate during the test
	th := newIPThrottler(time.Hour, defaultCleanTimeout)

	// First request must succeed
	require.NoError(t, th.registerNewRequest(addr))

	// Second request from the same IP must be rejected
	assert.ErrorIs(t, th.registerNewRequest(addr), errInvalidNumberOfRequests)
}

func TestIPThrottler_CleanupAllowsNewRequest(t *testing.T) {
	t.Parallel()

	cleanupInterval := time.Millisecond * 100

	addr := netip.MustParseAddr("127.0.0.1")

	// Rate interval is long so tokens won't regenerate on their own;
	// only cleanup (removing the stale entry) should allow a new request.
	th := newIPThrottler(time.Hour, cleanupInterval)

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	th.start(ctx)

	// First request succeeds, second is rejected
	require.NoError(t, th.registerNewRequest(addr))
	require.ErrorIs(t, th.registerNewRequest(addr), errInvalidNumberOfRequests)

	// Wait for the cleanup cycle to evict the stale entry
	time.Sleep(cleanupInterval * 3)

	// After cleanup the IP entry is gone, so a new request succeeds
	assert.NoError(t, th.registerNewRequest(addr))
}
