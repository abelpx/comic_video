package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Render 渲染任务实体
type Render struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
	ProjectID   uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	Project     Project        `json:"project" gorm:"foreignKey:ProjectID"`
	Name        string         `json:"name" gorm:"not null"`
	Status      string         `json:"status" gorm:"default:'pending'"` // pending, processing, completed, failed
	Progress    int            `json:"progress" gorm:"default:0"` // 进度百分比
	OutputPath  string         `json:"output_path"`
	OutputSize  int64          `json:"output_size"`
	Duration    float64        `json:"duration"` // 输出视频时长
	Resolution  string         `json:"resolution"`
	Format      string         `json:"format"`
	Quality     string         `json:"quality"` // high, medium, low
	Error       string         `json:"error"` // 错误信息
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Render) TableName() string {
	return "renders"
}

// BeforeCreate 创建前的钩子
func (r *Render) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
} 