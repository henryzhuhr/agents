package router

import (
	userhandler "example.com/gin-layout/internal/handler/user"
	"example.com/gin-layout/internal/svc"

	"github.com/gin-gonic/gin"
)

func registerUserRoutes(api *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	users := api.Group("/users")
	users.GET("/:id", userhandler.NewProfileHandler(svcCtx))
}
