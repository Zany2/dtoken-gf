package auth

import (
	"context"
	"dtoken-gf-example/internal/consts"
	"dtoken-gf-example/token"

	"github.com/gogf/gf/v2/frame/g"

	"dtoken-gf-example/api/auth/v1"
)

// Login handles user login 用户登录
func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	// Find demo user by user name 根据用户名查找示例用户
	userValue, ok := consts.UserDataMap.Search(req.UserName)
	if !ok {
		g.RequestFromCtx(ctx).Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "账号或密码错误",
			"data":    []interface{}{},
		})
		return
	}

	// Check password 验证密码
	user, _ := userValue.(consts.UserData)
	if user.Password != req.Password {
		g.RequestFromCtx(ctx).Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "账号或密码错误",
			"data":    []interface{}{},
		})
		return
	}

	// user data 生成data
	userData := g.Map{
		consts.CtxUserId:   user.Id,
		consts.CtxUserName: user.UserName,
	}

	// Generate token with user id 使用用户 ID 生成 Token
	tokenValue, err := token.DToken.Generate(ctx, user.Id, userData)
	//tokenValue, err := token.DToken.Generate(ctx, user.Id, nil)
	if err != nil {
		return nil, err
	}

	return &v1.LoginRes{
		Token: tokenValue,
	}, nil
}
