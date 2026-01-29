package cache

import (
	"context"
	"fmt"
	"github.com/iWuxc/go-wit/errors"
	"github.com/patrickmn/go-cache"
	"time"
)

const (
	defaultExpiration = time.Minute * 10
	cleanupInterval   = time.Minute * 20
)

var _ Cache = (*memoryCache)(nil)

type memoryCache struct {
	client *cache.Cache
}

// newMemoryCache .
func newMemoryCache() (Cache, error) {
	client := cache.New(defaultExpiration, cleanupInterval)
	return &memoryCache{client: client}, nil
}

func (m *memoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.client.Set(key, value, expiration)
	return nil
}

func (m *memoryCache) Get(ctx context.Context, key string) (string, error) {
	val, ok := m.client.Get(key)
	if !ok {
		return "", errors.ERRMissingCacheKey
	}

	switch val.(type) {
	case []byte:
		return string(val.([]byte)), nil
	case string:
		return val.(string), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", val), nil
	default:
		return "", errors.Errorf("cache key: %s type invalid", key)
	}
}

func (m *memoryCache) GetMulti(ctx context.Context, keys ...string) ([]interface{}, error) {
	var r []interface{}
	for _, key := range keys {
		v, _ := m.client.Get(key)
		r = append(r, v)
	}

	return r, nil
}

func (m *memoryCache) Delete(ctx context.Context, key string) error {
	m.client.Delete(key)
	return nil
}

func (m *memoryCache) Incr(ctx context.Context, key string) (int64, error) {
	return m.client.IncrementInt64(key, 1)
}

func (m *memoryCache) Decr(ctx context.Context, key string) (int64, error) {
	return m.client.DecrementInt64(key, 1)
}

func (m *memoryCache) IsExist(ctx context.Context, keys ...string) (int64, error) {
	var r []interface{}
	for _, key := range keys {
		if v, ok := m.client.Get(key); ok {
			r = append(r, v)
		}
	}

	return int64(len(r)), nil
}

func (m *memoryCache) Expire(ctx context.Context, key string, duration time.Duration) (time.Duration, error) {
	val, t, ok := m.client.GetWithExpiration(key)
	if !ok {
		return 0, errors.New("cache key not found")
	}

	if duration == 0 {
		return t.Sub(time.Now()), nil
	}

	return duration, m.client.Replace(key, val, duration)
}

func (m *memoryCache) Close() error {
	m.client.Flush()
	return nil
}
