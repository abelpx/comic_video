package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category" binding:"required"`
	Tags        string `json:"tags"`
	Thumbnail   string `json:"thumbnail"`
	Preview     string `json:"preview"`
	Config      map[string]interface{} `json:"config" binding:"required"`
	IsPublic    bool   `json:"is_public"`
	IsPremium   bool   `json:"is_premium"`
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Tags        string `json:"tags"`
	Thumbnail   string `json:"thumbnail"`
	Preview     string `json:"preview"`
	Config      map[string]interface{} `json:"config"`
	IsPublic    *bool  `json:"is_public"`
	IsPremium   *bool  `json:"is_premium"`
}

// TemplateResponse 模板详情响应
type TemplateResponse struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Tags        string                 `json:"tags"`
	Thumbnail   string                 `json:"thumbnail"`
	Preview     string                 `json:"preview"`
	Config      map[string]interface{} `json:"config"`
	IsPublic    bool                   `json:"is_public"`
	IsPremium   bool                   `json:"is_premium"`
	DownloadCount int                  `json:"download_count"`
	Rating      float64                `json:"rating"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ListTemplatesRequest 模板列表请求
type ListTemplatesRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Category string `form:"category"`
	Keyword  string `form:"keyword"`
	IsPublic *bool  `form:"is_public"`
	IsPremium *bool `form:"is_premium"`
}

// ListTemplatesResponse 模板列表响应
type ListTemplatesResponse struct {
	Templates []*TemplateResponse `json:"templates"`
	Total     int64               `json:"total"`
	Page      int                 `json:"page"`
	PageSize  int                 `json:"page_size"`
}

// ApplyTemplateRequest 应用模板请求
type ApplyTemplateRequest struct {
	ProjectID uuid.UUID `json:"project_id" binding:"required"`
	Config    map[string]interface{} `json:"config"`
}

// ApplyTemplateResponse 应用模板响应
type ApplyTemplateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
} 