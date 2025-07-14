package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"comic_video/internal/domain/entity"
)

// RenderRepository 渲染任务仓库接口
type RenderRepository interface {
	Create(ctx context.Context, render *entity.Render) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Render, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*entity.Render, int64, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID, page, pageSize int) ([]*entity.Render, int64, error)
	Update(ctx context.Context, render *entity.Render) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, progress int, error string) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetPendingRenders(ctx context.Context) ([]*entity.Render, error)
}

// renderRepository 渲染任务仓库实现
type renderRepository struct {
	db *gorm.DB
}

// NewRenderRepository 创建渲染任务仓库实例
func NewRenderRepository(db *gorm.DB) RenderRepository {
	return &renderRepository{db: db}
}

// Create 创建渲染任务
func (r *renderRepository) Create(ctx context.Context, render *entity.Render) error {
	return r.db.WithContext(ctx).Create(render).Error
}

// GetByID 根据ID获取渲染任务
func (r *renderRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Render, error) {
	var render entity.Render
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Project").
		Where("id = ?", id).
		First(&render).Error
	if err != nil {
		return nil, err
	}
	return &render, nil
}

// GetByUserID 根据用户ID获取渲染任务列表
func (r *renderRepository) GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*entity.Render, int64, error) {
	var renders []*entity.Render
	var total int64

	offset := (page - 1) * pageSize

	// 获取总数
	err := r.db.WithContext(ctx).Model(&entity.Render{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = r.db.WithContext(ctx).
		Preload("Project").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&renders).Error
	if err != nil {
		return nil, 0, err
	}

	return renders, total, nil
}

// GetByProjectID 根据项目ID获取渲染任务列表
func (r *renderRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID, page, pageSize int) ([]*entity.Render, int64, error) {
	var renders []*entity.Render
	var total int64

	offset := (page - 1) * pageSize

	// 获取总数
	err := r.db.WithContext(ctx).Model(&entity.Render{}).Where("project_id = ?", projectID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = r.db.WithContext(ctx).
		Preload("User").
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&renders).Error
	if err != nil {
		return nil, 0, err
	}

	return renders, total, nil
}

// Update 更新渲染任务
func (r *renderRepository) Update(ctx context.Context, render *entity.Render) error {
	return r.db.WithContext(ctx).Save(render).Error
}

// UpdateStatus 更新渲染任务状态
func (r *renderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, progress int, error string) error {
	updates := map[string]interface{}{
		"status":   status,
		"progress": progress,
	}

	if error != "" {
		updates["error"] = error
	}

	if status == "processing" {
		updates["started_at"] = gorm.Expr("CASE WHEN started_at IS NULL THEN NOW() ELSE started_at END")
	} else if status == "completed" || status == "failed" {
		updates["completed_at"] = gorm.Expr("NOW()")
	}

	return r.db.WithContext(ctx).Model(&entity.Render{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除渲染任务
func (r *renderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Render{}, id).Error
}

// GetPendingRenders 获取待处理的渲染任务
func (r *renderRepository) GetPendingRenders(ctx context.Context) ([]*entity.Render, error) {
	var renders []*entity.Render
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("User").
		Where("status = ?", "pending").
		Order("created_at ASC").
		Find(&renders).Error
	if err != nil {
		return nil, err
	}
	return renders, nil
} 