package core

import (
	"encoding/json"
	"time"

	"github.com/golang/groupcache/lru"

	"straas.io/sauron"
)

// NewCache creates local in-memory cache
func NewCache(maxEntries int, clock sauron.Clock) sauron.LocalCache {
	return &localCacheImpl{
		cache: lru.New(maxEntries),
		clock: clock,
	}
}

type localCacheImpl struct {
	cache *lru.Cache
	clock sauron.Clock
}

type entry struct {
	// expire unix timestamp
	expireTS int64
	// value store the value
	value interface{}
}

// Get returns data from store or generates
// new one if cache miss or expired
func (c *localCacheImpl) Get(ns, key string,
	ttl time.Duration, gen sauron.CacheGenerator) (interface{}, error) {
	k := c.makeKey(ns, key)
	raw, ok := c.cache.Get(k)
	if ok {
		ent := raw.(*entry)
		if ent.expireTS == 0 || c.clock.Now().Unix() < ent.expireTS {
			return ent.value, nil
		}
	}
	// expire or not exist
	v, err := gen()
	if err != nil {
		return nil, err
	}
	expireTS := int64(0)
	if ttl > 0 {
		expireTS = c.clock.Now().Add(ttl).Unix()
	}

	c.cache.Add(k, &entry{
		expireTS: expireTS,
		value:    v,
	})
	return v, nil
}

// makeKey generates map key according to namespace and key
func (c *localCacheImpl) makeKey(ns, key string) string {
	// use JSON marshal
	bs, _ := json.Marshal([]string{ns, key})
	return string(bs)
}
