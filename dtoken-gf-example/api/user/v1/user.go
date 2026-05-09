// @Author daixk 2026/5/9 22:39:00
package v1

import "github.com/gogf/gf/v2/frame/g"

// InfoReq profile request 用户信息请求
type InfoReq struct {
	g.Meta `path:"/info" method:"get" tags:"User" summary:"Get user info"`
}

// InfoRes profile response 用户信息响应
type InfoRes struct {
	Id       string `json:"id" dc:"User id"`               // User id 用户 ID
	UserName string `json:"userName" dc:"Login user name"` // Login user name 登录用户名
}
