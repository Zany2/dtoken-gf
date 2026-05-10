package dtoken_gf

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Renewer handles async token renewal Renewer 处理 Token 异步续期
type Renewer struct {
	options      *Options          // Options reference 配置引用
	store        Store             // Raw session store 原始会话存储
	sessionCodec SessionCodec      // Session codec Session 编解码器
	pool         *RenewPoolManager // Renewal pool 续期协程池
}

// NewRenewer creates a Renewer instance 创建 Renewer 实例
func NewRenewer(options *Options, store Store, sessionCodec SessionCodec, pool *RenewPoolManager) *Renewer {
	if sessionCodec == nil {
		sessionCodec = NewDefaultSessionCodec()
	}
	return &Renewer{
		options:      options,
		store:        store,
		sessionCodec: sessionCodec,
		pool:         pool,
	}
}

// ShouldRenew checks whether the session should be renewed 判断会话是否需要续期
func (r *Renewer) ShouldRenew(session *Session) bool {
	if session == nil {
		return false
	}
	now := gtime.Now().TimestampMilli()    // Current time 当前时间
	refTime := session.CreateTime          // Reference time 参考时间
	refreshNum := session.RefreshNum       // Renewal count 续期次数
	lastRenewTime := session.LastRenewTime // Last renewal time 上次续期时间
	if lastRenewTime > 0 {
		refTime = lastRenewTime
	}

	elapsed := now - refTime
	if r.options.MaxRefreshTimes > 0 && refreshNum >= r.options.MaxRefreshTimes {
		return false
	}
	remaining := r.options.Timeout - elapsed
	if remaining > r.options.MaxRefresh {
		return false
	}
	return true
}

// RenewAsync renews session asynchronously 异步续期会话
func (r *Renewer) RenewAsync(ctx context.Context, session *Session) {
	if session == nil || r.pool == nil || r.store == nil {
		return
	}
	err := r.pool.Submit(func() {
		sessionData, err := r.store.Load(ctx, session.UserKey)
		if err != nil {
			g.Log().Error(ctx, "[DToken] token renew load error", err) // Log renew load error 记录续期读取错误
			return
		}
		if sessionData == "" {
			return
		}
		currentSession, err := r.sessionCodec.Decode(ctx, sessionData)
		if err != nil {
			g.Log().Error(ctx, "[DToken] token renew decode error", err) // Log renew decode error 记录续期解码错误
			return
		}
		if currentSession == nil {
			return
		}
		if currentSession.Token != session.Token {
			return
		}
		if !r.ShouldRenew(currentSession) {
			return
		}

		currentSession.LastRenewTime = gtime.Now().TimestampMilli()
		currentSession.RefreshNum++
		sessionData, err = r.sessionCodec.Encode(ctx, currentSession)
		if err != nil {
			g.Log().Error(ctx, "[DToken] token renew encode error", err) // Log renew encode error 记录续期编码错误
			return
		}
		if err = r.store.Save(ctx, session.UserKey, sessionData); err != nil {
			g.Log().Error(ctx, "[DToken] token renew save error", err) // Log renew save error 记录续期保存错误
			return
		}
	})
	if err != nil {
		g.Log().Error(ctx, "[DToken] token renew submit error", err) // Log renew submit error 记录续期任务提交错误
	}
}

// Shutdown gracefully stops renewal pool 优雅关闭续期协程池
func (r *Renewer) Shutdown(ctx context.Context) {
	if r == nil || r.pool == nil {
		return
	}
	g.Log().Info(ctx, "[DToken] renew pool manager closed")
	r.pool.Stop()
}
