package order

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"example.com/gin-layout/internal/entity"
	"example.com/gin-layout/internal/middleware"
	orderservice "example.com/gin-layout/internal/service/order"
	"example.com/gin-layout/internal/svc"
)

// 请求与响应类型留在对应功能文件，不与业务实体共用绑定标签。
type createRequest struct {
	UserID uint64              `json:"user_id" binding:"required,gt=0"`
	Items  []createItemRequest `json:"items" binding:"required,min=1,dive"`
}

type createItemRequest struct {
	SKU      string `json:"sku" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,gt=0"`
}

type createResponse struct {
	ID     uint64         `json:"id"`
	UserID uint64         `json:"user_id"`
	Items  []itemResponse `json:"items"`
}

type itemResponse struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

// orderItems 把 HTTP 请求中的明细转成业务实体，保持 HTTP 类型不进入 Service。
func (r createRequest) orderItems() []entity.OrderItem {
	items := make([]entity.OrderItem, len(r.Items))
	for i, item := range r.Items {
		items[i] = entity.OrderItem{SKU: item.SKU, Quantity: item.Quantity}
	}
	return items
}

// newCreateResponse 把业务实体转成 HTTP 响应。
func newCreateResponse(order entity.Order) createResponse {
	items := make([]itemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = itemResponse{SKU: item.SKU, Quantity: item.Quantity}
	}
	return createResponse{ID: order.ID, UserID: order.UserID, Items: items}
}

func NewCreateHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var request createRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order request"})
			return
		}
		createService := orderservice.NewCreateService(ctx, svcCtx)
		order, err := createService.Create(request.UserID, request.orderItems())
		if err != nil {
			switch {
			case errors.Is(err, orderservice.ErrInvalidOrder):
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order"})
			case errors.Is(err, entity.ErrUserNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			default:
				svcCtx.Logger.ErrorContext(ctx, "create order",
					"request_id", middleware.RequestIDFromGin(c),
					"method", c.Request.Method,
					"path", c.FullPath(),
					"status", http.StatusInternalServerError,
					"error", err,
				)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			return
		}
		c.JSON(http.StatusCreated, newCreateResponse(order))
	}
}
