package router

import (
	"context"
	"net/http"
	"time"

	"example.com/gin-layout/internal/svc"

	"github.com/gin-gonic/gin"
)

func registerHealthRoutes(engine *gin.Engine, svcCtx *svc.ServiceContext) {
	liveness := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
	// liveness 与 startup 只表示进程已启动；readyz 才检查数据库依赖。
	engine.GET("/healthz", liveness)
	engine.GET("/livez", liveness)
	engine.GET("/startupz", liveness)
	engine.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := svcCtx.Health.Ping(ctx); err != nil {
			svcCtx.Logger.DebugContext(c.Request.Context(), "readiness check failed", "error", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
