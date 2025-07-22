package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Storyboard 分镜实体
type Storyboard struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID   uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	Project     Project        `json:"project" gorm:"foreignKey:ProjectID"`
	ScriptID    uuid.UUID      `json:"script_id" gorm:"type:uuid;not null"`
	Script      Script         `json:"script" gorm:"foreignKey:ScriptID"`
	Title       string         `json:"title" gorm:"not null"`
	Description string         `json:"description" gorm:"type:text"`
	TotalFrames int            `json:"total_frames"`
	Duration    int            `json:"duration"` // 总时长（秒）
	FrameRate   int            `json:"frame_rate" gorm:"default:24"`
	Resolution  string         `json:"resolution" gorm:"default:'1920x1080'"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Storyboard) TableName() string {
	return "storyboards"
}

// BeforeCreate 创建前的钩子
func (s *Storyboard) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// StoryboardFrame 分镜帧实体
type StoryboardFrame struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StoryboardID uuid.UUID      `json:"storyboard_id" gorm:"type:uuid;not null"`
	Storyboard   Storyboard     `json:"storyboard" gorm:"foreignKey:StoryboardID"`
	FrameNumber  int            `json:"frame_number" gorm:"not null"`
	SceneID      *uuid.UUID     `json:"scene_id" gorm:"type:uuid"`
	Scene        *Scene         `json:"scene" gorm:"foreignKey:SceneID"`
	Title        string         `json:"title"`
	Description  string         `json:"description" gorm:"type:text"`
	Dialogue     string         `json:"dialogue" gorm:"type:text"`
	Action       string         `json:"action" gorm:"type:text"`
	CameraAngle  string         `json:"camera_angle"`
	CameraMove   string         `json:"camera_move"`
	ShotType     string         `json:"shot_type"` // close-up, medium, wide, etc.
	Duration     float64        `json:"duration"` // 帧时长（秒）
	Transition   string         `json:"transition"` // 转场效果
	Characters   string         `json:"characters" gorm:"type:text"` // JSON数组
	Props        string         `json:"props" gorm:"type:text"` // JSON数组
	ImageURL     string         `json:"image_url"`
	Prompt       string         `json:"prompt" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (StoryboardFrame) TableName() string {
	return "storyboard_frames"
}

// BeforeCreate 创建前的钩子
func (sf *StoryboardFrame) BeforeCreate(tx *gorm.DB) error {
	if sf.ID == uuid.Nil {
		sf.ID = uuid.New()
	}
	return nil
}
