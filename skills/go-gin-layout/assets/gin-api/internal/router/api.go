package router

import (
	"net/http"

	orderhandler "example.com/gin-layout/internal/handler/order"
	userhandler "example.com/gin-layout/internal/handler/user"
	"github.com/gin-gonic/gin"
)

func New(users *userhandler.Handler, orders *orderhandler.Handler) (*gin.Engine, error) {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())
	if err := engine.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	api := engine.Group("/api/v1")
	api.GET("/users/:id", users.Get)
	api.POST("/orders", orders.Create)
	return engine, nil
}
