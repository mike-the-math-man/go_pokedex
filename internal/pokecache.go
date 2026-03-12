package internal

import (
	"sync"
	"time"
)

type CacheEntry struct {
	createdAt time.Time
	value     []byte
}

type Cache struct {
	mu         sync.Mutex
	cached_map map[string]CacheEntry
}

func NewCache(interval time.Duration) *Cache {
	new_cache := &Cache{
		sync.Mutex{},
		map[string]CacheEntry{},
	}
	go new_cache.ReapLoop(interval)
	return new_cache
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.cached_map[key]
	if !ok {
		return []byte{}, false
	} else {
		return value.value, true
	}
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cached_map[key] = CacheEntry{time.Now(), val}
}

func (c *Cache) ReapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)

	for range ticker.C {
		c.mu.Lock()

		for key, value := range c.cached_map {
			if time.Since(value.createdAt) > interval {
				delete(c.cached_map, key)
			}
		}

		c.mu.Unlock()
	}

}
