package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"example.com/gin-layout/internal/config"
	orderhandler "example.com/gin-layout/internal/handler/order"
	userhandler "example.com/gin-layout/internal/handler/user"
	orderrepo "example.com/gin-layout/internal/repository/order"
	userrepo "example.com/gin-layout/internal/repository/user"
	"example.com/gin-layout/internal/router"
	orderservice "example.com/gin-layout/internal/service/order"
	userservice "example.com/gin-layout/internal/service/user"
)

type App struct {
	server *http.Server
}

func New(cfg config.Config) (*App, error) {
	users := userservice.New(userrepo.New())
	// 用户 Service 满足订单用例所需的 UserReader 接口。
	orders := orderservice.New(orderrepo.New(), users)
	engine, err := router.New(userhandler.New(users), orderhandler.New(orders))
	if err != nil {
		return nil, fmt.Errorf("create router: %w", err)
	}
	return &App{server: &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}}, nil
}

func (a *App) Run(ctx context.Context) error {
	stopped := make(chan error, 1)
	go func() { stopped <- a.server.ListenAndServe() }()
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
