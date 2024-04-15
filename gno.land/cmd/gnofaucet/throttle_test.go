package main

import (
	"context"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPThrottler_RegisterNewRequest(t *testing.T) {
	t.Parallel()

	t.Run("valid number of requests", func(t *testing.T) {
		t.Parallel()

		addr, err := netip.ParseAddr("127.0.0.1")
		require.NoError(t, err)

		// Create the IP throttler
		th := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)

		// Register < max requests
		for i := uint64(0); i < maxRequestsPerMinute; i++ {
			assert.NoError(t, th.registerNewRequest(addr))
		}
	})

	t.Run("exceeded number of requests", func(t *testing.T) {
		t.Parallel()

		addr, err := netip.ParseAddr("127.0.0.1")
		require.NoError(t, err)

		// Create the IP throttler
		th := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)

		// Register max requests
		for i := uint64(0); i < maxRequestsPerMinute; i++ {
			assert.NoError(t, th.registerNewRequest(addr))
		}

		// Attempt to register an additional request
		assert.ErrorIs(t, th.registerNewRequest(addr), errInvalidNumberOfRequests)
	})
}

func TestIPThrottler_RequestsThrottled(t *testing.T) {
	t.Parallel()

	var (
		cleanupInterval = time.Millisecond * 100

		requestInterval = 3 * cleanupInterval      // requests triggered after ~5 cleans
		numRequests     = maxRequestsPerMinute * 2 // number of request loops
	)

	addr, err := netip.ParseAddr("127.0.0.1")
	require.NoError(t, err)

	// Create the IP throttler
	th := newIPThrottler(defaultRateLimitInterval, cleanupInterval)

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	// Start the throttler (async)
	th.start(ctx)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		var (
			requestsSent = 0
			ticker       = time.NewTicker(requestInterval)
		)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Fill out the request count for the address
				for i := uint64(0); i < maxRequestsPerMinute; i++ {
					require.NoError(t, th.registerNewRequest(addr))
				}

				requestsSent += maxRequestsPerMinute

				if requestsSent == numRequests {
					// Loops done
					return
				}
			}
		}
	}()

	wg.Wait()
}
