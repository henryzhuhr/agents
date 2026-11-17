package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"example.com/gin-layout/internal/model"
	userservice "example.com/gin-layout/internal/service/user"
)

type Handler struct {
	users *userservice.Service
}

func New(users *userservice.Service) *Handler {
	return &Handler{users: users}
}

// profileResponse 属于 HTTP 边界，业务实体无需依赖这些 JSON 字段。
type profileResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	user, err := h.users.Get(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, userservice.ErrInvalidID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		case errors.Is(err, model.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.JSON(http.StatusOK, profileResponse{ID: user.ID, Name: user.Name})
}
