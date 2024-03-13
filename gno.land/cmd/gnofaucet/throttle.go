package main

import (
	"context"
	"errors"
	"net/netip"
	"sync"
	"time"
)

const (
	maxRequestsPerIP = uint64(5)
)

var errInvalidNumberOfRequests = errors.New("invalid number of requests")

type requestMap map[netip.Addr]uint64

// iterate ranges over the request map
func (r requestMap) iterate(cb func(key netip.Addr, value uint64)) {
	for ip, requests := range r {
		cb(ip, requests)
	}
}

type ipThrottler struct {
	throttleTimeout time.Duration
	requestMap      requestMap

	sync.Mutex
}

// newIPThrottler creates a new ip throttler
func newIPThrottler(duration time.Duration) *ipThrottler {
	return &ipThrottler{
		throttleTimeout: duration,
		requestMap:      make(requestMap),
	}
}

// start starts the throttle routine
func (st *ipThrottler) start(ctx context.Context) {
	go st.runThrottler(ctx)
}

// runThrottler runs the main ip throttle loop
func (st *ipThrottler) runThrottler(ctx context.Context) {
	ticker := time.NewTicker(st.throttleTimeout)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			st.Lock()

			// Reset the individual subnet counters
			st.requestMap.iterate(func(ip netip.Addr, requests uint64) {
				newVal := requests / 2

				if newVal == 0 {
					delete(st.requestMap, ip)

					return
				}

				st.requestMap[ip] = newVal
			})

			st.Unlock()
		}
	}
}

// registerNewRequest registers a new IP request with the throttler
func (st *ipThrottler) registerNewRequest(ip netip.Addr) error {
	st.Lock()
	defer st.Unlock()

	// Get the request count for the address, if any
	requests := st.requestMap[ip]

	// Check if the IP exceeded the request count
	if requests >= maxRequestsPerIP {
		return errInvalidNumberOfRequests
	}

	// Update the request count
	st.requestMap[ip] = requests + 1

	return nil
}
