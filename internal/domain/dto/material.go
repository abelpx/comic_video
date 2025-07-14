package dto

import (
	"time"

	"github.com/google/uuid"
)

// UploadMaterialRequest 上传素材请求
type UploadMaterialRequest struct {
	Name        string `form:"name" binding:"required"`
	Description string `form:"description"`
	Category    string `form:"category" binding:"required,oneof=music image video effect"`
	Type        string `form:"type" binding:"required"`
	IsPublic    bool   `form:"is_public"`
	IsPremium   bool   `form:"is_premium"`
	Tags        string `form:"tags"`
	File        any    `form:"file" binding:"required"` // 实际上传文件
}

// UpdateMaterialRequest 更新素材请求
type UpdateMaterialRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    *bool  `json:"is_public"`
	IsPremium   *bool  `json:"is_premium"`
	Tags        string `json:"tags"`
}

// MaterialResponse 素材详情响应
type MaterialResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Type        string    `json:"type"`
	FileName    string    `json:"file_name"`
	FilePath    string    `json:"file_path"`
	FileSize    int64     `json:"file_size"`
	Duration    float64   `json:"duration"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	Format      string    `json:"format"`
	Thumbnail   string    `json:"thumbnail"`
	Tags        string    `json:"tags"`
	IsPublic    bool      `json:"is_public"`
	IsPremium   bool      `json:"is_premium"`
	DownloadCount int     `json:"download_count"`
	Rating      float64   `json:"rating"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListMaterialsRequest 素材列表请求
type ListMaterialsRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Category string `form:"category"`
	Type     string `form:"type"`
	Keyword  string `form:"keyword"`
	IsPublic *bool  `form:"is_public"`
	IsPremium *bool `form:"is_premium"`
}

// ListMaterialsResponse 素材列表响应
type ListMaterialsResponse struct {
	Materials []*MaterialResponse `json:"materials"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
} 