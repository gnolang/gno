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
		th := newIPThrottler(time.Second)

		// Register < max requests
		for i := uint64(0); i < maxRequestsPerIP; i++ {
			assert.NoError(t, th.registerNewRequest(addr))
		}
	})

	t.Run("exceeded number of requests", func(t *testing.T) {
		t.Parallel()

		addr, err := netip.ParseAddr("127.0.0.1")
		require.NoError(t, err)

		// Create the IP throttler
		th := newIPThrottler(time.Second)

		// Register max requests
		for i := uint64(0); i < maxRequestsPerIP; i++ {
			assert.NoError(t, th.registerNewRequest(addr))
		}

		// Attempt to register an additional request
		assert.ErrorIs(t, th.registerNewRequest(addr), errInvalidNumberOfRequests)
	})
}

func TestIPThrottler_RequestsHalved(t *testing.T) {
	t.Parallel()

	var (
		throttleDuration = time.Millisecond * 100

		requestInterval = 5 * throttleDuration // requests triggered after ~5 wipes
		numRequests     = maxRequestsPerIP * 2 // number of request loops
	)

	addr, err := netip.ParseAddr("127.0.0.1")
	require.NoError(t, err)

	// Create the IP throttler
	th := newIPThrottler(throttleDuration)

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	// Start the throttler (async)
	th.start(ctx)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		var (
			requestsSent = uint64(0)
			ticker       = time.NewTicker(requestInterval)
		)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Fill out the request count for the address
				for i := uint64(0); i < maxRequestsPerIP; i++ {
					require.NoError(t, th.registerNewRequest(addr))
				}

				requestsSent += maxRequestsPerIP

				if requestsSent == numRequests {
					// Loops done
					return
				}
			}
		}
	}()

	wg.Wait()
}
