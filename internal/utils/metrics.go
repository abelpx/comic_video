package utils

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsCollector 指标收集器
type MetricsCollector struct {
	mu                sync.RWMutex
	requestCount      map[string]int64
	requestDuration   map[string][]time.Duration
	errorCount        map[string]int64
	activeConnections int64
	startTime         time.Time
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		requestCount:    make(map[string]int64),
		requestDuration: make(map[string][]time.Duration),
		errorCount:      make(map[string]int64),
		startTime:       time.Now(),
	}
}

// 全局指标收集器实例
var GlobalMetrics = NewMetricsCollector()

// RecordRequest 记录请求指标
func (mc *MetricsCollector) RecordRequest(endpoint string, duration time.Duration, statusCode int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	// 记录请求次数
	mc.requestCount[endpoint]++
	
	// 记录请求耗时
	if mc.requestDuration[endpoint] == nil {
		mc.requestDuration[endpoint] = make([]time.Duration, 0)
	}
	mc.requestDuration[endpoint] = append(mc.requestDuration[endpoint], duration)
	
	// 保持最近1000次请求的耗时记录
	if len(mc.requestDuration[endpoint]) > 1000 {
		mc.requestDuration[endpoint] = mc.requestDuration[endpoint][1:]
	}
	
	// 记录错误次数
	if statusCode >= 400 {
		mc.errorCount[endpoint]++
	}
}

// RecordError 记录错误指标
func (mc *MetricsCollector) RecordError(endpoint string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.errorCount[endpoint]++
}

// IncrementActiveConnections 增加活跃连接数
func (mc *MetricsCollector) IncrementActiveConnections() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.activeConnections++
}

// DecrementActiveConnections 减少活跃连接数
func (mc *MetricsCollector) DecrementActiveConnections() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if mc.activeConnections > 0 {
		mc.activeConnections--
	}
}

// GetRequestCount 获取请求次数
func (mc *MetricsCollector) GetRequestCount(endpoint string) int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	return mc.requestCount[endpoint]
}

// GetErrorCount 获取错误次数
func (mc *MetricsCollector) GetErrorCount(endpoint string) int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	return mc.errorCount[endpoint]
}

// GetAverageResponseTime 获取平均响应时间
func (mc *MetricsCollector) GetAverageResponseTime(endpoint string) time.Duration {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	durations := mc.requestDuration[endpoint]
	if len(durations) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	
	return total / time.Duration(len(durations))
}

// GetPercentileResponseTime 获取百分位响应时间
func (mc *MetricsCollector) GetPercentileResponseTime(endpoint string, percentile float64) time.Duration {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	durations := mc.requestDuration[endpoint]
	if len(durations) == 0 {
		return 0
	}
	
	// 简单的百分位计算（实际应该排序）
	index := int(float64(len(durations)) * percentile / 100)
	if index >= len(durations) {
		index = len(durations) - 1
	}
	
	return durations[index]
}

// GetActiveConnections 获取活跃连接数
func (mc *MetricsCollector) GetActiveConnections() int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	return mc.activeConnections
}

// GetUptime 获取运行时间
func (mc *MetricsCollector) GetUptime() time.Duration {
	return time.Since(mc.startTime)
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	// CPU相关
	CPUUsage    float64 `json:"cpu_usage"`
	CPUCores    int     `json:"cpu_cores"`
	
	// 内存相关
	MemoryUsage     uint64  `json:"memory_usage"`
	MemoryTotal     uint64  `json:"memory_total"`
	MemoryPercent   float64 `json:"memory_percent"`
	
	// Goroutine相关
	GoroutineCount  int `json:"goroutine_count"`
	
	// GC相关
	GCCount         uint32        `json:"gc_count"`
	GCPauseTotal    time.Duration `json:"gc_pause_total"`
	GCPauseAverage  time.Duration `json:"gc_pause_average"`
	
	// 堆内存相关
	HeapAlloc       uint64 `json:"heap_alloc"`
	HeapSys         uint64 `json:"heap_sys"`
	HeapInuse       uint64 `json:"heap_inuse"`
	HeapReleased    uint64 `json:"heap_released"`
	HeapObjects     uint64 `json:"heap_objects"`
}

// GetSystemMetrics 获取系统指标
func GetSystemMetrics() *SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	metrics := &SystemMetrics{
		CPUCores:       runtime.NumCPU(),
		MemoryUsage:    m.Alloc,
		MemoryTotal:    m.Sys,
		MemoryPercent:  float64(m.Alloc) / float64(m.Sys) * 100,
		GoroutineCount: runtime.NumGoroutine(),
		GCCount:        m.NumGC,
		GCPauseTotal:   time.Duration(m.PauseTotalNs),
		HeapAlloc:      m.HeapAlloc,
		HeapSys:        m.HeapSys,
		HeapInuse:      m.HeapInuse,
		HeapReleased:   m.HeapReleased,
		HeapObjects:    m.HeapObjects,
	}
	
	// 计算GC平均暂停时间
	if m.NumGC > 0 {
		metrics.GCPauseAverage = time.Duration(m.PauseTotalNs) / time.Duration(m.NumGC)
	}
	
	return metrics
}

// ApplicationMetrics 应用指标
type ApplicationMetrics struct {
	// 请求相关
	TotalRequests    int64             `json:"total_requests"`
	RequestsPerSecond float64          `json:"requests_per_second"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	
	// 错误相关
	TotalErrors      int64   `json:"total_errors"`
	ErrorRate        float64 `json:"error_rate"`
	
	// 连接相关
	ActiveConnections int64 `json:"active_connections"`
	
	// 运行时间
	Uptime           time.Duration `json:"uptime"`
	
	// 端点统计
	EndpointStats    map[string]*EndpointMetrics `json:"endpoint_stats"`
}

// EndpointMetrics 端点指标
type EndpointMetrics struct {
	RequestCount    int64         `json:"request_count"`
	ErrorCount      int64         `json:"error_count"`
	AverageTime     time.Duration `json:"average_time"`
	P95Time         time.Duration `json:"p95_time"`
	P99Time         time.Duration `json:"p99_time"`
	ErrorRate       float64       `json:"error_rate"`
}

// GetApplicationMetrics 获取应用指标
func GetApplicationMetrics() *ApplicationMetrics {
	GlobalMetrics.mu.RLock()
	defer GlobalMetrics.mu.RUnlock()
	
	metrics := &ApplicationMetrics{
		ActiveConnections: GlobalMetrics.activeConnections,
		Uptime:           GlobalMetrics.GetUptime(),
		EndpointStats:    make(map[string]*EndpointMetrics),
	}
	
	var totalRequests, totalErrors int64
	
	// 计算各端点指标
	for endpoint, count := range GlobalMetrics.requestCount {
		errorCount := GlobalMetrics.errorCount[endpoint]
		avgTime := GlobalMetrics.GetAverageResponseTime(endpoint)
		p95Time := GlobalMetrics.GetPercentileResponseTime(endpoint, 95)
		p99Time := GlobalMetrics.GetPercentileResponseTime(endpoint, 99)
		
		errorRate := float64(0)
		if count > 0 {
			errorRate = float64(errorCount) / float64(count) * 100
		}
		
		metrics.EndpointStats[endpoint] = &EndpointMetrics{
			RequestCount: count,
			ErrorCount:   errorCount,
			AverageTime:  avgTime,
			P95Time:      p95Time,
			P99Time:      p99Time,
			ErrorRate:    errorRate,
		}
		
		totalRequests += count
		totalErrors += errorCount
	}
	
	metrics.TotalRequests = totalRequests
	metrics.TotalErrors = totalErrors
	
	// 计算总体错误率
	if totalRequests > 0 {
		metrics.ErrorRate = float64(totalErrors) / float64(totalRequests) * 100
	}
	
	// 计算RPS（简化计算）
	uptime := metrics.Uptime.Seconds()
	if uptime > 0 {
		metrics.RequestsPerSecond = float64(totalRequests) / uptime
	}
	
	return metrics
}

// MetricsMiddleware 指标收集中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 增加活跃连接数
		GlobalMetrics.IncrementActiveConnections()
		defer GlobalMetrics.DecrementActiveConnections()
		
		// 处理请求
		c.Next()
		
		// 记录指标
		duration := time.Since(start)
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}
		
		GlobalMetrics.RecordRequest(endpoint, duration, c.Writer.Status())
	}
}

// HealthCheck 健康检查结果
type HealthCheck struct {
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Uptime      time.Duration          `json:"uptime"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Checks      map[string]CheckResult `json:"checks"`
}

// CheckResult 检查结果
type CheckResult struct {
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency,omitempty"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	checks map[string]func() CheckResult
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]func() CheckResult),
	}
}

// AddCheck 添加检查项
func (hc *HealthChecker) AddCheck(name string, check func() CheckResult) {
	hc.checks[name] = check
}

// Check 执行健康检查
func (hc *HealthChecker) Check() *HealthCheck {
	result := &HealthCheck{
		Status:      "healthy",
		Timestamp:   time.Now(),
		Uptime:      GlobalMetrics.GetUptime(),
		Version:     "1.0.0", // 应该从配置或构建信息获取
		Environment: "development", // 应该从配置获取
		Checks:      make(map[string]CheckResult),
	}
	
	// 执行所有检查
	for name, check := range hc.checks {
		checkResult := check()
		result.Checks[name] = checkResult
		
		// 如果有任何检查失败，整体状态为不健康
		if checkResult.Status != "healthy" {
			result.Status = "unhealthy"
		}
	}
	
	return result
}

// 全局健康检查器
var GlobalHealthChecker = NewHealthChecker()

// RegisterDefaultHealthChecks 注册默认健康检查
func RegisterDefaultHealthChecks() {
	// 内存使用检查
	GlobalHealthChecker.AddCheck("memory", func() CheckResult {
		metrics := GetSystemMetrics()
		if metrics.MemoryPercent > 90 {
			return CheckResult{
				Status:  "unhealthy",
				Message: "Memory usage too high",
			}
		}
		return CheckResult{
			Status: "healthy",
		}
	})
	
	// Goroutine数量检查
	GlobalHealthChecker.AddCheck("goroutines", func() CheckResult {
		count := runtime.NumGoroutine()
		if count > 1000 {
			return CheckResult{
				Status:  "unhealthy",
				Message: "Too many goroutines",
			}
		}
		return CheckResult{
			Status: "healthy",
		}
	})
}
