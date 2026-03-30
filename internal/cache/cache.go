package cache

import (
	"sync"
	"time"
)

type entry struct {
	value     any
	expiresAt time.Time
}

// Cache is a simple in-memory TTL cache safe for concurrent use.
type Cache struct {
	mu   sync.RWMutex
	data map[string]entry
}

func New() *Cache {
	c := &Cache{data: make(map[string]entry)}
	go c.evictLoop()
	return c
}

func (c *Cache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	c.data[key] = entry{value: value, expiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	e, ok := c.data[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}

func (c *Cache) evictLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		c.mu.Lock()
		for k, e := range c.data {
			if now.After(e.expiresAt) {
				delete(c.data, k)
			}
		}
		c.mu.Unlock()
	}
}
