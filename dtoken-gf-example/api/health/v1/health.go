// @Author daixk 2026/5/9 22:39:00
package v1

import "github.com/gogf/gf/v2/frame/g"

// HealthReq health check request 健康检查请求
type HealthReq struct {
	g.Meta `path:"/" method:"get" tags:"Health" summary:"Health check"`
}

// HealthRes health check response 健康检查响应
type HealthRes struct {
	Service string `json:"service" dc:"Service name"`  // Service name 服务名称
	Status  string `json:"status" dc:"Service status"` // Service status 服务状态
	Time    string `json:"time" dc:"Server time"`      // Server time 服务时间
}
