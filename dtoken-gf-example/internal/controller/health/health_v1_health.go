package health

import (
	"context"

	"dtoken-gf-example/api/health/v1"
	"github.com/gogf/gf/v2/os/gtime"
)

// Health returns service health status 返回服务健康状态
func (c *ControllerV1) Health(ctx context.Context, req *v1.HealthReq) (res *v1.HealthRes, err error) {
	// Return service health status 返回服务健康状态
	return &v1.HealthRes{
		Service: "dtoken-gf-example",
		Status:  "ok",
		Time:    gtime.Now().String(),
	}, nil
}
