package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AudioType 音频类型
type AudioType string

const (
	AudioTypeVoice AudioType = "voice"
	AudioTypeMusic AudioType = "music"
	AudioTypeSFX   AudioType = "sfx"
)

// Audio 音频实体
type Audio struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID   uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	Project     Project        `json:"project" gorm:"foreignKey:ProjectID"`
	Name        string         `json:"name" gorm:"not null"`
	Type        AudioType      `json:"type" gorm:"not null"`
	Description string         `json:"description" gorm:"type:text"`
	FileURL     string         `json:"file_url" gorm:"not null"`
	Duration    float64        `json:"duration"` // 时长（秒）
	SampleRate  int            `json:"sample_rate"`
	Bitrate     int            `json:"bitrate"`
	Format      string         `json:"format"` // mp3, wav, etc.
	FileSize    int64          `json:"file_size"`
	Metadata    string         `json:"metadata" gorm:"type:text"` // JSON元数据
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Audio) TableName() string {
	return "audios"
}

// BeforeCreate 创建前的钩子
func (a *Audio) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// VoiceGeneration 语音生成实体
type VoiceGeneration struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID   uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	Project     Project        `json:"project" gorm:"foreignKey:ProjectID"`
	CharacterID *uuid.UUID     `json:"character_id" gorm:"type:uuid"`
	Character   *Character     `json:"character" gorm:"foreignKey:CharacterID"`
	Text        string         `json:"text" gorm:"type:text;not null"`
	VoiceModel  string         `json:"voice_model"`
	Language    string         `json:"language" gorm:"default:'zh-CN'"`
	Speed       float64        `json:"speed" gorm:"default:1.0"`
	Pitch       float64        `json:"pitch" gorm:"default:1.0"`
	Volume      float64        `json:"volume" gorm:"default:1.0"`
	Emotion     string         `json:"emotion"`
	AudioID     *uuid.UUID     `json:"audio_id" gorm:"type:uuid"`
	Audio       *Audio         `json:"audio" gorm:"foreignKey:AudioID"`
	Status      string         `json:"status" gorm:"default:'pending'"`
	Error       string         `json:"error" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (VoiceGeneration) TableName() string {
	return "voice_generations"
}

// BeforeCreate 创建前的钩子
func (vg *VoiceGeneration) BeforeCreate(tx *gorm.DB) error {
	if vg.ID == uuid.Nil {
		vg.ID = uuid.New()
	}
	return nil
}

// MusicGeneration 音乐生成实体
type MusicGeneration struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID   uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	Project     Project        `json:"project" gorm:"foreignKey:ProjectID"`
	Prompt      string         `json:"prompt" gorm:"type:text;not null"`
	Style       string         `json:"style"`
	Mood        string         `json:"mood"`
	Tempo       string         `json:"tempo"`
	Duration    int            `json:"duration"` // 时长（秒）
	AudioID     *uuid.UUID     `json:"audio_id" gorm:"type:uuid"`
	Audio       *Audio         `json:"audio" gorm:"foreignKey:AudioID"`
	Status      string         `json:"status" gorm:"default:'pending'"`
	Error       string         `json:"error" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (MusicGeneration) TableName() string {
	return "music_generations"
}

// BeforeCreate 创建前的钩子
func (mg *MusicGeneration) BeforeCreate(tx *gorm.DB) error {
	if mg.ID == uuid.Nil {
		mg.ID = uuid.New()
	}
	return nil
}
