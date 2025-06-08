package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

type MemoryCache struct {
	items   map[string]*CacheItem
	mutex   sync.RWMutex
	ttl     time.Duration
	cleanup *time.Ticker
}

func NewMemoryCache(ttl time.Duration) *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*CacheItem),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	cache.cleanup = time.NewTicker(ttl / 2)
	go cache.cleanupExpired()

	return cache
}

func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = &CacheItem{
		Data:      value,
		ExpiresAt: time.Now().Add(c.ttl),
	}

	return nil
}

func (c *MemoryCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists || time.Now().After(item.ExpiresAt) {
		return false, nil
	}

	// Use JSON marshalling for deep copy to avoid pointer issues
	data, err := json.Marshal(item.Data)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(data, dest)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
	return nil
}

func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*CacheItem)
	return nil
}

func (c *MemoryCache) cleanupExpired() {
	for range c.cleanup.C {
		c.mutex.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.ExpiresAt) {
				delete(c.items, key)
			}
		}
		c.mutex.Unlock()
	}
}

func (c *MemoryCache) Close() error {
	c.cleanup.Stop()
	return nil
}

// Cache interface for dependency injection
type Cache interface {
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string, dest interface{}) (bool, error)
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Close() error
}
