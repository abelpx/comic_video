package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
	WorkflowStatusPending    WorkflowStatus = "pending"
	WorkflowStatusRunning    WorkflowStatus = "running"
	WorkflowStatusCompleted  WorkflowStatus = "completed"
	WorkflowStatusFailed     WorkflowStatus = "failed"
	WorkflowStatusCancelled  WorkflowStatus = "cancelled"
)

// WorkflowStep 工作流步骤
type WorkflowStep string

const (
	StepTextInput       WorkflowStep = "text_input"
	StepScriptAdapt     WorkflowStep = "script_adapt"
	StepCharacterGen    WorkflowStep = "character_gen"
	StepSceneGen        WorkflowStep = "scene_gen"
	StepStoryboard      WorkflowStep = "storyboard"
	StepVoiceGen        WorkflowStep = "voice_gen"
	StepMusicGen        WorkflowStep = "music_gen"
	StepVideoEdit       WorkflowStep = "video_edit"
	StepPublish         WorkflowStep = "publish"
)

// Workflow 工作流实体
type Workflow struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID   uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	Project     Project        `json:"project" gorm:"foreignKey:ProjectID"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Status      WorkflowStatus `json:"status" gorm:"default:'pending'"`
	CurrentStep WorkflowStep   `json:"current_step" gorm:"default:'text_input'"`
	Config      string         `json:"config" gorm:"type:text"` // JSON配置
	Progress    int            `json:"progress" gorm:"default:0"` // 进度百分比
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Workflow) TableName() string {
	return "workflows"
}

// BeforeCreate 创建前的钩子
func (w *Workflow) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// WorkflowTask 工作流任务实体
type WorkflowTask struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkflowID uuid.UUID      `json:"workflow_id" gorm:"type:uuid;not null"`
	Workflow   Workflow       `json:"workflow" gorm:"foreignKey:WorkflowID"`
	Step       WorkflowStep   `json:"step" gorm:"not null"`
	Name       string         `json:"name" gorm:"not null"`
	Status     WorkflowStatus `json:"status" gorm:"default:'pending'"`
	Input      string         `json:"input" gorm:"type:text"`  // JSON输入数据
	Output     string         `json:"output" gorm:"type:text"` // JSON输出数据
	Error      string         `json:"error" gorm:"type:text"`  // 错误信息
	Progress   int            `json:"progress" gorm:"default:0"`
	StartedAt  *time.Time     `json:"started_at"`
	CompletedAt *time.Time    `json:"completed_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (WorkflowTask) TableName() string {
	return "workflow_tasks"
}

// BeforeCreate 创建前的钩子
func (wt *WorkflowTask) BeforeCreate(tx *gorm.DB) error {
	if wt.ID == uuid.Nil {
		wt.ID = uuid.New()
	}
	return nil
}
