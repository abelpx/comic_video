package gateway

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*ClientLimiter
	config   *RateLimitConfig
	cleanup  *time.Ticker
}

// ClientLimiter 客户端限制器
type ClientLimiter struct {
	tokens    int
	lastRefill time.Time
	requests  []time.Time
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	RequestsPerMinute int           `json:"requests_per_minute"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter() *RateLimiter {
	config := &RateLimitConfig{
		RequestsPerMinute: 60,  // 每分钟60个请求
		BurstSize:         10,  // 突发10个请求
		WindowSize:        1 * time.Minute,
		CleanupInterval:   5 * time.Minute,
	}

	rl := &RateLimiter{
		clients: make(map[string]*ClientLimiter),
		config:  config,
		cleanup: time.NewTicker(config.CleanupInterval),
	}

	// 启动清理协程
	go rl.cleanupLoop()

	return rl
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	
	// 获取或创建客户端限制器
	client, exists := rl.clients[clientID]
	if !exists {
		client = &ClientLimiter{
			tokens:     rl.config.BurstSize,
			lastRefill: now,
			requests:   make([]time.Time, 0),
		}
		rl.clients[clientID] = client
	}

	// 令牌桶算法
	elapsed := now.Sub(client.lastRefill)
	tokensToAdd := int(elapsed.Minutes() * float64(rl.config.RequestsPerMinute))
	if tokensToAdd > 0 {
		client.tokens += tokensToAdd
		if client.tokens > rl.config.BurstSize {
			client.tokens = rl.config.BurstSize
		}
		client.lastRefill = now
	}

	// 检查令牌
	if client.tokens > 0 {
		client.tokens--
		client.requests = append(client.requests, now)
		return true
	}

	return false
}

// GetStats 获取统计信息
func (rl *RateLimiter) GetStats(clientID string) map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	client, exists := rl.clients[clientID]
	if !exists {
		return map[string]interface{}{
			"tokens":         rl.config.BurstSize,
			"requests_count": 0,
			"last_request":   nil,
		}
	}

	// 清理过期请求
	now := time.Now()
	cutoff := now.Add(-rl.config.WindowSize)
	validRequests := make([]time.Time, 0)
	for _, req := range client.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	client.requests = validRequests

	var lastRequest *time.Time
	if len(client.requests) > 0 {
		lastRequest = &client.requests[len(client.requests)-1]
	}

	return map[string]interface{}{
		"tokens":         client.tokens,
		"requests_count": len(client.requests),
		"last_request":   lastRequest,
	}
}

// cleanupLoop 清理循环
func (rl *RateLimiter) cleanupLoop() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.config.WindowSize * 2) // 保留2个窗口的数据

		for clientID, client := range rl.clients {
			// 清理过期请求
			validRequests := make([]time.Time, 0)
			for _, req := range client.requests {
				if req.After(cutoff) {
					validRequests = append(validRequests, req)
				}
			}
			client.requests = validRequests

			// 删除长时间未活动的客户端
			if len(client.requests) == 0 && now.Sub(client.lastRefill) > rl.config.WindowSize*2 {
				delete(rl.clients, clientID)
			}
		}
		rl.mu.Unlock()
	}
}

// Stop 停止速率限制器
func (rl *RateLimiter) Stop() {
	rl.cleanup.Stop()
}

// AuthManager 认证管理器
type AuthManager struct {
	mu       sync.RWMutex
	tokens   map[string]*AuthToken
	users    map[string]*User
	config   *AuthConfig
}

// AuthToken 认证令牌
type AuthToken struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
	Scopes    []string  `json:"scopes"`
}

// User 用户
type User struct {
	ID          string            `json:"id"`
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	Role        string            `json:"role"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	LastLogin   time.Time         `json:"last_login"`
	Active      bool              `json:"active"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	TokenExpiry     time.Duration `json:"token_expiry"`
	RefreshExpiry   time.Duration `json:"refresh_expiry"`
	SecretKey       string        `json:"secret_key"`
	RequireAuth     bool          `json:"require_auth"`
	DefaultRole     string        `json:"default_role"`
}

// NewAuthManager 创建认证管理器
func NewAuthManager() *AuthManager {
	config := &AuthConfig{
		TokenExpiry:   24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
		SecretKey:     generateSecretKey(),
		RequireAuth:   false, // 开发环境默认不需要认证
		DefaultRole:   "user",
	}

	am := &AuthManager{
		tokens: make(map[string]*AuthToken),
		users:  make(map[string]*User),
		config: config,
	}

	// 创建默认用户
	am.createDefaultUsers()

	return am
}

// ValidateToken 验证令牌
func (am *AuthManager) ValidateToken(tokenStr string) (string, error) {
	if !am.config.RequireAuth {
		return "default_user", nil // 开发模式
	}

	// 移除 "Bearer " 前缀
	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}

	am.mu.RLock()
	token, exists := am.tokens[tokenStr]
	am.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("令牌不存在")
	}

	if time.Now().After(token.ExpiresAt) {
		// 删除过期令牌
		am.mu.Lock()
		delete(am.tokens, tokenStr)
		am.mu.Unlock()
		return "", fmt.Errorf("令牌已过期")
	}

	// 更新最后使用时间
	am.mu.Lock()
	token.LastUsed = time.Now()
	am.mu.Unlock()

	return token.UserID, nil
}

// CreateToken 创建令牌
func (am *AuthManager) CreateToken(userID string, scopes []string) (*AuthToken, error) {
	am.mu.RLock()
	user, exists := am.users[userID]
	am.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("用户不存在")
	}

	if !user.Active {
		return nil, fmt.Errorf("用户已禁用")
	}

	tokenStr := generateToken()
	token := &AuthToken{
		Token:     tokenStr,
		UserID:    userID,
		ExpiresAt: time.Now().Add(am.config.TokenExpiry),
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		Scopes:    scopes,
	}

	am.mu.Lock()
	am.tokens[tokenStr] = token
	user.LastLogin = time.Now()
	am.mu.Unlock()

	return token, nil
}

// RevokeToken 撤销令牌
func (am *AuthManager) RevokeToken(tokenStr string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.tokens[tokenStr]; !exists {
		return fmt.Errorf("令牌不存在")
	}

	delete(am.tokens, tokenStr)
	return nil
}

// CreateUser 创建用户
func (am *AuthManager) CreateUser(username, email, role string) (*User, error) {
	userID := uuid.New().String()
	
	user := &User{
		ID:          userID,
		Username:    username,
		Email:       email,
		Role:        role,
		Permissions: am.getRolePermissions(role),
		Metadata:    make(map[string]string),
		CreatedAt:   time.Now(),
		Active:      true,
	}

	am.mu.Lock()
	am.users[userID] = user
	am.mu.Unlock()

	return user, nil
}

// GetUser 获取用户
func (am *AuthManager) GetUser(userID string) (*User, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	user, exists := am.users[userID]
	if !exists {
		return nil, fmt.Errorf("用户不存在")
	}

	// 返回副本
	userCopy := *user
	return &userCopy, nil
}

// HasPermission 检查权限
func (am *AuthManager) HasPermission(userID, permission string) bool {
	am.mu.RLock()
	defer am.mu.RUnlock()

	user, exists := am.users[userID]
	if !exists {
		return false
	}

	for _, perm := range user.Permissions {
		if perm == permission || perm == "*" {
			return true
		}
	}

	return false
}

// CleanupExpiredTokens 清理过期令牌
func (am *AuthManager) CleanupExpiredTokens() {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	for tokenStr, token := range am.tokens {
		if now.After(token.ExpiresAt) {
			delete(am.tokens, tokenStr)
		}
	}
}

// GetStats 获取认证统计
func (am *AuthManager) GetStats() map[string]interface{} {
	am.mu.RLock()
	defer am.mu.RUnlock()

	activeTokens := 0
	expiredTokens := 0
	now := time.Now()

	for _, token := range am.tokens {
		if now.After(token.ExpiresAt) {
			expiredTokens++
		} else {
			activeTokens++
		}
	}

	activeUsers := 0
	for _, user := range am.users {
		if user.Active {
			activeUsers++
		}
	}

	return map[string]interface{}{
		"total_users":    len(am.users),
		"active_users":   activeUsers,
		"active_tokens":  activeTokens,
		"expired_tokens": expiredTokens,
		"auth_required":  am.config.RequireAuth,
	}
}

// 私有方法
func (am *AuthManager) createDefaultUsers() {
	// 创建管理员用户
	admin := &User{
		ID:          "admin",
		Username:    "admin",
		Email:       "admin@example.com",
		Role:        "admin",
		Permissions: []string{"*"}, // 所有权限
		Metadata:    make(map[string]string),
		CreatedAt:   time.Now(),
		Active:      true,
	}
	am.users["admin"] = admin

	// 创建默认用户
	defaultUser := &User{
		ID:          "default_user",
		Username:    "default",
		Email:       "default@example.com",
		Role:        "user",
		Permissions: am.getRolePermissions("user"),
		Metadata:    make(map[string]string),
		CreatedAt:   time.Now(),
		Active:      true,
	}
	am.users["default_user"] = defaultUser
}

func (am *AuthManager) getRolePermissions(role string) []string {
	permissions := map[string][]string{
		"admin": {"*"},
		"user": {
			"project:create",
			"project:read",
			"project:update",
			"workflow:execute",
			"quality:check",
			"monitoring:read",
		},
		"viewer": {
			"project:read",
			"monitoring:read",
		},
	}

	if perms, exists := permissions[role]; exists {
		return perms
	}

	return permissions["user"] // 默认用户权限
}

// 工具函数
func generateSecretKey() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("[AuthManager] 生成密钥失败: %v", err)
		return "default-secret-key-for-development"
	}
	return hex.EncodeToString(bytes)
}

func generateToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("[AuthManager] 生成令牌失败: %v", err)
		return fmt.Sprintf("token-%d", time.Now().Unix())
	}
	return hex.EncodeToString(bytes)
}

// WebSocketManager WebSocket管理器
type WebSocketManager struct {
	mu          sync.RWMutex
	connections map[string]*WebSocketConnection
	channels    map[string][]string // projectID -> connectionIDs
}

// WebSocketConnection WebSocket连接
type WebSocketConnection struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ProjectID string    `json:"project_id"`
	ConnectedAt time.Time `json:"connected_at"`
	LastPing  time.Time `json:"last_ping"`
	// 这里应该包含实际的WebSocket连接对象
}

// NewWebSocketManager 创建WebSocket管理器
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		connections: make(map[string]*WebSocketConnection),
		channels:    make(map[string][]string),
	}
}

// AddConnection 添加连接
func (wsm *WebSocketManager) AddConnection(userID, projectID string) string {
	connectionID := uuid.New().String()
	
	conn := &WebSocketConnection{
		ID:          connectionID,
		UserID:      userID,
		ProjectID:   projectID,
		ConnectedAt: time.Now(),
		LastPing:    time.Now(),
	}

	wsm.mu.Lock()
	wsm.connections[connectionID] = conn
	
	// 添加到项目频道
	if _, exists := wsm.channels[projectID]; !exists {
		wsm.channels[projectID] = make([]string, 0)
	}
	wsm.channels[projectID] = append(wsm.channels[projectID], connectionID)
	wsm.mu.Unlock()

	return connectionID
}

// RemoveConnection 移除连接
func (wsm *WebSocketManager) RemoveConnection(connectionID string) {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	conn, exists := wsm.connections[connectionID]
	if !exists {
		return
	}

	// 从项目频道移除
	if connections, exists := wsm.channels[conn.ProjectID]; exists {
		filtered := make([]string, 0)
		for _, id := range connections {
			if id != connectionID {
				filtered = append(filtered, id)
			}
		}
		wsm.channels[conn.ProjectID] = filtered
	}

	delete(wsm.connections, connectionID)
}

// BroadcastToProject 向项目广播消息
func (wsm *WebSocketManager) BroadcastToProject(projectID string, message interface{}) {
	wsm.mu.RLock()
	connectionIDs, exists := wsm.channels[projectID]
	wsm.mu.RUnlock()

	if !exists {
		return
	}

	for _, connectionID := range connectionIDs {
		// 这里应该发送实际的WebSocket消息
		log.Printf("[WebSocketManager] 发送消息到连接 %s: %v", connectionID, message)
	}
}

// GetStats 获取WebSocket统计
func (wsm *WebSocketManager) GetStats() map[string]interface{} {
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()

	return map[string]interface{}{
		"total_connections": len(wsm.connections),
		"active_channels":   len(wsm.channels),
	}
}
