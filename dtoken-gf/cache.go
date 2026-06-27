package dtoken_gf

import (
	"context"
	"errors"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
)

// DefaultCache implements the built-in raw store. 内置原始存储实现。
type DefaultCache struct {
	cache   *gcache.Cache // Cache instance. 缓存实例
	mode    int8          // Cache mode: 1 memory, 2 Redis. 缓存模式：1 内存，2 Redis
	preKey  string        // Cache key prefix. 缓存 key 前缀
	timeout int64         // Timeout in milliseconds. 超时时间，单位毫秒
}

// NewDefaultCache creates a new DefaultCache instance. 创建新的默认缓存实例。
// RedisName 为空时使用 g.Redis()，否则使用 g.Redis(redisName)。
func NewDefaultCache(mode int8, preKey string, timeout int64, redisName string) *DefaultCache {
	c := &DefaultCache{
		cache:   gcache.New(),
		mode:    mode,
		preKey:  preKey,
		timeout: timeout,
	}
	if c.mode == CacheModeRedis {
		if redisName == "" {
			c.cache.SetAdapter(gcache.NewAdapterRedis(g.Redis())) // Use default GoFrame Redis instance. 使用 GoFrame 默认 Redis 实例
		} else {
			c.cache.SetAdapter(gcache.NewAdapterRedis(g.Redis(redisName)))
		}
	}
	return c
}

// Save stores the encoded session data. 保存编码后的会话数据。
func (c *DefaultCache) Save(ctx context.Context, userKey string, data string) error {
	if data == "" {
		return errors.New(MsgErrDataEmpty)
	}
	return c.cache.Set(ctx, c.preKey+userKey, data, time.Duration(c.timeout)*time.Millisecond)
}

// Load retrieves the encoded session data. 获取编码后的会话数据。
func (c *DefaultCache) Load(ctx context.Context, userKey string) (string, error) {
	dataVar, err := c.cache.Get(ctx, c.preKey+userKey)
	if err != nil {
		return "", err
	}
	if dataVar.IsNil() {
		return "", nil
	}
	return dataVar.String(), nil
}

// Delete removes the token session. 删除 Token 会话。
func (c *DefaultCache) Delete(ctx context.Context, userKey string) error {
	_, err := c.cache.Remove(ctx, c.preKey+userKey)
	return err
}
