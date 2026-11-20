// Package database 负责数据库连接池的创建、健康检查和关闭约定。
package database

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"example.com/gin-layout/internal/config"
)

const (
	mysqlDefaultPort       = 3306                   // 默认默认端口
	defaultMaxOpenConns    = 20                     // 默认连接池最大连接数
	defaultMaxIdleConns    = 5                      // 默认连接池最大空闲连接数
	defaultConnMaxLifetime = 3 * time.Minute        // 默认连接最大复用时长
	defaultPingTimeout     = 3 * time.Second        // 默认启动 Ping 超时
	slowQueryThreshold     = 200 * time.Millisecond // GORM 慢查询阈值
)

// OpenMySQL 校验连接参数、补全连接池默认值并创建 GORM 连接。
// 连接的关闭由调用方（app）负责。
func OpenMySQL(ctx context.Context, cfg config.MySQLConfig) (*gorm.DB, error) {
	cfg, err := normalize(cfg)
	if err != nil {
		return nil, err
	}
	gormLogger := logger.New(
		log.New(os.Stderr, "", log.LstdFlags),
		logger.Config{
			SlowThreshold:             slowQueryThreshold,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
	db, err := gorm.Open(mysql.Open(dsn(cfg)), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               gormLogger,
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

func normalize(cfg config.MySQLConfig) (config.MySQLConfig, error) {
	if cfg.Host == "" {
		return cfg, fmt.Errorf("infra.mysql.host is required")
	}
	if cfg.Port == 0 {
		cfg.Port = mysqlDefaultPort
	}
	if cfg.User == "" {
		return cfg, fmt.Errorf("infra.mysql.user is required")
	}
	if cfg.Database == "" {
		return cfg, fmt.Errorf("infra.mysql.database is required")
	}
	if cfg.Pool.MaxOpenConns <= 0 {
		cfg.Pool.MaxOpenConns = defaultMaxOpenConns
	}
	if cfg.Pool.MaxIdleConns <= 0 {
		cfg.Pool.MaxIdleConns = defaultMaxIdleConns
	}
	if cfg.Pool.ConnMaxLifetime <= 0 {
		cfg.Pool.ConnMaxLifetime = config.Duration(defaultConnMaxLifetime)
	}
	if cfg.Pool.PingTimeout <= 0 {
		cfg.Pool.PingTimeout = config.Duration(defaultPingTimeout)
	}
	return cfg, nil
}

// Health 提供数据库就绪检查能力。
type Health struct {
	db *gorm.DB
}

func NewHealth(db *gorm.DB) *Health {
	return &Health{db: db}
}

func (h *Health) Ping(ctx context.Context) error {
	sqlDB, err := h.db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

func dsn(cfg config.MySQLConfig) string {
	driverCfg := mysqldriver.Config{
		User:                 cfg.User,
		Passwd:               cfg.Password,
		Net:                  "tcp",
		Addr:                 net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		DBName:               cfg.Database,
		Params:               cfg.Params,
		ParseTime:            true,
		Loc:                  time.Local,
		AllowNativePasswords: true,
	}
	return driverCfg.FormatDSN()
}
