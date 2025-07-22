package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ScriptType 剧本类型
type ScriptType string

const (
	ScriptTypeOriginal ScriptType = "original"
	ScriptTypeAdapted  ScriptType = "adapted"
)

// Script 剧本实体
type Script struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID   uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	Project     Project        `json:"project" gorm:"foreignKey:ProjectID"`
	Title       string         `json:"title" gorm:"not null"`
	Type        ScriptType     `json:"type" gorm:"default:'adapted'"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	SourceText  string         `json:"source_text" gorm:"type:text"` // 原始小说文本
	Metadata    string         `json:"metadata" gorm:"type:text"`    // JSON元数据
	Version     int            `json:"version" gorm:"default:1"`
	Status      string         `json:"status" gorm:"default:'draft'"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Script) TableName() string {
	return "scripts"
}

// BeforeCreate 创建前的钩子
func (s *Script) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// ScriptScene 剧本场景实体
type ScriptScene struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ScriptID    uuid.UUID      `json:"script_id" gorm:"type:uuid;not null"`
	Script      Script         `json:"script" gorm:"foreignKey:ScriptID"`
	SceneNumber int            `json:"scene_number" gorm:"not null"`
	Title       string         `json:"title"`
	Location    string         `json:"location"`
	TimeOfDay   string         `json:"time_of_day"`
	Description string         `json:"description" gorm:"type:text"`
	Dialogue    string         `json:"dialogue" gorm:"type:text"`
	Actions     string         `json:"actions" gorm:"type:text"`
	Characters  string         `json:"characters" gorm:"type:text"` // JSON数组
	Duration    int            `json:"duration"` // 预估时长（秒）
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (ScriptScene) TableName() string {
	return "script_scenes"
}

// BeforeCreate 创建前的钩子
func (ss *ScriptScene) BeforeCreate(tx *gorm.DB) error {
	if ss.ID == uuid.Nil {
		ss.ID = uuid.New()
	}
	return nil
}
