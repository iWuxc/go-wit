package captcha

import (
	"context"
	"fmt"
	"github.com/iWuxc/go-wit/cache"
	"strings"
	"time"
)

const (
	cacheKey = "go-kit:captcha:%s"
)

var store *captchaStore

type captchaStore struct {
	cacheKey string
	expire   time.Duration
}

func GetCaptchaStore(duration time.Duration) *captchaStore {
	if store == nil {
		store = &captchaStore{
			cacheKey: cacheKey,
			expire:   duration,
		}
	}

	store.expire = duration
	return store
}

func (c *captchaStore) Set(id string, value string) error {
	err := cache.Set(context.Background(), fmt.Sprintf(c.cacheKey, id), value, c.expire)
	if err != nil {
		return err
	}

	return nil
}

func (c *captchaStore) Get(id string, clear bool) string {
	val, _ := cache.Get(context.Background(), fmt.Sprintf(c.cacheKey, id))
	if clear {
		_ = cache.Delete(context.Background(), fmt.Sprintf(c.cacheKey, id))
	}

	return val
}

func (c *captchaStore) Verify(id, answer string, clear bool) (result bool) {
	result = strings.TrimSpace(c.Get(id, clear)) == strings.TrimSpace(answer)
	if clear {
		_ = cache.Delete(context.Background(), fmt.Sprintf(c.cacheKey, id))
	}

	return
}
