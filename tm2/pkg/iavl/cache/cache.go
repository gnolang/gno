package cache

import (
	lru "github.com/hashicorp/golang-lru/v2"

	ibytes "github.com/gnolang/gno/tm2/pkg/iavl/internal/bytes"
)

// Node represents a node eligible for caching.
type Node interface {
	GetKey() []byte
}

// Cache is an in-memory structure to persist nodes for quick access.
// Please see lruCache for more details about why we need a custom
// cache implementation.
type Cache interface {
	// Adds node to cache.
	// Returns true of eviction occurred, false otherwise.
	// CONTRACT: node can never be nil. Otherwise, cache panics.
	Add(node Node) bool

	// Returns Node for the key, if exists. nil otherwise.
	Get(key []byte) Node

	// Has returns true if node with key exists in cache, false otherwise.
	Has(key []byte) bool

	// Remove removes node with key from cache.
	// Returns true if removal occurred, false otherwise.
	Remove(key []byte) bool

	// Len returns the cache length.
	Len() int
}

// lruCache is an LRU cache implementation.
// The motivation for using a custom cache implementation is to
// allow for a custom max policy.
//
// Currently, the cache maximum is implemented in terms of the
// number of nodes which is not intuitive to configure.
// Instead, we are planning to add a byte maximum.
// The alternative implementations do not allow for
// customization and the ability to estimate the byte
// size of the cache.
type lruCache struct {
	c *lru.Cache[string, Node]
}

var _ Cache = (*lruCache)(nil)

func New(maxElementCount int) Cache {
	if maxElementCount <= 0 {
		return &lruCache{} // disabled cache
	}
	c, err := lru.New[string, Node](maxElementCount)
	if err != nil {
		panic(err)
	}
	return &lruCache{c}
}

func (c *lruCache) Add(node Node) bool {
	if c.c == nil {
		return false // cache is disabled
	}
	key := ibytes.UnsafeBytesToStr(node.GetKey())
	return c.c.Add(key, node)
}

func (c *lruCache) Get(key []byte) Node {
	if c.c == nil {
		return nil // cache is disabled
	}
	n, ok := c.c.Get(ibytes.UnsafeBytesToStr(key))
	if !ok {
		return nil
	}
	return n
}

func (c *lruCache) Has(key []byte) bool {
	if c.c == nil {
		return false // cache is disabled
	}
	return c.c.Contains(ibytes.UnsafeBytesToStr(key))
}

func (c *lruCache) Len() int {
	if c.c == nil {
		return 0 // cache is disabled
	}
	return c.c.Len()
}

func (c *lruCache) Remove(key []byte) bool {
	if c.c == nil {
		return false // cache is disabled
	}
	return c.c.Remove(ibytes.UnsafeBytesToStr(key))
}
