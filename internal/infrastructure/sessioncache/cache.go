package sessioncache

import (
	"sync"
	"time"
)

type Cache interface {
	Get(key string) (string, bool)
	Set(key, value string, ttl time.Duration)
}

type item struct {
	value     string
	expiresAt time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]item
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]item),
	}
}

func (c *MemoryCache) Get(key string) (string, bool) {
	c.mu.RLock()
	it, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return "", false
	}

	if time.Now().After(it.expiresAt) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return "", false
	}

	return it.value, true
}

func (c *MemoryCache) Set(key, value string, ttl time.Duration) {
	c.mu.Lock()
	c.items[key] = item{value: value, expiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
}
