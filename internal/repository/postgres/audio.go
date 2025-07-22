package postgres

import (
	"context"

	"comic_video/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AudioRepository 音频仓储
type AudioRepository struct {
	db *gorm.DB
}

// NewAudioRepository 创建音频仓储
func NewAudioRepository(db *gorm.DB) *AudioRepository {
	return &AudioRepository{db: db}
}

// Create 创建音频
func (r *AudioRepository) Create(ctx context.Context, audio *entity.Audio) error {
	return r.db.WithContext(ctx).Create(audio).Error
}

// GetByID 根据ID获取音频
func (r *AudioRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Audio, error) {
	var audio entity.Audio
	err := r.db.WithContext(ctx).
		Preload("Project").
		First(&audio, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &audio, nil
}

// Update 更新音频
func (r *AudioRepository) Update(ctx context.Context, audio *entity.Audio) error {
	return r.db.WithContext(ctx).Save(audio).Error
}

// Delete 删除音频
func (r *AudioRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Audio{}, "id = ?", id).Error
}

// GetByProjectID 根据项目ID获取音频列表
func (r *AudioRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Audio, error) {
	var audios []*entity.Audio
	err := r.db.WithContext(ctx).
		Preload("Project").
		Find(&audios, "project_id = ?", projectID).Error
	return audios, err
}

// GetByType 根据类型获取音频列表
func (r *AudioRepository) GetByType(ctx context.Context, projectID uuid.UUID, audioType entity.AudioType) ([]*entity.Audio, error) {
	var audios []*entity.Audio
	err := r.db.WithContext(ctx).
		Preload("Project").
		Find(&audios, "project_id = ? AND type = ?", projectID, audioType).Error
	return audios, err
}

// VoiceGenerationRepository 语音生成仓储
type VoiceGenerationRepository struct {
	db *gorm.DB
}

// NewVoiceGenerationRepository 创建语音生成仓储
func NewVoiceGenerationRepository(db *gorm.DB) *VoiceGenerationRepository {
	return &VoiceGenerationRepository{db: db}
}

// Create 创建语音生成记录
func (r *VoiceGenerationRepository) Create(ctx context.Context, voiceGen *entity.VoiceGeneration) error {
	return r.db.WithContext(ctx).Create(voiceGen).Error
}

// GetByID 根据ID获取语音生成记录
func (r *VoiceGenerationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.VoiceGeneration, error) {
	var voiceGen entity.VoiceGeneration
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Character").
		Preload("Audio").
		First(&voiceGen, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &voiceGen, nil
}

// Update 更新语音生成记录
func (r *VoiceGenerationRepository) Update(ctx context.Context, voiceGen *entity.VoiceGeneration) error {
	return r.db.WithContext(ctx).Save(voiceGen).Error
}

// Delete 删除语音生成记录
func (r *VoiceGenerationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.VoiceGeneration{}, "id = ?", id).Error
}

// GetByProjectID 根据项目ID获取语音生成记录列表
func (r *VoiceGenerationRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.VoiceGeneration, error) {
	var voiceGens []*entity.VoiceGeneration
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Character").
		Preload("Audio").
		Find(&voiceGens, "project_id = ?", projectID).Error
	return voiceGens, err
}

// MusicGenerationRepository 音乐生成仓储
type MusicGenerationRepository struct {
	db *gorm.DB
}

// NewMusicGenerationRepository 创建音乐生成仓储
func NewMusicGenerationRepository(db *gorm.DB) *MusicGenerationRepository {
	return &MusicGenerationRepository{db: db}
}

// Create 创建音乐生成记录
func (r *MusicGenerationRepository) Create(ctx context.Context, musicGen *entity.MusicGeneration) error {
	return r.db.WithContext(ctx).Create(musicGen).Error
}

// GetByID 根据ID获取音乐生成记录
func (r *MusicGenerationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.MusicGeneration, error) {
	var musicGen entity.MusicGeneration
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Audio").
		First(&musicGen, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &musicGen, nil
}

// Update 更新音乐生成记录
func (r *MusicGenerationRepository) Update(ctx context.Context, musicGen *entity.MusicGeneration) error {
	return r.db.WithContext(ctx).Save(musicGen).Error
}

// Delete 删除音乐生成记录
func (r *MusicGenerationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.MusicGeneration{}, "id = ?", id).Error
}

// GetByProjectID 根据项目ID获取音乐生成记录列表
func (r *MusicGenerationRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.MusicGeneration, error) {
	var musicGens []*entity.MusicGeneration
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Audio").
		Find(&musicGens, "project_id = ?", projectID).Error
	return musicGens, err
}
