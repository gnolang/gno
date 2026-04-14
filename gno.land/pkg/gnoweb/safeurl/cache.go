package safeurl

import (
	"container/list"
	"sync"
	"time"
)

// Cache implements a thread-safe LRU cache with TTL for scan results.
type Cache struct {
	mu      sync.RWMutex
	items   map[string]*cacheEntry
	order   *list.List // LRU order (front = most recent)
	maxSize int
	ttl     time.Duration
}

type cacheEntry struct {
	result  ScanResult
	element *list.Element
}

// NewCache creates a new cache with the specified maximum size and TTL.
func NewCache(maxSize int, ttl time.Duration) *Cache {
	if maxSize <= 0 {
		maxSize = 10000
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &Cache{
		items:   make(map[string]*cacheEntry),
		order:   list.New(),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// Get retrieves a scan result from the cache.
// Returns the result and true if found and not expired, otherwise zero value and false.
func (c *Cache) Get(url string) (ScanResult, bool) {
	c.mu.RLock()
	entry, ok := c.items[url]
	c.mu.RUnlock()

	if !ok {
		return ScanResult{}, false
	}

	// Check if expired
	if entry.result.IsExpired() {
		c.mu.Lock()
		c.removeEntry(url)
		c.mu.Unlock()
		return ScanResult{}, false
	}

	// Move to front (most recently used)
	c.mu.Lock()
	c.order.MoveToFront(entry.element)
	c.mu.Unlock()

	return entry.result, true
}

// Set stores a scan result in the cache.
func (c *Cache) Set(url string, result ScanResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Set expiration if not already set
	if result.ExpiresAt.IsZero() {
		result.ExpiresAt = time.Now().Add(c.ttl)
	}

	// Update existing entry
	if entry, ok := c.items[url]; ok {
		entry.result = result
		c.order.MoveToFront(entry.element)
		return
	}

	// Evict oldest entries if at capacity
	for c.order.Len() >= c.maxSize {
		c.evictOldest()
	}

	// Add new entry
	element := c.order.PushFront(url)
	c.items[url] = &cacheEntry{
		result:  result,
		element: element,
	}
}

// GetMulti retrieves multiple scan results from the cache.
// Returns a map of found results and a slice of missing URLs.
func (c *Cache) GetMulti(urls []string) (found map[string]ScanResult, missing []string) {
	found = make(map[string]ScanResult)
	missing = make([]string, 0)

	for _, url := range urls {
		if result, ok := c.Get(url); ok {
			found[url] = result
		} else {
			missing = append(missing, url)
		}
	}

	return found, missing
}

// SetMulti stores multiple scan results in the cache.
func (c *Cache) SetMulti(results map[string]ScanResult) {
	for url, result := range results {
		c.Set(url, result)
	}
}

// Len returns the number of entries in the cache.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*cacheEntry)
	c.order.Init()
}

// evictOldest removes the least recently used entry.
// Caller must hold the write lock.
func (c *Cache) evictOldest() {
	oldest := c.order.Back()
	if oldest == nil {
		return
	}
	url := oldest.Value.(string)
	c.removeEntry(url)
}

// removeEntry removes an entry by URL.
// Caller must hold the write lock.
func (c *Cache) removeEntry(url string) {
	entry, ok := c.items[url]
	if !ok {
		return
	}
	c.order.Remove(entry.element)
	delete(c.items, url)
}
