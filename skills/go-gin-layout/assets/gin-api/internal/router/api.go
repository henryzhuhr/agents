package router

import (
	"net/http"

	"example.com/gin-layout/internal/middleware"
	"example.com/gin-layout/internal/svc"

	"github.com/gin-gonic/gin"
)

func New(svcCtx *svc.ServiceContext) (*gin.Engine, error) {
	engine := gin.New()
	engine.Use(middleware.RequestID(), gin.Logger(), gin.Recovery())
	if err := engine.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	baseAPI := engine.Group("/api/v1")
	registerUserRoutes(baseAPI, svcCtx)
	registerOrderRoutes(baseAPI, svcCtx)
	return engine, nil
}
