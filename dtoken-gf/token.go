package dtoken_gf

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

// Token defines token interface Token 接口定义
type Token interface {
	Generate(ctx context.Context, userKey string, data any) (token string, err error) // Generate token 生成 Token
	Validate(ctx context.Context, token string) (data any, err error)                 // Validate token 验证 Token
	Get(ctx context.Context, userKey string) (token string, data any, err error)      // Get token by userKey 通过 userKey 获取 Token
	ParseToken(ctx context.Context, token string) (userKey string, data any, err error)
	Destroy(ctx context.Context, userKey string) error          // Destroy token 销毁 Token
	Renew(ctx context.Context, userKey string, userCache g.Map) // Asynchronously renew token 异步续期 Token
	Shutdown(ctx context.Context)                               // Gracefully shutdown renew pool 优雅关闭续期协程池
	GetOptions() Options                                        // Get config options 获取配置参数
}

// DTokenV2 main implementation dToken 主体结构体
type DTokenV2 struct {
	Options          Options           // Options configuration 配置参数
	Codec            Codec             // Codec implementation 编解码实现
	Cache            Cache             // Cache implementation 缓存实现
	RenewPoolManager *RenewPoolManager // Renew pool manager 续期协程池管理器
}

// NewDefaultTokenByConfig creates a token from global config 从全局配置创建 Token
func NewDefaultTokenByConfig() Token {
	var options Options
	err := g.Cfg().MustGet(gctx.New(), DTokenCfgName).Struct(&options)
	if err != nil {
		panic("dToken options init failed: " + err.Error())
	}
	return NewDefaultToken(options)
}

// NewDefaultToken creates token instance with options 使用配置创建 Token 实例
func NewDefaultToken(options Options) Token {
	// Apply defaults 应用默认配置
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

	// Validate MaxRefresh should be less than Timeout 校验 MaxRefresh 应小于 Timeout
	if options.MaxRefresh >= options.Timeout {
		g.Log().Warning(gctx.New(), "invalid config: MaxRefresh >= Timeout, reset to Timeout/2 | 已自动修正为 Timeout 的一半")
		options.MaxRefresh = options.Timeout / 2
	}

	// 2. RenewInterval should be less than Timeout RenewInterval 应小于 Timeout
	if options.RenewInterval >= options.Timeout {
		g.Log().Warning(gctx.New(), "invalid config: RenewInterval >= Timeout, reset to 0 | 已自动修正为 0 (无间隔限制)")
		options.RenewInterval = 0
	}

	// 3. PoolMaxSize must not be smaller than PoolMinSize PoolMaxSize 不能小于 PoolMinSize
	if options.PoolMaxSize < options.PoolMinSize {
		g.Log().Warningf(gctx.New(), "invalid config: PoolMaxSize < PoolMinSize, reset PoolMaxSize=%d | 已自动修正 PoolMaxSize 为 %d",
			options.PoolMinSize, options.PoolMinSize)
		options.PoolMaxSize = options.PoolMinSize
	}

	// 4. ScaleDownRate must be smaller than ScaleUpRate ScaleDownRate 必须小于 ScaleUpRate
	if options.PoolScaleDownRate >= options.PoolScaleUpRate {
		g.Log().Warning(gctx.New(), "invalid config: PoolScaleDownRate >= PoolScaleUpRate, reset to default values | 已自动修正为默认值")
		options.PoolScaleUpRate = DefaultScaleUpRate
		options.PoolScaleDownRate = DefaultScaleDownRate
	}

	// 5. EncryptKey length check, must panic if invalid EncryptKey 长度校验，非法时必须 panic
	if len(options.EncryptKey) != 16 && len(options.EncryptKey) != 24 && len(options.EncryptKey) != 32 {
		panic("invalid config: EncryptKey length must be 16, 24, or 32 bytes (AES key size) | EncryptKey 长度必须为 16、24 或 32 字节")
	}

	// 6. CacheMode check, must panic if invalid CacheMode 校验，非法时必须 panic
	if options.CacheMode != CacheModeCache && options.CacheMode != CacheModeRedis && options.CacheMode != CacheModeFile {
		panic("invalid config: CacheMode must be 1 (gcache), 2 (gredis), or 3 (gfile) | CacheMode 必须为 1(gcache)、2(gredis) 或 3(gfile)")
	}

	// Initialize renew pool 初始化续期协程池
	renewPoolManager, err := NewRenewPoolBuilder().
		MinSize(options.PoolMinSize).
		MaxSize(options.PoolMaxSize).
		ScaleUpRate(options.PoolScaleUpRate).
		ScaleDownRate(options.PoolScaleDownRate).
		Build()
	if err != nil {
		panic(err)
	}

	// Construct main token instance 构建主 Token 实例
	dToken := &DTokenV2{
		Options:          options,
		Codec:            NewDefaultCodec(options.TokenDelimiter, options.EncryptKey),
		Cache:            NewDefaultCache(options.CacheMode, options.CachePreKey, options.Timeout),
		RenewPoolManager: renewPoolManager,
	}

	PrintWithOptions(&dToken.Options)
	return dToken
}

// Generate creates a new token for user 生成 Token
func (m *DTokenV2) Generate(ctx context.Context, userKey string, data any) (token string, err error) {
	if userKey == "" {
		return "", gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}

	// Support multi-login by reusing existing token 支持多端重复登录，重用旧 Token
	if m.Options.MultiLogin {
		token, _, err = m.Get(ctx, userKey)
		if err == nil && token != "" {
			return token, nil
		}
	}

	// Encode userKey into token 编码用户唯一标识生成 Token
	token, err = m.Codec.Encode(ctx, userKey)
	if err != nil {
		return "", gerror.WrapCode(gcode.CodeInternalError, err)
	}

	// Cache structure for user token 构建用户缓存结构
	userCache := g.Map{
		KeyUserKey:       userKey,                      // User identifier 用户唯一标识
		KeyToken:         token,                        // Token value Token 值
		KeyData:          data,                         // Extra data 附加数据
		KeyRefreshNum:    0,                            // Renewal count 已续期次数
		KeyCreateTime:    gtime.Now().TimestampMilli(), // Creation time 创建时间
		KeyLastRenewTime: 0,                            // Last renewal time 续期时间
	}

	// Save token data to cache 将用户 Token 信息写入缓存
	if err = m.Cache.Set(ctx, userKey, userCache); err != nil {
		return "", gerror.WrapCode(gcode.CodeInternalError, err)
	}
	return token, nil
}

// Validate checks token validity and optionally triggers renewal 验证 Token 并触发续期
func (m *DTokenV2) Validate(ctx context.Context, token string) (data any, err error) {
	if token == "" {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, MsgErrTokenEmpty)
	}

	// Decode token to get user key 解码 Token 获取用户标识
	userKey, err := m.Codec.Decrypt(ctx, token)
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInvalidParameter, err)
	}

	// Retrieve cache info by user key 通过用户标识获取缓存信息
	userCache, err := m.Cache.Get(ctx, userKey)
	if err != nil {
		return nil, err
	}
	if userCache == nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, MsgErrDataEmpty)
	}

	// Verify token consistency 校验 Token 一致性
	if token != userCache[KeyToken] {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, MsgErrValidate)
	}

	// Check if renewal is needed 判断是否需要续期
	if m.shouldRenew(userCache) {
		m.Renew(gctx.NeverDone(ctx), userKey, userCache)
	}

	return userCache[KeyData], nil
}

// Renew asynchronously renews a token 异步续期 Token
func (m *DTokenV2) Renew(ctx context.Context, userKey string, userCache g.Map) {
	err := m.RenewPoolManager.Submit(func() {
		// Recheck token validity 再次确认 Token 是否依然有效
		currentCache, err := m.Cache.Get(ctx, userKey)
		if err != nil || currentCache == nil {
			// User logged out or cache was cleared 用户已登出或被清除
			return
		}

		// Verify token consistency to avoid replacement 校验 Token 一致性，防止被其他 Token 替换
		if currentCache[KeyToken] != userCache[KeyToken] {
			return
		}

		newMap := gconv.Map(userCache, gconv.MapOption{Deep: true})
		if newMap == nil {
			return
		}

		newMap[KeyLastRenewTime] = gtime.Now().TimestampMilli()
		newMap[KeyRefreshNum] = gconv.Int(newMap[KeyRefreshNum]) + 1
		_ = m.Cache.Set(ctx, userKey, newMap)
	})
	if err != nil {
		g.Log().Error(ctx, "[DToken]token renew submit error", err) // Log renew submit error 记录续期任务提交错误
	}
}

// shouldRenew checks whether the token should be renewed 判断是否需要续期
func (m *DTokenV2) shouldRenew(userCache g.Map) bool {
	now := gtime.Now().TimestampMilli()                       // Current time 当前时间
	createTime := gconv.Int64(userCache[KeyCreateTime])       // Token creation time Token 创建时间
	lastRenewTime := gconv.Int64(userCache[KeyLastRenewTime]) // Last renewal time, 0 if first 上次续期时间，第一次为 0
	refreshNum := gconv.Int(userCache[KeyRefreshNum])         // Number of renewals 已续期次数

	// 1. Skip renew logic if MaxRefresh is disabled 若未启用续期机制（MaxRefresh=0），则永不续期
	if m.Options.MaxRefresh == 0 {
		return false
	}

	// Determine reference time 确定参考时间，首次续期用创建时间，后续用上次续期时间
	refTime := createTime
	if lastRenewTime > 0 {
		refTime = lastRenewTime
	}

	// Calculate elapsed and remaining time 计算已过时间与剩余寿命
	elapsed := now - refTime
	remaining := m.Options.Timeout - elapsed

	// 2. Not in the refresh window 未进入续期判断窗口，剩余寿命大于 MaxRefresh 则不续期
	if remaining > m.Options.MaxRefresh {
		return false
	}

	// 3. Check renew interval limit, skip for first renewal 判断续期间隔，首次续期不受限制
	if refreshNum > 0 && m.Options.RenewInterval > 0 && elapsed < m.Options.RenewInterval {
		return false
	}

	// 4. Check max renew times, 0 means unlimited 判断最大续期次数，0 表示无限制
	if m.Options.MaxRefreshTimes > 0 && refreshNum >= m.Options.MaxRefreshTimes {
		return false
	}

	// Allow renewal 允许续期
	return true
}

// Get retrieves token and data by userKey 通过 userKey 获取 Token
func (m *DTokenV2) Get(ctx context.Context, userKey string) (token string, data any, err error) {
	if userKey == "" {
		return "", nil, gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}

	// Retrieve token and data from cache 从缓存中获取 Token 与附加数据
	userCache, err := m.Cache.Get(ctx, userKey)
	if err != nil {
		return "", nil, gerror.WrapCode(gcode.CodeInternalError, err)
	}
	if userCache == nil {
		return "", nil, gerror.NewCode(gcode.CodeInternalError, MsgErrDataEmpty)
	}
	return gconv.String(userCache[KeyToken]), userCache[KeyData], nil
}

// ParseToken parses token to retrieve userKey and data 解析 Token 获取 userKey 和数据
func (m *DTokenV2) ParseToken(ctx context.Context, token string) (userKey string, data any, err error) {
	if token == "" {
		return "", nil, gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}

	// Decode token 解密 Token
	userKey, err = m.Codec.Decrypt(ctx, token)
	if err != nil {
		return "", nil, gerror.WrapCode(gcode.CodeInvalidParameter, err)
	}

	// Fetch from cache 从缓存获取数据
	userCache, err := m.Cache.Get(ctx, userKey)
	if err != nil {
		return "", nil, gerror.WrapCode(gcode.CodeInternalError, err)
	}
	if userCache == nil {
		return "", nil, gerror.NewCode(gcode.CodeInternalError, MsgErrDataEmpty)
	}
	return userKey, userCache[KeyData], nil
}

// Destroy removes user token from cache 销毁 Token
func (m *DTokenV2) Destroy(ctx context.Context, userKey string) error {
	if userKey == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}
	// Remove cache entry 从缓存移除对应 Token 信息
	if err := m.Cache.Remove(ctx, userKey); err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err)
	}
	return nil
}

// Shutdown gracefully stops renew pool 优雅关闭续期协程池
func (m *DTokenV2) Shutdown(ctx context.Context) {
	if m.RenewPoolManager != nil {
		g.Log().Info(ctx, "Token RenewPoolManager closed")
		m.RenewPoolManager.Stop()
	}
}

// GetOptions returns current options 获取 Options 配置
func (m *DTokenV2) GetOptions() Options {
	return m.Options
}
