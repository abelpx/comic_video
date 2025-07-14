package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Video 视频实体
type Video struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
	ProjectID   *uuid.UUID     `json:"project_id" gorm:"type:uuid"`
	Project     *Project       `json:"project" gorm:"foreignKey:ProjectID"`
	FileName    string         `json:"file_name" gorm:"not null"`
	OriginalName string        `json:"original_name" gorm:"not null"`
	FilePath    string         `json:"file_path" gorm:"not null"`
	FileSize    int64          `json:"file_size"`
	Duration    float64        `json:"duration"` // 视频时长（秒）
	Width       int            `json:"width"`
	Height      int            `json:"height"`
	Format      string         `json:"format"`
	Codec       string         `json:"codec"`
	Bitrate     int            `json:"bitrate"`
	FPS         float64        `json:"fps"`
	Thumbnail   string         `json:"thumbnail"`
	Status      string         `json:"status" gorm:"default:'uploading'"`
	Type        string         `json:"type" gorm:"default:'video'"` // video, audio, image
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Video) TableName() string {
	return "videos"
}

// BeforeCreate 创建前的钩子
func (v *Video) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
} 