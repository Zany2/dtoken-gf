package dtoken_gf

import (
	"context"
	"errors"
	"strings"

	"github.com/gogf/gf/v2/crypto/gaes"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/util/grand"
)

// TokenEncoder defines the token encoder interface Token 编码器接口定义
type TokenEncoder interface {
	Encode(ctx context.Context, userKey string) (token string, err error) // Encode method to generate token 生成 Token 的编码方法
}

// TokenDecoder defines the token decoder interface Token 解码器接口定义
type TokenDecoder interface {
	Decode(ctx context.Context, token string) (userKey string, err error) // Decode method to extract userKey from token 解码方法，从 Token 中提取用户标识
}

// TokenCodec combines token encoder and decoder TokenCodec 组合 Token 编码器和解码器
type TokenCodec interface {
	TokenEncoder
	TokenDecoder
}

// DefaultTokenCodec default implementation of TokenCodec 默认 Token 编解码实现
type DefaultTokenCodec struct {
	// Delimiter used to separate userKey and random string 用于分隔用户标识和随机字符串的分隔符
	Delimiter string
	// EncryptKey for token encryption Token 加密密钥
	EncryptKey []byte
}

// NewDefaultTokenCodec creates a new DefaultTokenCodec instance 创建一个新的默认 Token 编解码器实例
func NewDefaultTokenCodec(delimiter string, encryptKey []byte) *DefaultTokenCodec {
	return &DefaultTokenCodec{
		Delimiter:  delimiter,
		EncryptKey: encryptKey,
	}
}

// Encode method to generate token 编码方法生成 Token
func (c *DefaultTokenCodec) Encode(ctx context.Context, userKey string) (token string, err error) {
	if userKey == "" {
		return "", errors.New(MsgErrUserKeyEmpty) // Error when userKey is empty 用户标识为空时返回错误
	}
	// Generate a random string 生成一个随机字符串
	randStr, err := gmd5.Encrypt(grand.Letters(10))
	if err != nil {
		return "", err
	}
	encryptBeforeStr := userKey + c.Delimiter + randStr
	// Encrypt the combined string 加密拼接后的字符串
	encryptByte, err := gaes.Encrypt([]byte(encryptBeforeStr), c.EncryptKey)
	if err != nil {
		return "", err
	}
	// Return base64 encoded token 返回 Base64 编码后的 Token
	return gbase64.EncodeToString(encryptByte), nil
}

// Decode method to extract userKey from token 解码方法从 Token 中提取用户标识
func (c *DefaultTokenCodec) Decode(ctx context.Context, token string) (userKey string, err error) {
	if token == "" {
		return "", errors.New(MsgErrTokenEmpty) // Error when token is empty Token 为空时返回错误
	}
	// Decode the base64 token 解码 Base64 编码的 Token
	token64, err := gbase64.Decode([]byte(token))
	if err != nil {
		return "", err
	}
	// Decrypt the decoded token 解密解码后的 Token
	decryptStr, err := gaes.Decrypt(token64, c.EncryptKey)
	if err != nil {
		return "", err
	}
	// Split by the last delimiter to allow delimiter in userKey 按最后一个分隔符拆分，允许 userKey 包含分隔符
	decryptText := string(decryptStr)
	delimiterIndex := strings.LastIndex(decryptText, c.Delimiter)
	if delimiterIndex <= 0 {
		return "", errors.New(MsgErrTokenLen) // Error when the token length is invalid Token 长度无效时返回错误
	}
	return decryptText[:delimiterIndex], nil
}
