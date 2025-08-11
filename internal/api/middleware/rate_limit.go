package middleware

import (
	"fmt"
	"strconv"
	"time"

	"comic_video/internal/repository/redis"
	"comic_video/internal/utils"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 每个时间窗口允许的请求数
	Requests int
	// 时间窗口大小
	Window time.Duration
	// 限流键生成函数
	KeyGenerator func(*gin.Context) string
	// 限流触发时的响应
	OnLimitReached func(*gin.Context)
}

// DefaultRateLimitConfig 默认限流配置
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Requests: 100,
		Window:   time.Minute,
		KeyGenerator: func(c *gin.Context) string {
			return c.ClientIP()
		},
		OnLimitReached: func(c *gin.Context) {
			utils.AbortWithError(c, utils.NewAppError(
				utils.ErrCodeTooManyRequests,
				"Too many requests",
				"Rate limit exceeded, please try again later",
			))
		},
	}
}

// RateLimitMiddleware 创建限流中间件
func RateLimitMiddleware(redisClient *redis.Client, config *RateLimitConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	return func(c *gin.Context) {
		// 生成限流键
		key := fmt.Sprintf("rate_limit:%s", config.KeyGenerator(c))
		
		// 获取当前时间窗口
		now := time.Now()
		window := now.Truncate(config.Window).Unix()
		windowKey := fmt.Sprintf("%s:%d", key, window)

		// 检查当前请求数
		current, err := redisClient.Get(c.Request.Context(), windowKey)
		if err != nil && err.Error() != "redis: nil" {
			// Redis错误，记录日志但不阻止请求
			utils.LogError(c.Request.Context(), err, "Failed to get rate limit counter", map[string]interface{}{
				"key": windowKey,
			})
			c.Next()
			return
		}

		var currentCount int
		if current != "" {
			currentCount, _ = strconv.Atoi(current)
		}

		// 检查是否超过限制
		if currentCount >= config.Requests {
			// 设置响应头
			c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(window+int64(config.Window.Seconds()), 10))
			
			// 触发限流响应
			config.OnLimitReached(c)
			return
		}

		// 增加计数器
		pipe := redisClient.Pipeline()
		pipe.Incr(c.Request.Context(), windowKey)
		pipe.Expire(c.Request.Context(), windowKey, config.Window)
		_, err = pipe.Exec(c.Request.Context())
		
		if err != nil {
			utils.LogError(c.Request.Context(), err, "Failed to update rate limit counter", map[string]interface{}{
				"key": windowKey,
			})
		}

		// 设置响应头
		remaining := config.Requests - currentCount - 1
		if remaining < 0 {
			remaining = 0
		}
		
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(window+int64(config.Window.Seconds()), 10))

		c.Next()
	}
}

// IPRateLimitMiddleware 基于IP的限流中间件
func IPRateLimitMiddleware(redisClient *redis.Client, requests int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyGenerator: func(c *gin.Context) string {
			return c.ClientIP()
		},
		OnLimitReached: func(c *gin.Context) {
			utils.AbortWithError(c, utils.NewAppError(
				utils.ErrCodeTooManyRequests,
				"Too many requests from this IP",
				fmt.Sprintf("Rate limit: %d requests per %v", requests, window),
			))
		},
	}
	return RateLimitMiddleware(redisClient, config)
}

// UserRateLimitMiddleware 基于用户的限流中间件
func UserRateLimitMiddleware(redisClient *redis.Client, requests int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyGenerator: func(c *gin.Context) string {
			// 优先使用用户ID，如果没有则使用IP
			if userID := utils.GetUserID(c); userID != "" {
				return fmt.Sprintf("user:%s", userID)
			}
			return fmt.Sprintf("ip:%s", c.ClientIP())
		},
		OnLimitReached: func(c *gin.Context) {
			utils.AbortWithError(c, utils.NewAppError(
				utils.ErrCodeTooManyRequests,
				"Too many requests",
				fmt.Sprintf("Rate limit: %d requests per %v", requests, window),
			))
		},
	}
	return RateLimitMiddleware(redisClient, config)
}

// APIKeyRateLimitMiddleware 基于API Key的限流中间件
func APIKeyRateLimitMiddleware(redisClient *redis.Client, requests int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyGenerator: func(c *gin.Context) string {
			// 从Header获取API Key
			apiKey := c.GetHeader("X-API-Key")
			if apiKey == "" {
				apiKey = c.GetHeader("Authorization")
			}
			if apiKey == "" {
				return fmt.Sprintf("ip:%s", c.ClientIP())
			}
			return fmt.Sprintf("apikey:%s", apiKey)
		},
		OnLimitReached: func(c *gin.Context) {
			utils.AbortWithError(c, utils.NewAppError(
				utils.ErrCodeTooManyRequests,
				"API rate limit exceeded",
				fmt.Sprintf("Rate limit: %d requests per %v", requests, window),
			))
		},
	}
	return RateLimitMiddleware(redisClient, config)
}

// EndpointRateLimitMiddleware 基于端点的限流中间件
func EndpointRateLimitMiddleware(redisClient *redis.Client, requests int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyGenerator: func(c *gin.Context) string {
			// 使用IP + 端点路径作为键
			return fmt.Sprintf("endpoint:%s:%s", c.ClientIP(), c.FullPath())
		},
		OnLimitReached: func(c *gin.Context) {
			utils.AbortWithError(c, utils.NewAppError(
				utils.ErrCodeTooManyRequests,
				"Endpoint rate limit exceeded",
				fmt.Sprintf("Rate limit for %s: %d requests per %v", c.FullPath(), requests, window),
			))
		},
	}
	return RateLimitMiddleware(redisClient, config)
}

// SlidingWindowRateLimitMiddleware 滑动窗口限流中间件
func SlidingWindowRateLimitMiddleware(redisClient *redis.Client, requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("sliding_rate_limit:%s", c.ClientIP())
		now := time.Now().Unix()
		
		// 使用Lua脚本实现滑动窗口限流
		luaScript := `
			local key = KEYS[1]
			local window = tonumber(ARGV[1])
			local limit = tonumber(ARGV[2])
			local now = tonumber(ARGV[3])
			
			-- 清理过期的记录
			redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
			
			-- 获取当前窗口内的请求数
			local current = redis.call('ZCARD', key)
			
			if current < limit then
				-- 添加当前请求
				redis.call('ZADD', key, now, now)
				redis.call('EXPIRE', key, window)
				return {1, limit - current - 1}
			else
				return {0, 0}
			end
		`
		
		result, err := redisClient.Eval(c.Request.Context(), luaScript, []string{key},
			int64(window.Seconds()), requests, now)
		
		if err != nil {
			utils.LogError(c.Request.Context(), err, "Failed to execute sliding window rate limit", map[string]interface{}{
				"key": key,
			})
			c.Next()
			return
		}
		
		resultSlice := result.([]interface{})
		allowed := resultSlice[0].(int64)
		remaining := resultSlice[1].(int64)
		
		// 设置响应头
		c.Header("X-RateLimit-Limit", strconv.Itoa(requests))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Header("X-RateLimit-Window", window.String())
		
		if allowed == 0 {
			utils.AbortWithError(c, utils.NewAppError(
				utils.ErrCodeTooManyRequests,
				"Rate limit exceeded",
				fmt.Sprintf("Sliding window rate limit: %d requests per %v", requests, window),
			))
			return
		}
		
		c.Next()
	}
}

// RateLimitStatus 获取限流状态
type RateLimitStatus struct {
	Limit     int   `json:"limit"`
	Remaining int   `json:"remaining"`
	Reset     int64 `json:"reset"`
	Window    int64 `json:"window"`
}

// GetRateLimitStatus 获取指定键的限流状态
func GetRateLimitStatus(redisClient *redis.Client, key string, config *RateLimitConfig) (*RateLimitStatus, error) {
	now := time.Now()
	window := now.Truncate(config.Window).Unix()
	windowKey := fmt.Sprintf("rate_limit:%s:%d", key, window)
	
	current, err := redisClient.Get(nil, windowKey)
	if err != nil && err.Error() != "redis: nil" {
		return nil, err
	}
	
	var currentCount int
	if current != "" {
		currentCount, _ = strconv.Atoi(current)
	}
	
	remaining := config.Requests - currentCount
	if remaining < 0 {
		remaining = 0
	}
	
	return &RateLimitStatus{
		Limit:     config.Requests,
		Remaining: remaining,
		Reset:     window + int64(config.Window.Seconds()),
		Window:    int64(config.Window.Seconds()),
	}, nil
}
