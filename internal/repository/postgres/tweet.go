package postgres

import (
	"context"

	"comic_video/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TweetRepository 推文仓储
type TweetRepository struct {
	db *gorm.DB
}

// NewTweetRepository 创建推文仓储
func NewTweetRepository(db *gorm.DB) *TweetRepository {
	return &TweetRepository{db: db}
}

// Create 创建推文
func (r *TweetRepository) Create(ctx context.Context, tweet *entity.Tweet) error {
	return r.db.WithContext(ctx).Create(tweet).Error
}

// GetByID 根据ID获取推文
func (r *TweetRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tweet, error) {
	var tweet entity.Tweet
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Project").
		First(&tweet, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tweet, nil
}

// Update 更新推文
func (r *TweetRepository) Update(ctx context.Context, tweet *entity.Tweet) error {
	return r.db.WithContext(ctx).Save(tweet).Error
}

// Delete 删除推文
func (r *TweetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Tweet{}, "id = ?", id).Error
}

// ListByUserID 根据用户ID获取推文列表
func (r *TweetRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Tweet, int64, error) {
	var tweets []*entity.Tweet
	var total int64

	// 获取总数
	err := r.db.WithContext(ctx).
		Model(&entity.Tweet{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	err = r.db.WithContext(ctx).
		Preload("User").
		Preload("Project").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tweets).Error

	return tweets, total, err
}

// ListByProjectID 根据项目ID获取推文列表
func (r *TweetRepository) ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Tweet, error) {
	var tweets []*entity.Tweet
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&tweets).Error
	return tweets, err
}

// ListByStatus 根据状态获取推文列表
func (r *TweetRepository) ListByStatus(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]*entity.Tweet, int64, error) {
	var tweets []*entity.Tweet
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Tweet{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	err = query.
		Preload("User").
		Preload("Project").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tweets).Error

	return tweets, total, err
}

// Search 搜索推文
func (r *TweetRepository) Search(ctx context.Context, userID uuid.UUID, keyword string, limit, offset int) ([]*entity.Tweet, int64, error) {
	var tweets []*entity.Tweet
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Tweet{}).Where("user_id = ?", userID)
	if keyword != "" {
		query = query.Where("content ILIKE ? OR title ILIKE ? OR theme ILIKE ?", 
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	err = query.
		Preload("User").
		Preload("Project").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tweets).Error

	return tweets, total, err
}

// UpdateStatus 更新推文状态
func (r *TweetRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&entity.Tweet{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// IncrementViewCount 增加查看次数
func (r *TweetRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.Tweet{}).
		Where("id = ?", id).
		Update("view_count", gorm.Expr("view_count + 1")).Error
}

// GetStatsByUserID 获取用户推文统计
func (r *TweetRepository) GetStatsByUserID(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	var stats struct {
		Total     int64 `json:"total"`
		Draft     int64 `json:"draft"`
		Published int64 `json:"published"`
		Archived  int64 `json:"archived"`
	}

	// 总数
	err := r.db.WithContext(ctx).
		Model(&entity.Tweet{}).
		Where("user_id = ?", userID).
		Count(&stats.Total).Error
	if err != nil {
		return nil, err
	}

	// 草稿数
	err = r.db.WithContext(ctx).
		Model(&entity.Tweet{}).
		Where("user_id = ? AND status = ?", userID, "draft").
		Count(&stats.Draft).Error
	if err != nil {
		return nil, err
	}

	// 已发布数
	err = r.db.WithContext(ctx).
		Model(&entity.Tweet{}).
		Where("user_id = ? AND status = ?", userID, "published").
		Count(&stats.Published).Error
	if err != nil {
		return nil, err
	}

	// 已归档数
	err = r.db.WithContext(ctx).
		Model(&entity.Tweet{}).
		Where("user_id = ? AND status = ?", userID, "archived").
		Count(&stats.Archived).Error
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"total":     stats.Total,
		"draft":     stats.Draft,
		"published": stats.Published,
		"archived":  stats.Archived,
	}

	return result, nil
}
