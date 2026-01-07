package gcache

import (
	"encoding/binary"
	"time"
)

type CacheWithTTL struct {
	ICache
}

func NewCacheWithTTL(maxBytes int) ICacheWithTTL {
	return &CacheWithTTL{
		ICache: NewCache(maxBytes),
	}
}

func (c *CacheWithTTL) Has(key string) bool {
	_, ok := unwrapCacheWithTTL(c.ICache.Get(key))
	return ok
}

func (c *CacheWithTTL) Get(key string) []byte {
	data, ok := unwrapCacheWithTTL(c.ICache.Get(key))
	if !ok {
		return nil
	}
	return data
}

func (c *CacheWithTTL) Set(key string, value []byte, ttl time.Duration) error {
	value = wrapCacheWithTTL(value, ttl)
	return c.ICache.Set(key, value)
}

func (c *CacheWithTTL) Delete(key string) error {
	return c.ICache.Delete(key)
}

func (c *CacheWithTTL) Close() error {
	return c.ICache.Close()
}

// wrapCacheWithTTL wrap data with ttl
func wrapCacheWithTTL(data []byte, ttl time.Duration) []byte {
	expireAt := time.Now().Add(ttl).UnixMilli()
	buf := make([]byte, 8+len(data))
	binary.BigEndian.PutUint64(buf[:8], uint64(expireAt))
	copy(buf[8:], data)
	return buf
}

// unwrapCacheWithTTL unwrap data with ttl
func unwrapCacheWithTTL(data []byte) ([]byte, bool) {
	if len(data) < 8 {
		return nil, false
	}

	expireAt := binary.BigEndian.Uint64(data[:8])
	if time.Now().UnixMilli() >= int64(expireAt) {
		return nil, false
	}
	return data[8:], true
}
