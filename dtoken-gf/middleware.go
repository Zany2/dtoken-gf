package dtoken_gf

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
)

// AuthFailHandler handles authentication failure response. 认证失败处理函数。
type AuthFailHandler func(r *ghttp.Request)

// Middleware defines the authentication middleware. 认证中间件。
type Middleware struct {
	Token           Token
	AuthFailHandler AuthFailHandler
}

// NewDefaultMiddleware creates a middleware instance. 创建默认中间件实例。
func NewDefaultMiddleware(token Token, authFailHandler ...AuthFailHandler) Middleware {
	handler := AuthFailHandler(defaultAuthFailHandler)
	if len(authFailHandler) > 0 && authFailHandler[0] != nil {
		handler = authFailHandler[0]
	}
	return Middleware{
		Token:           token,
		AuthFailHandler: handler,
	}
}

// defaultAuthFailHandler writes default auth failure response. 写入默认认证失败响应。
func defaultAuthFailHandler(r *ghttp.Request) {
	r.Response.WriteJson(ghttp.DefaultHandlerResponse{
		Code:    gcode.CodeNotAuthorized.Code(),
		Message: gcode.CodeNotAuthorized.Message(),
		Data:    []interface{}{},
	})
}

// handleAuthFail calls configured handler or default handler. 调用自定义或默认失败处理器。
func (m Middleware) handleAuthFail(r *ghttp.Request) {
	if m.AuthFailHandler != nil {
		m.AuthFailHandler(r)
		return
	}
	defaultAuthFailHandler(r)
}

// Auth performs token authentication and returns a unified response on failure. 执行认证，失败时返回统一响应。
func (m Middleware) Auth(r *ghttp.Request) {
	if m.Token == nil {
		m.handleAuthFail(r)
		return
	}

	if m.HasExcludePath(r) {
		r.Middleware.Next()
		return
	}

	token, err := GetRequestToken(r, m.Token.GetOptions().AuthHeaderKey)
	if err != nil {
		m.handleAuthFail(r)
		return
	}

	session, err := m.Token.ValidateSession(r.Context(), token)
	if err != nil {
		m.handleAuthFail(r)
		return
	}

	r.SetCtxVar(KeyUserKey, session.UserKey)
	r.SetCtxVar(KeyData, session.Data)
	r.Middleware.Next()
}

// HasExcludePath determines if the current request path should bypass authentication. 判断当前请求是否命中免认证路径。
func (m Middleware) HasExcludePath(r *ghttp.Request) bool {
	if m.Token == nil {
		return false
	}

	urlPath := r.URL.Path
	excludePaths := m.Token.GetOptions().AuthExcludePaths
	if len(excludePaths) == 0 {
		return false
	}

	if strings.HasSuffix(urlPath, "/") {
		urlPath = gstr.SubStr(urlPath, 0, len(urlPath)-1)
	}

	for _, excludePath := range excludePaths {
		tmpPath := excludePath
		if strings.HasSuffix(tmpPath, "/*") {
			tmpPath = gstr.SubStr(tmpPath, 0, len(tmpPath)-2)
			if urlPath == tmpPath || gstr.HasPrefix(urlPath, tmpPath+"/") {
				return true
			}
		} else {
			if strings.HasSuffix(tmpPath, "/") {
				tmpPath = gstr.SubStr(tmpPath, 0, len(tmpPath)-1)
			}
			if urlPath == tmpPath {
				return true
			}
		}
	}

	return false
}

// GetRequestToken extracts token from auth header or token param. 从请求头或 token 参数中提取 token。
func GetRequestToken(r *ghttp.Request, headerKey ...string) (string, error) {
	authHeaderKey := DefaultAuthHeaderKey
	if len(headerKey) > 0 && headerKey[0] != "" {
		authHeaderKey = headerKey[0]
	}

	authHeader := r.Header.Get(authHeaderKey)
	if authHeader != "" {
		return parseBearerToken(authHeader)
	}

	token := r.Get(KeyToken).String()
	if token == "" {
		return "", gerror.NewCode(gcode.CodeMissingParameter, "token empty | token 为空")
	}

	return token, nil
}

// parseBearerToken extracts token from Bearer auth header. 从 Bearer 请求头中提取 token。
func parseBearerToken(authHeader string) (string, error) {
	parts := strings.Fields(authHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "Bearer param invalid | Bearer 参数格式错误")
	}
	if parts[1] == "" {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "Bearer param empty | Bearer 参数为空")
	}
	return parts[1], nil
}
