package utils

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts     int           `json:"max_attempts"`     // 最大重试次数
	InitialDelay    time.Duration `json:"initial_delay"`    // 初始延迟
	MaxDelay        time.Duration `json:"max_delay"`        // 最大延迟
	BackoffFactor   float64       `json:"backoff_factor"`   // 退避因子
	Jitter          bool          `json:"jitter"`           // 是否添加随机抖动
	RetryableErrors []ErrorCode   `json:"retryable_errors"` // 可重试的错误码
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
		RetryableErrors: []ErrorCode{
			ErrCodeInternalError,
			ErrCodeAIServiceUnavailable,
			ErrCodeAITaskFailed,
		},
	}
}

// RetryableFunc 可重试的函数类型
type RetryableFunc func() error

// RetryWithResult 带返回值的可重试函数类型
type RetryWithResult[T any] func() (T, error)

// Retry 执行重试逻辑
func Retry(ctx context.Context, config *RetryConfig, fn RetryableFunc) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 执行函数
		err := fn()
		if err == nil {
			return nil // 成功，无需重试
		}

		lastErr = err

		// 检查是否为可重试错误
		if !isRetryableError(err, config.RetryableErrors) {
			LogWarn(ctx, "Error is not retryable, stopping retry", map[string]interface{}{
				"error":   err.Error(),
				"attempt": attempt,
			})
			return err
		}

		// 如果是最后一次尝试，不再延迟
		if attempt == config.MaxAttempts {
			break
		}

		// 计算延迟时间
		delay := calculateDelay(attempt, config)
		
		LogWarn(ctx, "Retrying after error", map[string]interface{}{
			"error":     err.Error(),
			"attempt":   attempt,
			"max_attempts": config.MaxAttempts,
			"delay":     delay,
		})

		// 等待延迟时间
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return WrapError(ErrCodeInternalError, 
		fmt.Sprintf("Max retry attempts (%d) exceeded", config.MaxAttempts), lastErr)
}

// RetryWithResultFunc 执行带返回值的重试逻辑
func RetryWithResultFunc[T any](ctx context.Context, config *RetryConfig, fn RetryWithResult[T]) (T, error) {
	var zero T
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
		}

		// 执行函数
		result, err := fn()
		if err == nil {
			return result, nil // 成功，无需重试
		}

		lastErr = err

		// 检查是否为可重试错误
		if !isRetryableError(err, config.RetryableErrors) {
			LogWarn(ctx, "Error is not retryable, stopping retry", map[string]interface{}{
				"error":   err.Error(),
				"attempt": attempt,
			})
			return zero, err
		}

		// 如果是最后一次尝试，不再延迟
		if attempt == config.MaxAttempts {
			break
		}

		// 计算延迟时间
		delay := calculateDelay(attempt, config)
		
		LogWarn(ctx, "Retrying after error", map[string]interface{}{
			"error":     err.Error(),
			"attempt":   attempt,
			"max_attempts": config.MaxAttempts,
			"delay":     delay,
		})

		// 等待延迟时间
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delay):
		}
	}

	return zero, WrapError(ErrCodeInternalError, 
		fmt.Sprintf("Max retry attempts (%d) exceeded", config.MaxAttempts), lastErr)
}

// calculateDelay 计算延迟时间
func calculateDelay(attempt int, config *RetryConfig) time.Duration {
	// 指数退避算法
	delay := float64(config.InitialDelay) * math.Pow(config.BackoffFactor, float64(attempt-1))
	
	// 限制最大延迟
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// 添加随机抖动避免雷群效应
	if config.Jitter {
		jitter := rand.Float64() * 0.1 * delay // 10%的随机抖动
		delay += jitter
	}

	return time.Duration(delay)
}

// isRetryableError 检查错误是否可重试
func isRetryableError(err error, retryableErrors []ErrorCode) bool {
	if appErr, ok := err.(*AppError); ok {
		for _, code := range retryableErrors {
			if appErr.Code == code {
				return true
			}
		}
	}
	return false
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	failureCount    int
	lastFailureTime time.Time
	state           CircuitState
}

// CircuitState 熔断器状态
type CircuitState int

const (
	CircuitClosed CircuitState = iota // 关闭状态，正常工作
	CircuitOpen                       // 开启状态，拒绝请求
	CircuitHalfOpen                   // 半开状态，尝试恢复
)

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
	}
}

// Execute 执行函数，带熔断保护
func (cb *CircuitBreaker) Execute(ctx context.Context, fn RetryableFunc) error {
	// 检查熔断器状态
	if cb.state == CircuitOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			cb.failureCount = 0
		} else {
			return NewAppError(ErrCodeServiceUnavailable, "Circuit breaker is open")
		}
	}

	// 执行函数
	err := fn()
	
	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

// onFailure 处理失败
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	
	if cb.failureCount >= cb.maxFailures {
		cb.state = CircuitOpen
	}
}

// onSuccess 处理成功
func (cb *CircuitBreaker) onSuccess() {
	cb.failureCount = 0
	cb.state = CircuitClosed
}

// GetState 获取熔断器状态
func (cb *CircuitBreaker) GetState() CircuitState {
	return cb.state
}

// TaskRetryManager 任务重试管理器
type TaskRetryManager struct {
	retryConfigs map[string]*RetryConfig
	circuitBreakers map[string]*CircuitBreaker
}

// NewTaskRetryManager 创建任务重试管理器
func NewTaskRetryManager() *TaskRetryManager {
	return &TaskRetryManager{
		retryConfigs: make(map[string]*RetryConfig),
		circuitBreakers: make(map[string]*CircuitBreaker),
	}
}

// SetRetryConfig 设置特定任务的重试配置
func (trm *TaskRetryManager) SetRetryConfig(taskType string, config *RetryConfig) {
	trm.retryConfigs[taskType] = config
}

// SetCircuitBreaker 设置特定任务的熔断器
func (trm *TaskRetryManager) SetCircuitBreaker(taskType string, cb *CircuitBreaker) {
	trm.circuitBreakers[taskType] = cb
}

// ExecuteTask 执行任务，带重试和熔断保护
func (trm *TaskRetryManager) ExecuteTask(ctx context.Context, taskType string, fn RetryableFunc) error {
	// 获取重试配置
	config := trm.retryConfigs[taskType]
	if config == nil {
		config = DefaultRetryConfig()
	}

	// 获取熔断器
	cb := trm.circuitBreakers[taskType]
	if cb == nil {
		cb = NewCircuitBreaker(5, 30*time.Second) // 默认熔断器
		trm.circuitBreakers[taskType] = cb
	}

	// 执行任务
	return cb.Execute(ctx, func() error {
		return Retry(ctx, config, fn)
	})
}

// 预定义的任务重试管理器
var GlobalTaskRetryManager = NewTaskRetryManager()

// 初始化默认重试配置
func init() {
	// AI服务重试配置
	aiRetryConfig := &RetryConfig{
		MaxAttempts:   5,
		InitialDelay:  2 * time.Second,
		MaxDelay:      60 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
		RetryableErrors: []ErrorCode{
			ErrCodeAIServiceUnavailable,
			ErrCodeAITaskFailed,
			ErrCodeInternalError,
		},
	}
	GlobalTaskRetryManager.SetRetryConfig("ai_generation", aiRetryConfig)

	// 图像生成重试配置
	imageRetryConfig := &RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  5 * time.Second,
		MaxDelay:      120 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
		RetryableErrors: []ErrorCode{
			ErrCodeAIServiceUnavailable,
			ErrCodeAITaskFailed,
		},
	}
	GlobalTaskRetryManager.SetRetryConfig("image_generation", imageRetryConfig)

	// 语音合成重试配置
	ttsRetryConfig := &RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  3 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 1.5,
		Jitter:        true,
		RetryableErrors: []ErrorCode{
			ErrCodeAIServiceUnavailable,
			ErrCodeInternalError,
		},
	}
	GlobalTaskRetryManager.SetRetryConfig("tts_generation", ttsRetryConfig)

	// 视频合成重试配置
	videoRetryConfig := &RetryConfig{
		MaxAttempts:   2,
		InitialDelay:  10 * time.Second,
		MaxDelay:      60 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        false,
		RetryableErrors: []ErrorCode{
			ErrCodeInternalError,
		},
	}
	GlobalTaskRetryManager.SetRetryConfig("video_composition", videoRetryConfig)
}
