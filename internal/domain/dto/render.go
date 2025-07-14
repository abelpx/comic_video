package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateRenderRequest 创建渲染任务请求
type CreateRenderRequest struct {
	ProjectID  uuid.UUID `json:"project_id" binding:"required"`
	Name       string    `json:"name" binding:"required"`
	Quality    string    `json:"quality" binding:"oneof=high medium low"`
	Format     string    `json:"format"`
	Resolution string    `json:"resolution"`
}

// RenderResponse 渲染任务详情响应
type RenderResponse struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	ProjectID   uuid.UUID  `json:"project_id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Progress    int        `json:"progress"`
	OutputPath  string     `json:"output_path"`
	OutputSize  int64      `json:"output_size"`
	Duration    float64    `json:"duration"`
	Resolution  string     `json:"resolution"`
	Format      string     `json:"format"`
	Quality     string     `json:"quality"`
	Error       string     `json:"error"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ListRendersRequest 渲染任务列表请求
type ListRendersRequest struct {
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
	ProjectID string `form:"project_id"`
	Status    string `form:"status"`
}

// ListRendersResponse 渲染任务列表响应
type ListRendersResponse struct {
	Renders  []*RenderResponse `json:"renders"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// RenderStatusResponse 渲染状态响应
type RenderStatusResponse struct {
	Status   string `json:"status"`
	Progress int    `json:"progress"`
	Error    string `json:"error"`
}

// DownloadRenderResponse 渲染结果下载响应
type DownloadRenderResponse struct {
	URL string `json:"url"`
} 