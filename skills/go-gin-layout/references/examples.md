# 项目代码示例

## 可直接运行的双模块 API

完整代码位于 [assets/gin-api](../assets/gin-api/)，使用 Go 1.25 或更高版本、Gin v1.12.0、GORM v1.31.1 和 MySQL。示例展示用户查询与订单创建。Handler 和 Service 按业务拆包。用户与订单关系紧密，Entity 和 Repository 各用一个包。

```text
gin-api/
├── cmd/api/main.go
├── internal/
│   ├── app/api.go
│   ├── config/config.go
│   ├── infra/
│   │   └── database/mysql.go
│   ├── router/
│   │   ├── api.go
│   │   ├── user.go
│   │   ├── order.go
│   │   └── health.go
│   ├── svc/servicecontext.go
│   ├── middleware/
│   │   ├── requestid.go
│   │   ├── accesslog.go
│   │   └── recovery.go
│   ├── handler/
│   │   ├── user/profile.go
│   │   └── order/create.go
│   ├── service/
│   │   ├── user/profile.go
│   │   └── order/create.go
│   ├── repository/
│   │   ├── user.go
│   │   └── order.go
│   └── entity/
│       ├── user.go
│       ├── order.go
│       └── order_item.go
├── migrations/
│   ├── 001_create_user_order_tables.sql
│   └── 002_seed_demo_user.sql
├── configs/
│   └── config.example.yaml
├── docker-compose.yml
├── README.md
├── go.mod
└── go.sum
```

示例通过 `ServiceContext` 共享结构化日志器以及用户和订单的仓储接口。仓储使用 GORM 访问 MySQL，表映射、查询和事务留在仓储层；关系型仓储的按表命名与 GORM 映射见 [Repository 仓储规范](../SKILL.md#repository-仓储规范)。

`migrations/001_create_user_order_tables.sql` 创建 `t_user`、`t_order`、`t_order_item`；`002_seed_demo_user.sql` 仅为本地示例预置用户 `1 / Alice`。用户由请求中的 `user_id` 指定，未实现身份认证；示例聚焦目录和依赖，不包含库存、支付等业务能力。迁移命名与执行规则见 [数据库迁移规范](migrations.md)。

### 启动与请求

在资产目录启动 MySQL（首次启动会执行 `migrations/` 下的 SQL）：

```bash
cd skills/go-gin-layout/assets/gin-api
docker compose up -d --wait
```

复制配置后启动 API：

```bash
cp configs/config.example.yaml configs/config.yaml
go run ./cmd/api
```

默认读取 `configs/config.yaml`，可用 `CONFIG_FILE` 指定其他路径。默认监听 `127.0.0.1:8080`。启动时会 Ping 数据库，连接失败则进程退出；按 Ctrl+C 停止时会等待在途 HTTP 请求结束，最多等待 15 秒，然后关闭数据库连接。

在另一个终端执行：

```bash
curl -i http://127.0.0.1:8080/api/v1/users/1

curl -i -X POST http://127.0.0.1:8080/api/v1/orders \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1,"items":[{"sku":"BOOK-001","quantity":2}]}'
```

用户查询返回 HTTP 200 和 `{"id":1,"name":"Alice"}`；首次创建订单返回 HTTP 201：

```json
{"id":1,"user_id":1,"items":[{"sku":"BOOK-001","quantity":2}]}
```

路径 ID 不合法、空订单或非正数量返回 HTTP 400；用户不存在返回 HTTP 404；未预期的内部错误返回 HTTP 500，响应不包含底层错误细节。订单与明细在同一事务中写入。

### 代码阅读顺序

| 文件 | 关注点 |
| --- | --- |
| [cmd/api/main.go](../assets/gin-api/cmd/api/main.go) | 创建共享日志器、加载配置并接收退出信号 |
| [internal/config/config.go](../assets/gin-api/internal/config/config.go) | YAML 加载、字段校验与日志级别解析 |
| [internal/infra/database/mysql.go](../assets/gin-api/internal/infra/database/mysql.go) | GORM 连接池、DSN 构造、Ping 与健康检查 |
| [internal/app/api.go](../assets/gin-api/internal/app/api.go) | 打开数据库、创建共享依赖、监听、优雅关闭与释放连接 |
| [internal/svc/servicecontext.go](../assets/gin-api/internal/svc/servicecontext.go) | 保存共享日志器、运行配置、仓储接口和健康检查能力 |
| [internal/router/api.go](../assets/gin-api/internal/router/api.go) | 引擎、中间件、版本分组与模块路由注册 |
| [internal/router/user.go](../assets/gin-api/internal/router/user.go) | 用户模块路由 |
| [internal/router/order.go](../assets/gin-api/internal/router/order.go) | 订单模块路由 |
| [internal/router/health.go](../assets/gin-api/internal/router/health.go) | liveness/readiness/startup 探针 |
| [internal/middleware/requestid.go](../assets/gin-api/internal/middleware/requestid.go) | `X-Request-ID` 的校验、生成、透传与响应头写入 |
| [internal/middleware/accesslog.go](../assets/gin-api/internal/middleware/accesslog.go) | 访问日志、跳过健康检查、注入 request_id |
| [internal/middleware/recovery.go](../assets/gin-api/internal/middleware/recovery.go) | panic 恢复与结构化 500 |
| [internal/handler/user/profile.go](../assets/gin-api/internal/handler/user/profile.go) | 路径参数、HTTP 响应结构与错误映射 |
| [internal/handler/order/create.go](../assets/gin-api/internal/handler/order/create.go) | JSON 请求结构、绑定校验与业务参数转换 |
| [internal/service/user/profile.go](../assets/gin-api/internal/service/user/profile.go) | 用户查询用例，不依赖 Gin 和 GORM |
| [internal/service/order/create.go](../assets/gin-api/internal/service/order/create.go) | 校验订单、通过仓储接口确认用户存在并保存订单 |
| [internal/repository/user.go](../assets/gin-api/internal/repository/user.go) | GORM 行映射、`ErrRecordNotFound` 到业务错误的转换 |
| [internal/repository/order.go](../assets/gin-api/internal/repository/order.go) | 订单与明细的事务写入、表映射与实体转换 |
| [migrations/001_create_user_order_tables.sql](../assets/gin-api/migrations/001_create_user_order_tables.sql) | 表结构与索引 |
| [docker-compose.yml](../assets/gin-api/docker-compose.yml) | 本地 MySQL 与初始化脚本挂载 |

应用入口创建 `*slog.Logger`，`app` 打开数据库并 Ping，调用 `svc.NewServiceContext(logger, cfg, db)`，经 Router 将同一个依赖指针传到两个 Handler 的闭包。`ServiceContext` 同时保存日志器、运行配置、仓储接口和数据库健康检查能力；Handler 通过 `svcCtx.Logger.ErrorContext` 记录意外错误，Service 可按需读取 `svcCtx.Config` 并调用 `svcCtx.Repo`，`/readyz` 通过 `svcCtx.Health` 检查数据库。每次请求由 Handler 把 `c.Request.Context()` 和共享依赖传给对应的业务构造函数（`NewProfileService` 或 `NewCreateService`）。Service 使用请求上下文调用仓储，仓储通过 `db.WithContext(ctx)` 执行查询；进程退出时 `app` 关闭数据库连接。

仓储实现的方法注释使用 `// GetByID implements [UserRepository.GetByID].` 形式定位接口方法，业务语义和错误约定写在接口声明处，具体规范见 [接口与实现的注释](../SKILL.md#接口与实现的注释)。

错误响应由 Handler 直接映射：每个响应带 `X-Request-ID`；访问日志由 middleware 输出，含 request_id 并跳过健康检查；4xx 不写结构化错误日志，但访问日志仍记录；5xx 以 `Error` 记录 `request_id`、method、path、status 和 error；panic 由 Recovery 中间件捕获并返回统一 JSON 500。完整规则见 [日志打印规范](logging.md)。

配置以只读值放入 `ServiceContext.Config`，由 Service 按需读取；模块配置增多时按业务拆分子结构，而不是汇总成 `BusinessConfig` 或为配置字段定义接口。行为依赖（仓储、缓存、消息发布器）才通过接口注入，具体取舍见 [配置与依赖注入](../SKILL.md#配置与依赖注入)。

## 可选基础设施

MySQL 已包含在示例中；Redis、消息队列和登录限流等能力仍按实际需求引入，接入位置见 [基础设施接入代码](infrastructure.md)。

复用时按目标项目调整 module 路径和 import，仅取实际需要的模块。Entity 与 Repository 按相同业务边界划分。接入关系型数据库时，按仓储规范组织表映射和存储实现。具体功能文件与可选基础设施按实际需求选择。
