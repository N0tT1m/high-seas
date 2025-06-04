package cache

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Cache represents an in-memory cache with TTL
type Cache struct {
	items   map[string]*CacheItem
	mutex   sync.RWMutex
	ttl     time.Duration
	cleanup chan bool
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	Items       int       `json:"items"`
	HitRate     float64   `json:"hit_rate"`
	LastCleanup time.Time `json:"last_cleanup"`
}

var (
	globalCache *Cache
	stats       CacheStats
	statsMutex  sync.RWMutex
)

// New creates a new cache instance
func New(ttl time.Duration) *Cache {
	cache := &Cache{
		items:   make(map[string]*CacheItem),
		ttl:     ttl,
		cleanup: make(chan bool),
	}

	// Start cleanup goroutine
	go cache.startCleanup()

	return cache
}

// GetGlobalCache returns the global cache instance
func GetGlobalCache() *Cache {
	if globalCache == nil {
		globalCache = New(1 * time.Hour) // Default 1 hour TTL
	}
	return globalCache
}

// InitGlobalCache initializes the global cache with custom TTL
func InitGlobalCache(ttl time.Duration) {
	if globalCache != nil {
		globalCache.Stop()
	}
	globalCache = New(ttl)
}

// Set stores an item in the cache
func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = &CacheItem{
		Data:      value,
		ExpiresAt: time.Now().Add(c.ttl),
		CreatedAt: time.Now(),
	}
}

// SetWithTTL stores an item in the cache with custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = &CacheItem{
		Data:      value,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		c.incrementMisses()
		return nil, false
	}

	// Check if item has expired
	if time.Now().After(item.ExpiresAt) {
		c.mutex.RUnlock()
		c.mutex.Lock()
		delete(c.items, key)
		c.mutex.Unlock()
		c.mutex.RLock()
		c.incrementMisses()
		return nil, false
	}

	c.incrementHits()
	return item.Data, true
}

// GetString retrieves a string from the cache
func (c *Cache) GetString(key string) (string, bool) {
	if data, exists := c.Get(key); exists {
		if str, ok := data.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetJSON retrieves and unmarshals JSON data from the cache
func (c *Cache) GetJSON(key string, target interface{}) bool {
	if data, exists := c.Get(key); exists {
		if jsonData, ok := data.([]byte); ok {
			if err := json.Unmarshal(jsonData, target); err == nil {
				return true
			}
		}
	}
	return false
}

// SetJSON marshals and stores JSON data in the cache
func (c *Cache) SetJSON(key string, value interface{}) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	c.Set(key, jsonData)
	return nil
}

// SetJSONWithTTL marshals and stores JSON data in the cache with custom TTL
func (c *Cache) SetJSONWithTTL(key string, value interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	c.SetWithTTL(key, jsonData, ttl)
	return nil
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.items, key)
}

// Exists checks if a key exists in the cache (without affecting hit/miss stats)
func (c *Cache) Exists(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return false
	}

	// Check if item has expired
	if time.Now().After(item.ExpiresAt) {
		return false
	}

	return true
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items = make(map[string]*CacheItem)
}

// Size returns the number of items in the cache
func (c *Cache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
}

// Keys returns all keys in the cache
func (c *Cache) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

// GetStats returns cache statistics
func (c *Cache) GetStats() CacheStats {
	statsMutex.RLock()
	defer statsMutex.RUnlock()

	currentStats := stats
	currentStats.Items = c.Size()

	// Calculate hit rate
	total := currentStats.Hits + currentStats.Misses
	if total > 0 {
		currentStats.HitRate = float64(currentStats.Hits) / float64(total) * 100
	}

	return currentStats
}

// ResetStats resets cache statistics
func (c *Cache) ResetStats() {
	statsMutex.Lock()
	defer statsMutex.Unlock()
	stats = CacheStats{}
}

// Stop stops the cache cleanup goroutine
func (c *Cache) Stop() {
	close(c.cleanup)
}

// startCleanup runs periodic cleanup of expired items
func (c *Cache) startCleanup() {
	ticker := time.NewTicker(5 * time.Minute) // Clean up every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.cleanup:
			return
		}
	}
}

// cleanupExpired removes expired items from the cache
func (c *Cache) cleanupExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	expired := make([]string, 0)

	for key, item := range c.items {
		if now.After(item.ExpiresAt) {
			expired = append(expired, key)
		}
	}

	for _, key := range expired {
		delete(c.items, key)
	}

	// Update cleanup stats
	statsMutex.Lock()
	stats.LastCleanup = now
	statsMutex.Unlock()
}

// incrementHits increments the hit counter
func (c *Cache) incrementHits() {
	statsMutex.Lock()
	defer statsMutex.Unlock()
	stats.Hits++
}

// incrementMisses increments the miss counter
func (c *Cache) incrementMisses() {
	statsMutex.Lock()
	defer statsMutex.Unlock()
	stats.Misses++
}

// Utility functions for global cache

// Set stores an item in the global cache
func Set(key string, value interface{}) {
	GetGlobalCache().Set(key, value)
}

// SetWithTTL stores an item in the global cache with custom TTL
func SetWithTTL(key string, value interface{}, ttl time.Duration) {
	GetGlobalCache().SetWithTTL(key, value, ttl)
}

// Get retrieves an item from the global cache
func Get(key string) (interface{}, bool) {
	return GetGlobalCache().Get(key)
}

// GetString retrieves a string from the global cache
func GetString(key string) (string, bool) {
	return GetGlobalCache().GetString(key)
}

// GetJSON retrieves and unmarshals JSON data from the global cache
func GetJSON(key string, target interface{}) bool {
	return GetGlobalCache().GetJSON(key, target)
}

// SetJSON marshals and stores JSON data in the global cache
func SetJSON(key string, value interface{}) error {
	return GetGlobalCache().SetJSON(key, value)
}

// SetJSONWithTTL marshals and stores JSON data in the global cache with custom TTL
func SetJSONWithTTL(key string, value interface{}, ttl time.Duration) error {
	return GetGlobalCache().SetJSONWithTTL(key, value, ttl)
}

// Delete removes an item from the global cache
func Delete(key string) {
	GetGlobalCache().Delete(key)
}

// Exists checks if a key exists in the global cache
func Exists(key string) bool {
	return GetGlobalCache().Exists(key)
}

// Clear removes all items from the global cache
func Clear() {
	GetGlobalCache().Clear()
}

// Size returns the number of items in the global cache
func Size() int {
	return GetGlobalCache().Size()
}

// Keys returns all keys in the global cache
func Keys() []string {
	return GetGlobalCache().Keys()
}

// GetStats returns global cache statistics
func GetStats() CacheStats {
	return GetGlobalCache().GetStats()
}

// ResetStats resets global cache statistics
func ResetStats() {
	GetGlobalCache().ResetStats()
}
