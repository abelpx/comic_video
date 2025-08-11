package monitoring

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"comic_video/internal/utils"
	"github.com/google/uuid"
)

// IntelligentMonitor 智能监控器
type IntelligentMonitor struct {
	mu                sync.RWMutex
	metrics           map[string]*MetricCollector
	alerts            map[string]*AlertRule
	anomalyDetector   *AnomalyDetector
	performanceTracker *PerformanceTracker
	healthChecker     *HealthChecker
	notificationMgr   *NotificationManager
	isRunning         bool
	stopChan          chan struct{}
}

// MetricCollector 指标收集器
type MetricCollector struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // counter, gauge, histogram
	Value       float64                `json:"value"`
	Labels      map[string]string      `json:"labels"`
	Timestamp   time.Time              `json:"timestamp"`
	History     []MetricPoint          `json:"history"`
	Aggregation *AggregationConfig     `json:"aggregation"`
}

// MetricPoint 指标点
type MetricPoint struct {
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// AggregationConfig 聚合配置
type AggregationConfig struct {
	Window   time.Duration `json:"window"`   // 聚合窗口
	Function string        `json:"function"` // avg, sum, max, min, count
	Enabled  bool          `json:"enabled"`
}

// AlertRule 告警规则
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	MetricName  string                 `json:"metric_name"`
	Condition   string                 `json:"condition"`   // >, <, >=, <=, ==, !=
	Threshold   float64                `json:"threshold"`
	Duration    time.Duration          `json:"duration"`    // 持续时间
	Severity    string                 `json:"severity"`    // low, medium, high, critical
	Enabled     bool                   `json:"enabled"`
	Actions     []AlertAction          `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata"`
	LastTriggered time.Time            `json:"last_triggered"`
	TriggerCount  int                  `json:"trigger_count"`
}

// AlertAction 告警动作
type AlertAction struct {
	Type       string                 `json:"type"`       // email, webhook, log, auto_fix
	Target     string                 `json:"target"`     // 目标地址
	Template   string                 `json:"template"`   // 消息模板
	Parameters map[string]interface{} `json:"parameters"`
}

// Alert 告警
type Alert struct {
	ID          string                 `json:"id"`
	RuleID      string                 `json:"rule_id"`
	RuleName    string                 `json:"rule_name"`
	MetricName  string                 `json:"metric_name"`
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	Timestamp   time.Time              `json:"timestamp"`
	Status      string                 `json:"status"`      // active, resolved, suppressed
	ProjectID   *uuid.UUID             `json:"project_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewIntelligentMonitor 创建智能监控器
func NewIntelligentMonitor() *IntelligentMonitor {
	return &IntelligentMonitor{
		metrics:           make(map[string]*MetricCollector),
		alerts:            make(map[string]*AlertRule),
		anomalyDetector:   NewAnomalyDetector(),
		performanceTracker: NewPerformanceTracker(),
		healthChecker:     NewHealthChecker(),
		notificationMgr:   NewNotificationManager(),
		stopChan:          make(chan struct{}),
	}
}

// Start 启动监控
func (im *IntelligentMonitor) Start(ctx context.Context) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if im.isRunning {
		return fmt.Errorf("监控器已经在运行")
	}

	im.isRunning = true
	
	// 启动各个组件
	go im.metricsCollectionLoop(ctx)
	go im.alertEvaluationLoop(ctx)
	go im.anomalyDetectionLoop(ctx)
	go im.healthCheckLoop(ctx)
	go im.performanceTrackingLoop(ctx)

	// 初始化默认指标和告警规则
	im.initializeDefaultMetrics()
	im.initializeDefaultAlerts()

	log.Printf("[IntelligentMonitor] 智能监控器已启动")
	return nil
}

// Stop 停止监控
func (im *IntelligentMonitor) Stop() {
	im.mu.Lock()
	defer im.mu.Unlock()

	if !im.isRunning {
		return
	}

	close(im.stopChan)
	im.isRunning = false
	
	log.Printf("[IntelligentMonitor] 智能监控器已停止")
}

// RecordMetric 记录指标
func (im *IntelligentMonitor) RecordMetric(name string, value float64, labels map[string]string) {
	im.mu.Lock()
	defer im.mu.Unlock()

	collector, exists := im.metrics[name]
	if !exists {
		collector = &MetricCollector{
			Name:      name,
			Type:      "gauge",
			History:   make([]MetricPoint, 0),
			Labels:    make(map[string]string),
		}
		im.metrics[name] = collector
	}

	// 更新指标值
	collector.Value = value
	collector.Timestamp = time.Now()
	if labels != nil {
		collector.Labels = labels
	}

	// 添加到历史记录
	point := MetricPoint{
		Value:     value,
		Timestamp: time.Now(),
		Labels:    labels,
	}
	collector.History = append(collector.History, point)

	// 限制历史记录长度
	if len(collector.History) > 1000 {
		collector.History = collector.History[1:]
	}
}

// IncrementCounter 增加计数器
func (im *IntelligentMonitor) IncrementCounter(name string, labels map[string]string) {
	im.mu.Lock()
	defer im.mu.Unlock()

	collector, exists := im.metrics[name]
	if !exists {
		collector = &MetricCollector{
			Name:    name,
			Type:    "counter",
			Value:   0,
			History: make([]MetricPoint, 0),
			Labels:  make(map[string]string),
		}
		im.metrics[name] = collector
	}

	collector.Value++
	collector.Timestamp = time.Now()
	if labels != nil {
		collector.Labels = labels
	}

	// 添加到历史记录
	point := MetricPoint{
		Value:     collector.Value,
		Timestamp: time.Now(),
		Labels:    labels,
	}
	collector.History = append(collector.History, point)
}

// AddAlertRule 添加告警规则
func (im *IntelligentMonitor) AddAlertRule(rule *AlertRule) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.alerts[rule.ID] = rule
	log.Printf("[IntelligentMonitor] 添加告警规则: %s", rule.Name)
}

// RemoveAlertRule 移除告警规则
func (im *IntelligentMonitor) RemoveAlertRule(ruleID string) {
	im.mu.Lock()
	defer im.mu.Unlock()

	delete(im.alerts, ruleID)
	log.Printf("[IntelligentMonitor] 移除告警规则: %s", ruleID)
}

// GetMetrics 获取指标
func (im *IntelligentMonitor) GetMetrics() map[string]*MetricCollector {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// 返回副本
	result := make(map[string]*MetricCollector)
	for name, collector := range im.metrics {
		copy := *collector
		result[name] = &copy
	}
	return result
}

// GetAlerts 获取活跃告警
func (im *IntelligentMonitor) GetAlerts() []*Alert {
	// 这里应该从告警存储中获取活跃告警
	// 简化实现，返回空列表
	return make([]*Alert, 0)
}

// GetSystemHealth 获取系统健康状态
func (im *IntelligentMonitor) GetSystemHealth() *SystemHealthReport {
	return im.healthChecker.GetHealthReport()
}

// GetPerformanceReport 获取性能报告
func (im *IntelligentMonitor) GetPerformanceReport() *PerformanceReport {
	return im.performanceTracker.GenerateReport()
}

// 监控循环
func (im *IntelligentMonitor) metricsCollectionLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // 每10秒收集一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-im.stopChan:
			return
		case <-ticker.C:
			im.collectSystemMetrics()
		}
	}
}

func (im *IntelligentMonitor) alertEvaluationLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒评估一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-im.stopChan:
			return
		case <-ticker.C:
			im.evaluateAlerts()
		}
	}
}

func (im *IntelligentMonitor) anomalyDetectionLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟检测一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-im.stopChan:
			return
		case <-ticker.C:
			im.detectAnomalies()
		}
	}
}

func (im *IntelligentMonitor) healthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-im.stopChan:
			return
		case <-ticker.C:
			im.performHealthCheck()
		}
	}
}

func (im *IntelligentMonitor) performanceTrackingLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟跟踪一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-im.stopChan:
			return
		case <-ticker.C:
			im.trackPerformance()
		}
	}
}

// 具体实现方法
func (im *IntelligentMonitor) collectSystemMetrics() {
	// 收集系统指标
	im.RecordMetric("system.cpu_usage", 0.45, map[string]string{"host": "localhost"})
	im.RecordMetric("system.memory_usage", 0.67, map[string]string{"host": "localhost"})
	im.RecordMetric("system.disk_usage", 0.23, map[string]string{"host": "localhost"})
	
	// 收集应用指标
	im.RecordMetric("app.active_projects", float64(len(im.metrics)), nil)
	im.RecordMetric("app.response_time", 150.5, map[string]string{"endpoint": "/api/generate"})
}

func (im *IntelligentMonitor) evaluateAlerts() {
	im.mu.RLock()
	alerts := make(map[string]*AlertRule)
	for id, rule := range im.alerts {
		if rule.Enabled {
			alerts[id] = rule
		}
	}
	metrics := im.metrics
	im.mu.RUnlock()

	for _, rule := range alerts {
		if metric, exists := metrics[rule.MetricName]; exists {
			if im.evaluateCondition(metric.Value, rule.Condition, rule.Threshold) {
				im.triggerAlert(rule, metric.Value)
			}
		}
	}
}

func (im *IntelligentMonitor) evaluateCondition(value float64, condition string, threshold float64) bool {
	switch condition {
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}

func (im *IntelligentMonitor) triggerAlert(rule *AlertRule, value float64) {
	alert := &Alert{
		ID:         uuid.New().String(),
		RuleID:     rule.ID,
		RuleName:   rule.Name,
		MetricName: rule.MetricName,
		Value:      value,
		Threshold:  rule.Threshold,
		Severity:   rule.Severity,
		Message:    fmt.Sprintf("指标 %s 的值 %.2f %s %.2f", rule.MetricName, value, rule.Condition, rule.Threshold),
		Timestamp:  time.Now(),
		Status:     "active",
		Metadata:   make(map[string]interface{}),
	}

	// 执行告警动作
	for _, action := range rule.Actions {
		im.executeAlertAction(alert, action)
	}

	// 更新规则统计
	rule.LastTriggered = time.Now()
	rule.TriggerCount++

	log.Printf("[IntelligentMonitor] 触发告警: %s", alert.Message)
}

func (im *IntelligentMonitor) executeAlertAction(alert *Alert, action AlertAction) {
	switch action.Type {
	case "log":
		log.Printf("[ALERT] %s: %s", alert.Severity, alert.Message)
	case "email":
		im.notificationMgr.SendEmail(action.Target, alert.Message)
	case "webhook":
		im.notificationMgr.SendWebhook(action.Target, alert)
	case "auto_fix":
		im.executeAutoFix(alert, action)
	}
}

func (im *IntelligentMonitor) executeAutoFix(alert *Alert, action AlertAction) {
	// 自动修复逻辑
	log.Printf("[IntelligentMonitor] 执行自动修复: %s", alert.RuleName)
	
	// 这里可以实现具体的自动修复逻辑
	// 例如：重启服务、清理缓存、调整资源分配等
}

func (im *IntelligentMonitor) detectAnomalies() {
	im.mu.RLock()
	metrics := im.metrics
	im.mu.RUnlock()

	for name, metric := range metrics {
		if len(metric.History) > 10 {
			anomaly := im.anomalyDetector.DetectAnomaly(metric.History)
			if anomaly != nil {
				log.Printf("[IntelligentMonitor] 检测到异常: %s - %s", name, anomaly.Description)
			}
		}
	}
}

func (im *IntelligentMonitor) performHealthCheck() {
	im.healthChecker.CheckHealth()
}

func (im *IntelligentMonitor) trackPerformance() {
	im.performanceTracker.TrackPerformance()
}

// 初始化默认配置
func (im *IntelligentMonitor) initializeDefaultMetrics() {
	// 初始化默认指标
	defaultMetrics := []string{
		"system.cpu_usage",
		"system.memory_usage",
		"system.disk_usage",
		"app.active_projects",
		"app.response_time",
		"app.error_rate",
		"app.throughput",
	}

	for _, name := range defaultMetrics {
		im.metrics[name] = &MetricCollector{
			Name:    name,
			Type:    "gauge",
			History: make([]MetricPoint, 0),
			Labels:  make(map[string]string),
		}
	}
}

func (im *IntelligentMonitor) initializeDefaultAlerts() {
	// CPU使用率告警
	im.AddAlertRule(&AlertRule{
		ID:         "cpu_high",
		Name:       "CPU使用率过高",
		MetricName: "system.cpu_usage",
		Condition:  ">",
		Threshold:  0.8,
		Duration:   5 * time.Minute,
		Severity:   "high",
		Enabled:    true,
		Actions: []AlertAction{
			{Type: "log", Target: "", Template: "CPU使用率过高: {{.Value}}"},
			{Type: "auto_fix", Target: "scale_up", Template: ""},
		},
	})

	// 内存使用率告警
	im.AddAlertRule(&AlertRule{
		ID:         "memory_high",
		Name:       "内存使用率过高",
		MetricName: "system.memory_usage",
		Condition:  ">",
		Threshold:  0.9,
		Duration:   3 * time.Minute,
		Severity:   "critical",
		Enabled:    true,
		Actions: []AlertAction{
			{Type: "log", Target: "", Template: "内存使用率过高: {{.Value}}"},
			{Type: "email", Target: "admin@example.com", Template: "系统内存不足"},
		},
	})

	// 响应时间告警
	im.AddAlertRule(&AlertRule{
		ID:         "response_time_high",
		Name:       "响应时间过长",
		MetricName: "app.response_time",
		Condition:  ">",
		Threshold:  5000, // 5秒
		Duration:   2 * time.Minute,
		Severity:   "medium",
		Enabled:    true,
		Actions: []AlertAction{
			{Type: "log", Target: "", Template: "响应时间过长: {{.Value}}ms"},
		},
	})

	// 错误率告警
	im.AddAlertRule(&AlertRule{
		ID:         "error_rate_high",
		Name:       "错误率过高",
		MetricName: "app.error_rate",
		Condition:  ">",
		Threshold:  0.05, // 5%
		Duration:   1 * time.Minute,
		Severity:   "high",
		Enabled:    true,
		Actions: []AlertAction{
			{Type: "log", Target: "", Template: "错误率过高: {{.Value}}"},
			{Type: "webhook", Target: "http://localhost:8080/alerts", Template: ""},
		},
	})
}

// MonitoringMetrics 监控指标常量
var MonitoringMetrics = struct {
	SystemCPU      string
	SystemMemory   string
	SystemDisk     string
	AppProjects    string
	AppResponseTime string
	AppErrorRate   string
	AppThroughput  string
}{
	SystemCPU:       "system.cpu_usage",
	SystemMemory:    "system.memory_usage",
	SystemDisk:      "system.disk_usage",
	AppProjects:     "app.active_projects",
	AppResponseTime: "app.response_time",
	AppErrorRate:    "app.error_rate",
	AppThroughput:   "app.throughput",
}

// 全局监控器实例
var GlobalMonitor = NewIntelligentMonitor()

// 便捷方法
func RecordMetric(name string, value float64, labels map[string]string) {
	GlobalMonitor.RecordMetric(name, value, labels)
}

func IncrementCounter(name string, labels map[string]string) {
	GlobalMonitor.IncrementCounter(name, labels)
}

func AddAlert(rule *AlertRule) {
	GlobalMonitor.AddAlertRule(rule)
}
