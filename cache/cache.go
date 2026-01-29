package cache

import (
	"context"
	"github.com/iWuxc/go-wit/redis"
	"github.com/pkg/errors"
	"sync"
	"time"
)

var (
	c   Cache
	one sync.Once
)

// Cache 缓存相关的实现 .
type Cache interface {
	// Set .
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Get .
	Get(ctx context.Context, key string) (string, error)

	// GetMulti .
	GetMulti(ctx context.Context, keys ...string) ([]interface{}, error)

	// Delete .
	Delete(ctx context.Context, key string) error

	// Incr . Increment a cached int value by key, as a counter.
	Incr(ctx context.Context, key string) (int64, error)

	// Decr . Decrement a cached int value by key, as a counter.
	Decr(ctx context.Context, key string) (int64, error)

	// IsExist .
	IsExist(ctx context.Context, key ...string) (int64, error)

	// Expire replace or get a cached expire by key .
	Expire(ctx context.Context, key string, duration time.Duration) (time.Duration, error)

	// Close .
	Close() error
}

// Set .
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Set(ctx, key, value, expiration)
}

// Get .
func Get(ctx context.Context, key string) (string, error) {
	return c.Get(ctx, key)
}

// GetMulti .
func GetMulti(ctx context.Context, keys ...string) ([]interface{}, error) {
	return c.GetMulti(ctx, keys...)
}

// Delete .
func Delete(ctx context.Context, key string) error {
	return c.Delete(ctx, key)
}

// Incr . Increment a cached int value by key, as a counter.
func Incr(ctx context.Context, key string) (int64, error) {
	return c.Incr(ctx, key)
}

// Decr . Decrement a cached int value by key, as a counter.
func Decr(ctx context.Context, key string) (int64, error) {
	return c.Decr(ctx, key)
}

// IsExist .
func IsExist(ctx context.Context, key ...string) (int64, error) {
	return c.IsExist(ctx, key...)
}

// Expire .
func Expire(ctx context.Context, key string, duration time.Duration) (time.Duration, error) {
	return c.Expire(ctx, key, duration)
}

// Close .
func Close() error {
	return c.Close()
}

func init() {
	one.Do(func() {
		if conn, err := newMemoryCache(); err == nil {
			c = conn
		}
	})
}

func NewCache(driver, conf string) (Cache, error) {
	if c != nil {
		_ = c.Close()
		c = nil
	}

	switch driver {
	case "memory":
		c, _ = newMemoryCache()
		return c, nil

	//case "file":
	//	f, err := newFileCache(conf)
	//	if err != nil {
	//		return nil, err
	//	}
	//	return f, nil

	case "redis":
		conn, err := redis.NewRedis(conf)
		if err != nil {
			return nil, err
		}
		c = conn
		return c, nil

	default:
		return nil, errors.Errorf("Unsupported Cache Driver: %s .", driver)
	}
}
