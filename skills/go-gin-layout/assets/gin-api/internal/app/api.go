package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"gorm.io/gorm"

	"example.com/gin-layout/internal/config"
	"example.com/gin-layout/internal/infra/database"
	"example.com/gin-layout/internal/router"
	"example.com/gin-layout/internal/svc"
)

type App struct {
	server *http.Server
	db     *gorm.DB
	logger *slog.Logger
}

func New(cfg config.Config, logger *slog.Logger) (*App, error) {
	db, err := database.OpenMySQL(context.Background(), cfg.MySQL.DSN)
	if err != nil {
		return nil, err
	}
	// 配置、日志器和数据库连接只装配一次，供所有请求共享。
	svcCtx := svc.NewServiceContext(logger, cfg, db)
	engine, err := router.New(svcCtx)
	if err != nil {
		_ = closeDatabase(db)
		return nil, fmt.Errorf("create router: %w", err)
	}
	return &App{
		server: &http.Server{
			Addr:              cfg.HTTP.Addr,
			Handler:           engine,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		db:     db,
		logger: logger,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer func() {
		if err := closeDatabase(a.db); err != nil {
			a.logger.Error("close database", "error", err)
		}
	}()

	// 先监听成功再对外宣告启动，端口占用时不会误报就绪。
	listener, err := net.Listen("tcp", a.server.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", a.server.Addr, err)
	}
	a.logger.Info("API listening", "addr", listener.Addr().String())

	stopped := make(chan error, 1)
	go func() { stopped <- a.server.Serve(listener) }()
	select {
	case err := <-stopped:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		// 请求取消后仍需一个独立的上下文等待在途请求结束。
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			_ = a.server.Close()
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		err := <-stopped
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

func closeDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	return sqlDB.Close()
}
