# DToken-GF

[English](README.md) | 简体中文

DToken-GF 是一个面向 [GoFrame v2](https://goframe.org/) 的轻量级 Token 认证组件，提供 Token 生成、Token 验证中间件、会话存储、自动续期和可插拔编解码能力，适合在 GoFrame HTTP 服务中快速接入登录认证。

## 功能特性

- 提供 GoFrame 中间件，可保护路由并将认证后的用户信息写入请求上下文变量。
- 默认使用 AES + Base64 进行 Token 编解码。
- 支持通过 GoFrame `gcache`、GoFrame Redis 适配器、文件缓存或自定义存储保存会话。
- 默认使用 `gjson` 序列化会话，也支持自定义会话编解码器。
- 当 Token 剩余有效期达到配置阈值时，异步自动续期。
- 支持配置认证请求头、免认证路径、Token 分隔符、过期时间、续期次数和续期协程池。

## 环境要求

- Go 1.23+
- GoFrame v2.9+

## 安装

```bash
go get github.com/Zany2/dtoken-gf/v2
```

由于实现代码位于 `dtoken-gf` 子目录，建议使用别名导入：

```go
import dtoken "github.com/Zany2/dtoken-gf/v2/dtoken-gf"
```

## 快速开始

```go
package main

import (
    "context"
    "time"

    dtoken "github.com/Zany2/dtoken-gf/v2/dtoken-gf"
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
                "userName": "demo",
            })
            if err != nil {
                r.Response.WriteJsonExit(g.Map{"code": 500, "message": err.Error()})
            }
            r.Response.WriteJsonExit(g.Map{
                "code":      0,
                "token":     tokenValue,
                "tokenType": "Bearer",
            })
        })

        group.GET("/profile", func(r *ghttp.Request) {
            userKey := r.GetCtxVar(dtoken.KeyUserKey).String()
            data := r.GetCtxVar(dtoken.KeyData).Map()
            r.Response.WriteJsonExit(g.Map{
                "code":    0,
                "userKey": userKey,
                "data":    data,
            })
        })
    })

    s.Run()
}
```

## 配置方式

DToken-GF 也可以从 GoFrame 全局配置节点 `dToken` 初始化。

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

从配置初始化：

```go
token, err := dtoken.NewFromConfig(ctx)
if err != nil {
    return err
}
defer token.Shutdown(ctx)
```

当 `cacheMode` 为 `2` 时，需要在创建 Token 实例前完成 GoFrame Redis 配置。若用于本地或轻量级场景，可将 `cacheMode` 设置为 `1` 使用内存缓存。

## 配置参数

| 参数 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `CacheMode` | `int8` | `1` | 存储模式：`1` = `gcache`，`2` = Redis，`3` = 文件缓存。 |
| `CachePreKey` | `string` | `DToken:` | 会话存储 key 前缀。 |
| `Timeout` | `int64` | `864000000` | Token/会话过期时间，单位毫秒。 |
| `MaxRefresh` | `int64` | `Timeout / 2` | 续期阈值。当剩余有效期小于等于该值时触发续期。 |
| `MaxRefreshTimes` | `int` | `0` | 最大续期次数，`0` 表示不限制。 |
| `EncryptKey` | `[]byte` | 内置密钥 | 默认 Token 编解码器使用的 AES 密钥，长度必须为 16、24 或 32 字节。 |
| `TokenDelimiter` | `string` | `_` | Token 内部分隔符，只能是 `_`、`-`、`.`、`:`、`\|`、`~` 之一。 |
| `MultiLogin` | `bool` | `false` | 设置为 `true` 时，同一个 `userKey` 重复登录会复用已有 Token。 |
| `AuthHeaderKey` | `string` | `Authorization` | 中间件读取 Token 的请求头名称。 |
| `AuthExcludePaths` | `[]string` | 空 | 免认证路径，支持精确路径和以 `/*` 结尾的前缀匹配。 |
| `PoolMinSize` | `int` | `20` | 异步续期协程池最小容量。 |
| `PoolMaxSize` | `int` | `2000` | 异步续期协程池最大容量。 |
| `PoolScaleUpRate` | `float64` | `0.8` | 协程池扩容阈值，需满足 `0 < down < up <= 1`。 |
| `PoolScaleDownRate` | `float64` | `0.3` | 协程池缩容阈值，需满足 `0 < down < up <= 1`。 |
| `PoolCheckInterval` | `int64` | `60000` | 协程池自动扩缩容检查间隔，单位毫秒。 |

生产环境应显式配置项目专用的 `EncryptKey`，不要直接使用内置默认密钥。

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

推荐通过 `New`、`NewToken`、`NewTokenWithConfig` 或 `NewFromConfig` 创建实例。

### Session

`Session` 保存单个 `userKey` 当前 Token 的会话状态。

```go
type Session struct {
    UserKey       string
    Token         string
    Data          g.Map
    RefreshNum    int
    CreateTime    int64
    LastRenewTime int64
}
```

### Store

`Store` 只负责保存、读取和删除编码后的会话字符串，不直接处理 `Session` 对象。

```go
type Store interface {
    Save(ctx context.Context, userKey string, data string) error
    Load(ctx context.Context, userKey string) (string, error)
    Delete(ctx context.Context, userKey string) error
}
```

如果项目已有存储层，可以接入自定义 Store：

```go
token, err := dtoken.New(
    dtoken.WithStore(myStore),
)
```

自定义 Store 需要自行处理 TTL 和过期策略。`CacheMode` 只影响内置 Store。

### TokenCodec

`TokenCodec` 负责 `userKey` 与 Token 字符串之间的编码和解码。

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

`SessionCodec` 负责 `Session` 与存储字符串之间的序列化和反序列化。

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

默认中间件会验证 Token，并将认证后的值写入 GoFrame Request CtxVar：

| Key | 内容 |
| --- | --- |
| `dtoken.KeyUserKey` | 当前认证用户标识。 |
| `dtoken.KeyData` | 调用 `Generate` 时写入的扩展数据。 |

读取示例：

```go
request := ghttp.RequestFromCtx(ctx)
userKey := request.GetCtxVar(dtoken.KeyUserKey).String()
data := request.GetCtxVar(dtoken.KeyData).Map()
```

这些值通过 `r.SetCtxVar` 写入，不是 `context.WithValue`，因此不要使用 `ctx.Value()` 读取。

可以自定义认证失败响应：

```go
middleware := dtoken.NewDefaultMiddleware(token, func(r *ghttp.Request) {
    r.Response.WriteJsonExit(g.Map{
        "code":    401,
        "message": "认证过期，请重新登录",
        "data":    []interface{}{},
    })
})
```

认证失败可能由以下原因导致：未传 Token、Bearer 格式错误、Token 无法解码、存储中的会话不存在，或请求 Token 与当前会话 Token 不一致。

## Token 提取规则

中间件按以下顺序读取 Token：

1. 请求头：`<AuthHeaderKey>: Bearer <token>`，默认是 `Authorization: Bearer <token>`。
2. 请求参数：`token`。

Bearer 请求头格式要求严格，必须是 `Bearer <token>`。

## 免认证路径

`AuthExcludePaths` 支持：

- 精确匹配：`/login`
- 前缀匹配：`/api/*`

前缀规则会去掉末尾的 `/*` 后按路径边界匹配。例如 `/api/*` 会匹配 `/api` 和 `/api/users`，不会匹配 `/apiary`。

## 自动续期

当认证成功的会话剩余有效期小于等于 `MaxRefresh` 时，DToken-GF 会提交异步续期任务。

- 续期不会阻塞当前请求。
- 续期写回前会重新读取当前会话并校验 Token 一致性，避免旧 Token 覆盖新 Token。
- `MaxRefreshTimes` 可限制最大续期次数。
- 文件缓存模式会在验证时额外检查过期时间，避免进程重启后恢复已过期会话。

应用退出时应调用 `Shutdown(ctx)`，以便优雅停止续期协程池。

## 缓存模式

- `CacheModeCache`：使用 GoFrame `gcache` 内存缓存。
- `CacheModeRedis`：使用 GoFrame Redis 缓存适配器，需要提前配置 GoFrame Redis。
- `CacheModeFile`：使用内存缓存，并将数据持久化到由缓存前缀和 `dtoken.dat` 组成的临时文件。

## 示例项目

仓库中包含 GoFrame 示例项目：`dtoken-gf-example`。

常用文件：

| 路径 | 说明 |
| --- | --- |
| `dtoken-gf-example/token/token.go` | 从 GoFrame 配置初始化 DToken-GF。 |
| `dtoken-gf-example/manifest/config/config.yaml` | 服务、Redis、日志和 `dToken` 示例配置。 |
| `dtoken-gf-example/internal/controller/auth` | 登录和退出登录示例。 |
| `dtoken-gf-example/internal/controller/user` | 受保护用户接口示例。 |

## 项目结构

| 文件 | 职责 |
| --- | --- |
| `factory.go` | Token 创建入口。 |
| `config.go` | 函数式配置和构造依赖。 |
| `options.go` | 配置结构、默认值和校验逻辑。 |
| `token.go` | Token 生成、验证、解析、会话读取和销毁。 |
| `store.go` | 编码后原始会话数据的存储接口。 |
| `session.go` | Session 数据结构。 |
| `codec_token.go` | Token 编解码接口和默认 AES + Base64 实现。 |
| `codec_session.go` | Session 编解码接口和默认 `gjson` 实现。 |
| `cache.go` | 内置 Store，实现内存、Redis 和文件缓存模式。 |
| `middleware.go` | GoFrame HTTP 认证中间件。 |
| `renewer.go` | 自动续期逻辑。 |
| `pool.go` | 续期协程池管理。 |
| `banner.go` | 启动横幅输出。 |
| `constants.go` | 常量和错误消息。 |

## 安全建议

- 使用强度足够的项目专用 AES 密钥，长度必须为 16、24 或 32 字节。
- 生产环境使用 HTTPS，避免 Token 在传输中泄露。
- 优先使用 `Authorization: Bearer <token>` 请求头，不建议通过查询参数传递 Token。
- 多实例部署时，建议使用 Redis 或自定义共享存储。
- 退出登录时调用 `Destroy` 或 `DestroyByToken` 删除服务端会话。

## License

DToken-GF 使用 Apache License 2.0 开源协议，详情请查看 [LICENSE](LICENSE)。
