package dtoken_gf

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/encoding/gjson"
)

// SessionEncoder defines the session encoder interface Session 编码器接口定义
type SessionEncoder interface {
	Encode(ctx context.Context, session *Session) (string, error) // Encode session to storage text 将 Session 编码为存储文本
}

// SessionDecoder defines the session decoder interface Session 解码器接口定义
type SessionDecoder interface {
	Decode(ctx context.Context, data string) (*Session, error) // Decode storage text to session 将存储文本解码为 Session
}

// SessionCodec combines session encoder and decoder SessionCodec 组合 Session 编码器和解码器
type SessionCodec interface {
	SessionEncoder
	SessionDecoder
}

// DefaultSessionCodec serializes session by gjson 默认使用 gjson 编解码 Session
type DefaultSessionCodec struct{}

// NewDefaultSessionCodec creates default session codec 创建默认 Session 编解码器
func NewDefaultSessionCodec() *DefaultSessionCodec {
	return &DefaultSessionCodec{}
}

// Encode encodes session with gjson 使用 gjson 编码 Session
func (c *DefaultSessionCodec) Encode(ctx context.Context, session *Session) (string, error) {
	if session == nil {
		return "", errors.New(MsgErrDataEmpty) // Error if session is empty 如果会话为空，返回错误
	}
	value, err := gjson.Encode(session.Map()) // Encode session map to JSON 将会话映射编码为 JSON
	if err != nil {
		return "", err
	}
	return string(value), nil
}

// Decode decodes session with gjson 使用 gjson 解码 Session
func (c *DefaultSessionCodec) Decode(ctx context.Context, data string) (*Session, error) {
	if data == "" {
		return nil, errors.New(MsgErrDataEmpty) // Error if storage data is empty 如果存储数据为空，返回错误
	}
	sessionJson, err := gjson.DecodeToJson(data) // Decode JSON text to map 将 JSON 文本解析为映射
	if err != nil {
		return nil, err
	}
	return NewSessionFromMap(sessionJson.Map()), nil
}
