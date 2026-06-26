package discovery

import "time"

type Option func(*Reactor)

// WithDiscoveryInterval sets the discovery crawl interval
func WithDiscoveryInterval(interval time.Duration) Option {
	return func(r *Reactor) {
		r.discoveryInterval = interval
	}
}
