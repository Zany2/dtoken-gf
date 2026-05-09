package dtoken_gf

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// TokenConfig defines construction dependencies TokenConfig 定义 Token 构造依赖
type TokenConfig struct {
	Options      Options      // Options configuration 基础配置参数
	Store        Store        // Raw store implementation 原始存储实现
	TokenCodec   TokenCodec   // Token codec implementation Token 编解码实现
	SessionCodec SessionCodec // Session codec implementation Session 编解码实现
	ShowBanner   bool         // Whether to print startup banner 是否打印启动横幅
}

// Option customizes token construction Option 自定义 Token 构造配置
type Option func(*TokenConfig)

// DefaultTokenConfig returns default construction config 返回默认 Token 构造配置
func DefaultTokenConfig() *TokenConfig {
	return &TokenConfig{
		Options:    NormalizeOptions(Options{}),
		ShowBanner: true,
	}
}

// WithCacheMode sets cache mode 设置缓存模式
func WithCacheMode(mode int8) Option {
	return func(config *TokenConfig) {
		config.Options.CacheMode = mode
	}
}

// WithCachePreKey sets cache key prefix 设置缓存 key 前缀
func WithCachePreKey(preKey string) Option {
	return func(config *TokenConfig) {
		config.Options.CachePreKey = preKey
	}
}

// WithTimeout sets token timeout 设置 Token 超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(config *TokenConfig) {
		config.Options.Timeout = timeout.Milliseconds()
	}
}

// WithMaxRefresh sets max refresh window 设置最大续期窗口
func WithMaxRefresh(maxRefresh time.Duration) Option {
	return func(config *TokenConfig) {
		config.Options.MaxRefresh = maxRefresh.Milliseconds()
	}
}

// WithMaxRefreshTimes sets max refresh times 设置最大续期次数
func WithMaxRefreshTimes(times int) Option {
	return func(config *TokenConfig) {
		config.Options.MaxRefreshTimes = times
	}
}

// WithRenewInterval sets minimum renew interval 设置最小续期间隔
func WithRenewInterval(interval time.Duration) Option {
	return func(config *TokenConfig) {
		config.Options.RenewInterval = interval.Milliseconds()
	}
}

// WithEncryptKey sets token encryption key 设置 Token 加密密钥
func WithEncryptKey(key []byte) Option {
	return func(config *TokenConfig) {
		config.Options.EncryptKey = key
	}
}

// WithTokenDelimiter sets token delimiter 设置 Token 分隔符
func WithTokenDelimiter(delimiter string) Option {
	return func(config *TokenConfig) {
		config.Options.TokenDelimiter = delimiter
	}
}

// WithReuseToken controls same-user token reuse 控制同一用户是否复用 Token
func WithReuseToken(reuse bool) Option {
	return func(config *TokenConfig) {
		config.Options.MultiLogin = reuse
	}
}

// WithAuthExcludePaths sets auth excluded paths 设置免认证路径
func WithAuthExcludePaths(paths ...string) Option {
	return func(config *TokenConfig) {
		config.Options.AuthExcludePaths = g.SliceStr(paths)
	}
}

// WithAuthHeaderKey sets auth header key 设置认证请求头名称
func WithAuthHeaderKey(headerKey string) Option {
	return func(config *TokenConfig) {
		config.Options.AuthHeaderKey = headerKey
	}
}

// WithRenewPool sets renewal pool options 设置续期协程池参数
func WithRenewPool(minSize, maxSize int, scaleUpRate, scaleDownRate float64) Option {
	return func(config *TokenConfig) {
		config.Options.PoolMinSize = minSize
		config.Options.PoolMaxSize = maxSize
		config.Options.PoolScaleUpRate = scaleUpRate
		config.Options.PoolScaleDownRate = scaleDownRate
	}
}

// WithStore sets custom raw session store 设置自定义原始会话存储
func WithStore(store Store) Option {
	return func(config *TokenConfig) {
		config.Store = store
	}
}

// WithTokenCodec sets custom token codec 设置自定义 Token 编解码器
func WithTokenCodec(tokenCodec TokenCodec) Option {
	return func(config *TokenConfig) {
		config.TokenCodec = tokenCodec
	}
}

// WithSessionCodec sets custom session codec 设置自定义 Session 编解码器
func WithSessionCodec(sessionCodec SessionCodec) Option {
	return func(config *TokenConfig) {
		config.SessionCodec = sessionCodec
	}
}

// WithBanner controls startup banner output 控制启动横幅输出
func WithBanner(show bool) Option {
	return func(config *TokenConfig) {
		config.ShowBanner = show
	}
}
