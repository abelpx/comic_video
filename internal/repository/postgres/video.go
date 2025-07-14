package postgres

import (
	"context"

	"comic_video/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VideoRepository struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

// Create 创建视频
func (r *VideoRepository) Create(ctx context.Context, video *entity.Video) error {
	return r.db.WithContext(ctx).Create(video).Error
}

// GetByID 根据ID获取视频
func (r *VideoRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Video, error) {
	var video entity.Video
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// GetByIDAndUser 根据ID和用户ID获取视频（权限校验）
func (r *VideoRepository) GetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*entity.Video, error) {
	var video entity.Video
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// ListByUser 获取用户的视频列表
func (r *VideoRepository) ListByUser(ctx context.Context, userID uuid.UUID, offset, limit int, filter map[string]interface{}) ([]*entity.Video, int64, error) {
	var videos []*entity.Video
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Video{}).Where("user_id = ?", userID)
	for k, v := range filter {
		query = query.Where(k+" = ?", v)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&videos).Error
	if err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

// ListByProject 获取项目的视频列表
func (r *VideoRepository) ListByProject(ctx context.Context, projectID uuid.UUID, offset, limit int) ([]*entity.Video, int64, error) {
	var videos []*entity.Video
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Video{}).Where("project_id = ?", projectID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&videos).Error
	if err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

// Update 更新视频
func (r *VideoRepository) Update(ctx context.Context, video *entity.Video) error {
	return r.db.WithContext(ctx).Save(video).Error
}

// Delete 删除视频
func (r *VideoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Video{}, id).Error
}

// UpdateStatus 更新视频状态
func (r *VideoRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&entity.Video{}).Where("id = ?", id).Update("status", status).Error
} 