package middleware

import (
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gvalid"
)

// Cors cors middleware 跨域中间件
func Cors() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		r.Response.CORSDefault()
		r.Middleware.Next()
	}
}

// HandlerResponseMiddleware unified response middleware 统一响应中间件
func HandlerResponseMiddleware() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		r.Middleware.Next()

		if r.Response.BufferLength() > 0 {
			return
		}

		var (
			err = r.GetError()
			res = r.GetHandlerResponse()
		)

		if err != nil {
			// HandlerResponseMiddleware validate error 参数校验错误
			if _, ok := err.(gvalid.Error); ok {
				r.Response.WriteJsonExit(g.Map{
					"code":    400,
					"message": err.Error(),
					"data":    []interface{}{},
				})
				return
			}

			// HandlerResponseMiddleware coded error 业务错误
			if code := gerror.Code(err); code != gcode.CodeNil {
				r.Response.WriteJsonExit(g.Map{
					"code":    code.Code(),
					"message": err.Error(),
					"data":    []interface{}{},
				})
				return
			}

			// HandlerResponseMiddleware unexpected error 非预期错误
			g.Log().Line().Errorf(r.Context(), "%+v", err)
			r.Response.WriteJsonExit(g.Map{
				"code":    500,
				"message": "服务异常",
				"data":    []interface{}{},
			})
			return
		}

		// EmptyResponseFallback empty response fallback 空响应回退为空数组
		if g.IsNil(res) {
			res = []interface{}{}
		}

		r.Response.WriteJsonExit(g.Map{
			"code":    200,
			"message": "",
			"data":    res,
		})
	}
}
