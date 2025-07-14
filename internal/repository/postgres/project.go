package postgres

import (
	"context"

	"comic_video/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectRepository 项目仓库接口
type ProjectRepository interface {
	Create(ctx context.Context, project *entity.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entity.Project, int64, error)
	Update(ctx context.Context, project *entity.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int, status, keyword string) ([]*entity.Project, int64, error)
	GetByStatus(ctx context.Context, status string) ([]*entity.Project, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

// projectRepository 项目仓库实现
type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

var _ ProjectRepository = (*projectRepository)(nil)

// Create 创建项目
func (r *projectRepository) Create(ctx context.Context, project *entity.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

// GetByID 根据ID获取项目
func (r *projectRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	var project entity.Project
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// ListByUserID 获取用户的项目列表
func (r *projectRepository) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entity.Project, int64, error) {
	var projects []*entity.Project
	var total int64

	// 获取总数
	err := r.db.WithContext(ctx).Model(&entity.Project{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取项目列表
	err = r.db.WithContext(ctx).Where("user_id = ?", userID).Offset(offset).Limit(limit).Order("updated_at DESC").Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// Update 更新项目
func (r *projectRepository) Update(ctx context.Context, project *entity.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

// Delete 删除项目
func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Project{}, id).Error
}

// List 获取项目列表（管理员用）
func (r *projectRepository) List(ctx context.Context, offset, limit int, status, keyword string) ([]*entity.Project, int64, error) {
	var projects []*entity.Project
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Project{})

	// 添加过滤条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取项目列表
	err = query.Offset(offset).Limit(limit).Order("updated_at DESC").Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// GetByStatus 根据状态获取项目
func (r *projectRepository) GetByStatus(ctx context.Context, status string) ([]*entity.Project, error) {
	var projects []*entity.Project
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&projects).Error
	return projects, err
}

// UpdateStatus 更新项目状态
func (r *projectRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&entity.Project{}).Where("id = ?", id).Update("status", status).Error
} 