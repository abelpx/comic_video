package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Scene 场景实体
type Scene struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID    uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	Project      Project        `json:"project" gorm:"foreignKey:ProjectID"`
	Name         string         `json:"name" gorm:"not null"`
	Description  string         `json:"description" gorm:"type:text"`
	Location     string         `json:"location"`
	TimeOfDay    string         `json:"time_of_day"`
	Weather      string         `json:"weather"`
	Season       string         `json:"season"`
	ArtStyle     string         `json:"art_style"`
	ColorPalette string         `json:"color_palette"`
	Lighting     string         `json:"lighting"`
	Atmosphere   string         `json:"atmosphere"`
	CameraAngle  string         `json:"camera_angle"`
	Composition  string         `json:"composition"`
	Background   string         `json:"background" gorm:"type:text"`
	Foreground   string         `json:"foreground" gorm:"type:text"`
	Props        string         `json:"props" gorm:"type:text"`
	Prompt       string         `json:"prompt" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Scene) TableName() string {
	return "scenes"
}

// BeforeCreate 创建前的钩子
func (s *Scene) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// SceneImage 场景图像实体
type SceneImage struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SceneID    uuid.UUID      `json:"scene_id" gorm:"type:uuid;not null"`
	Scene      Scene          `json:"scene" gorm:"foreignKey:SceneID"`
	ImageURL   string         `json:"image_url" gorm:"not null"`
	ImageType  string         `json:"image_type"` // concept, reference, final, etc.
	Prompt     string         `json:"prompt" gorm:"type:text"`
	Seed       int64          `json:"seed"`
	Parameters string         `json:"parameters" gorm:"type:text"` // JSON参数
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (SceneImage) TableName() string {
	return "scene_images"
}

// BeforeCreate 创建前的钩子
func (si *SceneImage) BeforeCreate(tx *gorm.DB) error {
	if si.ID == uuid.Nil {
		si.ID = uuid.New()
	}
	return nil
}
