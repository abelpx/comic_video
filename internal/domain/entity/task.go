package entity

import (
	"time"
	"github.com/google/uuid"
)

// TaskType 定义任务类型
const (
	TaskTypeRender = "render"
	TaskTypeAI     = "ai"
	TaskTypeVideo  = "video"
	// 可扩展更多类型
)

// TaskStatus 定义任务状态
const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

// Task 通用任务实体
// 支持多类型任务的进度、状态、参数、结果等统一管理
// 可用于队列、进度推送、历史记录等
// 参数和结果建议用 JSON 字符串存储，便于灵活扩展

type Task struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Type      string    `json:"type" gorm:"type:varchar(32);index"`      // 任务类型
	Status    string    `json:"status" gorm:"type:varchar(32);index"`    // 任务状态
	Progress  int       `json:"progress"`
	UserID    string    `json:"user_id" gorm:"type:varchar(64);index"`   // 用户ID
	Params    string    `json:"params" gorm:"type:text"`    // 任务参数（JSON字符串）
	Result    string    `json:"result" gorm:"type:text"`    // 任务结果（JSON字符串）
	Error     string    `json:"error" gorm:"type:text"`     // 错误信息
	Steps     string    `json:"steps" gorm:"type:text"`     // 步骤详情（JSON字符串）
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TaskStep 任务步骤
type TaskStep struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"` // pending, processing, completed, failed
	Progress    int       `json:"progress"`
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Result      string    `json:"result,omitempty"`
	Error       string    `json:"error,omitempty"`
	Description string    `json:"description,omitempty"`
}