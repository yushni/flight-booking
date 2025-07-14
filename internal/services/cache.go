package services

import (
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	"flight-booking/internal/config"
	"flight-booking/internal/interfaces"
)

// MemoryCacheService implements the CacheService interface using go-cache
type MemoryCacheService struct {
	cache   *cache.Cache
	enabled bool
	mu      sync.RWMutex
}

// NewMemoryCacheService creates a new memory cache service
func NewMemoryCacheService(config config.CacheConfig) interfaces.CacheService {
	if !config.Enabled {
		return &MemoryCacheService{
			enabled: false,
		}
	}

	return &MemoryCacheService{
		cache:   cache.New(config.DefaultTTL, config.CleanupInterval),
		enabled: true,
	}
}

// Get retrieves a value from the cache
func (c *MemoryCacheService) Get(key string) (interface{}, bool) {
	if !c.enabled {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.cache.Get(key)
}

// Set stores a value in the cache with TTL
func (c *MemoryCacheService) Set(key string, value interface{}, ttl time.Duration) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Set(key, value, ttl)
}

// Delete removes a value from the cache
func (c *MemoryCacheService) Delete(key string) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Delete(key)
}

// Clear removes all values from the cache
func (c *MemoryCacheService) Clear() {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Flush()
}

// GetStats returns cache statistics
func (c *MemoryCacheService) GetStats() (int, int) {
	if !c.enabled {
		return 0, 0
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.cache.ItemCount(), len(c.cache.Items())
}

// IsEnabled returns whether the cache is enabled
func (c *MemoryCacheService) IsEnabled() bool {
	return c.enabled
}
