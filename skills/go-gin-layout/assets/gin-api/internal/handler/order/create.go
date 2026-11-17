package order

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"example.com/gin-layout/internal/model"
	orderservice "example.com/gin-layout/internal/service/order"
)

type Handler struct {
	orders *orderservice.Service
}

func New(orders *orderservice.Service) *Handler {
	return &Handler{orders: orders}
}

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

func (h *Handler) Create(c *gin.Context) {
	var request createRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order request"})
		return
	}
	items := make([]model.OrderItem, len(request.Items))
	for i, item := range request.Items {
		items[i] = model.OrderItem{SKU: item.SKU, Quantity: item.Quantity}
	}
	order, err := h.orders.Create(c.Request.Context(), request.UserID, items)
	if err != nil {
		switch {
		case errors.Is(err, orderservice.ErrInvalidOrder):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order"})
		case errors.Is(err, model.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	response := createResponse{ID: order.ID, UserID: order.UserID, Items: make([]itemResponse, len(order.Items))}
	for i, item := range order.Items {
		response.Items[i] = itemResponse{SKU: item.SKU, Quantity: item.Quantity}
	}
	c.JSON(http.StatusCreated, response)
}
