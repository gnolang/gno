package store

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	// cacheTTL bounds how stale a store page can be. The realm's shelves
	// move by block height, so seconds of staleness are invisible.
	cacheTTL = 30 * time.Second

	// errorTTL remembers a failure briefly, so a store realm that is down
	// costs one timeout per key every few seconds, not one per reader.
	errorTTL = 5 * time.Second

	// cacheMaxEntries bounds memory. Keys come from request paths, so an
	// attacker can mint them; past the cap expired entries go first, then
	// everything.
	cacheMaxEntries = 512
)

type cacheEntry struct {
	val     any
	err     error
	expires time.Time
}

// responseCache memoises decoded, sanitised responses, and failures, and
// coalesces concurrent misses so a burst of readers costs one RPC and one
// decode. Cached values are shared: callers must treat them as read-only.
type responseCache struct {
	group   singleflight.Group
	mu      sync.Mutex
	entries map[string]cacheEntry
}

func newResponseCache() *responseCache {
	return &responseCache{entries: make(map[string]cacheEntry)}
}

// cached returns the value for key, loading it once for all concurrent
// callers. The shared load is detached from the request that started it, so
// one closed tab cannot fail the others.
func cached[T any](ctx context.Context, c *responseCache, key string, load func(context.Context) (T, error)) (T, error) {
	var zero T
	if e, ok := c.lookup(key); ok {
		if e.err != nil {
			return zero, e.err
		}
		return e.val.(T), nil
	}

	ch := c.group.DoChan(key, func() (any, error) {
		// A load that finished just before this one started is reused.
		if e, ok := c.lookup(key); ok {
			return e.val, e.err
		}
		lctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), fetchTimeout)
		defer cancel()
		v, err := load(lctx)
		c.put(key, v, err)
		return v, err
	})

	select {
	case res := <-ch:
		if res.Err != nil {
			return zero, res.Err
		}
		return res.Val.(T), nil
	case <-ctx.Done():
		return zero, ctx.Err()
	}
}

func (c *responseCache) lookup(key string) (cacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	return e, ok && time.Now().Before(e.expires)
}

func (c *responseCache) put(key string, v any, err error) {
	now := time.Now()
	ttl := cacheTTL
	if err != nil {
		ttl = errorTTL
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= cacheMaxEntries {
		for k, e := range c.entries {
			if now.After(e.expires) {
				delete(c.entries, k)
			}
		}
		if len(c.entries) >= cacheMaxEntries {
			clear(c.entries)
		}
	}
	c.entries[key] = cacheEntry{val: v, err: err, expires: now.Add(ttl)}
}
