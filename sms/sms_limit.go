package sms

import (
	"context"
	"fmt"
	"github.com/iWuxc/go-wit/cache"
	"github.com/iWuxc/go-wit/errors"
	"strconv"
	"time"
)

type RateLimitInterface interface {
	Limit(mobile string) error
	Clean(mobile string)
}

const (
	// go-kit:sms:minute:mobile  手机号x分钟内是否有发送过短信 (间隔一分钟)
	smsMinuteCacheKey = "go-kit:sms:minute:%s"
	// go-kit:sms:day:mobile  手机号某段时间内短信的发送次数 (每天限制发送 x 次)
	smsDayCacheKey = "go-kit:sms:%s:%s"
)

var (
	ERRLimit = errors.New("This Mobile has be limited")
)

type RateLimit struct {
	SendMinute   time.Duration `json:"send_minute"`
	RateForDay   time.Duration `json:"rate_for_day"`
	RateCacheKey string        `json:"rate_cache_key"`
	LimitOFDay   int           `json:"limit_of_day"`
}

func defaultRateLimit() *RateLimit {
	return &RateLimit{
		// 每分钟可发送一次
		SendMinute: time.Minute,
		// 每天最多可以发送十次
		RateForDay:   time.Minute * 60 * 24,
		RateCacheKey: time.Now().Format("2006-01-02"),
		LimitOFDay:   10,
	}
}

// Limit .
func (r *RateLimit) Limit(mobile string) error {
	if r.CanSend(mobile) {
		cache.Set(context.Background(), fmt.Sprintf(smsMinuteCacheKey, mobile), 1, r.SendMinute)
		count, _ := cache.IsExist(context.Background(), fmt.Sprintf(smsDayCacheKey, r.RateCacheKey, mobile))
		if count > 0 {
			cache.Incr(context.Background(), fmt.Sprintf(smsDayCacheKey, r.RateCacheKey, mobile))
			return nil
		}

		cache.Set(context.Background(), fmt.Sprintf(smsDayCacheKey, r.RateCacheKey, mobile), 1, r.RateForDay)
		return nil
	}

	return ERRLimit
}

// CanSend .
func (r *RateLimit) CanSend(mobile string) bool {
	count, _ := cache.IsExist(context.Background(), fmt.Sprintf(smsMinuteCacheKey, mobile))
	val, err := cache.Get(context.Background(), fmt.Sprintf(smsDayCacheKey, r.RateCacheKey, mobile))
	if err != nil && err != errors.ERRMissingCacheKey {
		return false
	}

	v, _ := strconv.Atoi(val)
	if count == 0 && v <= r.LimitOFDay {
		return true
	}

	return false
}

func (r *RateLimit) Clean(mobile string) {
	cache.Delete(context.Background(), fmt.Sprintf(smsMinuteCacheKey, mobile))
	cache.Delete(context.Background(), fmt.Sprintf(smsDayCacheKey, r.RateCacheKey, mobile))
}
