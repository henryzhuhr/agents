// Package svc 维护应用启动时装配的共享依赖。
package svc

import (
	"log/slog"

	"gorm.io/gorm"

	"example.com/gin-layout/internal/config"
	"example.com/gin-layout/internal/repository"
)

// ServiceContext 保存应用启动时装配、由各请求共享的依赖。
// 请求级状态不放入该结构，而是通过函数参数或请求上下文传递。
//

type ServiceContext struct {
	// Logger 是启动时创建的共享结构化日志器；记录请求日志时使用当前请求的上下文。
	Logger *slog.Logger

	// Config 保存应用运行配置，启动后只读；Service 按需读取自己负责的配置段。
	// 只读数据，不是行为依赖；Service 只读取自己负责的配置段。
	// 模块配置增多时按业务拆分 config 子结构，不额外汇总 BusinessConfig，也不为配置定义接口。
	Config config.Config

	// Repo 汇集各业务 Service 使用的仓储接口。
	Repo Repository
}

// Repository 汇集业务 Service 使用的仓储接口。
type Repository struct {
	User  repository.UserRepository
	Order repository.OrderRepository
}

// NewServiceContext 使用启动时创建的日志器、运行配置和数据库连接装配共享依赖。
func NewServiceContext(logger *slog.Logger, cfg config.Config, db *gorm.DB) *ServiceContext {
	return &ServiceContext{
		Logger: logger,
		Config: cfg,
		Repo: Repository{
			User:  repository.NewUserRepository(db),
			Order: repository.NewOrderRepository(db),
		},
	}
}
