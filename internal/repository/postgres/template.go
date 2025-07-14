package postgres

import (
	"context"

	"comic_video/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// Create 创建模板
func (r *TemplateRepository) Create(ctx context.Context, template *entity.Template) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// GetByID 根据ID获取模板
func (r *TemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Template, error) {
	var template entity.Template
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// List 获取模板列表
func (r *TemplateRepository) List(ctx context.Context, offset, limit int, filter map[string]interface{}) ([]*entity.Template, int64, error) {
	var templates []*entity.Template
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Template{})
	for k, v := range filter {
		query = query.Where(k+" = ?", v)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&templates).Error
	if err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// Update 更新模板
func (r *TemplateRepository) Update(ctx context.Context, template *entity.Template) error {
	return r.db.WithContext(ctx).Save(template).Error
}

// Delete 删除模板
func (r *TemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Template{}, id).Error
} 