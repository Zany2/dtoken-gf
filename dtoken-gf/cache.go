package dtoken_gf

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

// DefaultCache implements the default raw store 默认原始存储实现
type DefaultCache struct {
	cache   *gcache.Cache // Cache instance 缓存实例
	mode    int8          // Cache mode: 1 for gcache, 2 for gredis, 3 for gfile 缓存模式：1 为 gcache，2 为 gredis，3 为 gfile
	preKey  string        // Cache key prefix 缓存 key 前缀
	timeout int64         // Timeout in milliseconds 超时时间，单位毫秒
	fileMu  sync.Mutex    // File cache write lock 文件缓存写入锁
}

// NewDefaultCache creates a new DefaultCache instance 创建新的默认缓存实例
func NewDefaultCache(mode int8, preKey string, timeout int64) *DefaultCache {
	c := &DefaultCache{
		cache:   gcache.New(),
		mode:    mode,
		preKey:  preKey,
		timeout: timeout,
	}

	// Initialize the cache based on mode 根据模式初始化缓存
	if c.mode == CacheModeFile {
		c.initFileCache(gctx.New()) // Initialize file cache 初始化文件缓存
	} else if c.mode == CacheModeRedis {
		c.cache.SetAdapter(gcache.NewAdapterRedis(g.Redis())) // Initialize Redis cache 初始化 Redis 缓存
	}

	return c
}

// Save saves encoded session data 保存编码后的会话数据
func (c *DefaultCache) Save(ctx context.Context, userKey string, data string) error {
	if data == "" {
		return errors.New(MsgErrDataEmpty) // Error if encoded data is empty 编码数据为空时返回错误
	}
	return c.withFileLock(func() error {
		err := c.cache.Set(ctx, c.preKey+userKey, data, gconv.Duration(c.timeout)*time.Millisecond)
		if err != nil {
			return err
		}
		if c.mode == CacheModeFile {
			c.writeFileCache(ctx)
		}
		return nil
	})
}

// Load retrieves encoded session data 获取编码后的会话数据
func (c *DefaultCache) Load(ctx context.Context, userKey string) (string, error) {
	dataVar, err := c.cache.Get(ctx, c.preKey+userKey) // Get the cache value 获取缓存值
	if err != nil {
		return "", err
	}
	if dataVar.IsNil() {
		return "", nil // Return empty if cache value is empty 如果缓存值为空，则返回空字符串
	}
	return dataVar.String(), nil
}

// deleteKey deletes session key 删除会话 key
func (c *DefaultCache) deleteKey(ctx context.Context, cacheKey string) error {
	return c.withFileLock(func() error {
		_, err := c.cache.Remove(ctx, c.preKey+cacheKey)
		if c.mode == CacheModeFile {
			c.writeFileCache(ctx)
		}
		return err
	})
}

// Delete deletes token session 删除 Token 会话
func (c *DefaultCache) Delete(ctx context.Context, userKey string) error {
	return c.deleteKey(ctx, userKey)
}

// withFileLock serializes file cache mutation and persistence 串行化文件缓存变更和落盘
func (c *DefaultCache) withFileLock(f func() error) error {
	if c.mode != CacheModeFile {
		return f()
	}
	c.fileMu.Lock()
	defer c.fileMu.Unlock()
	return f()
}

// writeFileCache writes the cache data to a file 将缓存数据写入文件
func (c *DefaultCache) writeFileCache(ctx context.Context) {
	fileName := gstr.Replace(c.preKey, ":", "_") + CacheModeFileDat // Generate file name 生成文件名
	file := gfile.Temp(fileName)                                    // Create temporary file 创建临时文件
	data, e := c.cache.Data(ctx)                                    // Get cache data 获取缓存数据
	if e != nil {
		g.Log().Error(ctx, "[DToken] cache writeFileCache data error", e) // Log error if data retrieval fails 获取数据失败时记录错误
	}
	e = gfile.PutContents(file, gjson.New(data).MustToJsonString()) // Write data to file 将数据写入文件
	if e != nil {
		g.Log().Error(ctx, "[DToken] cache writeFileCache put error", e) // Log error if writing to file fails 写入文件失败时记录错误
	}
}

// initFileCache initializes the file cache 初始化文件缓存
func (c *DefaultCache) initFileCache(ctx context.Context) {
	fileName := gstr.Replace(c.preKey, ":", "_") + CacheModeFileDat // Generate file name 生成文件名
	file := gfile.Temp(fileName)                                    // Create temporary file 创建临时文件
	if !gfile.Exists(file) {
		return // Return if the file does not exist 如果文件不存在，则返回
	}
	data := gfile.GetContents(file)            // Read data from file 从文件读取数据
	cacheJson, err := gjson.DecodeToJson(data) // Decode cache JSON 解析缓存 JSON
	if err != nil {
		g.Log().Error(ctx, "[DToken] cache initFileCache decode error", err) // Log decode error 记录解析错误
		return
	}
	maps := cacheJson.Map() // Convert JSON to map 将 JSON 转换为 map
	if maps == nil || len(maps) <= 0 {
		return // Return if no data is found in the file 如果文件中没有数据，则返回
	}
	// Load the cache data from file 从文件加载缓存数据
	for k, v := range maps {
		_ = c.cache.Set(ctx, k, v, gconv.Duration(c.timeout)*time.Millisecond)
	}
}
