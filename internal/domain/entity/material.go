package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Material 素材实体
type Material struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Category    string         `json:"category" gorm:"not null"` // music, image, video, effect
	Type        string         `json:"type" gorm:"not null"` // 具体类型
	FileName    string         `json:"file_name" gorm:"not null"`
	FilePath    string         `json:"file_path" gorm:"not null"`
	FileSize    int64          `json:"file_size"`
	Duration    float64        `json:"duration"` // 音频/视频时长
	Width       int            `json:"width"` // 图片/视频宽度
	Height      int            `json:"height"` // 图片/视频高度
	Format      string         `json:"format"`
	Thumbnail   string         `json:"thumbnail"`
	Tags        string         `json:"tags"` // 标签，逗号分隔
	IsPublic    bool           `json:"is_public" gorm:"default:true"`
	IsPremium   bool           `json:"is_premium" gorm:"default:false"`
	DownloadCount int          `json:"download_count" gorm:"default:0"`
	Rating      float64        `json:"rating" gorm:"default:0"`
	Status      string         `json:"status" gorm:"default:'active'"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Material) TableName() string {
	return "materials"
}

// BeforeCreate 创建前的钩子
func (m *Material) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
} 