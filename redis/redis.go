package redis

import (
	"context"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/utils"
	"github.com/go-redis/redis/v8"
	"os"
	"time"
)

type Redis struct {
	client *redis.Client
}

// GetRedis .
func GetRedis(name ...string) *Redis {
	if len(name) > 0 {
		return getRedis(name[0])
	}

	if _redis != nil {
		return _redis
	}

	var (
		conf string
		err  error
	)
	if c := os.Getenv("redis_conf"); len(c) > 0 {
		conf = c
	}

	_redis, err = NewRedis(conf, global)
	if err != nil {
		panic(err.Error())
	}

	return _redis
}

func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *Redis) TTL(ctx context.Context, key string) time.Duration {
	return r.client.TTL(ctx, key).Val()
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	return err
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil && err == redis.Nil {
		return "", errors.ERRMissingCacheKey
	}
	return value, nil
}

func (r *Redis) GetMulti(ctx context.Context, keys ...string) ([]interface{}, error) {
	res, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *Redis) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

func (r *Redis) IsExist(ctx context.Context, key ...string) (int64, error) {
	return r.client.Exists(ctx, key...).Result()
}

func (r *Redis) Expire(ctx context.Context, key string, duration time.Duration) (time.Duration, error) {
	if duration == 0 {
		return r.TTL(ctx, key), nil
	}

	return duration, r.client.Expire(ctx, key, duration).Err()
}

type Sort struct {
	By            string
	Offset, Count int64
	Get           []string
	Order         string
	Alpha         bool
}

func (r *Redis) Sort(ctx context.Context, key string, sort Sort) ([]string, error) {
	var rSort redis.Sort
	if err := utils.Copy(sort, &rSort); err != nil {
		return nil, errors.Wrap(err, "Redis:Sort")
	}

	return r.client.Sort(ctx, key, &rSort).Result()
}

func (r *Redis) SortInterfaces(ctx context.Context, key string, sort Sort) ([]interface{}, error) {
	var rSort redis.Sort
	if err := utils.Copy(sort, &rSort); err != nil {
		return nil, errors.Wrap(err, "Redis:SortInterfaces")
	}

	return r.client.SortInterfaces(ctx, key, &rSort).Result()
}

func (r *Redis) SortStore(ctx context.Context, key, store string, sort Sort) (int64, error) {
	var rSort redis.Sort
	if err := utils.Copy(sort, &rSort); err != nil {
		return 0, errors.Wrap(err, "Redis:SortStore")
	}

	return r.client.SortStore(ctx, key, store, &rSort).Result()
}

// GetClient .
func (r *Redis) GetClient() *redis.Client {
	return r.client
}

func (r *Redis) Close() error {
	return r.client.Close()
}
