package auth

import (
	"context"
	"dtoken-gf-example/api/auth/v1"
	"dtoken-gf-example/token"
	dtoken "github.com/Zany2/dtoken-gf/dtoken-gf"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Logout handles current user logout 当前用户登出
func (c *ControllerV1) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	// Read current user from auth context 从认证上下文读取当前用户
	request := ghttp.RequestFromCtx(ctx)

	userKey := request.GetCtxVar(dtoken.KeyUserKey).String()
	err = token.DToken.Destroy(ctx, userKey)
	if err != nil {
		return nil, err
	}

	return
}
