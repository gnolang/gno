package main

import (
	"context"
	"errors"
	"net/netip"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	maxRequestsPerMinute = 5

	defaultCleanTimeout      = time.Minute * 3
	defaultRateLimitInterval = time.Minute / maxRequestsPerMinute
)

var errInvalidNumberOfRequests = errors.New("invalid number of requests")

type client struct {
	limiter *rate.Limiter
	seen    time.Time
}

type requestMap map[netip.Addr]*client

// iterate ranges over the request map (NOT thread safe)
func (r requestMap) iterate(cb func(key netip.Addr, value *client)) {
	for ip, requests := range r {
		cb(ip, requests)
	}
}

type ipThrottler struct {
	cleanupInterval   time.Duration
	rateLimitInterval time.Duration

	requestMap requestMap

	sync.Mutex
}

// newIPThrottler creates a new ip throttler
func newIPThrottler(rateLimitInterval, cleanupInterval time.Duration) *ipThrottler {
	return &ipThrottler{
		cleanupInterval:   cleanupInterval,
		rateLimitInterval: rateLimitInterval,
		requestMap:        make(requestMap),
	}
}

// start starts the throttle cleanup service
func (st *ipThrottler) start(ctx context.Context) {
	go st.runCleanup(ctx)
}

// runCleanup runs the main ip throttle cleanup loop
func (st *ipThrottler) runCleanup(ctx context.Context) {
	ticker := time.NewTicker(st.cleanupInterval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			st.Lock()

			// Clean up stale requests
			st.requestMap.iterate(func(ip netip.Addr, client *client) {
				// Check if the request was last seen a while ago
				if time.Since(client.seen) < st.cleanupInterval {
					return
				}

				delete(st.requestMap, ip)
			})

			st.Unlock()
		}
	}
}

// registerNewRequest registers a new IP request with the throttler
func (st *ipThrottler) registerNewRequest(ip netip.Addr) error {
	st.Lock()
	defer st.Unlock()

	// Get the client associated with the address, if any
	c := st.requestMap[ip]
	if c == nil {
		c = &client{
			limiter: rate.NewLimiter(rate.Every(st.rateLimitInterval), 5),
			seen:    time.Now(),
		}

		st.requestMap[ip] = c
	}

	// Check if the IP exceeded the request count
	if !c.limiter.Allow() {
		return errInvalidNumberOfRequests
	}

	// Update the last seen time
	c.seen = time.Now()

	return nil
}
