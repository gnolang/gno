package discovery

import "time"

type Option func(*Reactor)

// WithDiscoveryInterval sets the discovery crawl interval
func WithDiscoveryInterval(interval time.Duration) Option {
	return func(r *Reactor) {
		r.discoveryInterval = interval
	}
}

// WithStore sets the persistent peer store.
// When set, discovered peers are saved to disk and reloaded on startup.
func WithStore(store *Store) Option {
	return func(r *Reactor) {
		r.store = store
	}
}
