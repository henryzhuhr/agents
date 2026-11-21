# 可选基础设施接入示例

本文展示 MySQL、Redis 和 RabbitMQ 接入时的代码归属。MySQL 已随 [可运行示例](../assets/gin-api/) 落地并实际连接验证；Redis 和 RabbitMQ 仍是独立片段，不包含在默认示例中，也不会连接外部服务。只复制实际需要的部分，再在目标项目添加对应依赖。片段已使用 `gorm.io/gorm v1.31.1`、`gorm.io/driver/mysql v1.6.0`、`go-redis/v9 v9.23.0`、`amqp091-go v1.17.0` 编译验证，Redis 和 RabbitMQ 未连接实际服务。

| 路径 | 维护内容 |
| --- | --- |
| `internal/config` | 启动配置（HTTP、MySQL 连接参数、连接池、超时、日志级别）的加载与校验 |
| `internal/infra/database/mysql.go` | MySQL 连接池、DSN 构造与健康检查 |
| `internal/infra/cache/redis.go` | Redis 客户端初始化 |
| `internal/infra/mq/rabbitmq.go` | AMQP 连接及 Channel 生命周期 |
| `internal/repository/user_cache.go` | 用户缓存的 key、TTL、序列化 |
| `internal/messaging/order_created.go` | 订单消息的队列、编码、发布 |
| `internal/app` | 创建客户端、注入使用方、启动失败清理与退出关闭 |

配置值由 `internal/config` 从项目既有配置来源加载；真实凭据不写入源码、示例或日志。MySQL 的 host/port/user/password/database、连接池和 Ping 超时都在配置中定义，DSN 由 `internal/infra/database` 用驱动 API 构造，不手拼字符串。

## MySQL：GORM 连接池

`internal/infra/database/mysql.go`。`config.Infra.MySQL` 提供 host/port/user/password/database/params 和连接池、Ping 超时；`OpenMySQL` 负责必填项校验、默认值补全和 DSN 构造，DSN 由 `go-sql-driver/mysql` 的 `Config.FormatDSN()` 生成。示例使用 MySQL 8.4。关闭驱动的版本探测和 GORM 的自动 Ping，连接检查统一由带超时的 `PingContext` 执行。可运行示例的 GORM 日志启用 `ParameterizedQueries`，省略 SQL 绑定参数。相关 API 见 [GORM 文档](https://gorm.io/docs/)。表映射和 SQL 留在 Repository，业务层不导入 GORM。

```go
package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"example.com/gin-layout/internal/config"
)

func OpenMySQL(ctx context.Context, cfg config.MySQLConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn(cfg),
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DisableAutomaticPing: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.Pool.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Pool.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Pool.ConnMaxLifetime))
	pingCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Pool.PingTimeout))
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}
```

`dsn(cfg)` 使用 `mysqldriver.Config{User, Passwd, Net: "tcp", Addr, DBName, Params}.FormatDSN()` 生成 DSN。

`app` 将返回的 `*gorm.DB` 和运行配置传给 `svc.NewServiceContext`，由后者构造 Repository 和健康检查能力；Repository 使用 `WithContext` 传递请求上下文。连接池的关闭由 `app` 负责。

## Redis：连接与业务缓存分开

`internal/infra/cache/redis.go`。传入的 Options 已由配置层构造；开启 `ContextTimeoutEnabled`，让命令处理请求的超时和截止时间。相关 API 见 [go-redis 官方文档](https://pkg.go.dev/github.com/redis/go-redis/v9#Options)。

```go
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func OpenRedis(ctx context.Context, options *redis.Options) (*redis.Client, error) {
	opts := *options
	opts.ContextTimeoutEnabled = true
	client := redis.NewClient(&opts)
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}
```

这里的缓存是用户资料存储适配器。Redis 登录限流的 `Allow` 能力应放在 `internal/infra/ratelimit/`，由 `app` 装配供 Service 调用，不放入 `Repo` 或登录 middleware；详见 [仓储与限流职责](../SKILL.md#redis-登录限流器的归属)。

业务缓存放在 `internal/repository/user_cache.go`；下面的用户资料缓存复用示例项目的 `entity.User`，明确区分“未命中”和“操作失败”。

```go
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"example.com/gin-layout/internal/entity"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCache(client *redis.Client, ttl time.Duration) *Cache {
	return &Cache{client: client, ttl: ttl}
}

func (c *Cache) Get(ctx context.Context, id uint64) (entity.User, bool, error) {
	raw, err := c.client.Get(ctx, fmt.Sprintf("user:profile:%d", id)).Bytes()
	if errors.Is(err, redis.Nil) {
		return entity.User{}, false, nil
	}
	if err != nil {
		return entity.User{}, false, err
	}
	var user entity.User
	if err := json.Unmarshal(raw, &user); err != nil {
		return entity.User{}, false, err
	}
	return user, true, nil
}

func (c *Cache) Put(ctx context.Context, user entity.User) error {
	raw, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, fmt.Sprintf("user:profile:%d", user.ID), raw, c.ttl).Err()
}
```

`app` 将配置中的正数 TTL 传给 `NewCache`。需要查询加速时，可把它封装在 Repository 装饰器内；若 Service 确实需要缓存语义，在消费方定义 `Get`、`Put` 组成的小接口并注入。只有作为可降级的查询加速时，才可按业务要求把缓存故障记录后回源；资料更新后还需失效或更新缓存。本片段没有实现完整缓存一致性策略。

## RabbitMQ：通用连接与业务发布分开

`internal/infra/mq/rabbitmq.go`。创建 Channel 失败时释放已创建连接；共享一个长期连接，按职责配置 Channel。连接超时由所选客户端配置，下面仅演示资源所有权。

```go
package mq

import (
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func OpenRabbitMQ(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &Client{Conn: conn, Channel: channel}, nil
}

func (c *Client) Close() error {
	return errors.Join(c.Channel.Close(), c.Conn.Close())
}
```

`internal/messaging/order_created.go` 决定业务队列与消息格式，不把订单概念写入 `infra/mq`。

```go
package messaging

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderPublisher struct {
	channel *amqp.Channel
	queue   string
}

func NewOrderPublisher(channel *amqp.Channel) (*OrderPublisher, error) {
	queue, err := channel.QueueDeclare("order.created", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return &OrderPublisher{channel: channel, queue: queue.Name}, nil
}

func (p *OrderPublisher) PublishCreated(ctx context.Context, orderID uint64) error {
	body, err := json.Marshal(struct {
		OrderID uint64 `json:"order_id"`
	}{OrderID: orderID})
	if err != nil {
		return err
	}
	return p.channel.PublishWithContext(ctx, "", p.queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}
```

需要消息发布时，在 `internal/service/order` 定义消费方接口，并扩展该用例的构造函数保存它；这些改动未预先加入示例：

```go
type CreatedPublisher interface {
	PublishCreated(ctx context.Context, orderID uint64) error
}
```

`app` 依次创建连接、业务发布适配器、Service，并将发布适配器作为上述接口传入。任何后续初始化失败，都释放前面已创建的资源。停止时先停止接收 HTTP 请求和新消费任务，等待在途业务处理及待发送消息完成，再关闭 Channel、AMQP 连接、Redis 和数据库；设置停机等待上限，并记录关闭错误。

此发布示例使用持久队列和持久消息，**没有实现 publisher confirms**；发布返回 `nil` 不代表消息已持久化或业务已消费。需要可靠发布时，还应处理 broker 确认、不可路由消息以及重试；数据库写入与消息发布的一致性需求可通过 outbox 等方案解决。

`amqp091-go v1.17.0` 的 `PublishWithContext` 在调用发布前检查 Context，进入 `Publish` 后的在途网络写入不会由该 Context 中断；应另外按客户端配置控制网络超时，使用其他版本时重新核对行为。[amqp091-go 官方 API](https://pkg.go.dev/github.com/rabbitmq/amqp091-go#Channel.PublishWithContext)

消费者放在 `internal/consumer`，使用手动确认，在业务处理成功后再 Ack；失败按业务要求进行有限重试或转入失败队列，处理重复投递的幂等性。发布确认与消费确认的含义不同。[RabbitMQ 官方确认机制说明](https://www.rabbitmq.com/docs/confirms)
