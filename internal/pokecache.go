package internal

import (
	"sync"
	"time"
)

type Cache struct {
	mu       sync.Mutex
	cache    map[string]cacheEntry
	interval time.Duration
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		cache:    make(map[string]cacheEntry),
		interval: interval,
	}

	go c.reapLoop()

	return c
}

// Adds new entry to the cache
func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a new cache entry
	entry := cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}

	c.cache[key] = entry
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, exists := c.cache[key]

	if exists {
		return entry.val, true
	}

	return nil, false
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		<-ticker.C // wait for ticker to cik

		c.mu.Lock()
		now := time.Now()

		for key, entry := range c.cache {
			// Check if entry is older than reap timer
			if now.Sub(entry.createdAt) > c.interval {
				delete(c.cache, key)
			}
		}

		c.mu.Unlock()
	}
}
