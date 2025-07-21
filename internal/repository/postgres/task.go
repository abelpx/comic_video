package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"comic_video/internal/domain/entity"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// CreateTask 创建任务
func (r *TaskRepository) CreateTask(ctx context.Context, task *entity.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// GetTaskByID 根据ID获取任务
func (r *TaskRepository) GetTaskByID(ctx context.Context, taskID string) (*entity.Task, error) {
	var task entity.Task
	err := r.db.WithContext(ctx).Where("id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTask 更新任务
func (r *TaskRepository) UpdateTask(ctx context.Context, task *entity.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// UpdateTaskStatus 更新任务状态
func (r *TaskRepository) UpdateTaskStatus(ctx context.Context, taskID string, status string, progress int) error {
	return r.db.WithContext(ctx).Model(&entity.Task{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":     status,
			"progress":   progress,
			"updated_at": "NOW()",
		}).Error
}

// UpdateTaskResult 更新任务结果
func (r *TaskRepository) UpdateTaskResult(ctx context.Context, taskID string, result string) error {
	return r.db.WithContext(ctx).Model(&entity.Task{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"result":     result,
			"status":     entity.TaskStatusCompleted,
			"progress":   100,
			"updated_at": "NOW()",
		}).Error
}

// UpdateTaskError 更新任务错误
func (r *TaskRepository) UpdateTaskError(ctx context.Context, taskID string, errorMsg string) error {
	return r.db.WithContext(ctx).Model(&entity.Task{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"error":      errorMsg,
			"status":     entity.TaskStatusFailed,
			"updated_at": "NOW()",
		}).Error
}

// GetUserTasks 获取用户任务列表
func (r *TaskRepository) GetUserTasks(ctx context.Context, userID string, page, limit int) ([]*entity.Task, int64, error) {
	var tasks []*entity.Task
	var total int64

	// 计算偏移量
	offset := (page - 1) * limit

	// 查询总数
	err := r.db.WithContext(ctx).Model(&entity.Task{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 查询任务列表
	err = r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// GetTasksByStatus 根据状态获取任务列表
func (r *TaskRepository) GetTasksByStatus(ctx context.Context, userID string, status string, page, limit int) ([]*entity.Task, int64, error) {
	var tasks []*entity.Task
	var total int64

	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&entity.Task{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 查询总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 查询任务列表
	err = query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// GetTasksByType 根据类型获取任务列表
func (r *TaskRepository) GetTasksByType(ctx context.Context, userID string, taskType string, page, limit int) ([]*entity.Task, int64, error) {
	var tasks []*entity.Task
	var total int64

	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&entity.Task{}).Where("user_id = ?", userID)
	if taskType != "" {
		query = query.Where("type = ?", taskType)
	}

	// 查询总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 查询任务列表
	err = query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// DeleteTask 删除任务
func (r *TaskRepository) DeleteTask(ctx context.Context, taskID string, userID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", taskID, userID).
		Delete(&entity.Task{})
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("task not found or access denied")
	}
	
	return nil
}

// GetActiveTasks 获取用户活跃任务（处理中和等待中）
func (r *TaskRepository) GetActiveTasks(ctx context.Context, userID string) ([]*entity.Task, error) {
	var tasks []*entity.Task
	
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status IN (?)", userID, []string{entity.TaskStatusPending, entity.TaskStatusProcessing}).
		Order("created_at DESC").
		Find(&tasks).Error
	
	return tasks, err
}

// GetTaskStats 获取用户任务统计
func (r *TaskRepository) GetTaskStats(ctx context.Context, userID string) (map[string]int64, error) {
	stats := make(map[string]int64)
	
	// 统计各状态任务数量
	statuses := []string{entity.TaskStatusPending, entity.TaskStatusProcessing, entity.TaskStatusCompleted, entity.TaskStatusFailed}
	
	for _, status := range statuses {
		var count int64
		err := r.db.WithContext(ctx).Model(&entity.Task{}).
			Where("user_id = ? AND status = ?", userID, status).
			Count(&count).Error
		if err != nil {
			return nil, err
		}
		stats[status] = count
	}
	
	// 统计总任务数
	var total int64
	err := r.db.WithContext(ctx).Model(&entity.Task{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, err
	}
	stats["total"] = total
	
	return stats, nil
}
