package postgres

import (
	"context"

	"comic_video/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ScriptRepository 剧本仓储
type ScriptRepository struct {
	db *gorm.DB
}

// NewScriptRepository 创建剧本仓储
func NewScriptRepository(db *gorm.DB) *ScriptRepository {
	return &ScriptRepository{db: db}
}

// Create 创建剧本
func (r *ScriptRepository) Create(ctx context.Context, script *entity.Script) error {
	return r.db.WithContext(ctx).Create(script).Error
}

// GetByID 根据ID获取剧本
func (r *ScriptRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Script, error) {
	var script entity.Script
	err := r.db.WithContext(ctx).
		Preload("Project").
		First(&script, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &script, nil
}

// Update 更新剧本
func (r *ScriptRepository) Update(ctx context.Context, script *entity.Script) error {
	return r.db.WithContext(ctx).Save(script).Error
}

// Delete 删除剧本
func (r *ScriptRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Script{}, "id = ?", id).Error
}

// GetByProjectID 根据项目ID获取剧本
func (r *ScriptRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Script, error) {
	var scripts []*entity.Script
	err := r.db.WithContext(ctx).
		Preload("Project").
		Find(&scripts, "project_id = ?", projectID).Error
	return scripts, err
}

// ScriptSceneRepository 剧本场景仓储
type ScriptSceneRepository struct {
	db *gorm.DB
}

// NewScriptSceneRepository 创建剧本场景仓储
func NewScriptSceneRepository(db *gorm.DB) *ScriptSceneRepository {
	return &ScriptSceneRepository{db: db}
}

// Create 创建剧本场景
func (r *ScriptSceneRepository) Create(ctx context.Context, scene *entity.ScriptScene) error {
	return r.db.WithContext(ctx).Create(scene).Error
}

// GetByID 根据ID获取剧本场景
func (r *ScriptSceneRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ScriptScene, error) {
	var scene entity.ScriptScene
	err := r.db.WithContext(ctx).
		Preload("Script").
		First(&scene, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &scene, nil
}

// Update 更新剧本场景
func (r *ScriptSceneRepository) Update(ctx context.Context, scene *entity.ScriptScene) error {
	return r.db.WithContext(ctx).Save(scene).Error
}

// Delete 删除剧本场景
func (r *ScriptSceneRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.ScriptScene{}, "id = ?", id).Error
}

// GetByScriptID 根据剧本ID获取场景列表
func (r *ScriptSceneRepository) GetByScriptID(ctx context.Context, scriptID uuid.UUID) ([]*entity.ScriptScene, error) {
	var scenes []*entity.ScriptScene
	err := r.db.WithContext(ctx).
		Preload("Script").
		Order("scene_number ASC").
		Find(&scenes, "script_id = ?", scriptID).Error
	return scenes, err
}
