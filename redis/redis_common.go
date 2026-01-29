package redis

import (
	"context"
	"github.com/iWuxc/go-wit/errors"
	"github.com/go-redis/redis/v8"
	"sync"
)

const global = "common"

var (
	_redis     *Redis
	_redisMap  = map[string]*Redis{}
	_redisLock = sync.RWMutex{}
)

// NewRedis create new Cache .
// @params redisUrl redis://<user>:<password>@<host>:<port>/<db_number>
// 					redis://:password@localhost:6379/1?dial_timeout=3&read_timeout=6s&max_retries=2
func NewRedis(redisURL string, name ...string) (*Redis, error) {
	if len(name) == 0 {
		name = append(name, global)
	}

	_redisLock.Lock()
	defer _redisLock.Unlock()

	if r, ok := _redisMap[name[0]]; ok {
		return r, nil
	}

	r, err := newRedis(redisURL)
	if err != nil {
		return nil, err
	}

	_redisMap[name[0]] = r
	return r, nil
}

// NewRedis create new Cache .
// @params redisUrl redis://<user>:<password>@<host>:<port>/<db_number>
// 					redis://:password@localhost:6379/1?dial_timeout=3&read_timeout=6s&max_retries=2
func newRedis(redisURL string) (*Redis, error) {
	if len(redisURL) == 0 {
		return nil, errors.New("please set redis_conf environment variable. eg: export redis_conf=redis://127.0.0.1:6379/1?dial_timeout=3&read_timeout=6s&max_retries=2")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, errors.Wrap(err, "parse redis conf")
	}
	
	client := redis.NewClient(opt)
	if e := client.Ping(context.Background()).Err(); e != nil {
		return nil, errors.Wrap(e, "redis connect")
	}

	return &Redis{client: client}, nil
}

func getRedis(name string) *Redis {
	_redisLock.RLock()
	defer _redisLock.RUnlock()

	if r, ok := _redisMap[name]; ok {
		return r
	}

	return nil
}
