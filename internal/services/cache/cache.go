package cache

import (
	"sync"

	"flight-booking/config"
	"github.com/patrickmn/go-cache"
)

type Cache interface {
	GetOrLoad(key string, loader func() (any, error)) (any, error)
}

type inMemoryCache struct {
	cfg   config.CacheConfig
	cache *cache.Cache
	mu    sync.RWMutex
}

func New(cfg config.CacheConfig) Cache {
	return &inMemoryCache{
		cfg:   cfg,
		cache: cache.New(cfg.DefaultTTL, cfg.CleanupInterval),
	}
}

func (c *inMemoryCache) GetOrLoad(key string, loader func() (any, error)) (any, error) {
	if !c.cfg.Enabled {
		return nil, nil
	}

	c.mu.RLock()
	value, found := c.cache.Get(key)
	c.mu.RUnlock()

	if found {
		return value, nil
	}

	newValue, err := loader()
	if err != nil {
		return nil, err
	}

	c.cache.Set(key, newValue, c.cfg.DefaultTTL)
	return newValue, nil
}
