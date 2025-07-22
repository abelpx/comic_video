package postgres

import (
	"context"

	"comic_video/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowRepository 工作流仓储
type WorkflowRepository struct {
	db *gorm.DB
}

// NewWorkflowRepository 创建工作流仓储
func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

// Create 创建工作流
func (r *WorkflowRepository) Create(ctx context.Context, workflow *entity.Workflow) error {
	return r.db.WithContext(ctx).Create(workflow).Error
}

// GetByID 根据ID获取工作流
func (r *WorkflowRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Workflow, error) {
	var workflow entity.Workflow
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("User").
		First(&workflow, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

// Update 更新工作流
func (r *WorkflowRepository) Update(ctx context.Context, workflow *entity.Workflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}

// Delete 删除工作流
func (r *WorkflowRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Workflow{}, "id = ?", id).Error
}

// GetByProjectID 根据项目ID获取工作流
func (r *WorkflowRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Workflow, error) {
	var workflows []*entity.Workflow
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("User").
		Find(&workflows, "project_id = ?", projectID).Error
	return workflows, err
}

// GetByUserID 根据用户ID获取工作流
func (r *WorkflowRepository) GetByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entity.Workflow, int64, error) {
	var workflows []*entity.Workflow
	var total int64

	// 获取总数
	r.db.WithContext(ctx).Model(&entity.Workflow{}).Where("user_id = ?", userID).Count(&total)

	// 获取分页数据
	err := r.db.WithContext(ctx).
		Preload("Project").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&workflows).Error

	return workflows, total, err
}

// WorkflowTaskRepository 工作流任务仓储
type WorkflowTaskRepository struct {
	db *gorm.DB
}

// NewWorkflowTaskRepository 创建工作流任务仓储
func NewWorkflowTaskRepository(db *gorm.DB) *WorkflowTaskRepository {
	return &WorkflowTaskRepository{db: db}
}

// Create 创建工作流任务
func (r *WorkflowTaskRepository) Create(ctx context.Context, task *entity.WorkflowTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// GetByID 根据ID获取工作流任务
func (r *WorkflowTaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.WorkflowTask, error) {
	var task entity.WorkflowTask
	err := r.db.WithContext(ctx).
		Preload("Workflow").
		First(&task, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Update 更新工作流任务
func (r *WorkflowTaskRepository) Update(ctx context.Context, task *entity.WorkflowTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// Delete 删除工作流任务
func (r *WorkflowTaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.WorkflowTask{}, "id = ?", id).Error
}

// GetByWorkflowID 根据工作流ID获取任务列表
func (r *WorkflowTaskRepository) GetByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*entity.WorkflowTask, error) {
	var tasks []*entity.WorkflowTask
	err := r.db.WithContext(ctx).
		Preload("Workflow").
		Find(&tasks, "workflow_id = ?", workflowID).Error
	return tasks, err
}

// GetByStatus 根据状态获取任务列表
func (r *WorkflowTaskRepository) GetByStatus(ctx context.Context, status entity.WorkflowStatus) ([]*entity.WorkflowTask, error) {
	var tasks []*entity.WorkflowTask
	err := r.db.WithContext(ctx).
		Preload("Workflow").
		Find(&tasks, "status = ?", status).Error
	return tasks, err
}
