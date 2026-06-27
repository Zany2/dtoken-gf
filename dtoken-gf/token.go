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
	Generate(ctx context.Context, userKey string, data g.Map) (token string, err error)
	ValidateSession(ctx context.Context, token string) (*Session, error)
	GetSession(ctx context.Context, userKey string) (*Session, error)
	ParseUserKey(ctx context.Context, token string) (userKey string, err error)
	DestroyByToken(ctx context.Context, token string) error
	Destroy(ctx context.Context, userKey string) error
	Shutdown(ctx context.Context)
	GetOptions() Options
}

// DTokenV2 is the main implementation of Token.
type DTokenV2 struct {
	Options      Options
	TokenCodec   TokenCodec
	SessionCodec SessionCodec
	Store        Store
	Renewer      *Renewer
}

// Generate creates a token for the specified userKey.
func (m *DTokenV2) Generate(ctx context.Context, userKey string, data g.Map) (token string, err error) {
	if userKey == "" {
		return "", gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}

	if m.Options.MultiLogin {
		session, err := m.GetSession(ctx, userKey)
		if err == nil && session != nil && session.Token != "" {
			session.Data = data
			session.RefreshNum = 0
			session.CreateTime = gtime.Now().TimestampMilli()
			session.LastRenewTime = 0

			sessionData, err := m.SessionCodec.Encode(ctx, session)
			if err != nil {
				return "", gerror.WrapCode(gcode.CodeInternalError, err)
			}
			if err = m.Store.Save(ctx, userKey, sessionData); err != nil {
				g.Log().Errorf(ctx, "[DToken] save token session failed: userKey=%s error=%v", userKey, err)
				return "", gerror.WrapCode(gcode.CodeInternalError, err)
			}
			return session.Token, nil
		}
		if err != nil && gerror.Code(err).Code() != gcode.CodeNotAuthorized.Code() {
			return "", err
		}
	}

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

	sessionData, err := m.SessionCodec.Encode(ctx, session)
	if err != nil {
		return "", gerror.WrapCode(gcode.CodeInternalError, err)
	}
	if err = m.Store.Save(ctx, userKey, sessionData); err != nil {
		g.Log().Errorf(ctx, "[DToken] save token session failed: userKey=%s error=%v", userKey, err)
		return "", gerror.WrapCode(gcode.CodeInternalError, err)
	}

	return token, nil
}

// ValidateSession checks token validity and returns full session.
func (m *DTokenV2) ValidateSession(ctx context.Context, token string) (*Session, error) {
	if token == "" {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, MsgErrTokenEmpty)
	}

	userKey, err := m.ParseUserKey(ctx, token)
	if err != nil {
		return nil, err
	}

	session, err := m.GetSession(ctx, userKey)
	if err != nil {
		return nil, err
	}

	if token != session.Token {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, MsgErrTokenMismatch)
	}

	if m.Renewer != nil && m.Renewer.ShouldRenew(session) {
		m.Renewer.RenewAsync(gctx.NeverDone(ctx), session)
	}

	return session, nil
}

// GetSession retrieves session by userKey.
func (m *DTokenV2) GetSession(ctx context.Context, userKey string) (*Session, error) {
	if userKey == "" {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}

	sessionData, err := m.Store.Load(ctx, userKey)
	if err != nil {
		g.Log().Errorf(ctx, "[DToken] store load failed: userKey=%s error=%v", userKey, err)
		return nil, gerror.WrapCode(gcode.CodeInternalError, err)
	}
	if sessionData == "" {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, MsgErrDataEmpty)
	}

	session, err := m.SessionCodec.Decode(ctx, sessionData)
	if err != nil {
		g.Log().Errorf(ctx, "[DToken] session decode failed: userKey=%s error=%v", userKey, err)
		return nil, gerror.WrapCode(gcode.CodeInternalError, err)
	}
	if session == nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, MsgErrDataEmpty)
	}
	return session, nil
}

// ParseUserKey extracts userKey from token.
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

// DestroyByToken destroys the session using token.
func (m *DTokenV2) DestroyByToken(ctx context.Context, token string) error {
	session, err := m.ValidateSession(ctx, token)
	if err != nil {
		return err
	}
	return m.Destroy(ctx, session.UserKey)
}

// Destroy removes the session by userKey.
func (m *DTokenV2) Destroy(ctx context.Context, userKey string) error {
	if userKey == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}
	if err := m.Store.Delete(ctx, userKey); err != nil {
		g.Log().Errorf(ctx, "[DToken] destroy token failed: userKey=%s error=%v", userKey, err)
		return gerror.WrapCode(gcode.CodeInternalError, err)
	}
	return nil
}

// Shutdown gracefully stops renew pool.
func (m *DTokenV2) Shutdown(ctx context.Context) {
	if m.Renewer != nil {
		m.Renewer.Shutdown(ctx)
	}
}

// GetOptions returns current options.
func (m *DTokenV2) GetOptions() Options {
	return m.Options
}
