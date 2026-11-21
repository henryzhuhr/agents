package router

import (
	"example.com/gin-layout/internal/middleware"
	"example.com/gin-layout/internal/svc"

	"github.com/gin-gonic/gin"
)

func New(svcCtx *svc.ServiceContext) (*gin.Engine, error) {
	engine := gin.New()
	engine.RedirectTrailingSlash = false
	engine.Use(
		middleware.RequestID(),
		middleware.AccessLog(svcCtx.Logger),
		middleware.Recovery(svcCtx.Logger),
	)
	if err := engine.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	registerHealthRoutes(engine, svcCtx)

	baseAPI := engine.Group("/api/v1")
	registerUserRoutes(baseAPI, svcCtx)
	registerOrderRoutes(baseAPI, svcCtx)
	return engine, nil
}
