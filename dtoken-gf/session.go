package dtoken_gf

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

// Session stores token state for one userKey Session 保存单个 userKey 的 Token 状态
type Session struct {
	UserKey       string // User identifier 用户标识
	Token         string // Token value Token 值
	Data          g.Map  // Extra data 附加数据
	RefreshNum    int    // Renewal count 续期次数
	CreateTime    int64  // Creation timestamp in milliseconds 创建时间戳，单位毫秒
	LastRenewTime int64  // Last renewal timestamp in milliseconds 上次续期时间戳，单位毫秒
}

// NewSessionFromMap converts map to Session 将 map 转换为 Session
func NewSessionFromMap(cacheValue g.Map) *Session {
	if cacheValue == nil {
		return nil
	}
	return &Session{
		UserKey:       gconv.String(cacheValue[KeyUserKey]),
		Token:         gconv.String(cacheValue[KeyToken]),
		Data:          gconv.Map(cacheValue[KeyData]),
		RefreshNum:    gconv.Int(cacheValue[KeyRefreshNum]),
		CreateTime:    gconv.Int64(cacheValue[KeyCreateTime]),
		LastRenewTime: gconv.Int64(cacheValue[KeyLastRenewTime]),
	}
}

// Map converts Session to map 将 Session 转换为 map
func (s *Session) Map() g.Map {
	if s == nil {
		return nil
	}
	return g.Map{
		KeyUserKey:       s.UserKey,
		KeyToken:         s.Token,
		KeyData:          s.Data,
		KeyRefreshNum:    s.RefreshNum,
		KeyCreateTime:    s.CreateTime,
		KeyLastRenewTime: s.LastRenewTime,
	}
}
