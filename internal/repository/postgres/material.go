package postgres

import (
	"context"

	"comic_video/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MaterialRepository 素材仓库接口
type MaterialRepository interface {
	Create(ctx context.Context, material *entity.Material) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Material, error)
	List(ctx context.Context, offset, limit int, filter map[string]interface{}) ([]*entity.Material, int64, error)
	Update(ctx context.Context, material *entity.Material) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// materialRepository 素材仓库实现
type materialRepository struct {
	db *gorm.DB
}

func NewMaterialRepository(db *gorm.DB) MaterialRepository {
	return &materialRepository{db: db}
}

var _ MaterialRepository = (*materialRepository)(nil)

// Create 创建素材
func (r *materialRepository) Create(ctx context.Context, material *entity.Material) error {
	return r.db.WithContext(ctx).Create(material).Error
}

// GetByID 根据ID获取素材
func (r *materialRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Material, error) {
	var material entity.Material
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&material).Error
	if err != nil {
		return nil, err
	}
	return &material, nil
}

// List 获取素材列表
func (r *materialRepository) List(ctx context.Context, offset, limit int, filter map[string]interface{}) ([]*entity.Material, int64, error) {
	var materials []*entity.Material
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Material{})
	for k, v := range filter {
		query = query.Where(k+" = ?", v)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&materials).Error
	if err != nil {
		return nil, 0, err
	}

	return materials, total, nil
}

// Update 更新素材
func (r *materialRepository) Update(ctx context.Context, material *entity.Material) error {
	return r.db.WithContext(ctx).Save(material).Error
}

// Delete 删除素材
func (r *materialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Material{}, id).Error
} 