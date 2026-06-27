# DToken-GF

English | [简体中文](README_zh.md)

DToken-GF is a lightweight token authentication component for [GoFrame v2](https://goframe.org/). It provides token generation, token validation middleware, session storage, automatic renewal, and pluggable codecs for GoFrame HTTP services.

## Features

- GoFrame middleware for protecting routes and writing authenticated user data into request context variables.
- Token generation and validation with the default AES + Base64 token codec.
- Session persistence through GoFrame `gcache`, GoFrame Redis adapter, or a custom store.
- Session serialization through the default `gjson` codec or a custom session codec.
- Automatic asynchronous renewal when the remaining token lifetime reaches the configured threshold.
- Configurable authentication header, excluded paths, token delimiter, timeout, renewal limits, and renewal worker pool.

## Requirements

- Go 1.23+
- GoFrame v2.9+

## Installation

```bash
go get github.com/Zany2/dtoken-gf/v2
```

Import the package with an alias because the implementation lives in the `dtoken-gf` subdirectory:

```go
import dtoken "github.com/Zany2/dtoken-gf/v2/dtoken-gf"
```

## Quick Start

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

## Configuration

DToken-GF can also be initialized from the GoFrame global configuration node `dToken`.

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

Initialize from configuration:

```go
token, err := dtoken.NewFromConfig(ctx)
if err != nil {
    return err
}
defer token.Shutdown(ctx)
```

When `cacheMode` is `2`, configure GoFrame Redis before creating the token instance. For local or lightweight use, set `cacheMode` to `1` for in-memory cache.

## Options

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `CacheMode` | `int8` | `1` | Storage mode: `1` = `gcache`, `2` = Redis. |
| `CachePreKey` | `string` | `DToken:` | Prefix for session storage keys. |
| `Timeout` | `int64` | `864000000` | Token/session timeout in milliseconds. |
| `MaxRefresh` | `int64` | `Timeout / 2` | Renewal threshold. Renewal is triggered when the remaining lifetime is less than or equal to this value. |
| `MaxRefreshTimes` | `int` | `0` | Maximum number of renewals. `0` means unlimited. |
| `EncryptKey` | `[]byte` | built-in key | AES key used by the default token codec. Length must be 16, 24, or 32 bytes. |
| `TokenDelimiter` | `string` | `_` | Internal token delimiter. Must be one of `_`, `-`, `.`, `:`, `\|`, `~`. |
| `MultiLogin` | `bool` | `false` | Reuse an existing token for the same `userKey` when set to `true`. |
| `AuthHeaderKey` | `string` | `Authorization` | Header name used by the middleware to read the token. |
| `AuthExcludePaths` | `[]string` | empty | Paths that skip authentication. Supports exact paths and prefix patterns ending with `/*`. |
| `PoolMinSize` | `int` | `20` | Minimum size of the asynchronous renewal worker pool. |
| `PoolMaxSize` | `int` | `2000` | Maximum size of the asynchronous renewal worker pool. |
| `PoolScaleUpRate` | `float64` | `0.8` | Pool scale-up threshold. Must satisfy `0 < down < up <= 1`. |
| `PoolScaleDownRate` | `float64` | `0.3` | Pool scale-down threshold. Must satisfy `0 < down < up <= 1`. |
| `PoolCheckInterval` | `int64` | `60000` | Pool auto-scaling check interval in milliseconds. |

For production, always configure a project-specific `EncryptKey`. Do not rely on the built-in default key.

## Core API

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

Create instances through `New`, `NewToken`, `NewTokenWithConfig`, or `NewFromConfig`.

### Session

`Session` stores the current token state for one `userKey`.

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

`Store` handles encoded session strings. It does not directly manage `Session` objects.

```go
type Store interface {
    Save(ctx context.Context, userKey string, data string) error
    Load(ctx context.Context, userKey string) (string, error)
    Delete(ctx context.Context, userKey string) error
}
```

Use a custom store when your project already has a storage layer:

```go
token, err := dtoken.New(
    dtoken.WithStore(myStore),
)
```

Custom stores are responsible for their own TTL and expiration policy. `CacheMode` only affects the built-in store.

### TokenCodec

`TokenCodec` controls how `userKey` values are encoded into token strings and decoded back.

```go
type TokenCodec interface {
    Encode(ctx context.Context, userKey string) (token string, err error)
    Decode(ctx context.Context, token string) (userKey string, err error)
}
```

Default implementation:

```go
dtoken.NewDefaultTokenCodec(delimiter, encryptKey)
```

### SessionCodec

`SessionCodec` controls how `Session` values are serialized into storage strings and decoded back.

```go
type SessionCodec interface {
    Encode(ctx context.Context, session *Session) (string, error)
    Decode(ctx context.Context, data string) (*Session, error)
}
```

Default implementation:

```go
dtoken.NewDefaultSessionCodec()
```

## Middleware

The default middleware validates the token and writes authenticated values into the GoFrame request context variables:

| Key | Value |
| --- | --- |
| `dtoken.KeyUserKey` | Current authenticated user key. |
| `dtoken.KeyData` | Extra data passed to `Generate`. |

Read the values with `GetCtxVar`:

```go
request := ghttp.RequestFromCtx(ctx)
userKey := request.GetCtxVar(dtoken.KeyUserKey).String()
data := request.GetCtxVar(dtoken.KeyData).Map()
```

These values are set through `r.SetCtxVar`, not `context.WithValue`, so do not read them with `ctx.Value()`.

You can customize the authentication failure response:

```go
middleware := dtoken.NewDefaultMiddleware(token, func(r *ghttp.Request) {
    r.Response.WriteJsonExit(g.Map{
        "code":    401,
        "message": "authentication expired, please log in again",
        "data":    []interface{}{},
    })
})
```

Authentication can fail when the token is missing, the Bearer format is invalid, the token cannot be decoded, the stored session does not exist, or the token does not match the current session.

## Token Extraction

The middleware reads tokens in this order:

1. Header: `<AuthHeaderKey>: Bearer <token>`, defaulting to `Authorization: Bearer <token>`.
2. Request parameter: `token`.

The Bearer header format is strict and must be `Bearer <token>`.

## Excluded Paths

`AuthExcludePaths` supports:

- Exact match: `/login`
- Prefix match: `/api/*`

Prefix rules remove the trailing `/*` and then match by path boundary. For example, `/api/*` matches `/api` and `/api/users`, but does not match `/apiary`.

## Automatic Renewal

When a validated session has a remaining lifetime less than or equal to `MaxRefresh`, DToken-GF submits an asynchronous renewal task.

- Renewal does not block the current request.
- Before renewal writes the session back, it reloads the current session and checks token consistency to avoid replacing a newer token with an old one.
- `MaxRefreshTimes` can limit how many times a session may be renewed.

Call `Shutdown(ctx)` when the application exits to stop the renewal pool gracefully.

## Cache Modes

- `CacheModeCache`: uses GoFrame `gcache` in-memory cache.
- `CacheModeRedis`: uses GoFrame Redis cache adapter and requires GoFrame Redis configuration.

## Example Project

The repository includes a GoFrame example application under `dtoken-gf-example`.

Useful files:

| Path | Purpose |
| --- | --- |
| `dtoken-gf-example/token/token.go` | Initializes DToken-GF from GoFrame configuration. |
| `dtoken-gf-example/manifest/config/config.yaml` | Example server, Redis, logger, and `dToken` configuration. |
| `dtoken-gf-example/internal/controller/auth` | Login and logout examples. |
| `dtoken-gf-example/internal/controller/user` | Protected user endpoint examples. |

## Project Structure

| File | Responsibility |
| --- | --- |
| `factory.go` | Token construction entry points. |
| `config.go` | Functional options and construction dependencies. |
| `options.go` | Options, defaults, and validation. |
| `token.go` | Token generation, validation, parsing, session lookup, and destruction. |
| `store.go` | Raw encoded session storage interface. |
| `session.go` | Session data model. |
| `codec_token.go` | Token codec interfaces and default AES + Base64 implementation. |
| `codec_session.go` | Session codec interfaces and default `gjson` implementation. |
| `cache.go` | Built-in store implementation for cache and Redis modes. |
| `middleware.go` | GoFrame HTTP authentication middleware. |
| `renewer.go` | Automatic renewal logic. |
| `pool.go` | Renewal worker pool management. |
| `banner.go` | Startup banner output. |
| `constants.go` | Constants and error messages. |

## Security Notes

- Use a strong project-specific AES key with a valid length of 16, 24, or 32 bytes.
- Use HTTPS in production so tokens are not exposed in transit.
- Prefer the `Authorization: Bearer <token>` header over query parameters.
- Choose Redis or a custom shared store when running multiple application instances.
- Make logout call `Destroy` or `DestroyByToken` to remove the stored session.

## License

DToken-GF is licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.
