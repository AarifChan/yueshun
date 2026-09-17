package wxmp

import (
	"errors"
	"time"

	"zhizhang-server/internal/pkg/redis"
)

const tokenCacheKey = "wxmp:access_token"

var errCacheMiss = errors.New("cache miss")

// tokenCache access_token 缓存抽象（测试用内存实现注入）
type tokenCache interface {
	get(key string) (string, error)
	set(key, value string, ttl time.Duration) error
}

// redisTokenCache 基于 pkg/redis；redis.Client 为 nil 时退化为不缓存
type redisTokenCache struct{}

func (redisTokenCache) get(key string) (string, error) {
	if redis.Client == nil {
		return "", errCacheMiss
	}
	v, err := redis.Get(key)
	if err != nil {
		return "", errCacheMiss
	}
	return v, nil
}

func (redisTokenCache) set(key, value string, ttl time.Duration) error {
	if redis.Client == nil {
		return nil
	}
	return redis.Set(key, value, ttl)
}
