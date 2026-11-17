# 可选基础设施接入示例

本文展示 MySQL、Redis 和 RabbitMQ 接入时的代码归属。这些是独立的接入片段，不是完整可运行项目；默认的 `assets/gin-api` 仍使用内存存储，不包含这些客户端依赖，也不会连接外部服务。只复制实际需要的部分，再在目标项目添加对应依赖。片段已使用 `go-sql-driver/mysql v1.10.1`、`go-redis/v9 v9.23.0`、`amqp091-go v1.17.0` 编译验证，未连接实际服务。

| 路径 | 维护内容 |
| --- | --- |
| `internal/config` | 连接地址、认证、超时、连接池等配置的加载与校验 |
| `internal/infra/database/mysql.go` | MySQL 连接池初始化 |
| `internal/infra/cache/redis.go` | Redis 客户端初始化 |
| `internal/infra/mq/rabbitmq.go` | AMQP 连接及 Channel 生命周期 |
| `internal/repository/user/cache.go` | 用户缓存的 key、TTL、序列化 |
| `internal/messaging/order_created.go` | 订单消息的队列、编码、发布 |
| `internal/app` | 创建客户端、注入使用方、启动失败清理与退出关闭 |

配置值由 `internal/config` 从项目既有配置来源加载；真实凭据不写入源码、示例或日志。示例中的超时和池大小用于展示配置位置，使用时改为项目配置项。

## MySQL：初始化连接池

`internal/infra/database/mysql.go`。`dsn` 由配置层传入，包含项目所需的连接、读写超时等驱动参数；`sql.Open` 不等于建立连接，因此启动时还要执行 `PingContext`。相关行为见 [database/sql](https://pkg.go.dev/database/sql#Open) 和 [MySQL 驱动文档](https://github.com/go-sql-driver/mysql#usage)。

```go
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func OpenMySQL(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(3 * time.Minute)
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}
```

`app` 将返回的 `*sql.DB` 注入具体 Repository；SQL、行结构及业务实体转换留在 Repository，业务 Service 不导入数据库驱动。

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

业务缓存放在 `internal/repository/user/cache.go`；下面的用户资料缓存复用示例项目的 `model.User`，明确区分“未命中”和“操作失败”。

```go
package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"example.com/gin-layout/internal/model"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCache(client *redis.Client, ttl time.Duration) *Cache {
	return &Cache{client: client, ttl: ttl}
}

func (c *Cache) Get(ctx context.Context, id uint64) (model.User, bool, error) {
	raw, err := c.client.Get(ctx, fmt.Sprintf("user:profile:%d", id)).Bytes()
	if errors.Is(err, redis.Nil) {
		return model.User{}, false, nil
	}
	if err != nil {
		return model.User{}, false, err
	}
	var user model.User
	if err := json.Unmarshal(raw, &user); err != nil {
		return model.User{}, false, err
	}
	return user, true, nil
}

func (c *Cache) Put(ctx context.Context, user model.User) error {
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

需要消息发布时，在 `internal/service/order` 定义消费方接口，并扩展该用例的构造函数保存它；这些改动未预先加入内存示例：

```go
type CreatedPublisher interface {
	PublishCreated(ctx context.Context, orderID uint64) error
}
```

`app` 依次创建连接、业务发布适配器、Service，并将发布适配器作为上述接口传入。任何后续初始化失败，都释放前面已创建的资源。停止时先停止接收 HTTP 请求和新消费任务，等待在途业务处理及待发送消息完成，再关闭 Channel、AMQP 连接、Redis 和数据库；设置停机等待上限，并记录关闭错误。

此发布示例使用持久队列和持久消息，**没有实现 publisher confirms**；发布返回 `nil` 不代表消息已持久化或业务已消费。需要可靠发布时，还应处理 broker 确认、不可路由消息以及重试；数据库写入与消息发布的一致性需求可通过 outbox 等方案解决。

`amqp091-go v1.17.0` 的 `PublishWithContext` 在调用发布前检查 Context，进入 `Publish` 后的在途网络写入不会由该 Context 中断；应另外按客户端配置控制网络超时，使用其他版本时重新核对行为。[amqp091-go 官方 API](https://pkg.go.dev/github.com/rabbitmq/amqp091-go#Channel.PublishWithContext)

消费者放在 `internal/consumer`，使用手动确认，在业务处理成功后再 Ack；失败按业务要求进行有限重试或转入失败队列，处理重复投递的幂等性。发布确认与消费确认的含义不同。[RabbitMQ 官方确认机制说明](https://www.rabbitmq.com/docs/confirms)
