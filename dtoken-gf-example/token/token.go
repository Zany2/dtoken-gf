// @Author daixk 2026/5/9 23:02:00
package token

import (
	"context"

	dtoken "github.com/Zany2/dtoken-gf/v2/dtoken-gf"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var (
	DToken dtoken.Token // DToken default token instance 默认 dToken 实例

	// LoginExpiredHandler handles auth failure response 登录过期处理方法
	LoginExpiredHandler = func(r *ghttp.Request) {
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "认证过期，请重新登录",
			"data":    []interface{}{},
		})
	}
)

// Init initializes dToken from GoFrame config 从 GoFrame 配置初始化 dToken
func Init(ctx context.Context) error {
	// Load dToken options from config 从配置读取 dToken 参数
	token, err := dtoken.NewFromConfig(ctx)
	if err != nil {
		return err
	}
	DToken = token
	return nil
}
