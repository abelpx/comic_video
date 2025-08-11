package quota

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"comic_video/internal/repository/redis"
)

// QuotaManager 配额管理器
type QuotaManager struct {
	redis *redis.Client
}

// NewQuotaManager 创建配额管理器
func NewQuotaManager(redisClient *redis.Client) *QuotaManager {
	return &QuotaManager{
		redis: redisClient,
	}
}

// QuotaType 配额类型
type QuotaType string

const (
	QuotaTypeVideo    QuotaType = "video"    // 视频生成配额
	QuotaTypeImage    QuotaType = "image"    // 图片生成配额
	QuotaTypeStorage  QuotaType = "storage"  // 存储空间配额
	QuotaTypeAPI      QuotaType = "api"      // API调用配额
)

// UserQuota 用户配额信息
type UserQuota struct {
	UserID    string    `json:"user_id"`
	Type      QuotaType `json:"type"`
	Limit     int64     `json:"limit"`     // 配额限制
	Used      int64     `json:"used"`      // 已使用
	Remaining int64     `json:"remaining"` // 剩余
	ResetTime time.Time `json:"reset_time"` // 重置时间
}

// CheckQuota 检查用户配额
func (q *QuotaManager) CheckQuota(ctx context.Context, userID string, quotaType QuotaType, amount int64) error {
	quota, err := q.GetUserQuota(ctx, userID, quotaType)
	if err != nil {
		return fmt.Errorf("获取配额信息失败: %v", err)
	}

	if quota.Remaining < amount {
		return fmt.Errorf("配额不足: 需要%d，剩余%d", amount, quota.Remaining)
	}

	return nil
}

// ConsumeQuota 消费配额
func (q *QuotaManager) ConsumeQuota(ctx context.Context, userID string, quotaType QuotaType, amount int64) error {
	key := q.getQuotaKey(userID, quotaType)

	// 简化实现：直接使用IncrBy
	// TODO: 可以考虑使用Lua脚本实现原子操作

	// 增加使用量
	_, err := q.redis.IncrBy(ctx, key, amount)
	if err != nil {
		return fmt.Errorf("配额消费失败: %v", err)
	}

	// 设置过期时间（月底重置）
	now := time.Now()
	nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	err = q.redis.ExpireAt(ctx, key, nextMonth)
	if err != nil {
		return fmt.Errorf("设置过期时间失败: %v", err)
	}

	return nil
}

// GetUserQuota 获取用户配额信息
func (q *QuotaManager) GetUserQuota(ctx context.Context, userID string, quotaType QuotaType) (*UserQuota, error) {
	key := q.getQuotaKey(userID, quotaType)

	// 获取已使用量
	usedStr, err := q.redis.Get(ctx, key)
	var used int64 = 0
	if err == nil {
		used, _ = strconv.ParseInt(usedStr, 10, 64)
	}

	// 获取配额限制（根据用户类型）
	limit := q.getQuotaLimit(userID, quotaType)
	
	// 计算重置时间
	now := time.Now()
	resetTime := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())

	quota := &UserQuota{
		UserID:    userID,
		Type:      quotaType,
		Limit:     limit,
		Used:      used,
		Remaining: limit - used,
		ResetTime: resetTime,
	}

	if quota.Remaining < 0 {
		quota.Remaining = 0
	}

	return quota, nil
}

// getQuotaKey 生成配额Redis键
func (q *QuotaManager) getQuotaKey(userID string, quotaType QuotaType) string {
	now := time.Now()
	month := now.Format("2006-01")
	return fmt.Sprintf("quota:%s:%s:%s", quotaType, userID, month)
}

// getQuotaLimit 获取配额限制（这里简化处理，实际应该从数据库获取用户订阅信息）
func (q *QuotaManager) getQuotaLimit(userID string, quotaType QuotaType) int64 {
	// TODO: 从数据库获取用户订阅计划
	// 这里先使用默认的免费用户配额
	switch quotaType {
	case QuotaTypeVideo:
		return 10 // 免费用户每月10个视频
	case QuotaTypeImage:
		return 100 // 免费用户每月100张图片
	case QuotaTypeStorage:
		return 1024 * 1024 * 1024 // 1GB存储空间
	case QuotaTypeAPI:
		return 1000 // 每月1000次API调用
	default:
		return 0
	}
}

// GetAllUserQuotas 获取用户所有配额信息
func (q *QuotaManager) GetAllUserQuotas(ctx context.Context, userID string) (map[QuotaType]*UserQuota, error) {
	quotas := make(map[QuotaType]*UserQuota)
	
	quotaTypes := []QuotaType{QuotaTypeVideo, QuotaTypeImage, QuotaTypeStorage, QuotaTypeAPI}
	
	for _, quotaType := range quotaTypes {
		quota, err := q.GetUserQuota(ctx, userID, quotaType)
		if err != nil {
			return nil, fmt.Errorf("获取%s配额失败: %v", quotaType, err)
		}
		quotas[quotaType] = quota
	}

	return quotas, nil
}

// ResetUserQuota 重置用户配额（管理员功能）
func (q *QuotaManager) ResetUserQuota(ctx context.Context, userID string, quotaType QuotaType) error {
	key := q.getQuotaKey(userID, quotaType)
	return q.redis.Del(ctx, key)
}

// SetUserQuotaLimit 设置用户配额限制（管理员功能）
func (q *QuotaManager) SetUserQuotaLimit(ctx context.Context, userID string, quotaType QuotaType, limit int64) error {
	// TODO: 实现用户配额限制设置
	// 这需要在数据库中存储用户的订阅计划信息
	return fmt.Errorf("功能待实现")
}

// UsageStats 使用统计
type UsageStats struct {
	UserID     string                 `json:"user_id"`
	Period     string                 `json:"period"`     // 统计周期
	Quotas     map[QuotaType]*UserQuota `json:"quotas"`     // 各类配额使用情况
	TotalCost  float64                `json:"total_cost"` // 总成本（如果有的话）
}

// GetUsageStats 获取用户使用统计
func (q *QuotaManager) GetUsageStats(ctx context.Context, userID string) (*UsageStats, error) {
	quotas, err := q.GetAllUserQuotas(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	period := now.Format("2006-01")

	stats := &UsageStats{
		UserID:    userID,
		Period:    period,
		Quotas:    quotas,
		TotalCost: 0, // TODO: 计算实际成本
	}

	return stats, nil
}
