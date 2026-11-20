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
	// Gin 的 CustomRecovery 在返回的中间件里调用 c.Next()，执行后续处理链。
	// 这里传入的回调只在 Gin 捕获 panic 后执行，负责记录错误并返回 500。
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		attrs := requestAttrs(c, http.StatusInternalServerError)
		attrs = append(attrs,
			"error", fmt.Sprint(recovered),
			"stack", string(debug.Stack()),
		)
		logger.ErrorContext(c.Request.Context(), "panic recovered", attrs...)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	})
}
