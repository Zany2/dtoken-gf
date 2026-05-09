package dtoken_gf

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
)

// Token defines token interface Token 接口定义
type Token interface {
	Generate(ctx context.Context, userKey string, data g.Map) (token string, err error) // Generate token 生成 Token
	ValidateSession(ctx context.Context, token string) (*Session, error)                // Validate token and return session 验证 Token 并返回会话
	GetSession(ctx context.Context, userKey string) (*Session, error)                   // Get session by userKey 通过 userKey 获取会话
	ParseUserKey(ctx context.Context, token string) (userKey string, err error)         // Parse userKey from token 从 Token 解析 userKey
	DestroyByToken(ctx context.Context, token string) error                             // Destroy token by token value 通过 Token 销毁会话
	Destroy(ctx context.Context, userKey string) error                                  // Destroy token 销毁 Token
	Shutdown(ctx context.Context)                                                       // Gracefully shutdown renew pool 优雅关闭续期协程池
	GetOptions() Options                                                                // Get config options 获取配置参数
}

// DTokenV2 main implementation dToken 主体结构体
type DTokenV2 struct {
	Options      Options      // Options configuration 配置参数
	TokenCodec   TokenCodec   // Token codec implementation Token 编解码实现
	SessionCodec SessionCodec // Session codec implementation Session 编解码实现
	Store        Store        // Raw store implementation 原始存储实现
	Renewer      *Renewer     // Renewal component 续期组件
}

// Generate creates a new token for user 生成 Token
func (m *DTokenV2) Generate(ctx context.Context, userKey string, data g.Map) (token string, err error) {
	if userKey == "" {
		return "", gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}

	// Support multi-login by reusing existing token 支持多端重复登录，复用已有 Token
	if m.Options.MultiLogin {
		session, err := m.GetSession(ctx, userKey)
		if err == nil && session != nil && session.Token != "" {
			return session.Token, nil
		}
	}

	// Encode userKey into token 编码用户唯一标识生成 Token
	token, err = m.TokenCodec.Encode(ctx, userKey)
	if err != nil {
		return "", gerror.WrapCode(gcode.CodeInternalError, err)
	}

	session := &Session{
		UserKey:       userKey,
		Token:         token,
		Data:          data,
		RefreshNum:    0,
		CreateTime:    gtime.Now().TimestampMilli(),
		LastRenewTime: 0,
	}

	// Save token data to store 将用户 Token 信息写入存储
	sessionData, err := m.SessionCodec.Encode(ctx, session)
	if err != nil {
		return "", gerror.WrapCode(gcode.CodeInternalError, err)
	}
	if err = m.Store.Save(ctx, userKey, sessionData); err != nil {
		g.Log().Errorf(ctx, "[DToken] save token session failed: userKey=%s error=%v", userKey, err) // Log save failure 记录会话保存失败
		return "", gerror.WrapCode(gcode.CodeInternalError, err)
	}

	return token, nil
}

// ValidateSession checks token validity and returns full session 验证 Token 并返回完整会话
func (m *DTokenV2) ValidateSession(ctx context.Context, token string) (*Session, error) {
	if token == "" {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, MsgErrTokenEmpty)
	}

	// Decode token to get user key 解码 Token 获取用户标识
	userKey, err := m.ParseUserKey(ctx, token)
	if err != nil {
		return nil, err
	}

	// Retrieve session by user key 通过用户标识获取会话信息
	session, err := m.GetSession(ctx, userKey)
	if err != nil {
		return nil, err
	}

	// Verify token consistency 校验 Token 一致性
	if token != session.Token {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, MsgErrValidate)
	}

	// Check if renewal is needed 判断是否需要续期
	if m.Renewer.ShouldRenew(session) {
		m.Renewer.RenewAsync(gctx.NeverDone(ctx), session)
	}

	return session, nil
}

// GetSession retrieves session by userKey 通过 userKey 获取会话
func (m *DTokenV2) GetSession(ctx context.Context, userKey string) (*Session, error) {
	if userKey == "" {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}

	// Retrieve session from store 从存储中获取会话
	sessionData, err := m.Store.Load(ctx, userKey)
	if err != nil {
		g.Log().Errorf(ctx, "[DToken] store load failed: userKey=%s error=%v", userKey, err) // Log store load failure 记录存储读取失败
		return nil, gerror.WrapCode(gcode.CodeInternalError, err)
	}
	if sessionData == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, MsgErrDataEmpty)
	}
	session, err := m.SessionCodec.Decode(ctx, sessionData)
	if err != nil {
		g.Log().Errorf(ctx, "[DToken] session decode failed: userKey=%s error=%v", userKey, err) // Log session decode failure 记录会话解码失败
		return nil, gerror.WrapCode(gcode.CodeInternalError, err)
	}
	if session == nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, MsgErrDataEmpty)
	}
	return session, nil
}

// ParseUserKey parses userKey from token without loading session 仅从 Token 解析 userKey，不加载会话
func (m *DTokenV2) ParseUserKey(ctx context.Context, token string) (userKey string, err error) {
	if token == "" {
		return "", gerror.NewCode(gcode.CodeMissingParameter, MsgErrTokenEmpty)
	}

	userKey, err = m.TokenCodec.Decode(ctx, token)
	if err != nil {
		return "", gerror.WrapCode(gcode.CodeInvalidParameter, err)
	}
	return userKey, nil
}

// DestroyByToken removes session by token 通过 Token 销毁会话
func (m *DTokenV2) DestroyByToken(ctx context.Context, token string) error {
	session, err := m.ValidateSession(ctx, token)
	if err != nil {
		return err
	}
	return m.Destroy(ctx, session.UserKey)
}

// Destroy removes user token from store 销毁 Token
func (m *DTokenV2) Destroy(ctx context.Context, userKey string) error {
	if userKey == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}
	// Remove store entry 从存储移除对应 Token 信息
	if err := m.Store.Delete(ctx, userKey); err != nil {
		g.Log().Errorf(ctx, "[DToken] destroy token failed: userKey=%s error=%v", userKey, err) // Log destroy failure 记录 Token 销毁失败
		return gerror.WrapCode(gcode.CodeInternalError, err)
	}
	return nil
}

// Shutdown gracefully stops renew pool 优雅关闭续期协程池
func (m *DTokenV2) Shutdown(ctx context.Context) {
	if m.Renewer != nil {
		m.Renewer.Shutdown(ctx)
	}
}

// GetOptions returns current options 获取 Options 配置
func (m *DTokenV2) GetOptions() Options {
	return m.Options
}
