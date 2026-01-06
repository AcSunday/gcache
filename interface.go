package gcache

import "time"

type ICache interface {
	Has(key string) bool
	Get(key string) []byte
	Set(key string, value []byte) error
	Delete(key string) error

	Close() error
}

type ICacheWithTTL interface {
	Has(key string) bool
	Get(key string) []byte
	Set(key string, value []byte, ttl time.Duration) error
	Delete(key string) error

	Close() error
}
