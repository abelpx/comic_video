package orchestrator

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// ProgressTracker 进度跟踪器
type ProgressTracker struct {
	mu        sync.RWMutex
	projects  map[uuid.UUID]*ProjectProgress
	listeners map[uuid.UUID][]ProgressListener
}

// ProjectProgress 项目进度
type ProjectProgress struct {
	ProjectID       uuid.UUID              `json:"project_id"`
	CurrentStage    string                 `json:"current_stage"`
	Progress        float64                `json:"progress"`        // 0.0 - 1.0
	Message         string                 `json:"message"`
	StartTime       time.Time              `json:"start_time"`
	EstimatedEnd    time.Time              `json:"estimated_end"`
	CompletedStages []string               `json:"completed_stages"`
	TotalStages     int                    `json:"total_stages"`
	StageProgress   map[string]float64     `json:"stage_progress"`
	Metadata        map[string]interface{} `json:"metadata"`
	LastUpdate      time.Time              `json:"last_update"`
}

// ProgressListener 进度监听器
type ProgressListener func(progress *ProjectProgress)

// NewProgressTracker 创建进度跟踪器
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		projects:  make(map[uuid.UUID]*ProjectProgress),
		listeners: make(map[uuid.UUID][]ProgressListener),
	}
}

// Initialize 初始化项目进度
func (pt *ProgressTracker) Initialize(projectID uuid.UUID) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.projects[projectID] = &ProjectProgress{
		ProjectID:       projectID,
		CurrentStage:    "初始化",
		Progress:        0.0,
		Message:         "项目初始化中...",
		StartTime:       time.Now(),
		CompletedStages: make([]string, 0),
		StageProgress:   make(map[string]float64),
		Metadata:        make(map[string]interface{}),
		LastUpdate:      time.Now(),
	}
}

// UpdateProgress 更新进度
func (pt *ProgressTracker) UpdateProgress(projectID uuid.UUID, stage string, progress float64, message string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if proj, exists := pt.projects[projectID]; exists {
		proj.CurrentStage = stage
		proj.Progress = progress
		proj.Message = message
		proj.LastUpdate = time.Now()
		
		// 更新阶段进度
		proj.StageProgress[stage] = progress
		
		// 估算结束时间
		if progress > 0 {
			elapsed := time.Since(proj.StartTime)
			totalEstimated := time.Duration(float64(elapsed) / progress)
			proj.EstimatedEnd = proj.StartTime.Add(totalEstimated)
		}

		// 通知监听器
		pt.notifyListeners(projectID, proj)
	}
}

// CompleteStage 完成阶段
func (pt *ProgressTracker) CompleteStage(projectID uuid.UUID, stage string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if proj, exists := pt.projects[projectID]; exists {
		proj.CompletedStages = append(proj.CompletedStages, stage)
		proj.StageProgress[stage] = 1.0
		proj.LastUpdate = time.Now()

		// 通知监听器
		pt.notifyListeners(projectID, proj)
	}
}

// GetProgress 获取进度
func (pt *ProgressTracker) GetProgress(projectID uuid.UUID) *ProjectProgress {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	if proj, exists := pt.projects[projectID]; exists {
		// 返回副本
		copy := *proj
		return &copy
	}
	return nil
}

// AddListener 添加监听器
func (pt *ProgressTracker) AddListener(projectID uuid.UUID, listener ProgressListener) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if _, exists := pt.listeners[projectID]; !exists {
		pt.listeners[projectID] = make([]ProgressListener, 0)
	}
	pt.listeners[projectID] = append(pt.listeners[projectID], listener)
}

// RemoveListener 移除监听器
func (pt *ProgressTracker) RemoveListener(projectID uuid.UUID) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	delete(pt.listeners, projectID)
}

// Cleanup 清理项目数据
func (pt *ProgressTracker) Cleanup(projectID uuid.UUID) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	delete(pt.projects, projectID)
	delete(pt.listeners, projectID)
}

// notifyListeners 通知监听器
func (pt *ProgressTracker) notifyListeners(projectID uuid.UUID, progress *ProjectProgress) {
	if listeners, exists := pt.listeners[projectID]; exists {
		for _, listener := range listeners {
			go listener(progress) // 异步通知
		}
	}
}

// ResourceManager 资源管理器
type ResourceManager struct {
	mu                sync.RWMutex
	allocatedResources map[uuid.UUID]*ResourceAllocation
	globalLimits      *ResourceLimits
	currentUsage      *ResourceUsage
}

// ResourceAllocation 资源分配
type ResourceAllocation struct {
	ProjectID    uuid.UUID       `json:"project_id"`
	QualityLevel string          `json:"quality_level"`
	CPUCores     int             `json:"cpu_cores"`
	MemoryMB     int             `json:"memory_mb"`
	GPUMemoryMB  int             `json:"gpu_memory_mb"`
	DiskSpaceMB  int             `json:"disk_space_mb"`
	NetworkBandwidth int         `json:"network_bandwidth"` // Mbps
	Priority     int             `json:"priority"`          // 1-10, 10最高
	AllocatedAt  time.Time       `json:"allocated_at"`
	ExpiresAt    time.Time       `json:"expires_at"`
	Usage        *ResourceUsage  `json:"usage"`
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	MaxCPUCores      int `json:"max_cpu_cores"`
	MaxMemoryMB      int `json:"max_memory_mb"`
	MaxGPUMemoryMB   int `json:"max_gpu_memory_mb"`
	MaxDiskSpaceMB   int `json:"max_disk_space_mb"`
	MaxNetworkBandwidth int `json:"max_network_bandwidth"`
	MaxConcurrentProjects int `json:"max_concurrent_projects"`
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	CPUUsage        float64   `json:"cpu_usage"`         // 0.0 - 1.0
	MemoryUsageMB   int       `json:"memory_usage_mb"`
	GPUUsageMB      int       `json:"gpu_usage_mb"`
	DiskUsageMB     int       `json:"disk_usage_mb"`
	NetworkUsageMbps int      `json:"network_usage_mbps"`
	LastUpdated     time.Time `json:"last_updated"`
}

// ResourceUsageReport 资源使用报告
type ResourceUsageReport struct {
	ProjectID       uuid.UUID              `json:"project_id"`
	TotalDuration   time.Duration          `json:"total_duration"`
	PeakCPUUsage    float64                `json:"peak_cpu_usage"`
	PeakMemoryMB    int                    `json:"peak_memory_mb"`
	PeakGPUMB       int                    `json:"peak_gpu_mb"`
	TotalDiskMB     int                    `json:"total_disk_mb"`
	TotalNetworkMB  int                    `json:"total_network_mb"`
	CostEstimate    float64                `json:"cost_estimate"`
	EfficiencyScore float64                `json:"efficiency_score"`
	Recommendations []string               `json:"recommendations"`
	DetailedMetrics map[string]interface{} `json:"detailed_metrics"`
}

// NewResourceManager 创建资源管理器
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		allocatedResources: make(map[uuid.UUID]*ResourceAllocation),
		globalLimits: &ResourceLimits{
			MaxCPUCores:           16,
			MaxMemoryMB:           32768, // 32GB
			MaxGPUMemoryMB:        12288, // 12GB
			MaxDiskSpaceMB:        102400, // 100GB
			MaxNetworkBandwidth:   1000,   // 1Gbps
			MaxConcurrentProjects: 10,
		},
		currentUsage: &ResourceUsage{
			LastUpdated: time.Now(),
		},
	}
}

// AllocateResources 分配资源
func (rm *ResourceManager) AllocateResources(projectID uuid.UUID, qualityLevel string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 检查是否已经分配
	if _, exists := rm.allocatedResources[projectID]; exists {
		return nil // 已经分配
	}

	// 检查并发项目数限制
	if len(rm.allocatedResources) >= rm.globalLimits.MaxConcurrentProjects {
		return fmt.Errorf("达到最大并发项目数限制: %d", rm.globalLimits.MaxConcurrentProjects)
	}

	// 根据质量级别分配资源
	allocation := rm.calculateResourceAllocation(qualityLevel)
	allocation.ProjectID = projectID
	allocation.AllocatedAt = time.Now()
	allocation.ExpiresAt = time.Now().Add(2 * time.Hour) // 2小时过期
	allocation.Usage = &ResourceUsage{LastUpdated: time.Now()}

	// 检查资源是否足够
	if !rm.canAllocate(allocation) {
		return fmt.Errorf("资源不足，无法分配")
	}

	rm.allocatedResources[projectID] = allocation
	rm.updateCurrentUsage()

	return nil
}

// ReleaseResources 释放资源
func (rm *ResourceManager) ReleaseResources(projectID uuid.UUID) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.allocatedResources, projectID)
	rm.updateCurrentUsage()
}

// UpdateResourceUsage 更新资源使用情况
func (rm *ResourceManager) UpdateResourceUsage(projectID uuid.UUID, usage *ResourceUsage) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if allocation, exists := rm.allocatedResources[projectID]; exists {
		allocation.Usage = usage
		allocation.Usage.LastUpdated = time.Now()
	}
}

// GenerateUsageReport 生成使用报告
func (rm *ResourceManager) GenerateUsageReport(projectID uuid.UUID) *ResourceUsageReport {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	allocation, exists := rm.allocatedResources[projectID]
	if !exists {
		return &ResourceUsageReport{
			ProjectID: projectID,
			CostEstimate: 0,
			EfficiencyScore: 0,
			Recommendations: []string{"项目资源分配信息未找到"},
		}
	}

	duration := time.Since(allocation.AllocatedAt)
	
	report := &ResourceUsageReport{
		ProjectID:     projectID,
		TotalDuration: duration,
		PeakCPUUsage:  allocation.Usage.CPUUsage,
		PeakMemoryMB:  allocation.Usage.MemoryUsageMB,
		PeakGPUMB:     allocation.Usage.GPUUsageMB,
		TotalDiskMB:   allocation.Usage.DiskUsageMB,
		TotalNetworkMB: allocation.Usage.NetworkUsageMbps * int(duration.Minutes()),
		DetailedMetrics: make(map[string]interface{}),
	}

	// 计算成本估算
	report.CostEstimate = rm.calculateCost(allocation, duration)
	
	// 计算效率分数
	report.EfficiencyScore = rm.calculateEfficiency(allocation)
	
	// 生成建议
	report.Recommendations = rm.generateRecommendations(allocation, report)

	return report
}

// GetCurrentUsage 获取当前使用情况
func (rm *ResourceManager) GetCurrentUsage() *ResourceUsage {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	copy := *rm.currentUsage
	return &copy
}

// GetAllocation 获取分配信息
func (rm *ResourceManager) GetAllocation(projectID uuid.UUID) *ResourceAllocation {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if allocation, exists := rm.allocatedResources[projectID]; exists {
		copy := *allocation
		return &copy
	}
	return nil
}

// 私有方法
func (rm *ResourceManager) calculateResourceAllocation(qualityLevel string) *ResourceAllocation {
	allocation := &ResourceAllocation{
		QualityLevel: qualityLevel,
	}

	switch qualityLevel {
	case "basic":
		allocation.CPUCores = 2
		allocation.MemoryMB = 4096  // 4GB
		allocation.GPUMemoryMB = 2048 // 2GB
		allocation.DiskSpaceMB = 10240 // 10GB
		allocation.NetworkBandwidth = 100 // 100Mbps
		allocation.Priority = 3

	case "standard":
		allocation.CPUCores = 4
		allocation.MemoryMB = 8192  // 8GB
		allocation.GPUMemoryMB = 4096 // 4GB
		allocation.DiskSpaceMB = 20480 // 20GB
		allocation.NetworkBandwidth = 200 // 200Mbps
		allocation.Priority = 5

	case "premium":
		allocation.CPUCores = 8
		allocation.MemoryMB = 16384 // 16GB
		allocation.GPUMemoryMB = 8192 // 8GB
		allocation.DiskSpaceMB = 40960 // 40GB
		allocation.NetworkBandwidth = 500 // 500Mbps
		allocation.Priority = 8

	default: // medium
		allocation.CPUCores = 4
		allocation.MemoryMB = 8192
		allocation.GPUMemoryMB = 4096
		allocation.DiskSpaceMB = 20480
		allocation.NetworkBandwidth = 200
		allocation.Priority = 5
	}

	return allocation
}

func (rm *ResourceManager) canAllocate(allocation *ResourceAllocation) bool {
	totalCPU := allocation.CPUCores
	totalMemory := allocation.MemoryMB
	totalGPU := allocation.GPUMemoryMB
	totalDisk := allocation.DiskSpaceMB
	totalBandwidth := allocation.NetworkBandwidth

	// 计算当前已分配的资源
	for _, existing := range rm.allocatedResources {
		totalCPU += existing.CPUCores
		totalMemory += existing.MemoryMB
		totalGPU += existing.GPUMemoryMB
		totalDisk += existing.DiskSpaceMB
		totalBandwidth += existing.NetworkBandwidth
	}

	// 检查是否超过限制
	return totalCPU <= rm.globalLimits.MaxCPUCores &&
		totalMemory <= rm.globalLimits.MaxMemoryMB &&
		totalGPU <= rm.globalLimits.MaxGPUMemoryMB &&
		totalDisk <= rm.globalLimits.MaxDiskSpaceMB &&
		totalBandwidth <= rm.globalLimits.MaxNetworkBandwidth
}

func (rm *ResourceManager) updateCurrentUsage() {
	totalCPU := 0
	totalMemory := 0
	totalGPU := 0
	totalDisk := 0
	totalBandwidth := 0

	for _, allocation := range rm.allocatedResources {
		totalCPU += allocation.CPUCores
		totalMemory += allocation.MemoryMB
		totalGPU += allocation.GPUMemoryMB
		totalDisk += allocation.DiskSpaceMB
		totalBandwidth += allocation.NetworkBandwidth
	}

	rm.currentUsage.CPUUsage = float64(totalCPU) / float64(rm.globalLimits.MaxCPUCores)
	rm.currentUsage.MemoryUsageMB = totalMemory
	rm.currentUsage.GPUUsageMB = totalGPU
	rm.currentUsage.DiskUsageMB = totalDisk
	rm.currentUsage.NetworkUsageMbps = totalBandwidth
	rm.currentUsage.LastUpdated = time.Now()
}

func (rm *ResourceManager) calculateCost(allocation *ResourceAllocation, duration time.Duration) float64 {
	// 简化的成本计算
	hours := duration.Hours()
	
	cpuCost := float64(allocation.CPUCores) * 0.1 * hours      // $0.1/core/hour
	memoryCost := float64(allocation.MemoryMB) * 0.01 * hours  // $0.01/GB/hour
	gpuCost := float64(allocation.GPUMemoryMB) * 0.05 * hours  // $0.05/GB/hour
	diskCost := float64(allocation.DiskSpaceMB) * 0.001 * hours // $0.001/GB/hour
	
	return cpuCost + memoryCost + gpuCost + diskCost
}

func (rm *ResourceManager) calculateEfficiency(allocation *ResourceAllocation) float64 {
	if allocation.Usage == nil {
		return 0.5 // 默认效率
	}

	// 基于实际使用率计算效率
	cpuEfficiency := allocation.Usage.CPUUsage
	memoryEfficiency := float64(allocation.Usage.MemoryUsageMB) / float64(allocation.MemoryMB)
	gpuEfficiency := float64(allocation.Usage.GPUUsageMB) / float64(allocation.GPUMemoryMB)

	// 加权平均
	return (cpuEfficiency*0.4 + memoryEfficiency*0.3 + gpuEfficiency*0.3)
}

func (rm *ResourceManager) generateRecommendations(allocation *ResourceAllocation, report *ResourceUsageReport) []string {
	recommendations := make([]string, 0)

	if report.EfficiencyScore < 0.3 {
		recommendations = append(recommendations, "资源利用率较低，建议降低质量级别")
	}

	if report.PeakCPUUsage > 0.9 {
		recommendations = append(recommendations, "CPU使用率过高，建议增加CPU核心数")
	}

	if float64(report.PeakMemoryMB)/float64(allocation.MemoryMB) > 0.9 {
		recommendations = append(recommendations, "内存使用率过高，建议增加内存分配")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "资源使用情况良好")
	}

	return recommendations
}

// CleanupExpiredAllocations 清理过期分配
func (rm *ResourceManager) CleanupExpiredAllocations() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	now := time.Now()
	for projectID, allocation := range rm.allocatedResources {
		if now.After(allocation.ExpiresAt) {
			delete(rm.allocatedResources, projectID)
		}
	}
	
	rm.updateCurrentUsage()
}

// GetResourceStatistics 获取资源统计
func (rm *ResourceManager) GetResourceStatistics() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	stats := map[string]interface{}{
		"total_projects":     len(rm.allocatedResources),
		"cpu_utilization":    rm.currentUsage.CPUUsage,
		"memory_usage_mb":    rm.currentUsage.MemoryUsageMB,
		"gpu_usage_mb":       rm.currentUsage.GPUUsageMB,
		"disk_usage_mb":      rm.currentUsage.DiskUsageMB,
		"network_usage_mbps": rm.currentUsage.NetworkUsageMbps,
		"last_updated":       rm.currentUsage.LastUpdated,
	}

	// 按质量级别统计
	qualityStats := make(map[string]int)
	for _, allocation := range rm.allocatedResources {
		qualityStats[allocation.QualityLevel]++
	}
	stats["quality_distribution"] = qualityStats

	return stats
}
