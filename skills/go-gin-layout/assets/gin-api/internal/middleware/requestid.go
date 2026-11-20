// Package middleware 提供 Gin HTTP 中间件。
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestIDHeader 是请求和响应中传递请求 ID 的头部。
const RequestIDHeader = "X-Request-ID"

const requestIDKey = "request_id"

var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// RequestID 透传合法的入站请求 ID，否则生成新值；结果写入响应头和 gin.Context。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if !validRequestID.MatchString(id) {
			id = newRequestID()
		}
		c.Set(requestIDKey, id)
		c.Header(RequestIDHeader, id)
		c.Next()
	}
}

// RequestIDFromGin 返回当前请求的 ID；未启用中间件时返回空字符串。
func RequestIDFromGin(c *gin.Context) string {
	return c.GetString(requestIDKey)
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	// crypto/rand 失败时仍保证日志可关联，退化到时间戳。
	return strconv.FormatInt(time.Now().UnixNano(), 16)
}
