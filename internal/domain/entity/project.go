package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Project 项目实体
type Project struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
	TemplateID  *uuid.UUID     `json:"template_id" gorm:"type:uuid"`
	Template    *Template      `json:"template" gorm:"foreignKey:TemplateID"`
	Config      string         `json:"config" gorm:"type:text"` // JSON配置
	Status      string         `json:"status" gorm:"default:'draft'"`
	Thumbnail   string         `json:"thumbnail"`
	Duration    int            `json:"duration"` // 视频时长（秒）
	Resolution  string         `json:"resolution"` // 分辨率
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}

// BeforeCreate 创建前的钩子
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ProjectShare 项目分享实体
// 用于支持项目生成分享链接、有效期、密码保护等
type ProjectShare struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID  uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	Token      string         `json:"token" gorm:"uniqueIndex;not null"` // 分享唯一token
	Password   string         `json:"password"` // 密码hash，可为空
	ExpiresAt  *time.Time     `json:"expires_at"` // 过期时间，可为空
	CreatedBy  uuid.UUID      `json:"created_by" gorm:"type:uuid;not null"`
	Status     string         `json:"status" gorm:"default:'active'"` // active/expired/canceled
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

func (ProjectShare) TableName() string {
	return "project_shares"
} 