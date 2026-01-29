package cache

import (
	"context"
	"fmt"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/utils"
	"github.com/patrickmn/go-cache"
	"os"
	"time"
)

var cacheFileName string

var _ Cache = (*fileCache)(nil)

type fileCache struct {
	client *cache.Cache
}

// newFileCache .
func newFileCache(file string) (Cache, error) {
	client := cache.New(defaultExpiration, cleanupInterval)
	cacheFileName = file
	if err := client.LoadFile(cacheName()); err != nil {
		return nil, errors.Wrap(err, "file cache")
	}
	return &fileCache{client: client}, nil
}

func (f *fileCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	f.client.Set(key, value, expiration)
	return nil
}

func (f *fileCache) Get(ctx context.Context, key string) (string, error) {
	val, ok := f.client.Get(key)
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

func (f *fileCache) GetMulti(ctx context.Context, keys ...string) ([]interface{}, error) {
	var r []interface{}
	for _, key := range keys {
		v, _ := f.client.Get(key)
		r = append(r, v)
	}

	return r, nil
}

func (f *fileCache) Delete(ctx context.Context, key string) error {
	f.client.Delete(key)
	return nil
}

func (f *fileCache) Incr(ctx context.Context, key string) (int64, error) {
	return f.client.IncrementInt64(key, 1)
}

func (f *fileCache) Decr(ctx context.Context, key string) (int64, error) {
	return f.client.DecrementInt64(key, 1)
}

func (f *fileCache) IsExist(ctx context.Context, keys ...string) (int64, error) {
	var r []interface{}
	for _, key := range keys {
		if v, ok := f.client.Get(key); ok {
			r = append(r, v)
		}
	}

	return int64(len(r)), nil
}

func (f *fileCache) Close() error {
	if err := f.client.SaveFile(cacheName()); err != nil {
		return err
	}
	return nil
}

func (f *fileCache) Expire(ctx context.Context, key string, duration time.Duration) (time.Duration, error) {
	val, t, ok := f.client.GetWithExpiration(key)
	if !ok {
		return 0, errors.New("cache key not found")
	}

	if duration == 0 {
		return t.Sub(time.Now()), nil
	}

	return duration, f.client.Replace(key, val, duration)
}

func cacheName() string {
	if len(cacheFileName) == 0 {
		if cacheFileName = os.Getenv("FILE_CACHE_PATH"); len(cacheFileName) == 0 {
			cacheFileName = "cache.db"
		}
	}

	if !utils.FileExist(cacheFileName) {
		f, _ := os.OpenFile(cacheFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		_ = f.Close()
	}

	return cacheFileName
}
