package common

import (
	"context"
	"encoding/json"
	"time"

	"wetalk/db"
)

const cachePrefix = "wetalk:"

// cacheKey 构建带命名空间前缀的缓存键
func cacheKey(key string) string {
	return cachePrefix + key
}

// CacheGet JSON 反序列化读取缓存，返回是否命中
// Redis 不可用时静默降级（返回 miss），不阻塞业务
func CacheGet(key string, dest interface{}) (bool, error) {
	if db.RedisClient == nil {
		return false, nil
	}
	data, err := db.RedisClient.Get(context.Background(), cacheKey(key)).Bytes()
	if err != nil {
		return false, nil
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return false, nil
	}
	return true, nil
}

// CacheSet JSON 序列化写入缓存
func CacheSet(key string, value interface{}, ttl time.Duration) error {
	if db.RedisClient == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return db.RedisClient.Set(context.Background(), cacheKey(key), data, ttl).Err()
}

// CacheDel 删除一个或多个缓存键
func CacheDel(keys ...string) error {
	if db.RedisClient == nil || len(keys) == 0 {
		return nil
	}
	prefixed := make([]string, len(keys))
	for i, k := range keys {
		prefixed[i] = cacheKey(k)
	}
	return db.RedisClient.Del(context.Background(), prefixed...).Err()
}

// RedisCache 是 Cache 接口的生产环境实现，通过委托包级函数完成操作
type RedisCache struct{}

func (c *RedisCache) Get(key string, dest interface{}) (bool, error) { return CacheGet(key, dest) }

func (c *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	return CacheSet(key, value, ttl)
}

func (c *RedisCache) Del(keys ...string) error { return CacheDel(keys...) }
