package cmd

import (
	"context"
	"os"

	"dtoken-gf-example/internal/controller/auth"
	"dtoken-gf-example/internal/controller/health"
	"dtoken-gf-example/internal/controller/user"
	"dtoken-gf-example/middleware"
	"dtoken-gf-example/token"

	dtoken "github.com/Zany2/dtoken-gf/v2/dtoken-gf"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gproc"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()

			// Initialize dToken before binding routes 路由绑定前初始化 dToken
			if err = token.Init(ctx); err != nil {
				return err
			}

			s.Group("/api/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(
					middleware.Cors(), // Cors middleware handles cross-origin requests 跨域请求处理中间件
					dtoken.NewDefaultMiddleware(token.DToken, token.LoginExpiredHandler).Auth, // dToken auth middleware with custom expired handler dToken 认证中间件，使用自定义过期处理
					middleware.HandlerResponseMiddleware(),                                    // Unified response middleware 统一响应处理中间件
				)

				// Register health routes 注册健康检查路由
				group.Group("/health", func(group *ghttp.RouterGroup) {
					group.Bind(health.NewV1())
				})

				// Register auth routes 注册认证路由
				group.Group("/auth", func(group *ghttp.RouterGroup) {
					group.Bind(auth.NewV1())
				})

				// Register user routes 注册用户路由
				group.Group("/user", func(group *ghttp.RouterGroup) {
					group.Bind(user.NewV1())
				})
			})

			// Shutdown dToken renew pool on process exit 进程退出时回收 dToken 续期池
			gproc.AddSigHandlerShutdown(func(sig os.Signal) {
				token.DToken.Shutdown(ctx)
			})

			s.Run()
			return nil
		},
	}
)
