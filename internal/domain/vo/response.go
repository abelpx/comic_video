package vo

import (
	"time"
)

// SuccessResponse 成功响应
type SuccessResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Errors    interface{} `json:"errors,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	Size     int         `json:"size"`
	Pages    int         `json:"pages"`
	HasNext  bool        `json:"has_next"`
	HasPrev  bool        `json:"has_prev"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) SuccessResponse {
	return SuccessResponse{
		Code:      200,
		Message:   "success",
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) ErrorResponse {
	return ErrorResponse{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}
} 