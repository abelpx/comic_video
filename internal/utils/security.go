package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	// JWT配置
	JWTSecret           string        `json:"jwt_secret" yaml:"jwt_secret"`
	JWTExpiration       time.Duration `json:"jwt_expiration" yaml:"jwt_expiration"`
	JWTRefreshExpiration time.Duration `json:"jwt_refresh_expiration" yaml:"jwt_refresh_expiration"`
	
	// 密码策略
	MinPasswordLength    int  `json:"min_password_length" yaml:"min_password_length"`
	RequireUppercase     bool `json:"require_uppercase" yaml:"require_uppercase"`
	RequireLowercase     bool `json:"require_lowercase" yaml:"require_lowercase"`
	RequireNumbers       bool `json:"require_numbers" yaml:"require_numbers"`
	RequireSpecialChars  bool `json:"require_special_chars" yaml:"require_special_chars"`
	
	// 账户安全
	MaxLoginAttempts     int           `json:"max_login_attempts" yaml:"max_login_attempts"`
	LockoutDuration      time.Duration `json:"lockout_duration" yaml:"lockout_duration"`
	SessionTimeout       time.Duration `json:"session_timeout" yaml:"session_timeout"`
	
	// CORS配置
	AllowedOrigins       []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods       []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders       []string `json:"allowed_headers" yaml:"allowed_headers"`
	AllowCredentials     bool     `json:"allow_credentials" yaml:"allow_credentials"`
	
	// 其他安全设置
	EnableHTTPS          bool     `json:"enable_https" yaml:"enable_https"`
	TrustedProxies       []string `json:"trusted_proxies" yaml:"trusted_proxies"`
	MaxRequestSize       int64    `json:"max_request_size" yaml:"max_request_size"`
}

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		JWTExpiration:        24 * time.Hour,
		JWTRefreshExpiration: 7 * 24 * time.Hour,
		MinPasswordLength:    8,
		RequireUppercase:     true,
		RequireLowercase:     true,
		RequireNumbers:       true,
		RequireSpecialChars:  false,
		MaxLoginAttempts:     5,
		LockoutDuration:      15 * time.Minute,
		SessionTimeout:       30 * time.Minute,
		AllowedOrigins:       []string{"http://localhost:3000"},
		AllowedMethods:       []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:       []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		AllowCredentials:     true,
		EnableHTTPS:          false,
		TrustedProxies:       []string{"127.0.0.1"},
		MaxRequestSize:       10 << 20, // 10MB
	}
}

// PasswordValidator 密码验证器
type PasswordValidator struct {
	config *SecurityConfig
}

// NewPasswordValidator 创建密码验证器
func NewPasswordValidator(config *SecurityConfig) *PasswordValidator {
	return &PasswordValidator{config: config}
}

// ValidatePassword 验证密码强度
func (pv *PasswordValidator) ValidatePassword(password string) error {
	if len(password) < pv.config.MinPasswordLength {
		return NewAppError(ErrCodeInvalidParams, 
			fmt.Sprintf("Password must be at least %d characters long", pv.config.MinPasswordLength))
	}
	
	if pv.config.RequireUppercase && !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return NewAppError(ErrCodeInvalidParams, "Password must contain at least one uppercase letter")
	}
	
	if pv.config.RequireLowercase && !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return NewAppError(ErrCodeInvalidParams, "Password must contain at least one lowercase letter")
	}
	
	if pv.config.RequireNumbers && !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return NewAppError(ErrCodeInvalidParams, "Password must contain at least one number")
	}
	
	if pv.config.RequireSpecialChars && !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		return NewAppError(ErrCodeInvalidParams, "Password must contain at least one special character")
	}
	
	return nil
}

// HashPassword 哈希密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", WrapError(ErrCodeInternalError, "Failed to hash password", err)
	}
	return string(bytes), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateSecureToken 生成安全令牌
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", WrapError(ErrCodeInternalError, "Failed to generate secure token", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SanitizeInput 清理输入数据
func SanitizeInput(input string) string {
	// 移除潜在的XSS攻击字符
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#x27;")
	input = strings.ReplaceAll(input, "&", "&amp;")
	
	// 移除SQL注入相关字符
	sqlKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER",
		"UNION", "OR", "AND", "WHERE", "FROM", "JOIN", "HAVING", "GROUP BY",
		"ORDER BY", "LIMIT", "OFFSET", "--", "/*", "*/", ";",
	}
	
	upperInput := strings.ToUpper(input)
	for _, keyword := range sqlKeywords {
		if strings.Contains(upperInput, keyword) {
			input = strings.ReplaceAll(input, keyword, "")
			input = strings.ReplaceAll(input, strings.ToLower(keyword), "")
		}
	}
	
	return strings.TrimSpace(input)
}

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateUsername 验证用户名格式
func ValidateUsername(username string) bool {
	// 用户名只能包含字母、数字、下划线和连字符，长度3-20
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)
	return usernameRegex.MatchString(username)
}

// SecureCompare 安全比较字符串（防止时序攻击）
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// SecurityHeadersMiddleware 安全头中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止XSS攻击
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// 防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")
		
		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")
		
		// 强制HTTPS（如果启用）
		if c.Request.Header.Get("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		// 内容安全策略
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		
		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		c.Next()
	}
}

// RequestSizeLimitMiddleware 请求大小限制中间件
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			AbortWithError(c, NewAppError(
				ErrCodeInvalidParams,
				"Request entity too large",
				fmt.Sprintf("Maximum request size is %d bytes", maxSize),
			))
			return
		}
		
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// IPWhitelistMiddleware IP白名单中间件
func IPWhitelistMiddleware(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		allowed := false
		for _, ip := range allowedIPs {
			if ip == clientIP || ip == "*" {
				allowed = true
				break
			}
		}
		
		if !allowed {
			AbortWithError(c, NewAppError(
				ErrCodeForbidden,
				"Access denied",
				fmt.Sprintf("IP %s is not allowed", clientIP),
			))
			return
		}
		
		c.Next()
	}
}

// CSRFProtectionMiddleware CSRF保护中间件
func CSRFProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对于非安全方法，检查CSRF令牌
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" && c.Request.Method != "OPTIONS" {
			token := c.GetHeader("X-CSRF-Token")
			if token == "" {
				token = c.PostForm("_csrf_token")
			}
			
			// 这里应该验证CSRF令牌的有效性
			// 简化实现，实际应该与会话中的令牌进行比较
			if token == "" {
				AbortWithError(c, NewAppError(
					ErrCodeForbidden,
					"CSRF token missing",
					"CSRF protection requires a valid token",
				))
				return
			}
		}
		
		c.Next()
	}
}

// InputValidationMiddleware 输入验证中间件
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查常见的恶意输入模式
		maliciousPatterns := []string{
			"<script", "javascript:", "vbscript:", "onload=", "onerror=",
			"SELECT.*FROM", "INSERT.*INTO", "UPDATE.*SET", "DELETE.*FROM",
			"UNION.*SELECT", "DROP.*TABLE", "../", "..\\",
		}
		
		// 检查查询参数
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				upperValue := strings.ToUpper(value)
				for _, pattern := range maliciousPatterns {
					if strings.Contains(upperValue, strings.ToUpper(pattern)) {
						AbortWithError(c, NewAppError(
							ErrCodeInvalidParams,
							"Malicious input detected",
							fmt.Sprintf("Parameter %s contains potentially malicious content", key),
						))
						return
					}
				}
			}
		}
		
		c.Next()
	}
}
