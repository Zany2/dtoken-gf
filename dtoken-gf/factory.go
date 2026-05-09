package dtoken_gf

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// New creates token with functional options 使用函数式选项创建 Token
func New(opts ...Option) (Token, error) {
	config := DefaultTokenConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(config)
		}
	}
	return NewTokenWithConfig(config)
}

// NewFromConfig creates token from global config 从全局配置创建 Token
func NewFromConfig(ctx ...context.Context) (Token, error) {
	configCtx := gctx.New()
	if len(ctx) > 0 && ctx[0] != nil {
		configCtx = ctx[0]
	}
	var options Options
	if err := g.Cfg().MustGet(configCtx, DTokenCfgName).Struct(&options); err != nil {
		g.Log().Error(configCtx, "[DToken] load config failed", err) // Log config load failure 记录配置读取失败
		return nil, err
	}
	return NewToken(options)
}

// NewToken creates token instance and returns errors 创建 Token 实例并返回错误
func NewToken(options Options) (*DTokenV2, error) {
	config := DefaultTokenConfig()
	config.Options = options
	return NewTokenWithConfig(config)
}

// NewTokenWithConfig creates token instance from full config 使用完整配置创建 Token 实例
func NewTokenWithConfig(config *TokenConfig) (*DTokenV2, error) {
	if config == nil {
		config = DefaultTokenConfig()
	}
	normalized := NormalizeOptions(config.Options)
	if err := ValidateOptions(&normalized); err != nil {
		g.Log().Error(gctx.New(), "[DToken] validate options failed", err) // Log options validation failure 记录配置校验失败
		return nil, err
	}

	renewPoolManager, err := NewRenewPoolBuilder().
		MinSize(normalized.PoolMinSize).
		MaxSize(normalized.PoolMaxSize).
		ScaleUpRate(normalized.PoolScaleUpRate).
		ScaleDownRate(normalized.PoolScaleDownRate).
		Build()
	if err != nil {
		g.Log().Error(gctx.New(), "[DToken] renew pool init failed", err) // Log renew pool init failure 记录续期池初始化失败
		return nil, err
	}

	store := config.Store
	if store == nil {
		defaultCache := NewDefaultCache(normalized.CacheMode, normalized.CachePreKey, normalized.Timeout)
		store = defaultCache
	}

	tokenCodec := config.TokenCodec
	if tokenCodec == nil {
		tokenCodec = NewDefaultTokenCodec(normalized.TokenDelimiter, normalized.EncryptKey)
	}

	sessionCodec := config.SessionCodec
	if sessionCodec == nil {
		sessionCodec = NewDefaultSessionCodec()
	}

	dToken := &DTokenV2{
		Options:      normalized,
		TokenCodec:   tokenCodec,
		SessionCodec: sessionCodec,
		Store:        store,
	}
	dToken.Renewer = NewRenewer(&dToken.Options, dToken.Store, dToken.SessionCodec, renewPoolManager)

	if config.ShowBanner {
		PrintWithOptions(&dToken.Options)
	}
	return dToken, nil
}
