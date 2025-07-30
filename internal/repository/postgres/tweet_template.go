package postgres

import (
	"context"

	"comic_video/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TweetTemplateRepository 推文模板仓储
type TweetTemplateRepository struct {
	db *gorm.DB
}

// NewTweetTemplateRepository 创建推文模板仓储
func NewTweetTemplateRepository(db *gorm.DB) *TweetTemplateRepository {
	return &TweetTemplateRepository{db: db}
}

// Create 创建推文模板
func (r *TweetTemplateRepository) Create(ctx context.Context, template *entity.TweetTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// GetByID 根据ID获取推文模板
func (r *TweetTemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.TweetTemplate, error) {
	var template entity.TweetTemplate
	err := r.db.WithContext(ctx).First(&template, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// Update 更新推文模板
func (r *TweetTemplateRepository) Update(ctx context.Context, template *entity.TweetTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

// Delete 删除推文模板
func (r *TweetTemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.TweetTemplate{}, "id = ?", id).Error
}

// List 获取推文模板列表
func (r *TweetTemplateRepository) List(ctx context.Context, category, platform string, isPublic *bool, limit, offset int) ([]*entity.TweetTemplate, int64, error) {
	var templates []*entity.TweetTemplate
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.TweetTemplate{}).Where("status = ?", "active")

	// 添加筛选条件
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	err = query.
		Order("use_count DESC, rating DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&templates).Error

	return templates, total, err
}

// ListByCreator 根据创建者获取模板列表
func (r *TweetTemplateRepository) ListByCreator(ctx context.Context, creatorID uuid.UUID, limit, offset int) ([]*entity.TweetTemplate, int64, error) {
	var templates []*entity.TweetTemplate
	var total int64

	// 获取总数
	err := r.db.WithContext(ctx).
		Model(&entity.TweetTemplate{}).
		Where("created_by = ?", creatorID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	err = r.db.WithContext(ctx).
		Where("created_by = ?", creatorID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&templates).Error

	return templates, total, err
}

// Search 搜索推文模板
func (r *TweetTemplateRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*entity.TweetTemplate, int64, error) {
	var templates []*entity.TweetTemplate
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.TweetTemplate{}).
		Where("status = ? AND is_public = ?", "active", true)

	if keyword != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ? OR template ILIKE ?", 
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	err = query.
		Order("use_count DESC, rating DESC").
		Limit(limit).
		Offset(offset).
		Find(&templates).Error

	return templates, total, err
}

// GetPopular 获取热门模板
func (r *TweetTemplateRepository) GetPopular(ctx context.Context, platform string, limit int) ([]*entity.TweetTemplate, error) {
	var templates []*entity.TweetTemplate

	query := r.db.WithContext(ctx).
		Where("status = ? AND is_public = ?", "active", true)

	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	err := query.
		Order("use_count DESC, rating DESC").
		Limit(limit).
		Find(&templates).Error

	return templates, err
}

// GetByCategory 根据分类获取模板
func (r *TweetTemplateRepository) GetByCategory(ctx context.Context, category string, limit int) ([]*entity.TweetTemplate, error) {
	var templates []*entity.TweetTemplate

	err := r.db.WithContext(ctx).
		Where("category = ? AND status = ? AND is_public = ?", category, "active", true).
		Order("use_count DESC, rating DESC").
		Limit(limit).
		Find(&templates).Error

	return templates, err
}

// IncrementUseCount 增加使用次数
func (r *TweetTemplateRepository) IncrementUseCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.TweetTemplate{}).
		Where("id = ?", id).
		Update("use_count", gorm.Expr("use_count + 1")).Error
}

// UpdateRating 更新评分
func (r *TweetTemplateRepository) UpdateRating(ctx context.Context, id uuid.UUID, rating float64) error {
	return r.db.WithContext(ctx).
		Model(&entity.TweetTemplate{}).
		Where("id = ?", id).
		Update("rating", rating).Error
}

// GetCategories 获取所有分类
func (r *TweetTemplateRepository) GetCategories(ctx context.Context) ([]string, error) {
	var categories []string
	err := r.db.WithContext(ctx).
		Model(&entity.TweetTemplate{}).
		Where("status = ? AND is_public = ?", "active", true).
		Distinct("category").
		Pluck("category", &categories).Error
	return categories, err
}

// GetStats 获取模板统计信息
func (r *TweetTemplateRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	var stats struct {
		Total    int64 `json:"total"`
		Public   int64 `json:"public"`
		Premium  int64 `json:"premium"`
		Active   int64 `json:"active"`
	}

	// 总数
	err := r.db.WithContext(ctx).
		Model(&entity.TweetTemplate{}).
		Count(&stats.Total).Error
	if err != nil {
		return nil, err
	}

	// 公开模板数
	err = r.db.WithContext(ctx).
		Model(&entity.TweetTemplate{}).
		Where("is_public = ?", true).
		Count(&stats.Public).Error
	if err != nil {
		return nil, err
	}

	// 付费模板数
	err = r.db.WithContext(ctx).
		Model(&entity.TweetTemplate{}).
		Where("is_premium = ?", true).
		Count(&stats.Premium).Error
	if err != nil {
		return nil, err
	}

	// 活跃模板数
	err = r.db.WithContext(ctx).
		Model(&entity.TweetTemplate{}).
		Where("status = ?", "active").
		Count(&stats.Active).Error
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"total":   stats.Total,
		"public":  stats.Public,
		"premium": stats.Premium,
		"active":  stats.Active,
	}

	return result, nil
}
