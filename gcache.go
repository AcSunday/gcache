package gcache

import (
	"sync"

	"github.com/VictoriaMetrics/fastcache"
)

type Cache struct {
	pool  *sync.Pool
	cache *fastcache.Cache
}

func newSyncPool() *sync.Pool {
	return &sync.Pool{ // support memory 0 alloc
		New: func() any {
			b := make([]byte, 0, 1024)
			return &b
		},
	}
}

// NewCache based on fastcache, support small object < 64KB
func NewCache(maxBytes int) ICache {
	return &Cache{
		pool:  newSyncPool(),
		cache: fastcache.New(maxBytes),
	}
}

func (c *Cache) Has(key string) bool {
	return c.cache.Has([]byte(key))
}

func (c *Cache) Get(key string) []byte {
	bkey := []byte(key)

	// get buffer from pool
	buf := c.pool.Get().(*[]byte)
	dst := (*buf)[:0]

	dst = c.cache.Get(dst, bkey)
	if dst == nil {
		c.pool.Put(buf)
		return nil
	}

	// copy to new output buffer
	out := make([]byte, len(dst))
	copy(out, dst)

	c.pool.Put(buf)
	return out
}

func (c *Cache) Set(key string, value []byte) error {
	c.cache.Set([]byte(key), value)
	return nil
}

func (c *Cache) Delete(key string) error {
	c.cache.Del([]byte(key))
	return nil
}

func (c *Cache) Close() error {
	c.cache.Reset()
	return nil
}
