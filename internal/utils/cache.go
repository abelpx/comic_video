package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"comic_video/internal/repository/redis"
)

// CacheManager 缓存管理器
type CacheManager struct {
	redisClient *redis.Client
	prefix      string
	defaultTTL  time.Duration
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(redisClient *redis.Client, prefix string, defaultTTL time.Duration) *CacheManager {
	return &CacheManager{
		redisClient: redisClient,
		prefix:      prefix,
		defaultTTL:  defaultTTL,
	}
}

// buildKey 构建缓存键
func (cm *CacheManager) buildKey(key string) string {
	if cm.prefix != "" {
		return fmt.Sprintf("%s:%s", cm.prefix, key)
	}
	return key
}

// Set 设置缓存
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, ttl ...time.Duration) error {
	cacheKey := cm.buildKey(key)
	
	// 序列化值
	data, err := json.Marshal(value)
	if err != nil {
		return WrapError(ErrCodeInternalError, "Failed to marshal cache value", err)
	}
	
	// 确定TTL
	expiration := cm.defaultTTL
	if len(ttl) > 0 {
		expiration = ttl[0]
	}
	
	// 设置缓存
	err = cm.redisClient.Set(ctx, cacheKey, string(data), expiration)
	if err != nil {
		return WrapError(ErrCodeInternalError, "Failed to set cache", err)
	}
	
	return nil
}

// Get 获取缓存
func (cm *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	cacheKey := cm.buildKey(key)
	
	// 获取缓存值
	data, err := cm.redisClient.Get(ctx, cacheKey)
	if err != nil {
		if err.Error() == "redis: nil" {
			return NewAppError(ErrCodeNotFound, "Cache key not found")
		}
		return WrapError(ErrCodeInternalError, "Failed to get cache", err)
	}
	
	// 反序列化值
	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return WrapError(ErrCodeInternalError, "Failed to unmarshal cache value", err)
	}
	
	return nil
}

// GetOrSet 获取缓存，如果不存在则设置
func (cm *CacheManager) GetOrSet(ctx context.Context, key string, dest interface{}, setter func() (interface{}, error), ttl ...time.Duration) error {
	// 尝试获取缓存
	err := cm.Get(ctx, key, dest)
	if err == nil {
		return nil // 缓存命中
	}
	
	// 缓存未命中，调用setter函数获取值
	value, err := setter()
	if err != nil {
		return err
	}
	
	// 设置缓存
	err = cm.Set(ctx, key, value, ttl...)
	if err != nil {
		// 设置缓存失败不影响业务逻辑，只记录日志
		LogError(ctx, err, "Failed to set cache", map[string]interface{}{
			"key": key,
		})
	}
	
	// 将值复制到目标变量
	valueBytes, _ := json.Marshal(value)
	return json.Unmarshal(valueBytes, dest)
}

// Delete 删除缓存
func (cm *CacheManager) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	
	// 构建缓存键
	cacheKeys := make([]string, len(keys))
	for i, key := range keys {
		cacheKeys[i] = cm.buildKey(key)
	}
	
	// 删除缓存
	err := cm.redisClient.Del(ctx, cacheKeys...)
	if err != nil {
		return WrapError(ErrCodeInternalError, "Failed to delete cache", err)
	}
	
	return nil
}

// Exists 检查缓存是否存在
func (cm *CacheManager) Exists(ctx context.Context, key string) (bool, error) {
	cacheKey := cm.buildKey(key)
	
	count, err := cm.redisClient.Exists(ctx, cacheKey)
	if err != nil {
		return false, WrapError(ErrCodeInternalError, "Failed to check cache existence", err)
	}
	
	return count > 0, nil
}

// Expire 设置缓存过期时间
func (cm *CacheManager) Expire(ctx context.Context, key string, ttl time.Duration) error {
	cacheKey := cm.buildKey(key)
	
	err := cm.redisClient.Expire(ctx, cacheKey, ttl)
	if err != nil {
		return WrapError(ErrCodeInternalError, "Failed to set cache expiration", err)
	}
	
	return nil
}

// TTL 获取缓存剩余过期时间
func (cm *CacheManager) TTL(ctx context.Context, key string) (time.Duration, error) {
	cacheKey := cm.buildKey(key)
	
	ttl, err := cm.redisClient.TTL(ctx, cacheKey)
	if err != nil {
		return 0, WrapError(ErrCodeInternalError, "Failed to get cache TTL", err)
	}
	
	return ttl, nil
}

// Increment 递增缓存值
func (cm *CacheManager) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	cacheKey := cm.buildKey(key)
	
	result, err := cm.redisClient.IncrBy(ctx, cacheKey, delta)
	if err != nil {
		return 0, WrapError(ErrCodeInternalError, "Failed to increment cache", err)
	}
	
	return result, nil
}

// Decrement 递减缓存值
func (cm *CacheManager) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return cm.Increment(ctx, key, -delta)
}

// SetNX 仅在键不存在时设置
func (cm *CacheManager) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	cacheKey := cm.buildKey(key)
	
	// 序列化值
	data, err := json.Marshal(value)
	if err != nil {
		return false, WrapError(ErrCodeInternalError, "Failed to marshal cache value", err)
	}
	
	// 设置缓存（仅在不存在时）
	success, err := cm.redisClient.SetNX(ctx, cacheKey, string(data), ttl)
	if err != nil {
		return false, WrapError(ErrCodeInternalError, "Failed to set cache with NX", err)
	}
	
	return success, nil
}

// GetMulti 批量获取缓存
func (cm *CacheManager) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}
	
	// 构建缓存键
	cacheKeys := make([]string, len(keys))
	for i, key := range keys {
		cacheKeys[i] = cm.buildKey(key)
	}
	
	// 批量获取
	values, err := cm.redisClient.MGet(ctx, cacheKeys...)
	if err != nil {
		return nil, WrapError(ErrCodeInternalError, "Failed to get multiple cache values", err)
	}
	
	// 构建结果映射
	result := make(map[string]interface{})
	for i, value := range values {
		if value != nil {
			var data interface{}
			if err := json.Unmarshal([]byte(value.(string)), &data); err == nil {
				result[keys[i]] = data
			}
		}
	}
	
	return result, nil
}

// SetMulti 批量设置缓存
func (cm *CacheManager) SetMulti(ctx context.Context, items map[string]interface{}, ttl ...time.Duration) error {
	if len(items) == 0 {
		return nil
	}
	
	// 确定TTL
	expiration := cm.defaultTTL
	if len(ttl) > 0 {
		expiration = ttl[0]
	}
	
	// 使用管道批量设置
	pipe := cm.redisClient.Pipeline()
	
	for key, value := range items {
		cacheKey := cm.buildKey(key)
		
		// 序列化值
		data, err := json.Marshal(value)
		if err != nil {
			return WrapError(ErrCodeInternalError, "Failed to marshal cache value", err)
		}
		
		pipe.Set(ctx, cacheKey, string(data), expiration)
	}
	
	// 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil {
		return WrapError(ErrCodeInternalError, "Failed to set multiple cache values", err)
	}
	
	return nil
}

// Clear 清空指定前缀的所有缓存
func (cm *CacheManager) Clear(ctx context.Context, pattern ...string) error {
	searchPattern := cm.buildKey("*")
	if len(pattern) > 0 {
		searchPattern = cm.buildKey(pattern[0])
	}
	
	// 查找匹配的键
	keys, err := cm.redisClient.Keys(ctx, searchPattern)
	if err != nil {
		return WrapError(ErrCodeInternalError, "Failed to find cache keys", err)
	}
	
	if len(keys) == 0 {
		return nil
	}
	
	// 删除找到的键
	err = cm.redisClient.Del(ctx, keys...)
	if err != nil {
		return WrapError(ErrCodeInternalError, "Failed to clear cache", err)
	}
	
	return nil
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	HitRate     float64 `json:"hit_rate"`
	KeyCount    int64   `json:"key_count"`
	MemoryUsage int64   `json:"memory_usage"`
}

// GetStats 获取缓存统计信息
func (cm *CacheManager) GetStats(ctx context.Context) (*CacheStats, error) {
	// 这里简化实现，实际应该从Redis获取详细统计信息
	info, err := cm.redisClient.Info(ctx, "stats")
	if err != nil {
		return nil, WrapError(ErrCodeInternalError, "Failed to get cache stats", err)
	}
	
	// 解析统计信息（简化实现）
	stats := &CacheStats{
		Hits:     0,
		Misses:   0,
		HitRate:  0.0,
		KeyCount: 0,
	}
	
	// 实际实现应该解析info字符串获取真实统计数据
	_ = info
	
	return stats, nil
}

// 预定义的缓存管理器实例
var (
	// UserCache 用户缓存
	UserCache *CacheManager
	// ProjectCache 项目缓存
	ProjectCache *CacheManager
	// AICache AI服务缓存
	AICache *CacheManager
	// SessionCache 会话缓存
	SessionCache *CacheManager
)

// InitCacheManagers 初始化缓存管理器
func InitCacheManagers(redisClient *redis.Client) {
	UserCache = NewCacheManager(redisClient, "user", 30*time.Minute)
	ProjectCache = NewCacheManager(redisClient, "project", 15*time.Minute)
	AICache = NewCacheManager(redisClient, "ai", 1*time.Hour)
	SessionCache = NewCacheManager(redisClient, "session", 24*time.Hour)
}
