package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/macrolens/backend/internal/domain"
)

// cacheItem represents a single item in the cache with expiration
type cacheItem struct {
	Value      interface{}
	Expiration time.Time
}

// MemoryCache is a thread-safe in-memory cache with TTL support
type MemoryCache struct {
	data  map[string]cacheItem
	mutex sync.RWMutex
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]cacheItem),
	}

	// Start cleanup goroutine to remove expired entries every 10 minutes
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a value from the cache
func (c *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, domain.ErrCacheMiss
	}

	// Check if expired
	if time.Now().After(item.Expiration) {
		return nil, domain.ErrCacheMiss
	}

	return item.Value, nil
}

// Set stores a value in the cache with TTL
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Serialize to JSON and back to ensure consistent data structure
	// This mimics Redis behavior
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	var storedValue interface{}
	if err := json.Unmarshal(jsonData, &storedValue); err != nil {
		return err
	}

	c.data[key] = cacheItem{
		Value:      storedValue,
		Expiration: time.Now().Add(ttl),
	}

	return nil
}

// Delete removes a value from the cache
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.data, key)
	return nil
}

// Exists checks if a key exists in the cache and is not expired
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return false, nil
	}

	// Check if expired
	if time.Now().After(item.Expiration) {
		return false, nil
	}

	return true, nil
}

// cleanupExpired removes expired entries from the cache periodically
func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for key, item := range c.data {
			if now.After(item.Expiration) {
				delete(c.data, key)
			}
		}
		c.mutex.Unlock()
	}
}

// Size returns the current number of items in the cache (for debugging/monitoring)
func (c *MemoryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.data)
}

// Clear removes all items from the cache
func (c *MemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]cacheItem)
}
