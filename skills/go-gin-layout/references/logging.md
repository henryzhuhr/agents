# 日志打印规范

本规范约定 Go/Gin 服务的日志级别、请求关联、字段和错误记录方式。目标是让线上问题能从响应快速定位到日志，同时避免把预期错误记成系统故障。

## 基本原则

- 使用结构化日志（`log/slog`），使用稳定的 key，不拼接日志正文。
- 一次错误只在能决定 HTTP 响应的边界记录一次，避免 Handler、Service、Repository 重复打印。
- 只记录定位所需信息，不记录请求体、密码、令牌、Cookie、完整 SQL 参数等敏感内容。
- 对外响应只返回通用文案，内部错误只写日志。
- 预期内的 4xx 不按故障处理，5xx 和未知错误必须可告警。

## 日志级别

| 级别 | 使用场景 | 示例 |
| --- | --- | --- |
| `Debug` | 排查 4xx、开发期细节；默认不输出 | 绑定校验失败、非法路径参数 |
| `Info` | 生命周期和正常业务里程碑 | 启动监听、优雅关闭 |
| `Warn` | 可恢复的异常、降级、重试 | 依赖短暂失败但已降级 |
| `Error` | 5xx、启动失败、资源关闭失败等需要人工关注的问题 | 数据库查询失败、端口监听失败 |

配置文件中的 `log_level` 控制输出级别，默认 `info`；`debug` 需要显式开启。

## 请求关联

- 每个请求必须有 `X-Request-ID`，由 `internal/middleware/requestid.go` 处理。
- 合法入站 ID（`[A-Za-z0-9._-]{1,64}`）透传；缺失或非法时生成新的十六进制 ID。
- 该 ID 写入响应头 `X-Request-ID`，并放入 `gin.Context`；错误日志通过 `middleware.RequestIDFromGin(c)` 取出。
- 用户报告问题时，以响应头的 ID 搜索日志。

## 何时记录

| HTTP 状态 | 是否记录 | 级别 | 位置 |
| --- | --- | --- | --- |
| 4xx | 不写结构化 `Error`；访问日志仍记录 | `Debug`（可选） | Handler |
| 5xx | 必须 | `Error` | Handler |
| 2xx | 不记录业务日志 | — | 访问日志 |

- Handler 负责把业务错误映射为 HTTP 状态，并在 5xx 时记录内部错误。
- Service 和 Repository 返回错误，不打印日志；错误向上传递，由 Handler 决定是否记录。
- 不要对 400/404 使用 `Error`，否则告警会被用户输入错误淹没。
- 启动、监听、关闭等生命周期事件使用 `Info`；关闭数据库失败等使用 `Error`。

## 字段约定

| key | 类型 | 说明 |
| --- | --- | --- |
| `request_id` | string | 请求关联 ID |
| `method` | string | HTTP 方法 |
| `path` | string | `c.FullPath()`，不要使用含用户输入的原始 URL |
| `status` | int | HTTP 状态码 |
| `error` | error | 底层错误，仅 5xx 或 Debug 4xx 记录 |
| `<business_id>` | number/string | 如 `user_id`、`order_id`，按需记录 |

- message 使用稳定、简短的事件名，如 `create order`、`load user profile`，便于检索和聚合。
- 不要用 `fmt.Sprintf` 拼接日志正文；把变量放进 key/value。
- 不要记录 `c.Request.Body`、`Authorization`、Cookie、密码、完整 SQL 参数。
- 记录错误时使用原始 `error`，不要只写 `err.Error()` 而丢失类型信息。

## Handler 示例

### 5xx（必须记录）

```go
default:
    svcCtx.Logger.ErrorContext(ctx, "create order",
        "request_id", middleware.RequestIDFromGin(c),
        "method", c.Request.Method,
        "path", c.FullPath(),
        "status", http.StatusInternalServerError,
        "error", err,
    )
    c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
```

### 4xx（不写结构化 `Error`）

```go
case errors.Is(err, orderservice.ErrInvalidOrder):
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order"})
```

确需排查时，可在 `c.JSON` 前补一条 `Debug`：

```go
svcCtx.Logger.DebugContext(ctx, "request rejected",
    "request_id", middleware.RequestIDFromGin(c),
    "method", c.Request.Method,
    "path", c.FullPath(),
    "status", http.StatusBadRequest,
    "error", err,
)
```

## 访问日志与数据库日志

- 访问日志由 `internal/middleware/accesslog.go` 使用 `slog` 输出，包含 method、path、status、latency 和 `request_id`；它跳过 `/healthz`、`/livez`、`/readyz`、`/startupz`，也不记录 query string。
- 4xx 不写结构化 `Error` 日志，但请求仍会出现在访问日志中；需要排查校验细节时，在 Handler 中临时增加 `Debug`。
- panic 由 `internal/middleware/recovery.go` 捕获，使用 `Error` 记录 `request_id`、method、path、status、error 和堆栈，并返回统一的 JSON 500。
- GORM 日志用于慢查询和数据库错误，不替代业务日志。`ErrRecordNotFound` 属于预期查询结果，已配置为不输出。
- 不要在中间件、Service、Repository 中重复打印同一个错误。

## 禁止事项

- 使用 `fmt.Println` 或 `log.Printf` 输出业务日志。
- 把 `Error` 用在 400/404/参数校验等预期错误。
- 打印完整请求体、响应体、凭据或令牌。
- 同一错误在多层重复记录，或记录后吞掉错误。
- 只记录 `error` 而没有 `request_id`、method、path 等定位字段。
- 对健康检查等高频端点记录逐请求 `Info` 日志。

## 验证清单

- 400 和 404 不产生 `Error` 日志，但会出现在访问日志。
- 500 日志包含 `request_id`、method、path、status 和底层 `error`；panic 500 由 Recovery 中间件记录。
- 响应头带 `X-Request-ID`，合法入站值被原样透传。
- 健康检查端点不出现在访问日志中。
- 日志中没有请求体、令牌等敏感信息。
- `log_level: debug` 仅在需要时开启，开启后可看到额外 4xx 排查日志（如果代码中已添加）。
- 构建与静态检查通过：`gofmt -l .`、`go build ./...`、`go vet ./...`。
