package postgres

import (
	"context"

	"comic_video/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SceneRepository 场景仓储
type SceneRepository struct {
	db *gorm.DB
}

// NewSceneRepository 创建场景仓储
func NewSceneRepository(db *gorm.DB) *SceneRepository {
	return &SceneRepository{db: db}
}

// Create 创建场景
func (r *SceneRepository) Create(ctx context.Context, scene *entity.Scene) error {
	return r.db.WithContext(ctx).Create(scene).Error
}

// GetByID 根据ID获取场景
func (r *SceneRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Scene, error) {
	var scene entity.Scene
	err := r.db.WithContext(ctx).
		Preload("Project").
		First(&scene, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &scene, nil
}

// Update 更新场景
func (r *SceneRepository) Update(ctx context.Context, scene *entity.Scene) error {
	return r.db.WithContext(ctx).Save(scene).Error
}

// Delete 删除场景
func (r *SceneRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Scene{}, "id = ?", id).Error
}

// GetByProjectID 根据项目ID获取场景列表
func (r *SceneRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Scene, error) {
	var scenes []*entity.Scene
	err := r.db.WithContext(ctx).
		Preload("Project").
		Find(&scenes, "project_id = ?", projectID).Error
	return scenes, err
}

// SceneImageRepository 场景图像仓储
type SceneImageRepository struct {
	db *gorm.DB
}

// NewSceneImageRepository 创建场景图像仓储
func NewSceneImageRepository(db *gorm.DB) *SceneImageRepository {
	return &SceneImageRepository{db: db}
}

// Create 创建场景图像
func (r *SceneImageRepository) Create(ctx context.Context, image *entity.SceneImage) error {
	return r.db.WithContext(ctx).Create(image).Error
}

// GetByID 根据ID获取场景图像
func (r *SceneImageRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.SceneImage, error) {
	var image entity.SceneImage
	err := r.db.WithContext(ctx).
		Preload("Scene").
		First(&image, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// Update 更新场景图像
func (r *SceneImageRepository) Update(ctx context.Context, image *entity.SceneImage) error {
	return r.db.WithContext(ctx).Save(image).Error
}

// Delete 删除场景图像
func (r *SceneImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.SceneImage{}, "id = ?", id).Error
}

// GetBySceneID 根据场景ID获取图像列表
func (r *SceneImageRepository) GetBySceneID(ctx context.Context, sceneID uuid.UUID) ([]*entity.SceneImage, error) {
	var images []*entity.SceneImage
	err := r.db.WithContext(ctx).
		Preload("Scene").
		Find(&images, "scene_id = ?", sceneID).Error
	return images, err
}
