package dto

import (
	"time"

	"github.com/google/uuid"
)

// UploadVideoRequest 上传视频请求
type UploadVideoRequest struct {
	ProjectID   *uuid.UUID `form:"project_id"`
	Type        string     `form:"type" binding:"required,oneof=video audio image"`
	Description string     `form:"description"`
}

// UpdateVideoRequest 更新视频请求
type UpdateVideoRequest struct {
	Description string `json:"description"`
	Status      string `json:"status"`
	Thumbnail   string `json:"thumbnail"`
}

// VideoResponse 视频详情响应
type VideoResponse struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	ProjectID    *uuid.UUID `json:"project_id"`
	FileName     string     `json:"file_name"`
	OriginalName string     `json:"original_name"`
	FilePath     string     `json:"file_path"`
	FileSize     int64      `json:"file_size"`
	Duration     float64    `json:"duration"`
	Width        int        `json:"width"`
	Height       int        `json:"height"`
	Format       string     `json:"format"`
	Codec        string     `json:"codec"`
	Bitrate      int        `json:"bitrate"`
	FPS          float64    `json:"fps"`
	Thumbnail    string     `json:"thumbnail"`
	Status       string     `json:"status"`
	Type         string     `json:"type"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ListVideosRequest 视频列表请求
type ListVideosRequest struct {
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
	ProjectID string `form:"project_id"`
	Type      string `form:"type"`
	Status    string `form:"status"`
}

// ListVideosResponse 视频列表响应
type ListVideosResponse struct {
	Videos   []*VideoResponse `json:"videos"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
} 