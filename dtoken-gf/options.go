package dtoken_gf

import (
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// Options groups all configurable values for dToken. 配置项集合
type Options struct {
	// CacheMode selects the built-in cache backend. 缓存模式：1=内存，2=Redis。
	CacheMode int8

	// CachePreKey is the prefix used to build cache keys. 缓存 key 前缀。
	CachePreKey string

	// RedisName selects a named GoFrame Redis instance. Redis 实例名，空值时使用 g.Redis()。
	RedisName string

	// Timeout is the token lifetime in milliseconds. Token 有效期，单位毫秒。
	Timeout int64

	// MaxRefresh is the remaining lifetime threshold for renewal. 续期阈值，单位毫秒；剩余有效期小于等于该值时触发续期。
	MaxRefresh int64

	// MaxRefreshTimes limits how many times a token can be renewed. 最大续期次数，单位：次；超过后不再自动续期。
	MaxRefreshTimes int

	// TokenDelimiter is the delimiter used inside generated tokens. Token 内部分隔符。
	TokenDelimiter string

	// EncryptKey is the AES key used to protect token payloads. Token 加密密钥，长度需要是 16、24 或 32 字节。
	EncryptKey []byte

	// MultiLogin controls whether the same userKey reuses the existing token. 是否复用同一 userKey 的已有 Token。
	MultiLogin bool

	// AuthHeaderKey is the request header used to carry the token. 认证请求头名称。
	AuthHeaderKey string

	// AuthExcludePaths lists paths that bypass authentication. 认证豁免路径列表，支持精确匹配和以 /* 结尾的前缀匹配。
	AuthExcludePaths g.SliceStr

	// PoolMinSize is the minimum size of the renewal worker pool. 续期协程池最小值，单位：个。
	PoolMinSize int

	// PoolMaxSize is the maximum size of the renewal worker pool. 续期协程池最大值，单位：个。
	PoolMaxSize int

	// PoolScaleUpRate is the usage threshold for pool expansion. 协程池扩容阈值，范围 0~1。
	PoolScaleUpRate float64

	// PoolScaleDownRate is the usage threshold for pool shrinkage. 协程池缩容阈值，范围 0~1。
	PoolScaleDownRate float64

	// PoolCheckInterval is the renewal pool scaling check interval in milliseconds. 协程池检查间隔，单位毫秒。
	PoolCheckInterval int64
}

// NormalizeOptions applies default values only. 仅补全默认值。
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
	if options.Timeout < MinTimeout {
		options.Timeout = MinTimeout
	}
	if options.MaxRefresh <= 0 {
		options.MaxRefresh = options.Timeout / 2
	}
	if options.MaxRefresh <= 0 {
		options.MaxRefresh = 1
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
	if options.PoolCheckInterval <= 0 {
		options.PoolCheckInterval = DefaultCheckInterval.Milliseconds()
	}
	return options
}

// ValidateOptions validates and fixes a few safe invalid combinations. 校验并修正部分安全的非法组合。
func ValidateOptions(options *Options) error {
	if options == nil {
		return fmt.Errorf("invalid config: options is nil | 配置不能为空")
	}
	if options.CacheMode == 0 {
		options.CacheMode = CacheModeCache
	}
	if options.Timeout < MinTimeout {
		g.Log().Warning(gctx.New(), "[DToken] invalid config: Timeout < MinTimeout, reset to 10s | 已自动修正为 10 秒")
		options.Timeout = MinTimeout
	}

	if string(options.EncryptKey) == DefaultEncryptKey {
		g.Log().Warning(gctx.New(), "[DToken] SECURITY WARNING: using default EncryptKey, tokens can be forged! Please set a custom encryption key | 安全警告：正在使用默认加密密钥，Token 可能被伪造！请设置自定义加密密钥")
	}

	if options.MaxRefresh >= options.Timeout {
		g.Log().Warning(gctx.New(), "[DToken] invalid config: MaxRefresh >= Timeout, reset to Timeout/2 | 已自动修正为 Timeout 的一半")
		options.MaxRefresh = options.Timeout / 2
		if options.MaxRefresh <= 0 {
			options.MaxRefresh = 1
		}
	}
	if options.MaxRefreshTimes < 0 {
		g.Log().Warning(gctx.New(), "[DToken] invalid config: MaxRefreshTimes < 0, reset to 0 | 已自动修正为 0（不限制续期次数）")
		options.MaxRefreshTimes = 0
	}
	if options.PoolMaxSize < options.PoolMinSize {
		g.Log().Warningf(gctx.New(), "[DToken] invalid config: PoolMaxSize < PoolMinSize, reset PoolMaxSize=%d | 已自动修正为 %d",
			options.PoolMinSize, options.PoolMinSize)
		options.PoolMaxSize = options.PoolMinSize
	}
	if options.PoolScaleDownRate <= 0 || options.PoolScaleUpRate <= 0 || options.PoolScaleUpRate > 1 ||
		options.PoolScaleDownRate >= options.PoolScaleUpRate {
		g.Log().Warning(gctx.New(), "[DToken] invalid config: PoolScaleDownRate >= PoolScaleUpRate, reset to default values | 已自动修正为默认值")
		options.PoolScaleUpRate = DefaultScaleUpRate
		options.PoolScaleDownRate = DefaultScaleDownRate
	}

	if len(options.EncryptKey) != 16 && len(options.EncryptKey) != 24 && len(options.EncryptKey) != 32 {
		return fmt.Errorf("invalid config: EncryptKey length must be 16, 24, or 32 bytes (AES key size) | EncryptKey 长度必须是 16、24 或 32 字节")
	}
	if !isValidTokenDelimiter(options.TokenDelimiter) {
		return fmt.Errorf("invalid config: TokenDelimiter must be one of _ - . : | ~ | TokenDelimiter 必须是 _ - . : | ~ 之一")
	}
	if options.CacheMode != CacheModeCache && options.CacheMode != CacheModeRedis {
		return fmt.Errorf("invalid config: CacheMode must be 1 (memory) or 2 (redis) | CacheMode 必须是 1（内存）或 2（Redis）")
	}
	return nil
}

// isValidTokenDelimiter checks whether the delimiter is a safe single character. 判断分隔符是否是安全的单字符。
func isValidTokenDelimiter(delimiter string) bool {
	return len(delimiter) == 1 && strings.Contains("_-.:|~", delimiter)
}
