package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Template 模板实体
type Template struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Category    string         `json:"category" gorm:"not null"`
	Thumbnail   string         `json:"thumbnail"`
	Preview     string         `json:"preview"` // 预览视频URL
	Config      string         `json:"config" gorm:"type:text"` // JSON配置
	Duration    int            `json:"duration"` // 模板时长（秒）
	Resolution  string         `json:"resolution"` // 分辨率
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
func (Template) TableName() string {
	return "templates"
}

// BeforeCreate 创建前的钩子
func (t *Template) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
} 