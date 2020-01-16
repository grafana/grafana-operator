//Package memory provides an in-memory cache implementation for https://github.com/databus23/keystone
package memory

import (
	"encoding/json"
	"time"

	"github.com/pmylund/go-cache"
	"github.com/databus23/keystone"
)

type memoryCache struct {
	*cache.Cache
}

//New creates a new cache.
func New(cleanupInterval time.Duration) keystone.Cache {
	return &memoryCache{cache.New(5*time.Minute, cleanupInterval)}
}

func (m *memoryCache) Set(k string, x interface{}, ttl time.Duration) {
	if b, err := json.Marshal(x); err == nil {
		m.Cache.Set(k, b, ttl)
	}
}
func (m *memoryCache) Get(k string, x interface{}) bool {
	if b, ok := m.Cache.Get(k); ok {
		return json.Unmarshal(b.([]byte), x) == nil
	}
	return false
}
