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

新项目统一采用按层、层内再按业务模块拆包的默认布局，不按业务规模切换组织方式；已有项目的调整遵循其现行约定和用户要求。需要借鉴成熟项目时，阅读其实际入口、装配和模块代码，说明适用条件，不以目录相似作为生产验证的证据。

## 使用流程

1. 先读取项目的 `AGENTS.md`、`go.mod`、入口和现有目录，识别已有模块、框架版本、存储方案及启动方式。
2. 新项目默认采用下述布局；已有项目优先沿用一致且合理的结构。用户明确要求统一布局时，再规划迁移映射。
3. 仅创建当前功能需要的目录和文件。用户只要求设计或审查时，输出目录树、职责和建议，不生成业务实现。
4. 实施时沿一个实际请求串起路由、Handler、Service 和需要的存储实现，确保依赖可装配；不生成大量空壳模块。
5. 根据变更验证目录、依赖关系与构建，并报告实际执行结果；未执行的检查要说明原因。

## 默认目录结构

默认采用“按层组织、层内按业务模块拆包”，即使只有一个小模块，也采用相同结构。下面先展示顶层用途，再分段展开；各段属于同一个项目，`user` 和 `order` 仅为业务示例，功能文件按实际需求创建。

### 顶层概览

```text
project/
├── cmd/              # 服务进程入口
├── internal/         # 应用装配、业务代码和基础设施
├── configs/          # 脱敏配置示例
├── api/              # 可选：OpenAPI 等接口契约
├── migrations/       # 可选：数据库迁移
├── tests/            # pytest 端到端接口测试
├── scripts/          # 按需：开发与构建脚本
├── deployments/      # 按需：部署配置
├── go.mod
├── go.sum            # 由 Go 工具按实际依赖生成
└── README.md
```

### 启动、配置与应用装配

```text
cmd/api/main.go              # 进程入口
internal/app/app.go          # 依赖装配、启动和资源清理
internal/config/config.go    # 配置加载、默认值与校验
configs/config.example.yaml # 按实际配置方式提供脱敏示例
```

### HTTP 接入

```text
internal/
├── router/router.go      # Gin 引擎和路由分组
├── middleware/auth.go    # 鉴权等 HTTP 中间件
└── handler/
    ├── user/profile.go  # 用户资料接口及其请求、响应类型
    └── order/create.go  # 创建订单接口
```

### 业务用例

```text
internal/service/
├── user/profile.go      # 用户资料用例及其消费的能力接口
└── order/create.go      # 创建订单用例及其消费的能力接口
```

### 存储访问

```text
internal/repository/
├── user/mysql.go        # 用户存储实现
└── order/mysql.go       # 订单存储实现
```

Repository 文件按实际存储方案命名，例如 `mysql.go` 或 `memory.go`。业务接口统一经 Handler 调用 Service；涉及存储访问时，由 Service 通过接口调用对应 Repository。没有存储访问的功能无需制造空 Repository；健康检查等运行状态接口可直接响应，不作为业务模块。

### 模型、基础设施与测试

这些用途在各自章节展开，此处不重复完整层级：

| 用途 | 目录 | 展开位置 |
| --- | --- | --- |
| 业务实体与值对象 | `internal/model/` | [Model 的拆分](#model-的拆分) |
| 数据库、缓存与消息队列连接 | `internal/infra/` | [可选基础设施](#可选基础设施缓存与消息队列) |
| pytest 通过 HTTP 测试真实服务 | `tests/e2e/` | [端到端接口测试的目录与组织](#端到端接口测试的目录与组织) |

## 多业务模块的组织

以下是默认规范，从第一个业务模块开始执行，不以模块数量、代码行数或业务复杂度作为是否拆包的条件：

- **每层按业务拆包**：用户模块使用 `handler/user/`、`service/user/`，有存储访问时使用 `repository/user/`；订单等其他模块在各层建立对应子包。新增模块不把实现堆入层根目录的 `user.go`、`order.go`，也不跳过业务 Service 层。
- **文件按职责命名**：Handler 和 Service 子包内按具体功能或用例命名文件，如 `handler/user/profile.go`、`handler/order/create.go`；目录已表达所属层和模块，文件名不再重复为 `handler.go`、`service.go`。相关接口可按职责聚合，不要求每个 HTTP 路由各建一个文件，两层的文件也不必机械一一对应。
- **存储实现归对应模块**：Repository 在业务子包内按实际存储实现组织，例如 `repository/user/mysql.go`、`repository/user/memory.go`。消费方接口留在 Service 子包，可就近定义在用例文件中，需要集中维护时再使用 `repository.go` 等文件。
- **HTTP 类型就近维护**：请求和响应类型放在对应 Handler 功能文件中；确有共享需求时在该业务子包内提取类型文件，不默认建立集中式 `dto` 目录。已有项目使用 DTO 包时，按现行约定处理。
- **名称利用包上下文**：在 Handler、Service 子包内使用 `Handler`、`Service`、`New` 等名称；装配处遇到同名包时使用 `userhandler`、`userservice` 等导入别名区分。文件名和类型名分别表达各自职责。
- **模块间通过能力接口协作**：例如订单用例需要查询用户状态时，定义它需要的用户查询接口，由 `app` 注入；不调用其他模块的 Handler，不引入循环导入。
- **统一装配和路由**：`app` 装配各模块，`router` 注册模块路由；`config`、`middleware` 和 `infra` 等公共基础能力按用途复用，不为每个模块重复维护连接。

### Model 的拆分

`model` 按业务实体或值对象组织，不要求与 Handler、Service 的目录和文件一一对应。模型较少时保持平铺即可：

```text
internal/model/
├── user.go          # 用户实体
├── order.go         # 订单实体
└── order_item.go    # 订单明细实体
```

- 只有某个业务域包含较多相关模型、且需要独立包边界时，再按业务域拆包，例如 `model/user/account.go`、`model/user/profile.go`、`model/user/address.go`。只有一个 `User` 实体时，无需为它单独套一层目录。
- 文件名表达实体或值对象，不按 `create.go`、`query.go` 等接口动作拆分；同一实体的查询、创建和修改共用其业务定义，HTTP 请求与响应仍留在接口边界。
- `model` 不反向依赖 Handler、Service、Repository 或 `infra`。数据库专用行结构与 ORM 映射仍优先留在 Repository 中，不因拆包将协议或存储细节搬入业务模型。

## 分层职责与依赖

典型请求流向为 `router → handler → service → repository`。这是调用链；源码依赖通过消费方接口隔离：

- `app` 是装配入口：创建基础设施、Repository、Service 和 Handler，再注册路由；通过构造函数显式传入依赖。
- `router` 只处理 Gin 引擎、版本路径、路由分组和中间件挂载，不承载业务规则或 SQL。
- `handler` 绑定并校验 HTTP 输入，调用 Service，将业务错误映射为状态码和响应。HTTP DTO 留在边界，不将 Gin 对象传入业务层。
- `service` 承载业务规则、用例编排和事务边界。需要替换存储或隔离外部依赖时，在消费方定义小接口，例如 `UserRepository`；Repository 实现该接口，Service 不导入具体 Repository 包。
- `repository` 实现查询与持久化，不接收 `*gin.Context`、不写 HTTP 响应、不反向调用 Service。仅在实现接口确有需要时引用接口所属包，避免反向依赖形成环。
- `model` 放业务实体和值对象；默认不含 HTTP 绑定标签。数据库专用行结构和 ORM 映射优先放在 Repository 包内；简单 CRUD 若复用实体与持久化模型，应说明耦合取舍，不额外复制无意义结构。
- `infra` 仅封装实际使用的数据库、缓存、消息队列或外部客户端初始化；由 `app` 管理其生命周期，不依赖 Handler、Router 或业务 Service。

Service 的输入输出使用业务类型或用例参数，不依赖 `dto` 或 Handler 包中的 HTTP 协议结构。Handler 负责 DTO 与业务参数的转换。跨模块协作通过明确的 Service 能力或小接口完成，不让 Handler 相互调用。

避免全局数据库句柄、全局 Service 和在 `init()` 中建立外部连接。接口只围绕实际消费需求定义，不要求每个结构体都配一个接口。

## Gin 与资源边界

- `*gin.Context` 留在 HTTP 层；向 Service 和 Repository 传递 `c.Request.Context()`，保持超时和取消信号。
- 业务错误不包含 HTTP 状态码；在 Handler 或统一的 HTTP 错误处理中映射。沿用项目既有响应格式，不擅自把所有错误改成 HTTP 200。
- 身份校验可放在中间件；资源归属、业务权限和状态流转由业务层判断，不能只依赖路由是否挂载鉴权中间件。
- 多步写入的事务由用例决定，Repository 提供事务内执行能力；不在 Handler 中拼接事务，也不为目录规范强行引入事务框架。
- 连接初始化、启动失败清理、HTTP 优雅关闭和资源释放由入口与 `app` 协调；业务包不调用 `os.Exit`。
- 配置代码放 `internal/config`，配置示例放 `configs`；真实凭据通过项目既有配置机制注入，不写入示例或源码。

## 可选基础设施：缓存与消息队列

缓存和消息队列是常用扩展，按实际需求启用；没有使用时不创建对应目录、不引入客户端依赖。它们属于基础设施，`internal/middleware` 仍只放 Gin HTTP 中间件。

只展开基础设施相关目录，分别在使用对应依赖时创建：

```text
internal/infra/
├── database/
│   └── mysql.go         # 数据库连接池初始化
├── cache/
│   └── redis.go         # 缓存客户端初始化
└── mq/
    └── client.go       # 所选消息队列的连接与客户端初始化
```

### 连接配置与生命周期

- `internal/config` 定义并校验连接地址、认证信息、连接池及超时等配置；按所选产品补充配置项，`configs` 只保存脱敏示例。
- `internal/infra/cache` 初始化 Redis 等缓存客户端；`internal/infra/mq` 初始化所选消息队列的连接、生产者或消费者客户端。产品命名可用 `redis.go`、`rabbitmq.go` 或 `kafka.go`，不要求同时支持多种产品。
- `internal/app` 根据配置创建并复用客户端，通过构造函数注入实际使用方；负责启动失败时清理已创建资源，并在退出时按依赖顺序关闭。不要在每次请求中重新连接。
- “可选”表示项目可不引入该能力；引入后是启动必需依赖还是允许故障降级，应由业务要求明确，不因目录标为可选而自动忽略连接失败。

### 缓存代码归属

- `infra/cache` 只维护客户端和通用连接配置。业务缓存的 key、TTL、读写、失效和序列化逻辑放在对应业务适配器中，例如 `internal/repository/user/cache.go`，避免基础设施包引用业务 Service。
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
- 已有项目采用 `internal/user/{handler,service,repository}/` 等按业务聚合的布局时，先遵循其现行约定；用户要求统一布局后再规划迁移。新项目沿用层内业务子包的默认方式，只有用户明确指定其他布局时才调整，不因规模变化自动切换。
- 业务子包是默认边界，同包内继续拆文件不会产生新的包级隔离。模块内部确需进一步拆子包时，依据职责和依赖关系划分，避免循环导入。
- 单一服务默认一个 Go module；多入口可以新增 `cmd/worker` 等目录。只有独立发布或依赖隔离确有需求时才拆多 module，不因存在多个入口就引入 `go.work`。

## 已有项目迁移

先列出“现有路径 → 目标路径 → 原因”，再按业务模块迁移并同步 import、构造函数及受影响的构建和部署入口。保持接口路径、响应格式和业务行为不变；如需改变，单独说明并遵循用户任务范围。

不要仅为统一名称将既有 `controller` 全量重命名为 `handler`，也不要同时保留两套承担相同职责的目录。与仓库现行规则冲突时优先遵循仓库规则，并说明差异。

## 端到端接口测试的目录与组织

端到端接口测试使用 **pytest**，放在项目根目录的 `tests/e2e/`。先启动真实 Go/Gin 服务并等待就绪，再由 Python 测试通过 HTTP 访问服务，验证接口和完整业务流程。

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

- 需要查看完整目录如何对应实际代码、初始化新项目或实现业务模块时，阅读 [项目代码示例](references/examples.md)，按需查看 [可运行双模块项目](assets/gin-api/) 中对应层的源文件。
- 示例使用 `handler/user/profile.go`、`handler/order/create.go` 等具体功能文件，业务实体放在 `model`，HTTP 类型留在 Handler，展示依赖注入和跨模块能力接口；默认内存存储，不要求引入数据库、缓存或消息队列。
- 需要连接 MySQL、Redis 或消息队列时，再阅读 [可选基础设施接入代码](references/infrastructure.md)。这些片段说明代码归属和装配，不是默认项目必须启用的能力。

## 验证与交付

- 设计交付：给出适合当前规模的目录树、各层职责及关键依赖关系。
- 代码交付：确认 module/import 与目录一致、路由已注册、依赖已装配、错误有映射；检查实际涉及的配置和启动路径。
- Go 代码变更后，对修改文件执行 `gofmt`，运行 `go build ./...`，并遵循仓库既有检查。没有可运行的 Go 项目时不声称已验证构建。
- 不因使用本技能擅自引入 ORM、依赖注入框架、缓存、消息队列、Docker 或 Kubernetes；沿用项目技术栈，确有需求时再添加。

## 参考资料

- [Go 官方模块布局指南](https://go.dev/doc/modules/layout)：理解 `cmd`、`internal` 与服务器项目布局；本技能的分层名称属于项目约定。
- [Gin 官方文档](https://gin-gonic.com/en/docs/)：实现路由、中间件与绑定时，按项目使用版本核对具体 API。
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)：社区项目布局参考，非 Go 官方标准；按需参考，不将其目录列表视为必建清单。
