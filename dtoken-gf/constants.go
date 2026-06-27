package dtoken_gf

const (
	DTokenCfgName = "dToken" // Global configuration node name. 全局配置节点名称

	CacheModeCache = 1 // Cache mode using in-memory cache. 内存缓存模式
	CacheModeRedis = 2 // Cache mode using Redis. Redis 缓存模式

	DefaultTimeout        = 10 * 24 * 60 * 60 * 1000           // Default timeout, 10 days in milliseconds. 默认超时时间，10 天，单位毫秒
	MinTimeout            = 10 * 1000                          // Minimum timeout, 10 seconds in milliseconds. 最小超时时间，10 秒，单位毫秒
	DefaultCacheKey       = "DToken:"                          // Default prefix for cache keys. 默认缓存 key 前缀
	DefaultTokenDelimiter = "_"                                // Default delimiter for tokens. 默认 Token 分隔符
	DefaultEncryptKey     = "12345678912345678912345678912345" // Default encryption key for token. 默认 Token 加密密钥
	DefaultAuthHeaderKey  = "Authorization"                    // Default auth header key. 默认认证请求头名称

	// Cache key fields. 缓存 key 字段定义
	KeyUserKey       = "userKey"          // User identifier. 用户标识
	KeyCreateTime    = "createTime"       // Token creation time. Token 创建时间
	KeyRefreshNum    = "refreshNum"       // Token refresh count. Token 刷新次数
	KeyLastRenewTime = "keyLastRenewTime" // Last token renewal time. 上次续期时间
	KeyData          = "data"             // Custom data stored in cache. 缓存中的自定义数据
	KeyToken         = "token"            // The actual token value. 实际的 token 值
)

const (
	MsgErrUserKeyEmpty  = "userKey empty"       // Error message when userKey is empty. 用户标识为空时的错误信息
	MsgErrTokenEmpty    = "token is empty"      // Error message when token is empty. Token 为空时的错误信息
	MsgErrTokenLen      = "token len error"     // Error message when token length is incorrect. Token 长度不正确时的错误信息
	MsgErrValidate      = "user validate error" // Error message for user validation failure. 用户校验失败时的错误信息
	MsgErrTokenMismatch = "token mismatch"      // Error message when token does not match. Token 不一致时的错误信息
	MsgErrDataEmpty     = "cache value is nil"  // Error message when cache value is nil. 缓存值为空时的错误信息
)
