---
name: go-gin-layout
description: 规划、创建和审查 Go/Gin 后端项目目录结构，明确路由、Handler、Service、Repository 的职责、依赖方向与 pytest 端到端接口测试的组织方式。适用于初始化 Gin API 项目、为现有项目新增模块、整理分层目录或渐进重构；普通业务修改无需仅为套用本规范而调整目录。
---

# Go/Gin 后端项目目录规范

为 Go/Gin API 项目提供默认目录布局和代码归属规则。这是本技能的项目约定；Go 和 Gin 并未要求所有项目采用同一套分层目录。当前版本聚焦业务代码与基础设施的目录组织，并约定基于 pytest 的端到端 HTTP 接口测试目录与组织方式；单元测试布局暂不展开。

## 设计依据与适用边界

本模板综合 Go 项目组织习惯、分层架构和 Gin 的 HTTP 接口职责整理，不是某个仓库的完整复刻，也不是 Gin 官方规定的标准结构。

| 目录或设计 | 依据与约束 |
| --- | --- |
| `cmd/` | Go 官方模块布局指南介绍的可执行程序目录惯例，并非编译器要求 |
| `internal/` | Go 工具链支持的包导入访问限制，适合应用内部代码 |
| `pkg/` | 部分社区项目采用的可选目录；无对外复用需求时不创建，也不会自动使包成为公共 API |
| `handler/`、`router/`、`middleware/` | 对 HTTP 处理、路由注册和 Gin 中间件职责的组织选择，目录名称由项目决定 |
| `service/`、`repository/` | 分层架构与 Repository 模式的应用，帮助分离业务规则和存储实现 |
| `migrations/` | 数据库迁移文件的组织选择，按项目迁移工具约定使用 |

依赖隔离借鉴 Clean Architecture 的思想，但仅有 `handler → service → repository` 的目录或调用链并不等于完整实现整洁架构；应检查实际源码依赖和业务边界。`infra`、`dto` 等名称同样是本模板的组织约定，不应描述为 Go 社区统一标准。

新项目按层组织。Handler 和 Service 在层内按业务模块拆包，Entity 和 Repository 按同一业务边界一起拆包。已有项目的调整遵循其现行约定和用户要求。需要借鉴成熟项目时，阅读其实际入口、装配和模块代码，说明适用条件，不以目录相似作为生产验证的证据。

## 使用流程

1. 先读取项目的 `AGENTS.md`、`go.mod`、入口和现有目录，识别已有模块、框架版本、存储方案及启动方式。
2. 新项目默认采用下述布局；已有项目优先沿用一致且合理的结构。用户明确要求统一布局时，再规划迁移映射。
3. 仅创建当前功能需要的目录和文件。用户只要求设计或审查时，输出目录树、职责和建议，不生成业务实现。
4. 实施时沿一个实际请求串起路由、Handler、Service 和需要的存储实现，确保依赖可装配；不生成大量空壳模块。
5. 根据变更验证目录、依赖关系与构建，并报告实际执行结果；未执行的检查要说明原因。

## 默认目录结构

默认按层组织。Handler 和 Service 从第一个模块开始按业务拆包。单一业务域中，Entity 和 Repository 各用一个包。下面先展示顶层用途，再分段展开；各段属于同一个项目，`user` 和 `order` 仅为业务示例，功能文件按实际需求创建。

### 顶层概览

```text
project/
├── cmd/              # 服务进程入口，至少包含 cmd/api/
├── internal/         # 应用装配、业务代码和基础设施
│   ├── app/          # 启动与资源生命周期
│   ├── svc/          # Service 共享依赖与运行配置
│   ├── config/       # 配置加载与校验
│   ├── router/       # Gin 引擎与路由注册
│   ├── middleware/   # 按需：自定义 HTTP 中间件
│   ├── handler/      # HTTP 接口处理
│   ├── service/      # 业务用例与规则
│   ├── repository/   # 存储访问实现
│   ├── entity/       # 业务实体与值对象
│   └── infra/        # 按需：外部连接、限流等基础能力
├── configs/          # 按需：使用配置文件时放脱敏示例
├── api/              # 按需：OpenAPI 等接口契约
├── migrations/       # 按需：数据库迁移
├── tests/e2e/        # 按需：pytest 端到端接口测试
├── scripts/          # 按需：开发与构建脚本
├── deployments/      # 按需：部署配置
├── go.mod
├── go.sum            # 由 Go 工具按实际依赖生成
└── README.md         # 启动、配置、请求与验证说明
```

`cmd/`、`internal/`、`go.mod` 和 `README.md` 是新项目的基本组成，其余目录按当前功能实际需要创建。只用环境变量配置、没有配置文件的项目不必创建 `configs/`；骨架项目或纯目录规划不必创建 `tests/`，但应在交付说明中写明未包含端到端测试。

### 启动、配置与应用装配

```text
project/
├── cmd/
│   └── api/
│       └── main.go             # 进程入口
├── internal/
│   ├── app/
│   │   └── api.go              # 初始化基础设施、启动和资源清理
│   ├── svc/
│   │   └── servicecontext.go   # 装配 Service 共享依赖
│   └── config/
│       └── config.go           # 配置加载、默认值与校验
└── configs/
    └── config.example.yaml     # 使用配置文件时提供脱敏示例
```

`app` 和 `router` 的文件按进程或用途命名，不重复包名：API 进程使用 `app/api.go`、`router/api.go`；增加 worker 等入口时可加 `app/worker.go`。

#### ServiceContext 的传递

- `cmd` 加载并校验配置文件，用配置的日志级别创建共享日志器；`app` 接收已验证的 `config.Config`，初始化基础设施并调用 `svc.NewServiceContext(...)`，整个应用实例共享同一个 `*svc.ServiceContext`。连接的创建与关闭由 `app` 负责。
- `NewServiceContext` 使用已有连接构造仓储，将仓储接口汇集到 `Repo`。`app` 将日志器、运行配置和数据库连接传给 `NewServiceContext`，由 `ServiceContext.Logger`、`ServiceContext.Config` 提供给 Handler 和 Service，`ServiceContext.Health` 提供给探针。记录请求日志时使用当前请求的上下文。登录限流等独立能力单独注入。Service 只读取自己负责的配置段。
- `app` 将 `svcCtx` 显式传给 `router.New(svcCtx)`。注册路由时，Handler 构造函数接收该指针并返回持有它的 `gin.HandlerFunc`。
- 每次请求到来时，Handler 调用对应业务用例的构造函数，例如 `userservice.NewProfileService(c.Request.Context(), svcCtx)`，创建当前请求的 Service。Service 使用请求上下文传递超时与取消，通过 `svcCtx.Repo` 调用仓储。共享依赖在启动时装配，请求状态留在各自的请求中。

### HTTP 接入

```text
internal/
├── router/api.go         # Gin 引擎和路由分组
├── router/user.go        # 用户模块路由（路由增多时按业务拆分）
├── router/order.go       # 订单模块路由
├── router/health.go      # liveness/readiness/startup 探针
├── middleware/requestid.go # 请求 ID 生成与透传
├── middleware/accesslog.go # 访问日志、跳过健康检查
├── middleware/recovery.go  # panic 恢复与结构化 500
├── middleware/auth.go    # 鉴权等 HTTP 中间件（按需）
└── handler/
    ├── user/profile.go  # 用户资料接口及其请求、响应类型
    └── order/create.go  # 创建订单接口
```

路由仍归 `router` 包：`api.go` 只保留引擎、中间件、版本分组和模块注册，模块路由在 `router/<module>.go` 中注册；Handler 包继续只提供 `gin.HandlerFunc`。

### 业务用例

```text
internal/service/
├── user/profile.go      # 用户资料用例及其消费的能力接口
└── order/create.go      # 创建订单用例及其消费的能力接口
```

### 存储访问

```text
internal/repository/
├── user.go          # t_user：表映射、接口和私有实现
└── order.go         # t_order / t_order_item：表映射、接口和私有实现
```

仓储包的划分与 Entity 对齐。关系型仓储按表或紧密表组分文件，文件名省略表名的 `t_` 前缀。详见 [Repository 仓储规范](#repository-仓储规范)。业务接口统一经 Handler 调用 Service；涉及存储访问时，由 Service 通过接口调用对应 Repository。没有存储访问的功能无需制造空 Repository；健康检查等运行状态接口可直接响应，不作为业务模块。

### 实体、基础设施与测试

这些用途在各自章节展开，此处不重复完整层级：

| 用途 | 目录 | 展开位置 |
| --- | --- | --- |
| 业务实体与值对象 | `internal/entity/` | [业务实体的组织](#业务实体的组织) |
| 数据库、缓存与消息队列连接 | `internal/infra/` | [可选基础设施](#可选基础设施缓存与消息队列) |
| pytest 通过 HTTP 测试真实服务 | `tests/e2e/` | [端到端接口测试的目录与组织](#端到端接口测试的目录与组织) |

## 多业务模块的组织

Handler 和 Service 从第一个业务模块开始使用业务子包。Entity 和 Repository 采用下文的统一划分规则。

- **Handler 和 Service 按业务拆包**。用户模块使用 `handler/user/` 和 `service/user/`，订单等其他模块在这两层建立对应子包。存储访问由 Service 调用仓储接口。
- **文件按职责命名**：Handler 和 Service 子包内按具体功能或用例命名文件，如 `handler/user/profile.go`、`handler/order/create.go`；目录已表达所属层和模块，文件名不再重复包名，例如用 `app/api.go`、`router/api.go`，而不是 `app/app.go`、`router/router.go`。相关接口可按职责聚合，不要求每个 HTTP 路由各建一个文件，两层的文件也不必机械一一对应。
- **HTTP 类型就近维护**：请求和响应类型放在对应 Handler 功能文件中；确有共享需求时在该业务子包内提取类型文件，不默认建立集中式 `dto` 目录。已有项目使用 DTO 包时，按现行约定处理。
- **类型与构造函数体现业务用例**。Service 类型和构造函数按具体职责命名，例如 `user.ProfileService` 与 `user.NewProfileService`，以及 `order.CreateService` 与 `order.NewCreateService`。Handler 构造函数也可以体现对应功能，如 `NewProfileHandler`。装配处可用 `userhandler`、`userservice` 等导入别名区分同名包。
- **模块间通过能力接口协作**：订单用例需要查询用户状态时，在 `service/order` 内定义它消费的小接口（如 `UserReader`），由 `app` 装配进 `ServiceContext` 或构造参数；不调用其他模块的 Handler，不引入循环导入。不要与仓储接口混用：仓储接口描述存储能力，放在 repository 包并汇入 `svcCtx.Repo`；跨模块服务能力定义在消费方。
- **统一装配和路由**：`app` 装配各模块，`router` 注册模块路由；`config`、`middleware` 和 `infra` 等公共基础能力按用途复用，不为每个模块重复维护连接。

### 业务实体的组织

在单一业务域或模块彼此紧密相关时，`internal/entity/` 和 `internal/repository/` 各自使用一个包。Entity 按实体或值对象分文件，Repository 按表或紧密表组分文件。示例中订单依赖用户，因此两个目录都采用单包。

```text
internal/entity/
├── user.go          # 用户实体
├── order.go         # 订单实体
└── order_item.go    # 订单明细实体
```

- 存在多个职责独立、边界清晰的业务模块时，Entity 和 Repository 按相同的模块名一起拆包。例如账户与计费独立时，使用 `entity/account/` 对应 `repository/account/`，`entity/billing/` 对应 `repository/billing/`。仓储子包按实际存储需求创建。
- 文件名表达实体或值对象，不按 `create.go`、`query.go` 等接口动作拆分；同一实体的查询、创建和修改共用其业务定义，HTTP 请求与响应仍留在接口边界。
- `entity` 不反向依赖 Handler、Service、Repository 或 `infra`。数据库专用行结构与 ORM 映射仍优先留在 Repository 中，不因拆包将协议或存储细节搬入业务模型。

## 分层职责与依赖

典型请求流向为 `router → handler → service → repository`。Service 通过仓储包导出的接口调用存储能力，仓储不反向依赖 Service：

- `app` 管理基础设施生命周期，启动时将 `svc.NewServiceContext` 装配的共享依赖传给 Router。Router 注册持有共享依赖的 Handler，Handler 为每个请求创建 Service。
- `router` 只处理 Gin 引擎、版本路径、路由分组和中间件挂载，不承载业务规则或 SQL。
- `handler` 绑定并校验 HTTP 输入，调用 Service，将业务错误映射为状态码和响应。HTTP DTO 留在边界，不将 Gin 对象传入业务层。
- `service` 承载业务规则和用例编排，通过 `UserRepository` 等接口调用仓储。可导入仓储接口所属包，但不使用 GORM 表模型、`*gorm.DB` 或仓储私有实现；业务判断由 Service 负责，用例专用查询和存储事务实现留在仓储层。
- `repository` 定义仓储接口并提供私有实现，封装查询、持久化与存储事务，完成表映射和业务实体转换；不接收 `*gin.Context`、不写 HTTP 响应、不反向调用 Service。
- `entity` 放业务实体和值对象。HTTP 请求响应留在接口边界，数据库专用行结构和 ORM 映射放在 Repository 包内，仓储对业务层返回业务实体。
- `infra` 维护实际使用的数据库、缓存、消息队列客户端初始化，以及登录限流等基础能力；由 `app` 管理其生命周期，不依赖 Handler、Router 或业务 Service，不承载业务用例编排。

Service 的输入输出使用业务类型或用例参数，不依赖 `dto` 或 Handler 包中的 HTTP 协议结构。Handler 负责 DTO 与业务参数的转换。跨模块协作通过明确的 Service 能力或小接口完成，不让 Handler 相互调用。

避免全局数据库句柄、全局 Service 和在 `init()` 中建立外部连接。接口只围绕实际消费需求定义，不要求每个结构体都配一个接口。

## 配置与依赖注入

配置是启动时装配的只读数据，不是需要替换的行为依赖。默认把完整的 `config.Config` 放入 `svcCtx.Config`，由 Service 按需读取。

- **值为准**：配置以值传递，启动后视为只读；Service 只读取自己负责的配置段，不修改配置，也不把配置到处转发。
- **不包汇总类型**：不要建立包含所有业务配置的 `BusinessConfig`，它只是把同一个问题换个名字；需要分组时直接按模块拆 `OrderConfig`、`PaymentConfig` 等，再由 `Config` 组合。
- **接口只用于行为**：Repository、缓存、消息发布器、限流器和外部客户端值得用接口或构造注入；配置字段不要逐项定义接口或 getter。
- **隔离按需升级**：模块配置明显增多、需要独立部署或形成安全边界时，再把子配置显式传入构造函数；模块较少时，从 `svcCtx.Config.<模块>` 读取即可。

## Repository 仓储规范

Repository 与 [业务实体](#业务实体的组织) 按同一业务边界划分。关系型仓储按表或紧密表组分文件，表映射、实际需要的仓储接口、私有实现和构造函数放在一起。下面的 GORM 映射和构造函数规则适用于使用 GORM 的项目，不因应用目录规范自动引入 ORM。

### 按表命名与文件内容

以 `t_user` 为例：

| 职责 | 命名与归属 |
| --- | --- |
| 数据库表 | `t_user` |
| Go 文件 | `internal/repository/user.go`，省略 `t_` 前缀 |
| 表名常量 | 在同文件定义，例如 `tableNameUser`，值为 `t_user` |
| GORM 表映射 | `userRow`（私有），`TableName()` 返回表名常量 |
| 仓储接口 | `UserRepository`，与映射及实现同文件定义 |
| 私有实现 | `userRepository`，封装已有数据库连接 |
| 构造函数 | `NewUserRepository(db *gorm.DB) UserRepository`，接收已有连接并返回接口 |

其他表沿用同一套命名规则。同一业务包可以包含多张相关表，不要求一表一包。**只有存在实际存储操作时才定义仓储接口，不给每张表预建空 CRUD。**

仓储接口描述实际需要的存储能力，不把每张表自动扩展成通用 CRUD。构造函数不自行建立连接、不负责迁移表结构；连接的创建、复用和关闭由 `app` 管理。

### 实体、调用与事务边界

- `entity.User` 是业务实体，`userRow`（`internal/repository` 中的私有行结构）是 `t_user` 的 GORM 表映射。仓储在内部转换，对业务层返回业务实体，不返回 GORM 表模型、数据库句柄或查询构建器。
- `app` 将已有连接和运行配置传给 `svc.NewServiceContext`，由该构造函数创建仓储并将其接口放入 `svcCtx.Repo`，同时把配置放入 `svcCtx.Config`。Service 通过这些接口执行存储操作，Handler 将业务请求交给 Service。
- 用例专用查询和存储事务留在仓储层。Service 负责业务条件和流程，调用仓储提供的操作，不直接操作 GORM，也不在业务层拼接表查询或事务。

### Redis 登录限流器的归属

登录限流器提供的是 `Allow` 能力，属于基础设施能力。实现放在 `internal/infra/ratelimit/`，例如 `redis.go`；由 `app` 装配并注入 Service，业务流程在相应位置调用 `Allow`。

登录限流器不放进 `svcCtx.Repo`，也不作为登录业务的 HTTP middleware。Redis 客户端仍由 `infra/cache` 等现有连接初始化模块提供；限流算法、key 和计数规则由 `infra/ratelimit` 维护，不重复建立连接。用户资料缓存等存储适配器与登录限流器按职责分别组织。

## 数据库迁移

`migrations/` 使用按文件名升序执行的 SQL 文件，推荐 `NNN_<动词>_<对象>.sql`，例如 `001_create_user_order_tables.sql`。

- 已应用的迁移不可修改、重命名或删除；修复用新文件，一个文件只做一个逻辑变更。
- 应用启动不执行迁移；由迁移工具、部署流水线或运维显式执行并记录版本。Docker 的 initdb 只用于本地空库初始化。
- 生产迁移不包含示例数据；本地 seed 在文件名中明确标识并在部署时排除，当前示例的 `002_seed_demo_user.sql` 仅用于本地。
- 破坏性变更先备份，按 expand-contract 分步上线，保证新旧代码与 schema 兼容。
- 迁移工具不限定，但同一项目只选一种；完整规则见 [数据库迁移规范](references/migrations.md)。

## Gin 与资源边界

- `*gin.Context` 留在 HTTP 层；向 Service 和 Repository 传递 `c.Request.Context()`，保持超时和取消信号。
- 业务错误不包含 HTTP 状态码；在 Handler 或统一的 HTTP 错误处理中映射。沿用项目既有响应格式，不擅自把所有错误改成 HTTP 200。
- 错误响应统一经 Handler 映射：4xx 属于预期结果，默认不写 `Error`，需要排查时可在 Handler 中加 `Debug` 日志；5xx 或未知错误必须写 `Error`，字段包含 `request_id`、method、path、status 和 error，不记录请求体、凭据或令牌。panic 由 `internal/middleware/recovery.go` 捕获并记录结构化 500。
- 请求进入时生成或透传 `X-Request-ID`，校验其字符集与长度，并写入响应头和日志；访问日志由 `internal/middleware/accesslog.go` 输出，包含 `request_id`、跳过健康检查、不记录 query。4xx 不写结构化 `Error`，但仍会出现在访问日志。错误响应与日志字段在 Handler 中直接映射；多个 Handler 出现重复时再提取公共函数，不为少量接口预建错误处理包。完整规则见 [日志打印规范](references/logging.md)。
- 身份校验可放在中间件；资源归属、业务权限和状态流转由业务层判断，不能只依赖路由是否挂载鉴权中间件。
- Service 决定业务流程与条件，用例专用的查询、跨表写入及存储事务封装在仓储层；Handler 不查库，Service 不持有或操作 GORM 事务对象，不为目录规范强行引入事务框架。
- 连接初始化、启动失败清理、HTTP 优雅关闭和资源释放由入口与 `app` 协调；业务包不调用 `os.Exit`。
- 启动日志在监听成功后输出：先 `net.Listen` 再 `Serve`，端口占用等错误直接返回；端到端测试以 `/readyz` 等就绪接口判断可用，不依赖固定等待时间。
- 配置代码放 `internal/config`；使用配置文件时，脱敏示例放 `configs/`。真实凭据通过项目既有配置机制注入，不写入示例或源码。
- MySQL 的 `infra.mysql` 配置（host/port/user/password/database、驱动参数、连接池和 Ping 超时）由 `internal/config` 定义；DSN 由 `internal/infra/database` 使用驱动 API 构造。

## 可选基础设施：缓存与消息队列

缓存和消息队列是常用扩展，按实际需求启用；没有使用时不创建对应目录、不引入客户端依赖。它们属于基础设施，`internal/middleware` 仍只放 Gin HTTP 中间件。

只展开基础设施相关目录，分别在使用对应依赖时创建：

```text
internal/infra/
├── database/
│   └── mysql.go         # 数据库连接池初始化
├── cache/
│   └── redis.go         # 缓存客户端初始化
├── mq/
│   └── client.go       # 所选消息队列的连接与客户端初始化
└── ratelimit/
    └── redis.go        # 按需：Redis 登录限流 Allow 能力
```

### 连接配置与生命周期

- `internal/config` 定义 `infra.mysql` 的 host/port/user/password/database、驱动参数、连接池和 Ping 超时；必填项校验、默认值和 DSN 构造在 `internal/infra/database` 初始化时完成；存在配置文件时由 `configs` 保存脱敏示例。
- `internal/infra/cache` 初始化 Redis 等缓存客户端；`internal/infra/mq` 初始化所选消息队列的连接、生产者或消费者客户端。产品命名可用 `redis.go`、`rabbitmq.go` 或 `kafka.go`，不要求同时支持多种产品。
- `internal/app` 根据配置创建并复用客户端，通过构造函数注入实际使用方；负责启动失败时清理已创建资源，并在退出时按依赖顺序关闭。不要在每次请求中重新连接。
- “可选”表示项目可不引入该能力；引入后是启动必需依赖还是允许故障降级，应由业务要求明确，不因目录标为可选而自动忽略连接失败。

### 缓存代码归属

Redis 登录限流的 `Allow` 能力按 [仓储规范中的职责划分](#redis-登录限流器的归属) 放在 `internal/infra/ratelimit/`，不作为用户资料缓存或仓储接口。

- `infra/cache` 只维护客户端和通用连接配置。业务缓存的 key、TTL、读写、失效和序列化逻辑放在对应业务适配器中，例如 `internal/repository/user_cache.go`，避免基础设施包引用业务 Service。
- Service 需要直接表达缓存能力时，在消费方定义小接口，由具体缓存适配器实现；只有加速存储查询的缓存也可封装为 Repository 装饰器，保持 Service 的存储接口不变。
- 按实际业务明确缓存未命中、故障降级和写后失效策略。缓存用作会话、幂等或其他正确性依赖时，不把连接错误当作普通未命中。

### 消息代码归属

- 按需新增 `internal/messaging/`，存放业务消息的发布适配器，例如 `order_publisher.go`；业务事件结构可放 `internal/event/`。队列名称、路由键、消息编解码与具体客户端操作属于消息适配器，平台连接包不承载订单等业务流程。
- Service 在消费方定义需要的发布接口，例如 `OrderEventPublisher`，由 `app` 注入具体实现；HTTP Handler 不直接向消息队列发布业务消息。
- 消费入口放 `internal/consumer/`，例如 `order.go`，负责解码、校验消息、调用 Service，并根据处理结果执行所选客户端支持的确认或提交操作。消费者不依赖 Gin，也不复用 HTTP DTO 作为消息契约。
- 需要独立运行或扩缩容时增加 `cmd/worker/main.go`；与 API 同进程运行时由 `app` 统一启动和停止消费任务。退出时先停止接收新任务、等待在途处理和待发送消息完成，再关闭依赖客户端，并设置等待超时。
- 涉及消息处理时明确幂等、失败重试和失败消息去向；不要假设只投递一次。数据库写入与发布消息需要一致性保证时，按业务选择 outbox 等方案，不把普通数据库事务当作跨数据库与消息队列的原子事务。

## 命名与规模演进

- Go 包名采用简短的小写词，避免下划线和无意义复数；文件使用 `snake_case.go`。
- 不默认创建 `utils`、`common`、`base` 等杂物包；共享能力按用途命名，并放在最小合理作用域。
- 应用私有代码优先放 `internal`。只有确实需要被其他模块导入且准备维护其 API 时，才建立对外包；不默认创建 `pkg`。
- 已有项目采用 `internal/user/{handler,service,repository}/` 等按业务聚合的布局时，先遵循其现行约定；用户要求统一布局后再规划迁移。新项目使用上述分层规则，Entity 与 Repository 随业务边界一起调整。
- 同包内拆文件只影响代码组织。进一步拆包时，按职责和依赖关系划分，避免循环导入。
- 单一服务默认一个 Go module；多入口可以新增 `cmd/worker` 等目录。只有独立发布或依赖隔离确有需求时才拆多 module，不因存在多个入口就引入 `go.work`。

## 接口与实现的注释

这是强制约定，审查时必须检查。

- 接口声明是业务语义的唯一权威说明：用途、参数约束、返回值、错误以及事务或并发假设都写在接口方法上，不能只写在实现处。
- 实现注释只承担定位职责，不重复接口语义；接口方法重命名或迁移后同步更新链接。
- 同一实现方法满足多个接口、或需要引用其他包的接口时，使用可解析的完整链接，例如 `[repository.UserRepository.GetByID]`。

实现项目内接口的方法注释必须使用 Go doc 链接指向所实现的接口方法，格式固定为：

```go
// GetByID implements [UserRepository.GetByID].
```

## 已有项目迁移

先列出“现有路径 → 目标路径 → 原因”，再按业务模块迁移并同步 import、构造函数及受影响的构建和部署入口。保持接口路径、响应格式和业务行为不变；如需改变，单独说明并遵循用户任务范围。

不要仅为统一名称将既有 `controller` 全量重命名为 `handler`，也不要同时保留两套承担相同职责的目录。与仓库现行规则冲突时优先遵循仓库规则，并说明差异。

## 端到端接口测试的目录与组织

端到端接口测试使用 **pytest**，放在项目根目录的 `tests/e2e/`。项目包含 HTTP 接口、需要验证真实服务和完整业务流程时创建；只做目录规划、纯骨架或尚未提供测试时，说明未包含测试即可，不伪造测试文件。测试先启动真实 Go/Gin 服务并等待就绪，再由 Python 测试通过 HTTP 访问服务，验证接口和完整业务流程。

测试与业务代码分开维护，默认仍在同一仓库；Python 测试依赖和 pytest 配置由 `tests/e2e/pyproject.toml` 独立管理，不放入业务的 `go.mod`。只有测试覆盖多个独立仓库或需要独立维护、发布时，再考虑拆成测试仓库。这是本模板的项目约定，不是 Go/Gin 的强制目录。

以下仅为目录规划，不表示已提供对应文件或实现；只规划目录时，不自动创建测试代码、依赖配置、启动脚本或 CI 配置：

```text
tests/
└── e2e/
    ├── pyproject.toml         # Python 测试依赖与 pytest 配置
    ├── conftest.py            # HTTP 客户端、认证与测试数据等共享 fixture
    ├── test_user_profile.py   # 用户资料接口场景
    ├── test_order_create.py   # 订单创建接口场景
    ├── test_order_flow.py     # 跨接口、跨模块的订单业务流程
    └── testdata/              # 请求样例与初始化数据
```

- **按业务场景组织**：使用 `test_<业务>_<场景>.py` 命名文件，跨接口流程单独维护，不与 Handler、Service 文件机械对应。用例增多时可按业务子目录组织；共享准备和清理由 `conftest.py` 中的 fixture 承担，作用域按资源生命周期选择，不让用例依赖执行顺序。
- **真实服务与真实 HTTP**：服务由本地测试环境或 CI 预先启动，健康检查通过后再运行 pytest。通过配置的 `BASE_URL` 访问服务，使用 HTTPX 或 requests 等 HTTP 客户端发送请求；不在测试进程内直接调用 Handler、Service 或 Repository，也不替换被测业务组件。
- **配置与辅助能力归测试侧**：服务地址、测试账号和认证信息由测试配置或环境变量注入；HTTP 客户端、认证和通用响应检查集中维护，复杂时再提取辅助模块。Go 业务代码不依赖 Python 测试代码，测试依赖也不混入业务依赖管理。
- **隔离数据和外部依赖**：使用独立的测试数据库、缓存及队列资源，按场景准备并清理数据；`testdata` 只保存可复用样例和初始化数据，不保存真实凭据或运行产物。业务流程涉及存储或消息处理时，连接相应的真实测试实例并验证实际结果；内存或替代实现的检查需单独说明，不能替代真实依赖的集成验证。
- **独立执行与生命周期**：pytest 使用独立的 Python 环境和 CI 阶段运行，按需通过 pytest 标记区分场景。明确环境准备、服务启动、就绪等待、用例执行与清理的责任；fixture 负责自己管理的客户端和测试数据，不擅自停止外部已启动的服务。
- **断言业务结果**：验证状态码、响应内容及后续可观察的业务结果，例如创建后查询、实际数据写入或消息处理结果。异步结果采用有超时的轮询判断，不以请求发送成功或固定等待一段时间作为通过依据。

## 项目示例代码

- 需要查看 HTTP 调用链、依赖装配和 MySQL 持久化示例时，阅读 [项目代码示例](references/examples.md)，按需查看 [可运行双模块项目](assets/gin-api/) 中对应层的源文件。示例使用 `ServiceContext` 传递共享日志器、运行配置和仓储接口，`repository/user.go` 和 `repository/order.go` 演示 GORM 表映射与事务；MySQL 连接、建表脚本和 Docker Compose 也在资产中。关系型仓储的按表命名与 GORM 映射见 [Repository 仓储规范](#repository-仓储规范)。
- 示例使用 `handler/user/profile.go`、`handler/order/create.go` 等具体功能文件，业务实体放在 `internal/entity/`，HTTP 类型留在 Handler，展示依赖注入。订单与用户属于同一紧密业务域，示例为保持最小化直接复用 `svcCtx.Repo.User`，并由 order Handler 用 `errors.Is` 映射 `entity.ErrUserNotFound`，未单独定义跨模块能力接口，此写法仅适合示例规模。模块边界独立或需要替换用户来源时，按 [多业务模块的组织](#多业务模块的组织) 在 `service/order` 中抽取消费方接口。示例默认使用 MySQL + GORM，启动前需按 README 准备数据库；缓存和消息队列仍按需引入，未包含 pytest 端到端测试。
- 需要连接 MySQL、Redis 或消息队列时，再阅读 [可选基础设施接入代码](references/infrastructure.md)。这些片段说明代码归属和装配，不是默认项目必须启用的能力。

## 审查清单

审查目录结构和代码归属时，逐项核对并按“问题 → 依据 → 建议”输出；用户只要求审查时，不直接改代码。

- **入口与装配**：`cmd` 加载并校验配置、按配置创建日志器并启动；`app` 管理基础设施生命周期与优雅关闭；`svc` 只装配共享依赖，不含业务规则。
- **依赖方向**：调用链为 `router → handler → service → repository`，`entity` 不反向依赖其他层，包之间无循环导入。
- **HTTP 边界**：`*gin.Context` 不进入 Service/Repository；HTTP 请求响应类型留在 Handler；业务错误在 Handler 映射状态码，不泄漏内部细节。
- **模块边界**：Handler 和 Service 按业务拆包；跨模块通过消费方定义的接口协作，不调用其他模块的 Handler；同一紧密业务域可复用仓储接口并共享该域的实体错误，跨模块时通过消费方接口或错误翻译隔离，不要让 Handler 判断无关领域的错误。
- **存储职责**：仓储接口按实际存储需求定义并与实现同包，无存储访问的模块不建空 Repository；表映射、用例查询和事务留在仓储层。
- **命名与规模**：包名简短小写；文件名不重复包名；不建 `utils`、`common` 等杂物包；未使用的基础设施、`pkg`、`middleware` 不为“看起来完整”而创建。
- **接口注释**：接口声明写清用途、参数、返回值、错误与事务/并发语义；项目内接口的实现方法必须使用 `// Method implements [Interface.Method].` 链接，接口重命名后同步更新。
- **日志与错误响应**：4xx 默认不写 `Error`、需要时用 `Debug`；5xx 必须写 `Error` 并带 `request_id`、method、path、status、error；对外响应不泄露内部错误。
- **配置注入**：配置以只读值放入 `ServiceContext`；不增加汇总的 `BusinessConfig`，不为配置字段造接口；行为依赖才用接口注入，配置隔离按模块边界演进。
- **数据库迁移**：`migrations/` 按 `NNN_<动词>_<对象>.sql` 排序执行；已应用文件不改，生产迁移不含 seed；破坏性变更先备份并按 expand-contract 上线。
- **配置与测试**：配置代码在 `internal/config`，配置文件示例在 `configs/`（仅在存在配置文件时）；端到端测试在 `tests/e2e/`（仅在提供测试时），Go 代码不依赖 Python 测试代码。
- **交付完整性**：README 说明启动、配置与接口示例；用 `gofmt`、`go build ./...` 等实际检查支撑结论，未执行的检查明确标注。

## 验证与交付

- 设计交付：给出适合当前规模的目录树、各层职责及关键依赖关系。
- 审查交付：按[审查清单](#审查清单)输出问题、依据和建议，不把审查与修改混在一步。
- 代码交付：确认 module/import 与目录一致、路由已注册、依赖已装配、错误有映射；检查实际涉及的配置和启动路径。
- 注释交付：项目内接口的方法有具体语义说明，实现方法带有 `// Method implements [Interface.Method].` 链接，且链接与接口名一致。
- 文档交付：`README.md` 说明启动命令、关键配置项或环境变量、健康检查和主要接口示例；现有项目缺少 README 时按任务范围补充。
- Go 代码变更后，对修改文件执行 `gofmt`，运行 `go build ./...`，并遵循仓库既有检查。没有可运行的 Go 项目时不声称已验证构建。
- 不因使用本技能擅自引入 ORM、依赖注入框架、缓存、消息队列、Docker 或 Kubernetes；沿用项目技术栈，确有需求时再添加。

## 参考资料

- [Go 官方模块布局指南](https://go.dev/doc/modules/layout)：理解 `cmd`、`internal` 与服务器项目布局；本技能的分层名称属于项目约定。
- [Gin 官方文档](https://gin-gonic.com/en/docs/)：实现路由、中间件与绑定时，按项目使用版本核对具体 API。
- [日志打印规范](references/logging.md)：日志级别、请求 ID、字段约定、错误记录与验证清单。
- [数据库迁移规范](references/migrations.md)：迁移命名、执行、seed 与破坏性变更规则。
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)：社区项目布局参考，非 Go 官方标准；按需参考，不将其目录列表视为必建清单。
