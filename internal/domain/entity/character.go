package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Character 角色实体
type Character struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID       uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	Project         Project        `json:"project" gorm:"foreignKey:ProjectID"`
	Name            string         `json:"name" gorm:"not null"`
	Description     string         `json:"description" gorm:"type:text"`
	Age             string         `json:"age"`
	Gender          string         `json:"gender"`
	FacialFeatures  string         `json:"facial_features" gorm:"type:text"`
	HairStyle       string         `json:"hair_style"`
	Clothing        string         `json:"clothing" gorm:"type:text"`
	BodyType        string         `json:"body_type"`
	Personality     string         `json:"personality" gorm:"type:text"`
	Importance      string         `json:"importance"`      // 角色重要性：主角/重要配角/次要角色/群众角色
	ScreenTime      string         `json:"screen_time"`     // 预估出场频率：高/中/低
	ArtStyle        string         `json:"art_style"`
	VisualKeywords  string         `json:"visual_keywords" gorm:"type:text"`
	ColorScheme     string         `json:"color_scheme"`
	ConsistencyPrompt string       `json:"consistency_prompt" gorm:"type:text"`
	Seed            int64          `json:"seed"`
	ReferenceImages string         `json:"reference_images" gorm:"type:text"` // JSON数组
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Character) TableName() string {
	return "characters"
}

// BeforeCreate 创建前的钩子
func (c *Character) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// CharacterImage 角色图像实体
type CharacterImage struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CharacterID uuid.UUID      `json:"character_id" gorm:"type:uuid;not null"`
	Character   Character      `json:"character" gorm:"foreignKey:CharacterID"`
	ImageURL    string         `json:"image_url" gorm:"not null"`
	ImageType   string         `json:"image_type"` // portrait, full_body, expression, etc.
	Prompt      string         `json:"prompt" gorm:"type:text"`
	Seed        int64          `json:"seed"`
	Parameters  string         `json:"parameters" gorm:"type:text"` // JSON参数
	IsReference bool           `json:"is_reference" gorm:"default:false"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (CharacterImage) TableName() string {
	return "character_images"
}

// BeforeCreate 创建前的钩子
func (ci *CharacterImage) BeforeCreate(tx *gorm.DB) error {
	if ci.ID == uuid.Nil {
		ci.ID = uuid.New()
	}
	return nil
}
