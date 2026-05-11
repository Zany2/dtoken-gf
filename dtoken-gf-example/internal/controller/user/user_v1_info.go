package user

import (
	"context"
	"dtoken-gf-example/internal/consts"

	dtoken "github.com/Zany2/dtoken-gf/v2/dtoken-gf"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"dtoken-gf-example/api/user/v1"
)

// Info returns current user information 获取当前用户信息
func (c *ControllerV1) Info(ctx context.Context, req *v1.InfoReq) (res *v1.InfoRes, err error) {
	// Read current user from auth context 从认证上下文读取当前用户
	request := ghttp.RequestFromCtx(ctx)
	userKey := request.GetCtxVar(dtoken.KeyUserKey).String()
	data := request.GetCtxVar(dtoken.KeyData).Map()

	return &v1.InfoRes{
		Id:       userKey,
		UserName: gconv.String(data[consts.CtxUserName]),
	}, nil
}
