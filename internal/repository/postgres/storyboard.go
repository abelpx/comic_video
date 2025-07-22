package postgres

import (
	"context"

	"comic_video/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StoryboardRepository 分镜仓储
type StoryboardRepository struct {
	db *gorm.DB
}

// NewStoryboardRepository 创建分镜仓储
func NewStoryboardRepository(db *gorm.DB) *StoryboardRepository {
	return &StoryboardRepository{db: db}
}

// Create 创建分镜
func (r *StoryboardRepository) Create(ctx context.Context, storyboard *entity.Storyboard) error {
	return r.db.WithContext(ctx).Create(storyboard).Error
}

// GetByID 根据ID获取分镜
func (r *StoryboardRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Storyboard, error) {
	var storyboard entity.Storyboard
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Script").
		First(&storyboard, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &storyboard, nil
}

// Update 更新分镜
func (r *StoryboardRepository) Update(ctx context.Context, storyboard *entity.Storyboard) error {
	return r.db.WithContext(ctx).Save(storyboard).Error
}

// Delete 删除分镜
func (r *StoryboardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Storyboard{}, "id = ?", id).Error
}

// GetByProjectID 根据项目ID获取分镜列表
func (r *StoryboardRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Storyboard, error) {
	var storyboards []*entity.Storyboard
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Script").
		Find(&storyboards, "project_id = ?", projectID).Error
	return storyboards, err
}

// GetByScriptID 根据剧本ID获取分镜
func (r *StoryboardRepository) GetByScriptID(ctx context.Context, scriptID uuid.UUID) (*entity.Storyboard, error) {
	var storyboard entity.Storyboard
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Script").
		First(&storyboard, "script_id = ?", scriptID).Error
	if err != nil {
		return nil, err
	}
	return &storyboard, nil
}

// StoryboardFrameRepository 分镜帧仓储
type StoryboardFrameRepository struct {
	db *gorm.DB
}

// NewStoryboardFrameRepository 创建分镜帧仓储
func NewStoryboardFrameRepository(db *gorm.DB) *StoryboardFrameRepository {
	return &StoryboardFrameRepository{db: db}
}

// Create 创建分镜帧
func (r *StoryboardFrameRepository) Create(ctx context.Context, frame *entity.StoryboardFrame) error {
	return r.db.WithContext(ctx).Create(frame).Error
}

// GetByID 根据ID获取分镜帧
func (r *StoryboardFrameRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.StoryboardFrame, error) {
	var frame entity.StoryboardFrame
	err := r.db.WithContext(ctx).
		Preload("Storyboard").
		Preload("Scene").
		First(&frame, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &frame, nil
}

// Update 更新分镜帧
func (r *StoryboardFrameRepository) Update(ctx context.Context, frame *entity.StoryboardFrame) error {
	return r.db.WithContext(ctx).Save(frame).Error
}

// Delete 删除分镜帧
func (r *StoryboardFrameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.StoryboardFrame{}, "id = ?", id).Error
}

// GetByStoryboardID 根据分镜ID获取帧列表
func (r *StoryboardFrameRepository) GetByStoryboardID(ctx context.Context, storyboardID uuid.UUID) ([]*entity.StoryboardFrame, error) {
	var frames []*entity.StoryboardFrame
	err := r.db.WithContext(ctx).
		Preload("Storyboard").
		Preload("Scene").
		Order("frame_number ASC").
		Find(&frames, "storyboard_id = ?", storyboardID).Error
	return frames, err
}
