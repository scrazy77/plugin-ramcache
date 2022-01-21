// Package plugin-ramcache is a plugin to cache responses to memory.
package plugin_ramcache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type ramCache struct {
	expire  int
	handler *cache.Cache
}

func newRAMCache(expire int) (*ramCache, error) {
	cacheHandler := cache.New(time.Duration(expire)*time.Second, 5*time.Minute)
	rc := &ramCache{
		expire:  expire,
		handler: cacheHandler,
	}
	return rc, nil
}

func (c *ramCache) Get(key string) ([]byte, bool) {
	value, found := c.handler.Get(key)
	if found {
		return value.([]byte), found
	} else {
		return nil, found
	}
}

func (c *ramCache) Set(key string, val []byte, expire int) error {
	c.handler.Set(key, val, time.Duration(expire)*time.Second)
	return nil
}
