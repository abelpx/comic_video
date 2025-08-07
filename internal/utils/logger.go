package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger 全局日志实例
var Logger *logrus.Logger

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
	LogLevelPanic LogLevel = "panic"
)

// LogConfig 日志配置
type LogConfig struct {
	Level      LogLevel `json:"level" yaml:"level"`
	Format     string   `json:"format" yaml:"format"`     // json, text
	Output     string   `json:"output" yaml:"output"`     // stdout, file, both
	FilePath   string   `json:"file_path" yaml:"file_path"`
	MaxSize    int      `json:"max_size" yaml:"max_size"`       // MB
	MaxBackups int      `json:"max_backups" yaml:"max_backups"` // 保留的日志文件数量
	MaxAge     int      `json:"max_age" yaml:"max_age"`         // 保留天数
	Compress   bool     `json:"compress" yaml:"compress"`       // 是否压缩
}

// DefaultLogConfig 默认日志配置
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:      LogLevelInfo,
		Format:     "json",
		Output:     "both",
		FilePath:   "logs/app.log",
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   true,
	}
}

// InitLogger 初始化日志系统
func InitLogger(config *LogConfig) error {
	Logger = logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(string(config.Level))
	if err != nil {
		return fmt.Errorf("invalid log level: %s", config.Level)
	}
	Logger.SetLevel(level)

	// 设置日志格式
	if config.Format == "json" {
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
		})
	}

	// 设置输出
	var writers []io.Writer

	if config.Output == "stdout" || config.Output == "both" {
		writers = append(writers, os.Stdout)
	}

	if config.Output == "file" || config.Output == "both" {
		// 确保日志目录存在
		logDir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		writers = append(writers, file)
	}

	if len(writers) > 1 {
		Logger.SetOutput(io.MultiWriter(writers...))
	} else if len(writers) == 1 {
		Logger.SetOutput(writers[0])
	}

	// 添加调用者信息
	Logger.SetReportCaller(true)

	return nil
}

// ContextLogger 带上下文的日志记录器
type ContextLogger struct {
	*logrus.Entry
}

// NewContextLogger 创建带上下文的日志记录器
func NewContextLogger(ctx context.Context) *ContextLogger {
	entry := Logger.WithContext(ctx)
	
	// 添加请求ID
	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		entry = entry.WithField("request_id", requestID)
	}
	
	// 添加用户ID
	if userID := GetUserIDFromContext(ctx); userID != "" {
		entry = entry.WithField("user_id", userID)
	}
	
	return &ContextLogger{Entry: entry}
}

// WithFields 添加字段
func (cl *ContextLogger) WithFields(fields map[string]interface{}) *ContextLogger {
	return &ContextLogger{Entry: cl.Entry.WithFields(fields)}
}

// WithField 添加单个字段
func (cl *ContextLogger) WithField(key string, value interface{}) *ContextLogger {
	return &ContextLogger{Entry: cl.Entry.WithField(key, value)}
}

// WithError 添加错误信息
func (cl *ContextLogger) WithError(err error) *ContextLogger {
	return &ContextLogger{Entry: cl.Entry.WithError(err)}
}

// LoggerMiddleware Gin日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)

		// 获取请求ID
		requestID := GetRequestID(c)

		// 构建日志字段
		fields := logrus.Fields{
			"request_id":   requestID,
			"method":       c.Request.Method,
			"path":         path,
			"query":        raw,
			"status":       c.Writer.Status(),
			"latency":      latency,
			"latency_ms":   float64(latency.Nanoseconds()) / 1e6,
			"client_ip":    c.ClientIP(),
			"user_agent":   c.Request.UserAgent(),
			"request_size": c.Request.ContentLength,
			"response_size": c.Writer.Size(),
		}

		// 添加用户ID（如果存在）
		if userID := GetUserID(c); userID != "" {
			fields["user_id"] = userID
		}

		// 根据状态码选择日志级别
		entry := Logger.WithFields(fields)
		
		if len(c.Errors) > 0 {
			// 有错误时记录错误日志
			entry.WithField("errors", c.Errors.String()).Error("Request completed with errors")
		} else if c.Writer.Status() >= 500 {
			entry.Error("Request completed with server error")
		} else if c.Writer.Status() >= 400 {
			entry.Warn("Request completed with client error")
		} else {
			entry.Info("Request completed successfully")
		}
	}
}

// 便捷的日志函数
func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

// LogError 记录错误日志
func LogError(ctx context.Context, err error, message string, fields ...map[string]interface{}) {
	logger := NewContextLogger(ctx)
	
	entry := logger.WithError(err)
	
	// 添加调用者信息
	if pc, file, line, ok := runtime.Caller(1); ok {
		funcName := runtime.FuncForPC(pc).Name()
		entry = entry.WithFields(logrus.Fields{
			"caller_file": filepath.Base(file),
			"caller_line": line,
			"caller_func": filepath.Base(funcName),
		})
	}
	
	// 添加额外字段
	for _, fieldMap := range fields {
		entry = entry.WithFields(fieldMap)
	}
	
	entry.Error(message)
}

// LogInfo 记录信息日志
func LogInfo(ctx context.Context, message string, fields ...map[string]interface{}) {
	logger := NewContextLogger(ctx)
	
	entry := logger.Entry
	
	// 添加额外字段
	for _, fieldMap := range fields {
		entry = entry.WithFields(fieldMap)
	}
	
	entry.Info(message)
}

// LogWarn 记录警告日志
func LogWarn(ctx context.Context, message string, fields ...map[string]interface{}) {
	logger := NewContextLogger(ctx)
	
	entry := logger.Entry
	
	// 添加额外字段
	for _, fieldMap := range fields {
		entry = entry.WithFields(fieldMap)
	}
	
	entry.Warn(message)
}

// 辅助函数：从上下文获取请求ID
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// 辅助函数：从上下文获取用户ID
func GetUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// 辅助函数：从Gin上下文获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// 辅助函数：从Gin上下文获取用户ID
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	return ""
}
