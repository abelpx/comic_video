package middleware

import (
	"net/http"
	"strings"

	"comic_video/internal/service/auth"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证格式",
			})
			c.Abort()
			return
		}

		// 提取令牌
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证令牌
		userID, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "令牌无效或已过期",
			})
			c.Abort()
			return
		}

		// 将用户ID存储到上下文中
		c.Set("user_id", userID)
		c.Next()
	}
}

// OptionalAuthMiddleware 可选的JWT认证中间件
func OptionalAuthMiddleware(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		// 提取令牌
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证令牌
		userID, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.Next()
			return
		}

		// 将用户ID存储到上下文中
		c.Set("user_id", userID)
		c.Next()
	}
}

// RoleMiddleware 角色权限中间件
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		// 检查用户角色是否在允许的角色列表中
		hasRole := false
		for _, role := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
} 