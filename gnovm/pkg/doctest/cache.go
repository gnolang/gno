package doctest

import (
	"container/list"
	"sync"
)

const maxCacheSize = 25

type cacheItem struct {
	key   string
	value string
}

type lruCache struct {
	capacity int
	items    map[string]*list.Element
	order    *list.List
	mutex    sync.RWMutex
}

func newCache(capacity int) *lruCache {
	return &lruCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

func (c *lruCache) get(key string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		return elem.Value.(cacheItem).value, true
	}
	return "", false
}

func (c *lruCache) set(key, value string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		elem.Value = cacheItem{key, value}
	} else {
		if c.order.Len() >= c.capacity {
			oldest := c.order.Back()
			if oldest != nil {
				delete(c.items, oldest.Value.(cacheItem).key)
				c.order.Remove(oldest)
			}
		}
		elem := c.order.PushFront(cacheItem{key, value})
		c.items[key] = elem
	}
}
