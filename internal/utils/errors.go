package utils

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorCode 错误码类型
type ErrorCode int

// 定义错误码常量
const (
	// 通用错误码
	ErrCodeSuccess         ErrorCode = 0
	ErrCodeInternalError   ErrorCode = 1000
	ErrCodeInvalidParams   ErrorCode = 1001
	ErrCodeUnauthorized    ErrorCode = 1002
	ErrCodeForbidden       ErrorCode = 1003
	ErrCodeNotFound        ErrorCode = 1004
	ErrCodeConflict        ErrorCode = 1005
	ErrCodeTooManyRequests ErrorCode = 1006

	// 用户相关错误码
	ErrCodeUserNotFound     ErrorCode = 2001
	ErrCodeUserExists       ErrorCode = 2002
	ErrCodeInvalidPassword  ErrorCode = 2003
	ErrCodeUserDisabled     ErrorCode = 2004
	ErrCodeEmailExists      ErrorCode = 2005
	ErrCodeUsernameExists   ErrorCode = 2006

	// 项目相关错误码
	ErrCodeProjectNotFound ErrorCode = 3001
	ErrCodeProjectExists   ErrorCode = 3002
	ErrCodeProjectAccess   ErrorCode = 3003

	// AI服务相关错误码
	ErrCodeAIServiceUnavailable ErrorCode = 4001
	ErrCodeAIQuotaExceeded      ErrorCode = 4002
	ErrCodeAITaskFailed         ErrorCode = 4003
	ErrCodeAIInvalidPrompt      ErrorCode = 4004

	// 文件相关错误码
	ErrCodeFileNotFound    ErrorCode = 5001
	ErrCodeFileUploadFailed ErrorCode = 5002
	ErrCodeFileTooBig      ErrorCode = 5003
	ErrCodeInvalidFileType ErrorCode = 5004
)

// AppError 应用错误结构
type AppError struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
	Stack     string    `json:"stack,omitempty"`
	cause     error
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *AppError) Unwrap() error {
	return e.cause
}

// NewAppError 创建新的应用错误
func NewAppError(code ErrorCode, message string, details ...string) *AppError {
	err := &AppError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}
	
	if len(details) > 0 {
		err.Details = details[0]
	}
	
	// 获取调用栈信息
	if _, file, line, ok := runtime.Caller(1); ok {
		err.Stack = fmt.Sprintf("%s:%d", file, line)
	}
	
	return err
}

// WrapError 包装现有错误
func WrapError(code ErrorCode, message string, cause error) *AppError {
	err := &AppError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		cause:     cause,
	}
	
	if cause != nil {
		err.Details = cause.Error()
	}
	
	// 获取调用栈信息
	if _, file, line, ok := runtime.Caller(1); ok {
		err.Stack = fmt.Sprintf("%s:%d", file, line)
	}
	
	return err
}

// WithRequestID 添加请求ID
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// WithDetails 添加详细信息
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// 错误码到HTTP状态码的映射
var errorCodeToHTTPStatus = map[ErrorCode]int{
	ErrCodeSuccess:              http.StatusOK,
	ErrCodeInternalError:        http.StatusInternalServerError,
	ErrCodeInvalidParams:        http.StatusBadRequest,
	ErrCodeUnauthorized:         http.StatusUnauthorized,
	ErrCodeForbidden:           http.StatusForbidden,
	ErrCodeNotFound:            http.StatusNotFound,
	ErrCodeConflict:            http.StatusConflict,
	ErrCodeTooManyRequests:     http.StatusTooManyRequests,
	ErrCodeUserNotFound:        http.StatusNotFound,
	ErrCodeUserExists:          http.StatusConflict,
	ErrCodeInvalidPassword:     http.StatusUnauthorized,
	ErrCodeUserDisabled:        http.StatusForbidden,
	ErrCodeEmailExists:         http.StatusConflict,
	ErrCodeUsernameExists:      http.StatusConflict,
	ErrCodeProjectNotFound:     http.StatusNotFound,
	ErrCodeProjectExists:       http.StatusConflict,
	ErrCodeProjectAccess:       http.StatusForbidden,
	ErrCodeAIServiceUnavailable: http.StatusServiceUnavailable,
	ErrCodeAIQuotaExceeded:     http.StatusTooManyRequests,
	ErrCodeAITaskFailed:        http.StatusInternalServerError,
	ErrCodeAIInvalidPrompt:     http.StatusBadRequest,
	ErrCodeFileNotFound:        http.StatusNotFound,
	ErrCodeFileUploadFailed:    http.StatusInternalServerError,
	ErrCodeFileTooBig:          http.StatusRequestEntityTooLarge,
	ErrCodeInvalidFileType:     http.StatusBadRequest,
}

// GetHTTPStatus 获取HTTP状态码
func (e *AppError) GetHTTPStatus() int {
	if status, exists := errorCodeToHTTPStatus[e.Code]; exists {
		return status
	}
	return http.StatusInternalServerError
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Success   bool      `json:"success"`
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
	Path      string    `json:"path,omitempty"`
}

// HandleError 统一错误处理中间件
func HandleError() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			var appErr *AppError
			var httpStatus int
			var response ErrorResponse

			// 判断错误类型
			if ae, ok := err.Err.(*AppError); ok {
				appErr = ae
				httpStatus = ae.GetHTTPStatus()
			} else {
				// 包装未知错误
				appErr = WrapError(ErrCodeInternalError, "Internal server error", err.Err)
				httpStatus = http.StatusInternalServerError
			}

			// 添加请求ID
			requestID := GetRequestID(c)
			if requestID != "" {
				appErr = appErr.WithRequestID(requestID)
			}

			// 构建响应
			response = ErrorResponse{
				Success:   false,
				Code:      appErr.Code,
				Message:   appErr.Message,
				Details:   appErr.Details,
				Timestamp: appErr.Timestamp,
				RequestID: appErr.RequestID,
				Path:      c.Request.URL.Path,
			}

			c.JSON(httpStatus, response)
			c.Abort()
		}
	}
}

// AbortWithError 中止请求并返回错误
func AbortWithError(c *gin.Context, err *AppError) {
	c.Error(err)
	c.Abort()
}

// 预定义的常用错误
var (
	ErrInternalServer = NewAppError(ErrCodeInternalError, "Internal server error")
	ErrInvalidParams  = NewAppError(ErrCodeInvalidParams, "Invalid parameters")
	ErrUnauthorized   = NewAppError(ErrCodeUnauthorized, "Unauthorized")
	ErrForbidden      = NewAppError(ErrCodeForbidden, "Forbidden")
	ErrNotFound       = NewAppError(ErrCodeNotFound, "Resource not found")
	ErrConflict       = NewAppError(ErrCodeConflict, "Resource conflict")
	
	ErrUserNotFound    = NewAppError(ErrCodeUserNotFound, "User not found")
	ErrUserExists      = NewAppError(ErrCodeUserExists, "User already exists")
	ErrInvalidPassword = NewAppError(ErrCodeInvalidPassword, "Invalid password")
	ErrUserDisabled    = NewAppError(ErrCodeUserDisabled, "User account is disabled")
	
	ErrProjectNotFound = NewAppError(ErrCodeProjectNotFound, "Project not found")
	ErrProjectExists   = NewAppError(ErrCodeProjectExists, "Project already exists")
	ErrProjectAccess   = NewAppError(ErrCodeProjectAccess, "Access denied to project")
	
	ErrAIServiceUnavailable = NewAppError(ErrCodeAIServiceUnavailable, "AI service is unavailable")
	ErrAIQuotaExceeded      = NewAppError(ErrCodeAIQuotaExceeded, "AI quota exceeded")
	ErrAITaskFailed         = NewAppError(ErrCodeAITaskFailed, "AI task failed")
	
	ErrFileNotFound     = NewAppError(ErrCodeFileNotFound, "File not found")
	ErrFileUploadFailed = NewAppError(ErrCodeFileUploadFailed, "File upload failed")
	ErrFileTooBig       = NewAppError(ErrCodeFileTooBig, "File size too large")
	ErrInvalidFileType  = NewAppError(ErrCodeInvalidFileType, "Invalid file type")
)
