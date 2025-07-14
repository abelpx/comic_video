package postgres

import (
	"context"
	"comic_video/internal/domain/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectShareRepository 项目分享仓库接口
// 支持创建、查询、失效分享
//
type ProjectShareRepository interface {
	Create(ctx context.Context, share *entity.ProjectShare) error
	GetByToken(ctx context.Context, token string) (*entity.ProjectShare, error)
	ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.ProjectShare, error)
	ExpireByID(ctx context.Context, id uuid.UUID) error
}

// projectShareRepository 实现

type projectShareRepository struct {
	db *gorm.DB
}

func NewProjectShareRepository(db *gorm.DB) ProjectShareRepository {
	return &projectShareRepository{db: db}
}

// Create 创建分享
func (r *projectShareRepository) Create(ctx context.Context, share *entity.ProjectShare) error {
	return r.db.WithContext(ctx).Create(share).Error
}

// GetByToken 通过token查询分享
func (r *projectShareRepository) GetByToken(ctx context.Context, token string) (*entity.ProjectShare, error) {
	var share entity.ProjectShare
	err := r.db.WithContext(ctx).Where("token = ? AND status = 'active'", token).First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// ListByProjectID 查询项目下所有分享
func (r *projectShareRepository) ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.ProjectShare, error) {
	var shares []*entity.ProjectShare
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at DESC").Find(&shares).Error
	return shares, err
}

// ExpireByID 失效分享
func (r *projectShareRepository) ExpireByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entity.ProjectShare{}).Where("id = ?", id).Update("status", "expired").Error
} 