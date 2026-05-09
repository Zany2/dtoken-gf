package dtoken_gf

import (
	"context"
)

// Store defines raw session storage 原始会话存储接口
type Store interface {
	// Save saves encoded session data 保存编码后的会话数据
	Save(ctx context.Context, userKey string, data string) error
	// Load retrieves encoded session data 获取编码后的会话数据
	Load(ctx context.Context, userKey string) (string, error)
	// Delete deletes encoded session data 删除编码后的会话数据
	Delete(ctx context.Context, userKey string) error
}
