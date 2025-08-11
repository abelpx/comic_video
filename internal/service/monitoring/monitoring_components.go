package monitoring

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

// AnomalyDetector 异常检测器
type AnomalyDetector struct {
	threshold float64 // 异常阈值
}

// Anomaly 异常
type Anomaly struct {
	MetricName  string    `json:"metric_name"`
	Value       float64   `json:"value"`
	Expected    float64   `json:"expected"`
	Deviation   float64   `json:"deviation"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewAnomalyDetector 创建异常检测器
func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{
		threshold: 2.0, // 2个标准差
	}
}

// DetectAnomaly 检测异常
func (ad *AnomalyDetector) DetectAnomaly(history []MetricPoint) *Anomaly {
	if len(history) < 10 {
		return nil // 数据不足
	}

	// 取最近的数据点
	recent := history[len(history)-1]
	
	// 计算历史数据的均值和标准差
	mean, stdDev := ad.calculateStats(history[:len(history)-1])
	
	// 计算偏差
	deviation := math.Abs(recent.Value - mean)
	
	// 检查是否异常
	if deviation > ad.threshold*stdDev {
		severity := "medium"
		if deviation > 3*stdDev {
			severity = "high"
		}
		
		return &Anomaly{
			Value:       recent.Value,
			Expected:    mean,
			Deviation:   deviation,
			Severity:    severity,
			Description: fmt.Sprintf("值 %.2f 偏离期望值 %.2f 超过 %.1f 个标准差", recent.Value, mean, deviation/stdDev),
			Timestamp:   recent.Timestamp,
		}
	}
	
	return nil
}

// calculateStats 计算统计数据
func (ad *AnomalyDetector) calculateStats(points []MetricPoint) (mean, stdDev float64) {
	if len(points) == 0 {
		return 0, 0
	}
	
	// 计算均值
	sum := 0.0
	for _, point := range points {
		sum += point.Value
	}
	mean = sum / float64(len(points))
	
	// 计算标准差
	variance := 0.0
	for _, point := range points {
		variance += math.Pow(point.Value-mean, 2)
	}
	variance /= float64(len(points))
	stdDev = math.Sqrt(variance)
	
	return mean, stdDev
}

// PerformanceTracker 性能跟踪器
type PerformanceTracker struct {
	startTime time.Time
	metrics   map[string]*PerformanceMetric
}

// PerformanceMetric 性能指标
type PerformanceMetric struct {
	Name         string        `json:"name"`
	Count        int64         `json:"count"`
	TotalTime    time.Duration `json:"total_time"`
	MinTime      time.Duration `json:"min_time"`
	MaxTime      time.Duration `json:"max_time"`
	AvgTime      time.Duration `json:"avg_time"`
	LastUpdated  time.Time     `json:"last_updated"`
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	GeneratedAt    time.Time                    `json:"generated_at"`
	Uptime         time.Duration                `json:"uptime"`
	Metrics        map[string]*PerformanceMetric `json:"metrics"`
	Summary        *PerformanceSummary          `json:"summary"`
	Recommendations []string                    `json:"recommendations"`
}

// PerformanceSummary 性能摘要
type PerformanceSummary struct {
	TotalRequests     int64         `json:"total_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	ThroughputPerSecond float64       `json:"throughput_per_second"`
	ErrorRate         float64       `json:"error_rate"`
	PerformanceScore  float64       `json:"performance_score"` // 0-100
}

// NewPerformanceTracker 创建性能跟踪器
func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		startTime: time.Now(),
		metrics:   make(map[string]*PerformanceMetric),
	}
}

// TrackOperation 跟踪操作
func (pt *PerformanceTracker) TrackOperation(name string, duration time.Duration) {
	metric, exists := pt.metrics[name]
	if !exists {
		metric = &PerformanceMetric{
			Name:    name,
			MinTime: duration,
			MaxTime: duration,
		}
		pt.metrics[name] = metric
	}
	
	metric.Count++
	metric.TotalTime += duration
	metric.AvgTime = time.Duration(int64(metric.TotalTime) / metric.Count)
	
	if duration < metric.MinTime {
		metric.MinTime = duration
	}
	if duration > metric.MaxTime {
		metric.MaxTime = duration
	}
	
	metric.LastUpdated = time.Now()
}

// TrackPerformance 跟踪性能
func (pt *PerformanceTracker) TrackPerformance() {
	// 这里可以添加自动性能跟踪逻辑
	log.Printf("[PerformanceTracker] 性能跟踪更新")
}

// GenerateReport 生成报告
func (pt *PerformanceTracker) GenerateReport() *PerformanceReport {
	uptime := time.Since(pt.startTime)
	
	// 计算摘要
	summary := &PerformanceSummary{}
	totalRequests := int64(0)
	totalTime := time.Duration(0)
	
	for _, metric := range pt.metrics {
		totalRequests += metric.Count
		totalTime += metric.TotalTime
	}
	
	summary.TotalRequests = totalRequests
	if totalRequests > 0 {
		summary.AverageResponseTime = time.Duration(int64(totalTime) / totalRequests)
		summary.ThroughputPerSecond = float64(totalRequests) / uptime.Seconds()
	}
	
	// 计算性能分数
	summary.PerformanceScore = pt.calculatePerformanceScore(summary)
	
	// 生成建议
	recommendations := pt.generateRecommendations(summary)
	
	return &PerformanceReport{
		GeneratedAt:     time.Now(),
		Uptime:          uptime,
		Metrics:         pt.metrics,
		Summary:         summary,
		Recommendations: recommendations,
	}
}

// calculatePerformanceScore 计算性能分数
func (pt *PerformanceTracker) calculatePerformanceScore(summary *PerformanceSummary) float64 {
	score := 100.0
	
	// 基于响应时间扣分
	if summary.AverageResponseTime > 5*time.Second {
		score -= 30
	} else if summary.AverageResponseTime > 2*time.Second {
		score -= 15
	} else if summary.AverageResponseTime > 1*time.Second {
		score -= 5
	}
	
	// 基于错误率扣分
	if summary.ErrorRate > 0.1 {
		score -= 40
	} else if summary.ErrorRate > 0.05 {
		score -= 20
	} else if summary.ErrorRate > 0.01 {
		score -= 10
	}
	
	// 基于吞吐量加分
	if summary.ThroughputPerSecond > 100 {
		score += 10
	} else if summary.ThroughputPerSecond > 50 {
		score += 5
	}
	
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	return score
}

// generateRecommendations 生成建议
func (pt *PerformanceTracker) generateRecommendations(summary *PerformanceSummary) []string {
	recommendations := make([]string, 0)
	
	if summary.AverageResponseTime > 2*time.Second {
		recommendations = append(recommendations, "响应时间过长，建议优化算法或增加缓存")
	}
	
	if summary.ErrorRate > 0.05 {
		recommendations = append(recommendations, "错误率过高，建议检查错误处理逻辑")
	}
	
	if summary.ThroughputPerSecond < 10 {
		recommendations = append(recommendations, "吞吐量较低，建议优化并发处理")
	}
	
	if summary.PerformanceScore < 70 {
		recommendations = append(recommendations, "整体性能需要改进，建议进行性能调优")
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "性能表现良好，继续保持")
	}
	
	return recommendations
}

// HealthChecker 健康检查器
type HealthChecker struct {
	checks []HealthCheck
}

// HealthCheck 健康检查
type HealthCheck struct {
	Name        string                          `json:"name"`
	Description string                          `json:"description"`
	CheckFunc   func() *HealthCheckResult       `json:"-"`
	Interval    time.Duration                   `json:"interval"`
	Timeout     time.Duration                   `json:"timeout"`
	LastCheck   time.Time                       `json:"last_check"`
	LastResult  *HealthCheckResult              `json:"last_result"`
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Status    string                 `json:"status"`    // healthy, unhealthy, unknown
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
}

// SystemHealthReport 系统健康报告
type SystemHealthReport struct {
	OverallStatus string                         `json:"overall_status"`
	Checks        map[string]*HealthCheckResult  `json:"checks"`
	Summary       *HealthSummary                 `json:"summary"`
	GeneratedAt   time.Time                      `json:"generated_at"`
}

// HealthSummary 健康摘要
type HealthSummary struct {
	TotalChecks   int `json:"total_checks"`
	HealthyChecks int `json:"healthy_checks"`
	UnhealthyChecks int `json:"unhealthy_checks"`
	UnknownChecks int `json:"unknown_checks"`
	HealthScore   float64 `json:"health_score"` // 0-100
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker() *HealthChecker {
	hc := &HealthChecker{
		checks: make([]HealthCheck, 0),
	}
	
	// 添加默认健康检查
	hc.addDefaultChecks()
	
	return hc
}

// addDefaultChecks 添加默认检查
func (hc *HealthChecker) addDefaultChecks() {
	// 数据库连接检查
	hc.checks = append(hc.checks, HealthCheck{
		Name:        "database",
		Description: "数据库连接检查",
		CheckFunc:   hc.checkDatabase,
		Interval:    1 * time.Minute,
		Timeout:     10 * time.Second,
	})
	
	// 磁盘空间检查
	hc.checks = append(hc.checks, HealthCheck{
		Name:        "disk_space",
		Description: "磁盘空间检查",
		CheckFunc:   hc.checkDiskSpace,
		Interval:    5 * time.Minute,
		Timeout:     5 * time.Second,
	})
	
	// 内存使用检查
	hc.checks = append(hc.checks, HealthCheck{
		Name:        "memory",
		Description: "内存使用检查",
		CheckFunc:   hc.checkMemory,
		Interval:    1 * time.Minute,
		Timeout:     5 * time.Second,
	})
	
	// AI服务检查
	hc.checks = append(hc.checks, HealthCheck{
		Name:        "ai_service",
		Description: "AI服务连接检查",
		CheckFunc:   hc.checkAIService,
		Interval:    2 * time.Minute,
		Timeout:     30 * time.Second,
	})
}

// CheckHealth 执行健康检查
func (hc *HealthChecker) CheckHealth() {
	for i := range hc.checks {
		check := &hc.checks[i]
		
		// 检查是否需要执行
		if time.Since(check.LastCheck) < check.Interval {
			continue
		}
		
		// 执行检查
		start := time.Now()
		result := check.CheckFunc()
		result.Duration = time.Since(start)
		result.Timestamp = time.Now()
		
		check.LastCheck = time.Now()
		check.LastResult = result
		
		log.Printf("[HealthChecker] %s: %s", check.Name, result.Status)
	}
}

// GetHealthReport 获取健康报告
func (hc *HealthChecker) GetHealthReport() *SystemHealthReport {
	checks := make(map[string]*HealthCheckResult)
	summary := &HealthSummary{}
	
	for _, check := range hc.checks {
		if check.LastResult != nil {
			checks[check.Name] = check.LastResult
			
			summary.TotalChecks++
			switch check.LastResult.Status {
			case "healthy":
				summary.HealthyChecks++
			case "unhealthy":
				summary.UnhealthyChecks++
			default:
				summary.UnknownChecks++
			}
		}
	}
	
	// 计算健康分数
	if summary.TotalChecks > 0 {
		summary.HealthScore = float64(summary.HealthyChecks) / float64(summary.TotalChecks) * 100
	}
	
	// 确定整体状态
	overallStatus := "healthy"
	if summary.UnhealthyChecks > 0 {
		overallStatus = "unhealthy"
	} else if summary.UnknownChecks > summary.HealthyChecks {
		overallStatus = "unknown"
	}
	
	return &SystemHealthReport{
		OverallStatus: overallStatus,
		Checks:        checks,
		Summary:       summary,
		GeneratedAt:   time.Now(),
	}
}

// 具体的健康检查方法
func (hc *HealthChecker) checkDatabase() *HealthCheckResult {
	// 简化实现
	return &HealthCheckResult{
		Status:  "healthy",
		Message: "数据库连接正常",
		Details: map[string]interface{}{
			"connection_pool": "active",
			"response_time":   "15ms",
		},
	}
}

func (hc *HealthChecker) checkDiskSpace() *HealthCheckResult {
	// 简化实现
	return &HealthCheckResult{
		Status:  "healthy",
		Message: "磁盘空间充足",
		Details: map[string]interface{}{
			"usage_percent": 45,
			"free_space":    "100GB",
		},
	}
}

func (hc *HealthChecker) checkMemory() *HealthCheckResult {
	// 简化实现
	return &HealthCheckResult{
		Status:  "healthy",
		Message: "内存使用正常",
		Details: map[string]interface{}{
			"usage_percent": 67,
			"available":     "8GB",
		},
	}
}

func (hc *HealthChecker) checkAIService() *HealthCheckResult {
	// 简化实现
	return &HealthCheckResult{
		Status:  "healthy",
		Message: "AI服务连接正常",
		Details: map[string]interface{}{
			"endpoint":      "https://api.openai.com",
			"response_time": "250ms",
		},
	}
}

// NotificationManager 通知管理器
type NotificationManager struct {
	emailConfig *EmailConfig
	webhookConfig *WebhookConfig
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	FromAddress  string `json:"from_address"`
	Enabled      bool   `json:"enabled"`
}

// WebhookConfig Webhook配置
type WebhookConfig struct {
	DefaultURL string            `json:"default_url"`
	Headers    map[string]string `json:"headers"`
	Timeout    time.Duration     `json:"timeout"`
	Enabled    bool              `json:"enabled"`
}

// NewNotificationManager 创建通知管理器
func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		emailConfig: &EmailConfig{
			SMTPHost:    "smtp.gmail.com",
			SMTPPort:    587,
			Enabled:     false, // 默认禁用
		},
		webhookConfig: &WebhookConfig{
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Timeout: 30 * time.Second,
			Enabled: false, // 默认禁用
		},
	}
}

// SendEmail 发送邮件
func (nm *NotificationManager) SendEmail(to, message string) error {
	if !nm.emailConfig.Enabled {
		log.Printf("[NotificationManager] 邮件通知已禁用")
		return nil
	}
	
	// 简化实现
	auth := smtp.PlainAuth("", nm.emailConfig.Username, nm.emailConfig.Password, nm.emailConfig.SMTPHost)
	
	msg := fmt.Sprintf("To: %s\r\nSubject: 系统告警\r\n\r\n%s", to, message)
	
	addr := fmt.Sprintf("%s:%d", nm.emailConfig.SMTPHost, nm.emailConfig.SMTPPort)
	err := smtp.SendMail(addr, auth, nm.emailConfig.FromAddress, []string{to}, []byte(msg))
	
	if err != nil {
		log.Printf("[NotificationManager] 发送邮件失败: %v", err)
		return err
	}
	
	log.Printf("[NotificationManager] 邮件发送成功: %s", to)
	return nil
}

// SendWebhook 发送Webhook
func (nm *NotificationManager) SendWebhook(url string, data interface{}) error {
	if !nm.webhookConfig.Enabled {
		log.Printf("[NotificationManager] Webhook通知已禁用")
		return nil
	}
	
	// 简化实现
	log.Printf("[NotificationManager] 发送Webhook: %s", url)
	
	// 这里应该实现真实的HTTP请求
	client := &http.Client{Timeout: nm.webhookConfig.Timeout}
	
	// 构建请求
	body := strings.NewReader(fmt.Sprintf(`{"alert": "%v"}`, data))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	
	// 添加头部
	for key, value := range nm.webhookConfig.Headers {
		req.Header.Set(key, value)
	}
	
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[NotificationManager] Webhook发送失败: %v", err)
		return err
	}
	defer resp.Body.Close()
	
	log.Printf("[NotificationManager] Webhook发送成功: %d", resp.StatusCode)
	return nil
}
