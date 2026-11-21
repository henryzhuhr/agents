package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery 捕获 panic，使用结构化日志记录并返回统一的 JSON 500。
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	// 关闭 Gin 自带的 panic 日志，统一由结构化回调记录。
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		attrs := requestAttrs(c, http.StatusInternalServerError)
		attrs = append(attrs,
			"error", fmt.Sprint(recovered),
			"stack", string(debug.Stack()),
		)
		logger.ErrorContext(c.Request.Context(), "panic recovered", attrs...)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	})
}
