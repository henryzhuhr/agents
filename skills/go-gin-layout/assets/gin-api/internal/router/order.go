package router

import (
	orderhandler "example.com/gin-layout/internal/handler/order"
	"example.com/gin-layout/internal/svc"

	"github.com/gin-gonic/gin"
)

func registerOrderRoutes(api *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	orders := api.Group("/orders")
	orders.POST("", orderhandler.NewCreateHandler(svcCtx))
}
