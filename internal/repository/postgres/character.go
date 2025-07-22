package postgres

import (
	"context"

	"comic_video/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CharacterRepository 角色仓储
type CharacterRepository struct {
	db *gorm.DB
}

// NewCharacterRepository 创建角色仓储
func NewCharacterRepository(db *gorm.DB) *CharacterRepository {
	return &CharacterRepository{db: db}
}

// Create 创建角色
func (r *CharacterRepository) Create(ctx context.Context, character *entity.Character) error {
	return r.db.WithContext(ctx).Create(character).Error
}

// GetByID 根据ID获取角色
func (r *CharacterRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Character, error) {
	var character entity.Character
	err := r.db.WithContext(ctx).
		Preload("Project").
		First(&character, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &character, nil
}

// Update 更新角色
func (r *CharacterRepository) Update(ctx context.Context, character *entity.Character) error {
	return r.db.WithContext(ctx).Save(character).Error
}

// Delete 删除角色
func (r *CharacterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Character{}, "id = ?", id).Error
}

// GetByProjectID 根据项目ID获取角色列表
func (r *CharacterRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Character, error) {
	var characters []*entity.Character
	err := r.db.WithContext(ctx).
		Preload("Project").
		Find(&characters, "project_id = ?", projectID).Error
	return characters, err
}

// CharacterImageRepository 角色图像仓储
type CharacterImageRepository struct {
	db *gorm.DB
}

// NewCharacterImageRepository 创建角色图像仓储
func NewCharacterImageRepository(db *gorm.DB) *CharacterImageRepository {
	return &CharacterImageRepository{db: db}
}

// Create 创建角色图像
func (r *CharacterImageRepository) Create(ctx context.Context, image *entity.CharacterImage) error {
	return r.db.WithContext(ctx).Create(image).Error
}

// GetByID 根据ID获取角色图像
func (r *CharacterImageRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.CharacterImage, error) {
	var image entity.CharacterImage
	err := r.db.WithContext(ctx).
		Preload("Character").
		First(&image, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// Update 更新角色图像
func (r *CharacterImageRepository) Update(ctx context.Context, image *entity.CharacterImage) error {
	return r.db.WithContext(ctx).Save(image).Error
}

// Delete 删除角色图像
func (r *CharacterImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.CharacterImage{}, "id = ?", id).Error
}

// GetByCharacterID 根据角色ID获取图像列表
func (r *CharacterImageRepository) GetByCharacterID(ctx context.Context, characterID uuid.UUID) ([]*entity.CharacterImage, error) {
	var images []*entity.CharacterImage
	err := r.db.WithContext(ctx).
		Preload("Character").
		Find(&images, "character_id = ?", characterID).Error
	return images, err
}

// GetReferenceImages 获取参考图像
func (r *CharacterImageRepository) GetReferenceImages(ctx context.Context, characterID uuid.UUID) ([]*entity.CharacterImage, error) {
	var images []*entity.CharacterImage
	err := r.db.WithContext(ctx).
		Preload("Character").
		Find(&images, "character_id = ? AND is_reference = ?", characterID, true).Error
	return images, err
}
