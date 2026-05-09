package dtoken_gf

import (
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// Options defines all configuration for dToken dToken 全局配置参数
type Options struct {
	CacheMode        int8       // Cache mode: 1-gcache 2-gredis 3-gfile 缓存模式：1 gcache 2 gredis 3 gfile
	CachePreKey      string     // Cache key prefix 缓存 key 前缀
	Timeout          int64      // Token expiration time in milliseconds Token 超时时间（毫秒）
	MaxRefresh       int64      // Max auto-refresh interval in milliseconds 最大自动刷新间隔（毫秒）
	MaxRefreshTimes  int        // Maximum number of refresh times, 0 means unlimited 最大刷新次数，0 表示不限制
	TokenDelimiter   string     // Token delimiter Token 分隔符
	EncryptKey       []byte     // Token encryption key Token 加密密钥
	MultiLogin       bool       // Reuse existing token for the same userKey 是否复用同一 userKey 的已有 Token
	AuthHeaderKey    string     // Auth header key 认证请求头名称
	AuthExcludePaths g.SliceStr // Paths excluded from authentication 免认证路径列表

	PoolMinSize       int     // Minimum pool size 最小协程数
	PoolMaxSize       int     // Maximum pool size 最大协程数
	PoolScaleUpRate   float64 // Scale-up threshold, expands when usage exceeds this ratio 扩容阈值，使用率超过此比例时扩容
	PoolScaleDownRate float64 // Scale-down threshold, shrinks when usage falls below this ratio 缩容阈值，使用率低于此比例时缩容
	RenewInterval     int64   // Minimum renewal interval in milliseconds 最小续期间隔（毫秒）
}

// NormalizeOptions applies default options 应用默认配置
func NormalizeOptions(options Options) Options {
	if options.CacheMode == 0 {
		options.CacheMode = CacheModeCache
	}
	if options.CachePreKey == "" {
		options.CachePreKey = DefaultCacheKey
	}
	if options.Timeout <= 0 {
		options.Timeout = DefaultTimeout
	}
	if options.MaxRefresh <= 0 {
		options.MaxRefresh = options.Timeout / 2
	}
	if len(options.EncryptKey) == 0 {
		options.EncryptKey = []byte(DefaultEncryptKey)
	}
	if options.TokenDelimiter == "" {
		options.TokenDelimiter = DefaultTokenDelimiter
	}
	if options.AuthHeaderKey == "" {
		options.AuthHeaderKey = DefaultAuthHeaderKey
	}
	if options.PoolMinSize <= 0 {
		options.PoolMinSize = DefaultMinSize
	}
	if options.PoolMaxSize <= 0 {
		options.PoolMaxSize = DefaultMaxSize
	}
	if options.PoolScaleUpRate <= 0 {
		options.PoolScaleUpRate = DefaultScaleUpRate
	}
	if options.PoolScaleDownRate <= 0 {
		options.PoolScaleDownRate = DefaultScaleDownRate
	}
	if options.RenewInterval < 0 {
		options.RenewInterval = 0
	}
	return options
}

// ValidateOptions validates and normalizes related options 校验并修正关联配置
func ValidateOptions(options *Options) error {
	if options.MaxRefresh >= options.Timeout {
		g.Log().Warning(gctx.New(), "[DToken] invalid config: MaxRefresh >= Timeout, reset to Timeout/2 | 已自动修正为 Timeout 的一半")
		options.MaxRefresh = options.Timeout / 2
	}
	if options.RenewInterval >= options.Timeout {
		g.Log().Warning(gctx.New(), "[DToken] invalid config: RenewInterval >= Timeout, reset to 0 | 已自动修正为 0 (无间隔限制)")
		options.RenewInterval = 0
	}
	if options.PoolMaxSize < options.PoolMinSize {
		g.Log().Warningf(gctx.New(), "[DToken] invalid config: PoolMaxSize < PoolMinSize, reset PoolMaxSize=%d | 已自动修正 PoolMaxSize 为 %d",
			options.PoolMinSize, options.PoolMinSize)
		options.PoolMaxSize = options.PoolMinSize
	}
	if options.PoolScaleDownRate >= options.PoolScaleUpRate {
		g.Log().Warning(gctx.New(), "[DToken] invalid config: PoolScaleDownRate >= PoolScaleUpRate, reset to default values | 已自动修正为默认值")
		options.PoolScaleUpRate = DefaultScaleUpRate
		options.PoolScaleDownRate = DefaultScaleDownRate
	}
	if len(options.EncryptKey) != 16 && len(options.EncryptKey) != 24 && len(options.EncryptKey) != 32 {
		return fmt.Errorf("invalid config: EncryptKey length must be 16, 24, or 32 bytes (AES key size) | EncryptKey 长度必须为 16、24 或 32 字节")
	}
	if options.CacheMode != CacheModeCache && options.CacheMode != CacheModeRedis && options.CacheMode != CacheModeFile {
		return fmt.Errorf("invalid config: CacheMode must be 1 (gcache), 2 (gredis), or 3 (gfile) | CacheMode 必须为 1(gcache)、2(gredis) 或 3(gfile)")
	}
	return nil
}
