package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader 请求ID的Header名称
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey 在Gin上下文中存储请求ID的键
	RequestIDKey = "request_id"
)

// RequestIDMiddleware 请求ID中间件
// 为每个请求生成唯一的ID，用于日志追踪和问题排查
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取请求ID
		requestID := c.GetHeader(RequestIDHeader)
		
		// 如果请求头中没有请求ID，则生成一个新的
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// 将请求ID存储到上下文中
		c.Set(RequestIDKey, requestID)
		
		// 将请求ID添加到响应头中
		c.Header(RequestIDHeader, requestID)
		
		// 继续处理请求
		c.Next()
	}
}

// GetRequestID 从Gin上下文中获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		return requestID.(string)
	}
	return ""
}
