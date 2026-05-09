# DToken-GF

基于 [GoFrame](https://goframe.org) 框架的轻量级 Token 认证插件，支持 AES 加密、多缓存模式、自动续期和动态协程池。

## 特性

- Token 生成与校验：基于 AES 加密 + Base64 编码，安全可靠
- 多缓存模式：支持内存缓存（gcache）、Redis（gredis）、文件缓存（gfile）
- 自动续期：请求验证时自动异步续期，基于 [ants](https://github.com/panjf2000/ants) 动态协程池
- 中间件集成：开箱即用的 GoFrame HTTP 中间件，支持路径排除（精确匹配 / 前缀通配）
- 多端登录控制：可配置是否允许同一用户多端复用 Token
- 灵活配置：支持代码配置和 GoFrame 配置文件两种方式

## 安装

```bash
go get github.com/Zany2/dtoken-gf
```

要求：Go 1.23+，GoFrame v2.9+

## 快速开始

### 1. 代码方式初始化

推荐使用 v2 函数式配置，时间参数直接使用 `time.Duration`：

```go
package main

import (
    "time"

    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/net/ghttp"
    dtoken "github.com/Zany2/dtoken-gf/dtoken-gf"
)

func main() {
    s := g.Server()

    // 创建 Token 实例
    token, err := dtoken.New(
        dtoken.WithCacheMode(dtoken.CacheModeCache),
        dtoken.WithCachePreKey("MyApp:"),
        dtoken.WithTimeout(10*24*time.Hour),
        dtoken.WithMaxRefresh(5*24*time.Hour),
        dtoken.WithEncryptKey([]byte("12345678912345678912345678912345")), // 必须 16/24/32 字节
        dtoken.WithReuseToken(true),
        dtoken.WithAuthExcludePaths("/login", "/register", "/public/*"),
    )
    if err != nil {
        panic(err)
    }

    // 创建中间件
    middleware := dtoken.NewDefaultMiddleware(token)

    // 注册路由
    s.Group("/", func(group *ghttp.RouterGroup) {
        group.Middleware(middleware.Auth)

        group.POST("/login", func(r *ghttp.Request) {
            // 生成 Token
            t, err := token.Generate(r.Context(), "user_001", g.Map{"role": "admin"})
            if err != nil {
                r.Response.WriteJsonExit(g.Map{"code": 1, "msg": err.Error()})
            }
            r.Response.WriteJson(g.Map{"code": 0, "token": t})
        })

        group.POST("/logout", func(r *ghttp.Request) {
            // 销毁 Token
            _ = token.Destroy(r.Context(), "user_001")
            r.Response.WriteJson(g.Map{"code": 0, "msg": "ok"})
        })

        group.GET("/profile", func(r *ghttp.Request) {
            // 从上下文获取用户数据
            userKey := r.GetCtxVar(dtoken.KeyUserKey)
            data := r.GetCtxVar(dtoken.KeyData)
            r.Response.WriteJson(g.Map{"code": 0, "userKey": userKey, "data": data})
        })
    })

    // 优雅关闭
    defer token.Shutdown(nil)

    s.Run()
}
```

### 2. 配置文件方式初始化

在 GoFrame 配置文件（如 `config.yaml`）中添加：

```yaml
dToken:
  cacheMode: 2                # 1-gcache 2-gredis 3-gfile
  cachePreKey: "MyApp:"
  timeout: 864000000          # 10天（毫秒）
  maxRefresh: 432000000       # 5天（毫秒）
  maxRefreshTimes: 0          # 最大续期次数，0=不限
  tokenDelimiter: "_"
  encryptKey: "12345678912345678912345678912345"
  multiLogin: true
  authHeaderKey: "Authorization"
  authExcludePaths:
    - "/login"
    - "/public/*"
  poolMinSize: 100
  poolMaxSize: 2000
  poolScaleUpRate: 0.8
  poolScaleDownRate: 0.3
  renewInterval: 60000        # 最小续期间隔（毫秒）
```

```go
token, err := dtoken.NewFromConfig()
if err != nil {
    panic(err)
}
```

### 3. 自定义校验失败响应

```go
middleware := dtoken.NewDefaultMiddleware(token, func(r *ghttp.Request) {
    r.Response.WriteJsonExit(g.Map{
        "code": 401,
        "msg":  "未授权，请重新登录",
    })
})
```

## 配置参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| CacheMode | int8 | 1 | 缓存模式：1=gcache 2=gredis 3=gfile |
| CachePreKey | string | `DToken:` | 缓存 key 前缀 |
| Timeout | int64 | 864000000 (10天) | Token 过期时间（毫秒） |
| MaxRefresh | int64 | Timeout/2 | 剩余时间低于此值时触发续期（毫秒） |
| MaxRefreshTimes | int | 0 | 最大续期次数，0=不限 |
| RenewInterval | int64 | 0 | 最小续期间隔（毫秒），防止频繁续期 |
| EncryptKey | []byte | 内置默认值 | AES 加密密钥，长度必须为 16/24/32 字节 |
| TokenDelimiter | string | `_` | Token 内部分隔符 |
| MultiLogin | bool | false | 是否允许多端登录复用 Token |
| AuthHeaderKey | string | `Authorization` | 认证请求头名称 |
| AuthExcludePaths | []string | 空 | 免认证路径，支持 `/*` 通配 |
| PoolMinSize | int | 100 | 续期协程池最小容量 |
| PoolMaxSize | int | 2000 | 续期协程池最大容量 |
| PoolScaleUpRate | float64 | 0.8 | 协程池扩容阈值（使用率） |
| PoolScaleDownRate | float64 | 0.3 | 协程池缩容阈值（使用率） |

## v2 代码结构

核心代码按职责拆分：

| 文件 | 职责 |
|------|------|
| `factory.go` | Token 创建入口 |
| `config.go` | 函数式配置选项 |
| `options.go` | 基础配置结构、默认值和校验 |
| `token.go` | Token 生成、校验、解析、销毁主流程 |
| `store.go` | 原始会话数据存储接口 |
| `session.go` | Token 会话数据结构 |
| `session_codec.go` | Session 编解码实现 |
| `renewer.go` | 自动续期判断和异步续期 |
| `cache.go` | 默认缓存实现 |
| `codec.go` | Token 编解码实现 |
| `middleware.go` | GoFrame HTTP 中间件 |
| `pool.go` | 续期协程池 |
| `banner.go` | 启动横幅和配置输出 |
| `constants.go` | 常量和错误消息 |

## 核心接口

### Token 接口

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

认证成功后返回完整 `Session`：

```go
session, err := token.ValidateSession(ctx, tokenValue)
if err != nil {
    return err
}

userKey := session.UserKey
data := session.Data
```

`Session.Data` 统一使用 `g.Map`，建议把扩展信息以键值对形式写入：

```go
tokenValue, err := token.Generate(ctx, "user_001", g.Map{
    "role": "admin",
    "name": "Tom",
})
```

中间件认证成功后会写入三个上下文变量：

| Key | 内容 |
|-----|------|
| `dtoken.KeyUserKey` | 当前用户标识 |
| `dtoken.KeyData` | 登录时写入的扩展数据 |
| `dtoken.KeySession` | 完整 Session |

### Store 接口

`Store` 只负责保存和读取编码后的会话字符串：

```go
type Store interface {
    Save(ctx context.Context, userKey string, data string) error
    Load(ctx context.Context, userKey string) (string, error)
    Delete(ctx context.Context, userKey string) error
}
```

通过函数式选项注入：

```go
token, err := dtoken.New(
    dtoken.WithStore(myStore),
)
```

### TokenCodec 接口

`TokenCodec` 负责把 `userKey` 编进 Token，并从 Token 中解析 `userKey`：

```go
type TokenCodec interface {
    Encode(ctx context.Context, userKey string) (token string, err error)
    Decode(ctx context.Context, token string) (userKey string, err error)
}
```

### SessionCodec 接口

`SessionCodec` 负责 `Session` 和存储字符串之间的编解码，默认实现使用 `gjson`：

```go
type SessionCodec interface {
    Encode(ctx context.Context, session *Session) (string, error)
    Decode(ctx context.Context, data string) (*Session, error)
}
```

## Token 提取方式

中间件按以下优先级从请求中提取 Token：

1. `<AuthHeaderKey>: Bearer <token>` 请求头，默认是 `Authorization`
2. `token` 请求参数（Query / Form）

## 自动续期机制

当 Token 剩余有效期低于 `MaxRefresh` 时，验证请求会自动触发异步续期：

- 续期通过 ants 协程池异步执行，不阻塞请求
- 协程池支持动态扩缩容（基于使用率自动调整容量）
- 可通过 `MaxRefreshTimes` 限制最大续期次数
- 可通过 `RenewInterval` 设置最小续期间隔，避免高并发下频繁续期
- 续期前会二次校验缓存一致性，防止并发冲突

## 多端登录复用

`MultiLogin`/`WithReuseToken(true)` 表示同一个 `userKey` 重复登录时复用已有 Token。复用时不会自动覆盖已有 `Session.Data`，如果用户角色或扩展信息变化，建议先 `Destroy` 后重新 `Generate`。

## 路径排除规则

`AuthExcludePaths` 支持两种匹配模式：

- 精确匹配：`/login` — 仅匹配该路径
- 前缀通配：`/api/*` — 匹配所有以 `/api/` 开头的路径

## 依赖

- [GoFrame v2](https://github.com/gogf/gf) — Web 框架及工具库
- [ants v2](https://github.com/panjf2000/ants) — 高性能协程池

## License

MIT
