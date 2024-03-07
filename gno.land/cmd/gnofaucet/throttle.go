package main

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

const (
	maxRequestsPerIP = 5
)

var (
	errInvalidIP               = errors.New("invalid IP")
	errInvalidNumberOfRequests = errors.New("invalid number of requests")
)

type subnetThrottler struct {
	throttleDuration time.Duration
	subnets          sync.Map // unique IP identifier -> request count (<=5)
}

// newSubnetThrottler creates a new subnet throttler
func newSubnetThrottler(duration time.Duration) *subnetThrottler {
	return &subnetThrottler{
		throttleDuration: duration,
	}
}

// start starts the throttle routine
func (st *subnetThrottler) start(ctx context.Context) {
	go st.runThrottler(ctx)
}

// runThrottler runs the main subnet throttle loop
func (st *subnetThrottler) runThrottler(ctx context.Context) {
	ticker := time.NewTicker(st.throttleDuration)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Reset the individual subnet counters
			st.subnets.Range(func(key, value any) bool {
				newVal, _ := value.(uint8)

				newVal /= 2
				if newVal == 0 {
					// Drop the key from the subnet map
					st.subnets.Delete(key)
				} else {
					// Save the new subnet value
					st.subnets.Store(key, newVal)
				}

				return true
			})
		}
	}
}

// registerNewRequest verifies the given request IP
func (st *subnetThrottler) registerNewRequest(requestIP net.IP) error {
	// Check if the IP is valid
	ip := requestIP.To4()
	if ip == nil {
		return errInvalidIP
	}

	// Get the unique IP identifier
	key := int64(ip[0])<<16 + int64(ip[1])<<8 + int64(ip[2])

	// Increment the request count for the IP
	count, loaded := st.subnets.LoadOrStore(key, 1)
	if !loaded {
		// Request is the first one from this IP key
		return nil
	}

	// Check if the IP exceeded the request count
	requestCount := count.(int8)
	if requestCount > maxRequestsPerIP {
		return errInvalidNumberOfRequests
	}

	// Update the request count
	st.subnets.Store(key, requestCount+1)

	return nil
}
