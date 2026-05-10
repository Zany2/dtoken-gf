# DToken-GF

基于 GoFrame v2 的轻量级 Token 认证组件，提供 Token 生成、认证中间件、会话存储、自动续期和可插拔编解码能力。

## 特性

- Token 编解码：默认使用 AES 加密 + Base64 编码，将 `userKey` 编入 Token。
- 会话存储：默认支持 `gcache`、`gredis`、`gfile`。
- 会话编解码：默认使用 `gjson` 编解码 `Session`。
- 自动续期：Token 剩余有效期小于等于 `MaxRefresh` 时异步续期。
- GoFrame 中间件：支持认证拦截、自定义认证失败响应、免认证路径。
- 可配置请求头：默认从 `Authorization: Bearer <token>` 获取 Token，也可自定义 Header Key。

## 安装

```bash
go get github.com/Zany2/dtoken-gf
```

要求：Go 1.23+，GoFrame v2.9+。

## 快速开始

### 代码方式

```go
package main

import (
    "context"
    "time"

    dtoken "github.com/Zany2/dtoken-gf/dtoken-gf"
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/net/ghttp"
)

func main() {
    token, err := dtoken.New(
        dtoken.WithCacheMode(dtoken.CacheModeCache),
        dtoken.WithCachePreKey("MyApp:"),
        dtoken.WithTimeout(10*24*time.Hour),
        dtoken.WithMaxRefresh(5*24*time.Hour),
        dtoken.WithMaxRefreshTimes(0),
        dtoken.WithEncryptKey([]byte("12345678912345678912345678912345")),
        dtoken.WithReuseToken(true),
        dtoken.WithAuthHeaderKey("Authorization"),
        dtoken.WithAuthExcludePaths("/login", "/public/*"),
        dtoken.WithRenewPool(10, 200, 0.8, 0.3),
        dtoken.WithRenewPoolCheckInterval(time.Minute),
    )
    if err != nil {
        panic(err)
    }
    defer token.Shutdown(context.Background())

    authMiddleware := dtoken.NewDefaultMiddleware(token)
    s := g.Server()

    s.Group("/", func(group *ghttp.RouterGroup) {
        group.Middleware(authMiddleware.Auth)

        group.POST("/login", func(r *ghttp.Request) {
            tokenValue, err := token.Generate(r.Context(), "user_001", g.Map{
                "userName": "test1",
            })
            if err != nil {
                r.Response.WriteJsonExit(g.Map{"code": 500, "message": err.Error()})
            }
            r.Response.WriteJsonExit(g.Map{"code": 0, "token": tokenValue, "tokenType": "Bearer"})
        })

        group.GET("/profile", func(r *ghttp.Request) {
            userKey := r.GetCtxVar(dtoken.KeyUserKey).String()
            data := r.GetCtxVar(dtoken.KeyData).Map()
            r.Response.WriteJsonExit(g.Map{"code": 0, "userKey": userKey, "data": data})
        })
    })

    s.Run()
}
```

### 配置文件方式

GoFrame 配置示例：

```yaml
redis:
  default:
    address: 127.0.0.1:6379
    db: 0
    pass: ""

dToken:
  cacheMode: 2
  cachePreKey: "DToken:"
  timeout: 60000
  maxRefresh: 30000
  maxRefreshTimes: 0
  tokenDelimiter: "_"
  encryptKey: "12345678912345678912345678912345"
  multiLogin: false
  authHeaderKey: "Authorization"
  authExcludePaths:
    - "/api/v1/health"
    - "/api/v1/auth/login"
  poolMinSize: 10
  poolMaxSize: 200
  poolScaleUpRate: 0.8
  poolScaleDownRate: 0.3
  poolCheckInterval: 60000
```

初始化：

```go
token, err := dtoken.NewFromConfig(ctx)
if err != nil {
    return err
}
```

`cacheMode: 2` 使用 Redis，需要先配置 GoFrame Redis；没有 Redis 环境时可改为 `cacheMode: 1` 使用内存缓存。

## 配置参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| CacheMode | int8 | 1 | 缓存模式：1=gcache，2=gredis，3=gfile；0 会重置为 1 |
| CachePreKey | string | `DToken:` | 存储 key 前缀 |
| Timeout | int64 | 864000000 | Token 过期时间，单位毫秒 |
| MaxRefresh | int64 | Timeout/2 | 剩余有效期小于等于该值时触发续期；0 或负数使用默认值 |
| MaxRefreshTimes | int | 0 | 最大续期次数，0 表示不限；负数会重置为 0 |
| EncryptKey | []byte | 内置默认值 | AES 密钥，长度必须为 16、24 或 32 字节 |
| TokenDelimiter | string | `_` | Token 内部分隔符，仅支持 `_ - . : \| ~` 中的单字符 |
| MultiLogin | bool | false | 同一 `userKey` 重复登录时是否复用已有 Token |
| AuthHeaderKey | string | `Authorization` | 读取 Token 的请求头名称 |
| AuthExcludePaths | []string | 空 | 免认证路径 |
| PoolMinSize | int | 20 | 续期协程池最小容量 |
| PoolMaxSize | int | 2000 | 续期协程池最大容量，小于最小容量时会自动修正 |
| PoolScaleUpRate | float64 | 0.8 | 协程池扩容阈值，需满足 `0 < down < up <= 1` |
| PoolScaleDownRate | float64 | 0.3 | 协程池缩容阈值，需满足 `0 < down < up <= 1` |
| PoolCheckInterval | int64 | 60000 | 协程池扩缩容检查间隔，单位毫秒 |

生产环境建议显式配置 `EncryptKey`，不要直接使用默认密钥。

## 核心接口

### Token

```go
type Token interface {
    Generate(ctx context.Context, userKey string, data g.Map) (token string, err error)
    ValidateSession(ctx context.Context, token string) (*Session, error)
    GetSession(ctx context.Context, userKey string) (*Session, error)
    ParseUserKey(ctx context.Context, token string) (userKey string, err error)
    DestroyByToken(ctx context.Context, token string) error
    Destroy(ctx context.Context, userKey string) error
    Shutdown(ctx context.Context)
    GetOptions() Options
}
```

推荐通过 `New`、`NewToken` 或 `NewFromConfig` 创建实例，不建议手动构造 `DTokenV2`。

### Store

`Store` 只负责保存、读取和删除编码后的会话字符串，不直接处理 `Session`：

```go
type Store interface {
    Save(ctx context.Context, userKey string, data string) error
    Load(ctx context.Context, userKey string) (string, error)
    Delete(ctx context.Context, userKey string) error
}
```

自定义存储：

```go
token, err := dtoken.New(
    dtoken.WithStore(myStore),
)
```

自定义 `Store` 需要自行处理 TTL/过期策略；框架只会按接口读写会话字符串，不会替自定义存储自动设置过期时间，`CacheMode` 也不会影响自定义存储的过期行为。

### TokenCodec

`TokenCodec` 负责 `userKey` 和 Token 之间的编解码：

```go
type TokenCodec interface {
    Encode(ctx context.Context, userKey string) (token string, err error)
    Decode(ctx context.Context, token string) (userKey string, err error)
}
```

默认实现：

```go
dtoken.NewDefaultTokenCodec(delimiter, encryptKey)
```

### SessionCodec

`SessionCodec` 负责 `Session` 和存储字符串之间的编解码，默认使用 `gjson`：

```go
type SessionCodec interface {
    Encode(ctx context.Context, session *Session) (string, error)
    Decode(ctx context.Context, data string) (*Session, error)
}
```

默认实现：

```go
dtoken.NewDefaultSessionCodec()
```

## 中间件

默认认证失败返回 GoFrame 的 `gcode.CodeNotAuthorized`。也可以传入自定义处理方法：

```go
middleware := dtoken.NewDefaultMiddleware(token, func(r *ghttp.Request) {
    r.Response.WriteJsonExit(g.Map{
        "code":    401,
        "message": "认证过期，请重新登录",
        "data":    []interface{}{},
    })
})
```

认证失败包括：未传 Token、Token 格式错误、Token 无效、Token 过期或会话不存在。

认证成功后，中间件会写入 GoFrame Request CtxVar：

| Key | 内容 |
|-----|------|
| `dtoken.KeyUserKey` | 当前用户标识 |
| `dtoken.KeyData` | `Generate` 时写入的扩展数据 |

读取示例：

```go
request := ghttp.RequestFromCtx(ctx)
userKey := request.GetCtxVar(dtoken.KeyUserKey).String()
data := request.GetCtxVar(dtoken.KeyData).Map()
```

这些值是通过 `r.SetCtxVar` 写入的，不是 `context.WithValue`，因此不要用 `ctx.Value()` 读取。

## Token 提取方式

中间件按以下顺序提取 Token：

1. 请求头：`<AuthHeaderKey>: Bearer <token>`，默认是 `Authorization: Bearer <token>`
2. 请求参数：`token`

当前请求头格式要求严格匹配 `Bearer <token>`。

## 免认证路径

`AuthExcludePaths` 支持：

- 精确匹配：`/login`
- 前缀通配：`/api/*`

当前前缀通配会去掉末尾的 `/*` 后做路径边界匹配，例如 `/api/*` 会匹配 `/api` 和 `/api/...`，不会匹配 `/apiary`。

## 自动续期

当 Token 剩余有效期小于等于 `MaxRefresh` 时，认证请求会触发异步续期。

- 续期不会阻塞当前请求。
- 续期前会重新读取当前会话，并校验 Token 一致性，避免旧 Token 覆盖新 Token。
- `MaxRefreshTimes` 可限制最大续期次数。
- `CacheModeFile` 会在认证时额外检查会话时间，避免文件缓存重启后恢复过期会话。

## 缓存模式

- `CacheModeCache`：使用 GoFrame `gcache` 内存缓存。
- `CacheModeRedis`：使用 GoFrame Redis 适配器，需要提前配置 Redis。
- `CacheModeFile`：使用内存缓存并将数据落盘到临时目录文件，适合轻量场景。

## 日志策略

框架日志遵循：

- 遇到错误时打印日志。
- 关键节点打印日志，例如配置被自动修正、续期池关闭。
- 普通成功路径不打印日志，例如 Token 生成成功、销毁成功、续期成功、协程池扩缩容。

启动 banner 可通过 `WithBanner(false)` 关闭。

## 代码结构

| 文件 | 职责 |
|------|------|
| `factory.go` | Token 创建入口 |
| `config.go` | 函数式配置选项 |
| `options.go` | 配置结构、默认值和校验 |
| `token.go` | Token 生成、校验、解析、销毁 |
| `store.go` | 原始会话数据存储接口 |
| `session.go` | Session 数据结构 |
| `codec_token.go` | Token 编解码 |
| `codec_session.go` | Session 编解码 |
| `cache.go` | 默认 Store 实现 |
| `middleware.go` | GoFrame HTTP 认证中间件 |
| `renewer.go` | 自动续期 |
| `pool.go` | 续期协程池 |
| `banner.go` | 启动 banner |
| `constants.go` | 常量和错误消息 |

## License

MIT
