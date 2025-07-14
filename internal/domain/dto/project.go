package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name        string                 `json:"name" binding:"required,min=1,max=100"`
	Description string                 `json:"description" binding:"max=500"`
	Settings    map[string]interface{} `json:"settings"`
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	Name        string                 `json:"name" binding:"max=100"`
	Description string                 `json:"description" binding:"max=500"`
	Settings    map[string]interface{} `json:"settings"`
}

// ProjectResponse 项目响应
type ProjectResponse struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	UserID      uuid.UUID              `json:"user_id"`
	Status      string                 `json:"status"`
	Settings    map[string]interface{} `json:"settings"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ShareProjectRequest 分享项目请求
type ShareProjectRequest struct {
	ExpiresAt *time.Time `json:"expires_at"`
	Password  string     `json:"password"`
}

// ShareResponse 分享响应
type ShareResponse struct {
	ShareURL  string     `json:"share_url"`
	Token     string     `json:"token"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// ListProjectsRequest 获取项目列表请求
type ListProjectsRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
}

// ListProjectsResponse 项目列表响应
type ListProjectsResponse struct {
	Projects []*ProjectResponse `json:"projects"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
} 