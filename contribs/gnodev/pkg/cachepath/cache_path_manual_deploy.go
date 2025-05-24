package cachepath

import (
	"sync"
)

type CachePath struct {
	data map[string]bool
	mu   sync.RWMutex
}

var cache *CachePath

func init() {
	cache = &CachePath{
		data: make(map[string]bool),
	}
}

func Set(key string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.data[key] = true
}

func Get(key string) bool {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	return cache.data[key]
}
