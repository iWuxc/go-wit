package redis

import (
	"context"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/utils"
	"github.com/go-redis/redis/v8"
)

func (r *Redis) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return r.client.MGet(ctx, keys...).Result()
}

// MSet is like Set but accepts multiple values:
//   - MSet("key1", "value1", "key2", "value2")
//   - MSet([]string{"key1", "value1", "key2", "value2"})
//   - MSet(map[string]interface{}{"key1": "value1", "key2": "value2"})
func (r *Redis) MSet(ctx context.Context, val ...interface{}) (string, error) {
	return r.client.MSet(ctx, val...).Result()
}

// HSet accepts values in following formats:
//   - HSet("myhash", "key1", "value1", "key2", "value2")
//   - HSet("myhash", []string{"key1", "value1", "key2", "value2"})
//   - HSet("myhash", map[string]interface{}{"key1": "value1", "key2": "value2"})
//
// Note that it requires Redis v4 for multiple field/value pairs support.
func (r *Redis) HSet(ctx context.Context, key string, val ...interface{}) (int64, error) {
	return r.client.HSet(ctx, key, val...).Result()
}

func (r *Redis) HGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return r.client.MGet(ctx, keys...).Result()
}

// HMSet is a deprecated version of HSet left for compatibility with Redis 3.
func (r *Redis) HMSet(ctx context.Context, key string, values ...interface{}) (bool, error) {
	return r.client.HMSet(ctx, key, values...).Result()
}

// HMGet returns the values for the specified fields in the hash stored at key.
// It returns an interface{} to distinguish between empty string and nil value.
func (r *Redis) HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, error) {
	return r.client.HMGet(ctx, key, fields...).Result()
}

// HGetAll .
func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

func (r *Redis) Hook(hook redis.Hook) {
	r.client.AddHook(hook)
}

func (r *Redis) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return r.client.LPush(ctx, key, values...).Result()
}

func (r *Redis) LPop(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

func (r *Redis) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return r.client.RPush(ctx, key, values...).Result()
}

func (r *Redis) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

func (r *Redis) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

// Publish posts the message to the channel.
func (r *Redis) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	return r.client.Publish(ctx, channel, message).Result()
}

// Subscribe subscribes the client to the specified channels.
// Channels can be omitted to create empty subscription.
// Note that this method does not wait on a response from Redis, so the
// subscription may not be active immediately. To force the connection to wait,
// you may call the Receive() method on the returned *PubSub like so:
//
//	sub := client.Subscribe(queryResp)
//	iface, err := sub.Receive()
//	if err != nil {
//	    // handle error
//	}
//
//	// Should be *Subscription, but others are possible if other actions have been
//	// taken on sub since it was created.
//	switch iface.(type) {
//	case *Subscription:
//	    // subscribe succeeded
//	case *Message:
//	    // received first message
//	case *Pong:
//	    // pong received
//	default:
//	    // handle error
//	}
//
//	ch := sub.Channel()
func (r *Redis) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

func (r *Redis) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return r.client.SAdd(ctx, key, members...).Result()
}

// ZAdd `ZADD key score member [score member ...]` command.
func (r *Redis) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	return r.client.ZAdd(ctx, key, members...).Result()
}

// ZAddXX `ZADD key XX score member [score member ...]` command.
func (r *Redis) ZAddXX(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	return r.client.ZAddXX(ctx, key, members...).Result()
}

func (r *Redis) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return r.client.ZRem(ctx, key, members...).Result()
}

func (r *Redis) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

type RangeBy struct {
	Min, Max      string
	Offset, Count int64
}

func (r *Redis) ZRangeByScore(ctx context.Context, key string, opt RangeBy) ([]string, error) {
	var rangeBy redis.ZRangeBy
	if err := utils.Copy(opt, &rangeBy); err != nil {
		return nil, errors.Wrap(err, "Redis:ZRangeByScore")
	}
	return r.client.ZRangeByScore(ctx, key, &rangeBy).Result()
}

func (r *Redis) RedisZ(members map[string]float64) []*redis.Z {
	var rz []*redis.Z
	for member, score := range members {
		rz = append(rz, &redis.Z{Member: member, Score: score})
	}
	return rz
}

// Scripts .
func (r *Redis) Scripts(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	res, err := redis.NewScript(script).Run(ctx, r.client, keys, args...).Result()
	return res, err
}
