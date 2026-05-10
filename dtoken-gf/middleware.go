package dtoken_gf

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
)

// AuthFailHandler handles authentication failure response 认证失败响应处理方法
type AuthFailHandler func(r *ghttp.Request)

// Middleware defines the authentication middleware 认证中间件结构体
type Middleware struct {
	Token           Token           // Token instance Token 实例
	AuthFailHandler AuthFailHandler // Custom authentication failure handler 自定义认证失败处理方法
}

// NewDefaultMiddleware creates a middleware instance and uses authFailHandler when provided 创建默认中间件实例，传入 authFailHandler 时使用自定义认证失败处理方法
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

// defaultAuthFailHandler writes default auth failure response 默认认证失败响应
func defaultAuthFailHandler(r *ghttp.Request) {
	r.Response.WriteJson(ghttp.DefaultHandlerResponse{
		Code:    gcode.CodeNotAuthorized.Code(),
		Message: gcode.CodeNotAuthorized.Message(),
		Data:    []interface{}{},
	})
}

// handleAuthFail calls configured handler or default handler 调用配置的认证失败处理方法或默认方法
func (m Middleware) handleAuthFail(r *ghttp.Request) {
	if m.AuthFailHandler != nil {
		m.AuthFailHandler(r)
		return
	}
	defaultAuthFailHandler(r)
}

// Auth performs token authentication and returns a unified response on failure 执行请求认证拦截，校验失败时返回统一错误响应
func (m Middleware) Auth(r *ghttp.Request) {
	if m.Token == nil {
		// Fail safely when token instance is missing Token 实例为空时安全失败
		m.handleAuthFail(r)
		return
	}

	// Skip authentication if path is excluded 路径在排除列表中则跳过认证
	if m.HasExcludePath(r) {
		r.Middleware.Next()
		return
	}

	// Extract token from request 从请求中获取 Token
	token, err := GetRequestToken(r, m.Token.GetOptions().AuthHeaderKey)
	if err != nil {
		m.handleAuthFail(r)
		return
	}

	// Validate token and retrieve session 校验 Token 并获取会话
	session, err := m.Token.ValidateSession(r.Context(), token)
	if err != nil {
		m.handleAuthFail(r)
		return
	}

	// Store session info in request context 将会话信息存入请求上下文
	r.SetCtxVar(KeyUserKey, session.UserKey)
	r.SetCtxVar(KeyData, session.Data)

	// Continue request 执行后续中间件链
	r.Middleware.Next()
}

// HasExcludePath determines if the current request path should bypass authentication 判断路径是否应跳过认证
func (m Middleware) HasExcludePath(r *ghttp.Request) bool {
	if m.Token == nil {
		return false
	}

	var (
		urlPath      = r.URL.Path
		excludePaths = m.Token.GetOptions().AuthExcludePaths
	)

	// No exclusion rules configured 未配置排除路径
	if len(excludePaths) == 0 {
		return false
	}

	// Remove trailing slash 去除路径末尾斜杠
	if strings.HasSuffix(urlPath, "/") {
		urlPath = gstr.SubStr(urlPath, 0, len(urlPath)-1)
	}

	// Iterate through exclude paths 遍历排除路径规则
	for _, excludePath := range excludePaths {
		tmpPath := excludePath

		// Prefix match, for example "/api/*" 前缀匹配，例如 /api/*
		if strings.HasSuffix(tmpPath, "/*") {
			tmpPath = gstr.SubStr(tmpPath, 0, len(tmpPath)-2)
			if urlPath == tmpPath || gstr.HasPrefix(urlPath, tmpPath+"/") {
				// Path matches prefix, skip authentication 匹配前缀路径则跳过认证
				return true
			}
		} else {
			// Full path match 全路径匹配
			if strings.HasSuffix(tmpPath, "/") {
				tmpPath = gstr.SubStr(tmpPath, 0, len(tmpPath)-1)
			}
			if urlPath == tmpPath {
				// Exact match, skip authentication 精确匹配则跳过认证
				return true
			}
		}
	}

	// No exclusion match, require authentication 未匹配排除规则则需认证
	return false
}

// GetRequestToken extracts token from auth header or token param 从认证请求头或 token 参数中提取 Token
func GetRequestToken(r *ghttp.Request, headerKey ...string) (string, error) {
	authHeaderKey := DefaultAuthHeaderKey // Default auth header key 默认认证请求头名称
	if len(headerKey) > 0 && headerKey[0] != "" {
		authHeaderKey = headerKey[0]
	}

	// 1. Try auth header 优先从认证请求头中获取
	authHeader := r.Header.Get(authHeaderKey)
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)

		// Validate Bearer token format 校验 Bearer 格式是否正确
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "Bearer param invalid | Bearer 参数格式错误")
		} else if parts[1] == "" {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "Bearer param empty | Bearer 参数为空")
		}

		return parts[1], nil
	}

	// 2. Fallback to token parameter 尝试从请求参数中读取 Token
	authHeader = r.Get(KeyToken).String()
	if authHeader == "" {
		return "", gerror.NewCode(gcode.CodeMissingParameter, "token empty | 缺少 token 参数")
	}

	return authHeader, nil
}
