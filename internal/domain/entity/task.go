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
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`      // 任务类型
	Status    string    `json:"status"`    // 任务状态
	Progress  int       `json:"progress"`  // 进度百分比
	Params    string    `json:"params"`    // 任务参数（JSON字符串）
	Result    string    `json:"result"`    // 任务结果（JSON字符串）
	Error     string    `json:"error"`     // 错误信息
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
} 