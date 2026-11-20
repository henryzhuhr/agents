package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

var skippedHealthPaths = map[string]struct{}{
	"/healthz":  {},
	"/livez":    {},
	"/readyz":   {},
	"/startupz": {},
}

// AccessLog 使用 slog 记录 HTTP 访问日志，不记录 query，也跳过健康检查端点。
// 应在 Recovery 之前注册。这样下游 handler 发生 panic 时，内层 Recovery 会先捕获
// panic 并处理响应，随后 c.Next() 正常返回，访问日志才能读到处理后的状态码。
// 反过来注册时，panic 会先跳过 c.Next() 后的日志代码，再由外层 Recovery 捕获。
func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// c.Next() 执行后续中间件和路由 handler。返回后再读取状态码并计算耗时。
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		if _, skip := skippedHealthPaths[path]; skip {
			return
		}
		attrs := requestAttrs(c, c.Writer.Status())
		attrs = append(attrs, "latency_ms", time.Since(start).Milliseconds())
		logger.InfoContext(c.Request.Context(), "http request", attrs...)
	}
}

func requestAttrs(c *gin.Context, status int) []any {
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	attrs := []any{
		"method", c.Request.Method,
		"path", path,
		"status", status,
	}
	if id := RequestIDFromGin(c); id != "" {
		attrs = append(attrs, "request_id", id)
	}
	return attrs
}
