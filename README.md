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

```go
package main

import (
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/net/ghttp"
    dtoken "github.com/Zany2/dtoken-gf/dtoken-gf"
)

func main() {
    s := g.Server()

    // 创建 Token 实例
    token := dtoken.NewDefaultToken(dtoken.Options{
        CacheMode:   dtoken.CacheModeCache, // 1-内存 2-Redis 3-文件
        CachePreKey: "MyApp:",
        Timeout:     10 * 24 * 60 * 60 * 1000, // 10天（毫秒）
        MaxRefresh:  5 * 24 * 60 * 60 * 1000,  // 5天内自动续期
        EncryptKey:  []byte("12345678912345678912345678912345"), // 必须 16/24/32 字节
        MultiLogin:  true,
        AuthExcludePaths: g.SliceStr{"/login", "/register", "/public/*"},
    })

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
            data := r.GetCtxVar(dtoken.KeyUserKey)
            r.Response.WriteJson(g.Map{"code": 0, "data": data})
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
token := dtoken.NewDefaultTokenByConfig()
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
| AuthExcludePaths | []string | 空 | 免认证路径，支持 `/*` 通配 |
| PoolMinSize | int | 100 | 续期协程池最小容量 |
| PoolMaxSize | int | 2000 | 续期协程池最大容量 |
| PoolScaleUpRate | float64 | 0.8 | 协程池扩容阈值（使用率） |
| PoolScaleDownRate | float64 | 0.3 | 协程池缩容阈值（使用率） |

## 核心接口

### Token 接口

```go
type Token interface {
    Generate(ctx context.Context, userKey string, data any) (token string, err error)
    Validate(ctx context.Context, token string) (data any, err error)
    Get(ctx context.Context, userKey string) (token string, data any, err error)
    ParseToken(ctx context.Context, token string) (userKey string, data any, err error)
    Destroy(ctx context.Context, userKey string) error
    Renew(ctx context.Context, userKey string, userCache g.Map)
    Shutdown(ctx context.Context)
    GetOptions() Options
}
```

### Cache 接口

支持自定义缓存实现：

```go
type Cache interface {
    Set(ctx context.Context, cacheKey string, cacheValue g.Map) error
    Get(ctx context.Context, cacheKey string) (g.Map, error)
    Remove(ctx context.Context, cacheKey string) error
}
```

### Codec 接口

支持自定义编解码实现：

```go
type Codec interface {
    Encode(ctx context.Context, userKey string) (token string, err error)
    Decrypt(ctx context.Context, token string) (userKey string, err error)
}
```

## Token 提取方式

中间件按以下优先级从请求中提取 Token：

1. `Authorization: Bearer <token>` 请求头
2. `token` 请求参数（Query / Form）

## 自动续期机制

当 Token 剩余有效期低于 `MaxRefresh` 时，验证请求会自动触发异步续期：

- 续期通过 ants 协程池异步执行，不阻塞请求
- 协程池支持动态扩缩容（基于使用率自动调整容量）
- 可通过 `MaxRefreshTimes` 限制最大续期次数
- 可通过 `RenewInterval` 设置最小续期间隔，避免高并发下频繁续期
- 续期前会二次校验缓存一致性，防止并发冲突

## 路径排除规则

`AuthExcludePaths` 支持两种匹配模式：

- 精确匹配：`/login` — 仅匹配该路径
- 前缀通配：`/api/*` — 匹配所有以 `/api/` 开头的路径

## 依赖

- [GoFrame v2](https://github.com/gogf/gf) — Web 框架及工具库
- [ants v2](https://github.com/panjf2000/ants) — 高性能协程池

## License

MIT
