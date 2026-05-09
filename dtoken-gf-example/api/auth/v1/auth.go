// @Author daixk 2026/5/9 22:39:00
package v1

import "github.com/gogf/gf/v2/frame/g"

// LoginReq login request 登录请求
type LoginReq struct {
	g.Meta   `path:"/login" method:"post" tags:"Auth" summary:"Login"`
	UserName string `json:"userName" v:"required#userName is required" dc:"Login user name"` // Login user name 登录用户名
	Password string `json:"password" v:"required#password is required" dc:"Login password"`  // Login password 登录密码
}

// LoginRes login response 登录响应
type LoginRes struct {
	Token string `json:"token" dc:"Generated token"` // Generated token 生成的 Token
}

// LogoutReq logout request 登出请求
type LogoutReq struct {
	g.Meta `path:"/logout" method:"post" tags:"Auth" summary:"Logout"`
}

// LogoutRes logout response 登出响应
type LogoutRes struct{}
