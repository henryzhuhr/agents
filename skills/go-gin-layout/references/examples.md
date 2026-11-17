# 项目代码示例

## 可直接运行的双模块 API

完整代码位于 [assets/gin-api](../assets/gin-api/)，使用 Go 1.25 或更高版本，固定 Gin v1.12.0。示例采用本技能默认的“按层、再按业务子包”结构，展示用户查询与订单创建两个模块；只有一个小模块时也沿用这种组织方式。

```text
gin-api/
├── cmd/api/main.go
├── internal/
│   ├── app/api.go
│   ├── config/config.go
│   ├── router/api.go
│   ├── handler/
│   │   ├── user/profile.go
│   │   └── order/create.go
│   ├── service/
│   │   ├── user/profile.go
│   │   └── order/create.go
│   ├── repository/
│   │   ├── user/memory.go
│   │   └── order/memory.go
│   └── model/
│       ├── user.go
│       ├── order.go
│       └── order_item.go
├── go.mod
└── go.sum
```

默认使用内存 Repository，预置用户 `1 / Alice`，订单只保存在当前进程中，重启后清空。用户由请求中的 `user_id` 指定，未实现身份认证；示例聚焦目录和依赖，不包含库存、支付等业务能力。

### 启动与请求

在本仓库根目录执行：

```bash
cd skills/go-gin-layout/assets/gin-api
go run ./cmd/api
```

默认监听 `127.0.0.1:8080`，可通过 `HTTP_ADDR` 调整。按 Ctrl+C 停止时会等待在途 HTTP 请求结束，最多等待 5 秒。首次运行需下载 `go.mod` 中的依赖。

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

路径 ID 不合法、空订单或非正数量返回 HTTP 400；用户不存在返回 HTTP 404；未预期的内部错误返回 HTTP 500，响应不包含底层错误细节。

### 代码阅读顺序

| 文件 | 关注点 |
| --- | --- |
| [cmd/api/main.go](../assets/gin-api/cmd/api/main.go) | 加载配置、接收退出信号并调用应用 |
| [internal/config/config.go](../assets/gin-api/internal/config/config.go) | 环境变量、默认值与配置校验 |
| [internal/app/api.go](../assets/gin-api/internal/app/api.go) | 显式装配 Repository、Service、Handler，并管理 HTTP 生命周期 |
| [internal/router/api.go](../assets/gin-api/internal/router/api.go) | 统一注册两个模块的路由，健康检查直接响应 |
| [internal/handler/user/profile.go](../assets/gin-api/internal/handler/user/profile.go) | 路径参数、HTTP 响应结构与错误映射 |
| [internal/handler/order/create.go](../assets/gin-api/internal/handler/order/create.go) | JSON 请求结构、绑定校验与业务参数转换 |
| [internal/service/user/profile.go](../assets/gin-api/internal/service/user/profile.go) | 消费方定义存储接口，业务层不依赖 Gin |
| [internal/service/order/create.go](../assets/gin-api/internal/service/order/create.go) | 消费方定义 `UserReader`，通过注入的用户能力完成跨模块协作 |
| [internal/repository/order/memory.go](../assets/gin-api/internal/repository/order/memory.go) | 存储实现、并发保护与数据副本 |
| [internal/model/order_item.go](../assets/gin-api/internal/model/order_item.go) | 独立业务实体，无 JSON 或绑定标签 |

订单 Service 不直接导入用户 Service；`app` 将满足 `UserReader` 接口的用户 Service 注入订单 Service。Handler 之间没有相互调用。HTTP 请求与响应类型就近放在功能文件中，因此该示例不额外建立 `dto` 目录。

## 可选基础设施

需要使用数据库、缓存或消息队列时，再读取 [基础设施接入代码](infrastructure.md)。默认项目不包含这些客户端依赖，也不会尝试连接外部服务。

复用时按目标项目调整 module 路径和 import，仅取实际需要的模块。层内业务子包划分遵循本技能默认规范；具体功能文件和可选基础设施按实际需求选择。
