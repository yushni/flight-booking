package cache

import (
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

type Cache interface {
	GetOrLoad(key string, ttl time.Duration, loader func() (any, error)) (any, error)
}

type inMemoryCache struct {
	cache *cache.Cache
	mu    sync.RWMutex
}

func New() Cache {
	return &inMemoryCache{
		cache: cache.New(time.Second, time.Second),
	}
}

func (c *inMemoryCache) GetOrLoad(key string, ttl time.Duration, loader func() (any, error)) (any, error) {
	c.mu.RLock()
	value, found := c.cache.Get(key)
	c.mu.RUnlock()

	if found {
		return value, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	newValue, err := loader()
	if err != nil {
		return nil, err
	}

	c.cache.Set(key, newValue, ttl)

	return newValue, nil
}
