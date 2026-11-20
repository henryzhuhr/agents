package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"example.com/gin-layout/internal/entity"
	"example.com/gin-layout/internal/middleware"
	userservice "example.com/gin-layout/internal/service/user"
	"example.com/gin-layout/internal/svc"
)

// profileResponse 属于 HTTP 边界，业务实体无需依赖这些 JSON 字段。
type profileResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func NewProfileHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}
		profileService := userservice.NewProfileService(ctx, svcCtx)
		user, err := profileService.Get(id)
		if err != nil {
			switch {
			case errors.Is(err, userservice.ErrInvalidID):
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			case errors.Is(err, entity.ErrUserNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			default:
				svcCtx.Logger.ErrorContext(ctx, "load user profile",
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
		c.JSON(http.StatusOK, profileResponse{ID: user.ID, Name: user.Name})
	}
}
