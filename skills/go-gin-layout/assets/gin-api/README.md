# gin-api

Go/Gin 分层 API 示例：路由、Handler、Service、Repository 分层，使用 MySQL + GORM 持久化用户和订单。

## 目录

```text
cmd/api/main.go                  # 进程入口
internal/app/api.go              # 启动、优雅关闭、数据库生命周期
internal/config/config.go        # YAML 配置加载；MySQL 参数初始化时校验
internal/infra/database/mysql.go # MySQL 连接池、DSN 构造与 Ping
internal/router/api.go           # 引擎、中间件与模块路由注册
internal/router/user.go          # 用户模块路由
internal/router/order.go         # 订单模块路由
internal/router/health.go        # liveness/readiness/startup 探针
internal/handler/                # HTTP 输入输出与错误映射
internal/middleware/requestid.go # X-Request-ID 生成与透传
internal/middleware/accesslog.go # 访问日志
internal/middleware/recovery.go  # panic 恢复与结构化 500
internal/service/                # 业务用例
internal/repository/             # GORM 表映射、仓储接口与实现
internal/entity/                 # 业务实体
internal/svc/servicecontext.go   # 共享依赖装配
configs/config.example.yaml      # YAML 配置示例
migrations/                      # 建表与本地示例数据
docker-compose.yml               # 本地 MySQL
```

## 启动

1. 启动 MySQL（首次启动会自动执行 `migrations/` 下的 SQL）：

```bash
docker compose up -d --wait
```

2. 复制配置并按需修改 `configs/config.yaml`：

```bash
cp configs/config.example.yaml configs/config.yaml
go run ./cmd/api
```

默认读取 `configs/config.yaml`，可通过 `CONFIG_FILE` 指定其他路径。默认监听 `127.0.0.1:8080`；启动时会 Ping 数据库，连接失败则进程退出。

## 请求示例

```bash
curl -i http://127.0.0.1:8080/livez
curl -i http://127.0.0.1:8080/readyz

curl -i http://127.0.0.1:8080/api/v1/users/1

curl -i -X POST http://127.0.0.1:8080/api/v1/orders \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1,"items":[{"sku":"BOOK-001","quantity":2}]}'
```

`002_seed_demo_user.sql` 预置了用户 `1 / Alice`。用户不存在返回 404；请求体不合法、SKU 超过 64 字符、quantity 超出 INT 范围或 items 超过 100 条返回 400；请求体超过 1 MiB 返回 413。

探针：`/livez` 和 `/startupz` 只表示进程已启动，`/readyz` 会带超时检查数据库并返回 200/503，`/healthz` 保留为 liveness 别名。

## 配置

配置来自 YAML 文件，默认路径 `configs/config.yaml`，可用 `CONFIG_FILE` 指定其他文件。`configs/config.yaml` 不提交，`configs/config.example.yaml` 是脱敏示例。

| YAML 字段 | 必填 | 说明 |
| --- | --- | --- |
| `http.addr` | 否 | 监听地址，默认 `127.0.0.1:8080` |
| `infra.mysql.host` / `infra.mysql.port` | 是 | MySQL 地址；Compose 内部端口 3306，宿主机映射为 3307 |
| `infra.mysql.user` / `infra.mysql.password` | 是 | 数据库账号；密码允许为空 |
| `infra.mysql.database` | 是 | 数据库名 |
| `infra.mysql.params` | 否 | 驱动参数，如 `charset: utf8mb4` |
| `infra.mysql.pool.*` | 否 | 连接池与超时：`max_open_conns`、`max_idle_conns`、`conn_max_lifetime`、`ping_timeout` |
| `log_level` | 否 | 日志级别：`debug`、`info`、`warn`、`error`，默认 `info` |

DSN 由 `internal/infra/database` 根据上述字段构造，不在配置文件里拼字符串。YAML 使用严格字段校验，未知字段会导致启动失败；MySQL 必填项和连接池默认值在 `database.OpenMySQL` 初始化时校验/补全。`docker-compose.yml` 中的 `root/root` 与固定端口仅用于本地开发，生产环境通过部署配置注入真实凭据。

## 日志与错误响应

- 每个响应带 `X-Request-ID`：合法入站值会透传，否则自动生成；访问日志和错误日志都用该 ID 关联请求。
- 访问日志由 `internal/middleware/accesslog.go` 输出，包含 method/path/status/latency/request_id，跳过健康检查，不记录 query。带尾斜杠但未注册的路径按 404 处理，也经过请求 ID 和访问日志中间件。
- 400/404 等预期错误不写结构化 `Error` 日志，但访问日志仍会记录；需要排查时在 Handler 中临时增加 `Debug` 日志。
- 500/未知错误和就绪检查失败（503）以 `Error` 记录 `request_id`、method、path、status、error；panic 由 `internal/middleware/recovery.go` 捕获并返回统一 JSON 500。
- 错误响应在 Handler 中直接映射；多个 Handler 重复相同日志字段时再提取公共函数。
- 完整规范见 [日志打印规范](../../references/logging.md)。

## 数据库

- `migrations/001_create_user_order_tables.sql` 创建 `t_user`、`t_order`、`t_order_item`。
- `migrations/002_seed_demo_user.sql` 仅预置本地示例用户；生产环境不要执行。
- 迁移命名与执行规则见 [数据库迁移规范](../../references/migrations.md)。
- 对 Compose 内置 MySQL 执行迁移（容器内端口 3306，宿主机映射 3307）：

```bash
docker compose exec -T mysql mysql -uroot -proot gin_layout < migrations/001_create_user_order_tables.sql
docker compose exec -T mysql mysql -uroot -proot gin_layout < migrations/002_seed_demo_user.sql
```

- 使用宿主机已有 MySQL 时，按 `infra.mysql.host/port` 配置连接（常见为 3306）：

```bash
mysql -h 127.0.0.1 -P 3306 -uroot -p -e 'CREATE DATABASE IF NOT EXISTS gin_layout CHARACTER SET utf8mb4'
mysql -h 127.0.0.1 -P 3306 -uroot -p gin_layout < migrations/001_create_user_order_tables.sql
mysql -h 127.0.0.1 -P 3306 -uroot -p gin_layout < migrations/002_seed_demo_user.sql
```

重置本地数据：

```bash
docker compose down -v
```

## 开发约定

- 接口方法在 `interface` 声明处写清业务语义、参数、返回值和错误。
- 实现项目内接口的方法使用 `// Method implements [Interface.Method].` 注释链接接口方法，例如 `// GetByID implements [UserRepository.GetByID].`。
- 配置作为只读数据放入 `ServiceContext`；需要分组时按模块拆配置结构，不建汇总的 `BusinessConfig`，也不为配置字段定义接口。

## 验证

```bash
gofmt -l .
go build ./...
go vet ./...
```
