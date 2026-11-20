# 数据库迁移规范

本规范约定 `migrations/` 的命名、内容和执行方式。目标是让任意环境都能从空库按顺序构建，并且已应用的历史不会被改写。

## 目录与命名

```text
migrations/
├── 001_create_user_order_tables.sql
├── 002_seed_demo_user.sql   # 仅本地示例，生产排除
└── 003_add_order_status.sql # 新增迁移从下一个序号继续
```

- 文件放在项目根目录 `migrations/`，按文件名升序执行。
- 使用 `NNN_<动词>_<对象>.sql`：三位零填充序号 + 下划线 + 英文小写描述，例如 `001_create_user_order_tables.sql`。
- 同一项目固定一种编号方式；多人并行开发需要时间戳时，改用 UTC 时间戳（如 `20261120153000_add_order_status.sql`），不要混用。
- 一个文件只做一个逻辑变更；不要把无关变更塞进同一个迁移。

## 基本规则

- 已应用的迁移不可修改、重命名或删除；修复用新的迁移文件。
- 序号不复用，迁移按文件名顺序执行，不依赖执行顺序之外的手工步骤。
- 迁移只描述 schema 和相关数据变更，不读取业务配置、不调用外部服务。
- 应用启动不自动执行迁移；由迁移工具、部署流水线或运维显式执行，并记录已应用版本。同一项目只选一种迁移工具。
- Docker Compose 的 `/docker-entrypoint-initdb.d` 只用于本地空库初始化，不能替代生产迁移工具。
- 生产迁移不包含示例或开发数据。本地 seed 在文件名中明确标识（如 `002_seed_demo_user.sql`），部署时排除；当前示例的 seed 仅用于本地。
- 破坏性变更（删表、删列、改类型）先备份，采用 expand-contract：先加新结构并双写/回填，再切读，最后删除旧结构。
- 文件中的表名、列名与 Repository 映射保持一致；不在迁移文件里配置 ORM 自动迁移。

## 本地执行

清空本地数据并重新初始化：

```bash
docker compose down -v
docker compose up -d --wait
```

已有数据库手动执行（以下命令在 Compose 容器内使用 3306；宿主机映射端口是 3307）：

```bash
docker compose exec -T mysql mysql -uroot -proot gin_layout < migrations/001_create_user_order_tables.sql
# 仅本地示例；生产环境不要执行
docker compose exec -T mysql mysql -uroot -proot gin_layout < migrations/002_seed_demo_user.sql
```

对宿主机自装 MySQL 执行时，使用 `mysql -h 127.0.0.1 -P 3306 ...`，并确保 `configs/config.yaml` 的 `infra.mysql.port` 与之一致。

## 验证清单

- 空库按文件名顺序执行成功。
- 在已有数据的库上向前执行成功；破坏性变更已评估备份与回滚。
- 已应用文件未被修改，新变更有新的文件。
- 代码与 schema 的部署顺序明确：先兼容旧代码，再切新结构。
- 本地 Compose 初始化只用于开发；生产由迁移工具执行并记录版本。
