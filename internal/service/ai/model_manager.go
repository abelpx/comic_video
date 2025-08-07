package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ModelManager 模型管理器
type ModelManager struct {
	ModelsPath    string
	LoadedModels  map[string]*LoadedModel
	LoRACache     map[string]*LoRAModel
	mutex         sync.RWMutex
	maxMemoryGB   float64 // 最大显存使用限制
	currentMemory float64 // 当前显存使用
}

// LoadedModel 已加载的模型
type LoadedModel struct {
	Name         string    `json:"name"`
	Type         string    `json:"type"`         // "base", "turbo", "lightning", "lcm"
	Path         string    `json:"path"`
	MemoryUsage  float64   `json:"memory_usage"` // GB
	LoadTime     time.Time `json:"load_time"`
	LastUsed     time.Time `json:"last_used"`
	IsActive     bool      `json:"is_active"`
	Config       ModelConfig `json:"config"`
}

// ModelConfig 模型配置
type ModelConfig struct {
	Resolution    []int   `json:"resolution"`     // [width, height]
	Steps         []int   `json:"steps"`          // 支持的步数范围
	CFGScale      []float64 `json:"cfg_scale"`    // CFG范围
	Samplers      []string `json:"samplers"`      // 支持的采样器
	MemoryOptim   bool    `json:"memory_optim"`   // 内存优化
	Precision     string  `json:"precision"`      // "fp16", "fp32", "int8"
	BatchSize     int     `json:"batch_size"`     // 最大批次大小
}

// LoRAModel LoRA模型
type LoRAModel struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Strength    float64   `json:"strength"`
	MemoryUsage float64   `json:"memory_usage"`
	LoadTime    time.Time `json:"load_time"`
	IsLoaded    bool      `json:"is_loaded"`
	Tags        []string  `json:"tags"`
	Description string    `json:"description"`
}

// NewModelManager 创建模型管理器
func NewModelManager(modelsPath string, maxMemoryGB float64) *ModelManager {
	return &ModelManager{
		ModelsPath:    modelsPath,
		LoadedModels:  make(map[string]*LoadedModel),
		LoRACache:     make(map[string]*LoRAModel),
		maxMemoryGB:   maxMemoryGB,
		currentMemory: 0,
	}
}

// InitializeModels 初始化模型库
func (mm *ModelManager) InitializeModels() error {
	log.Printf("[ModelManager] 初始化模型库: %s", mm.ModelsPath)
	
	// 创建模型目录结构
	dirs := []string{
		filepath.Join(mm.ModelsPath, "base"),
		filepath.Join(mm.ModelsPath, "turbo"),
		filepath.Join(mm.ModelsPath, "lightning"),
		filepath.Join(mm.ModelsPath, "lcm"),
		filepath.Join(mm.ModelsPath, "lora"),
		filepath.Join(mm.ModelsPath, "embeddings"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %v", dir, err)
		}
	}
	
	// 扫描现有模型
	return mm.ScanModels()
}

// ScanModels 扫描模型文件
func (mm *ModelManager) ScanModels() error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	modelTypes := map[string]ModelConfig{
		"base": {
			Resolution:  []int{512, 768, 1024},
			Steps:       []int{20, 30, 50},
			CFGScale:    []float64{7.0, 9.0, 12.0},
			Samplers:    []string{"DPM++ 2M Karras", "Euler a", "DDIM"},
			MemoryOptim: true,
			Precision:   "fp16",
			BatchSize:   2,
		},
		"turbo": {
			Resolution:  []int{512, 768},
			Steps:       []int{1, 2, 4},
			CFGScale:    []float64{1.0, 2.0},
			Samplers:    []string{"Euler a"},
			MemoryOptim: true,
			Precision:   "fp16",
			BatchSize:   4,
		},
		"lightning": {
			Resolution:  []int{512, 768, 1024},
			Steps:       []int{2, 4, 8},
			CFGScale:    []float64{1.0, 2.0, 4.0},
			Samplers:    []string{"Euler a", "DPM++ 2M"},
			MemoryOptim: true,
			Precision:   "fp16",
			BatchSize:   3,
		},
		"lcm": {
			Resolution:  []int{512, 768},
			Steps:       []int{4, 8},
			CFGScale:    []float64{1.0, 2.0},
			Samplers:    []string{"LCM"},
			MemoryOptim: true,
			Precision:   "fp16",
			BatchSize:   4,
		},
	}
	
	// 扫描每种类型的模型
	for modelType, config := range modelTypes {
		typeDir := filepath.Join(mm.ModelsPath, modelType)
		files, err := os.ReadDir(typeDir)
		if err != nil {
			log.Printf("[ModelManager] 扫描目录失败 %s: %v", typeDir, err)
			continue
		}
		
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".safetensors" || filepath.Ext(file.Name()) == ".ckpt" {
				modelName := file.Name()
				modelPath := filepath.Join(typeDir, modelName)
				
				// 估算内存使用
				memoryUsage := mm.estimateModelMemory(modelType)
				
				mm.LoadedModels[modelName] = &LoadedModel{
					Name:        modelName,
					Type:        modelType,
					Path:        modelPath,
					MemoryUsage: memoryUsage,
					IsActive:    false,
					Config:      config,
				}
				
				log.Printf("[ModelManager] 发现模型: %s (%s, %.1fGB)", modelName, modelType, memoryUsage)
			}
		}
	}
	
	// 扫描LoRA模型
	return mm.scanLoRAModels()
}

// scanLoRAModels 扫描LoRA模型
func (mm *ModelManager) scanLoRAModels() error {
	loraDir := filepath.Join(mm.ModelsPath, "lora")
	files, err := os.ReadDir(loraDir)
	if err != nil {
		return nil // LoRA目录可选
	}
	
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".safetensors" {
			loraName := file.Name()
			loraPath := filepath.Join(loraDir, loraName)
			
			mm.LoRACache[loraName] = &LoRAModel{
				Name:        loraName,
				Path:        loraPath,
				Strength:    1.0,
				MemoryUsage: 0.2, // LoRA通常很小
				IsLoaded:    false,
				Tags:        mm.parseLoRATags(loraName),
			}
			
			log.Printf("[ModelManager] 发现LoRA: %s", loraName)
		}
	}
	
	return nil
}

// LoadModel 加载模型
func (mm *ModelManager) LoadModel(modelName string) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	model, exists := mm.LoadedModels[modelName]
	if !exists {
		return fmt.Errorf("模型不存在: %s", modelName)
	}
	
	if model.IsActive {
		model.LastUsed = time.Now()
		return nil // 已经加载
	}
	
	// 检查内存限制
	if mm.currentMemory+model.MemoryUsage > mm.maxMemoryGB {
		// 卸载最久未使用的模型
		if err := mm.unloadLeastRecentlyUsed(model.MemoryUsage); err != nil {
			return fmt.Errorf("释放内存失败: %v", err)
		}
	}
	
	log.Printf("[ModelManager] 加载模型: %s (%.1fGB)", modelName, model.MemoryUsage)
	
	// 模拟加载过程
	startTime := time.Now()
	
	// 这里会调用实际的模型加载逻辑
	if err := mm.loadModelToGPU(model); err != nil {
		return fmt.Errorf("加载模型到GPU失败: %v", err)
	}
	
	model.IsActive = true
	model.LoadTime = startTime
	model.LastUsed = time.Now()
	mm.currentMemory += model.MemoryUsage
	
	log.Printf("[ModelManager] 模型加载完成: %s (耗时: %v)", modelName, time.Since(startTime))
	return nil
}

// UnloadModel 卸载模型
func (mm *ModelManager) UnloadModel(modelName string) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	model, exists := mm.LoadedModels[modelName]
	if !exists || !model.IsActive {
		return nil
	}
	
	log.Printf("[ModelManager] 卸载模型: %s", modelName)
	
	// 这里会调用实际的模型卸载逻辑
	if err := mm.unloadModelFromGPU(model); err != nil {
		return fmt.Errorf("从GPU卸载模型失败: %v", err)
	}
	
	model.IsActive = false
	mm.currentMemory -= model.MemoryUsage
	
	return nil
}

// LoadLoRA 加载LoRA
func (mm *ModelManager) LoadLoRA(loraName string, strength float64) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	lora, exists := mm.LoRACache[loraName]
	if !exists {
		return fmt.Errorf("LoRA不存在: %s", loraName)
	}
	
	if lora.IsLoaded {
		lora.Strength = strength
		return nil
	}
	
	log.Printf("[ModelManager] 加载LoRA: %s (强度: %.2f)", loraName, strength)
	
	// 这里会调用实际的LoRA加载逻辑
	if err := mm.loadLoRAToGPU(lora, strength); err != nil {
		return fmt.Errorf("加载LoRA到GPU失败: %v", err)
	}
	
	lora.IsLoaded = true
	lora.Strength = strength
	lora.LoadTime = time.Now()
	mm.currentMemory += lora.MemoryUsage
	
	return nil
}

// GetAvailableModels 获取可用模型列表
func (mm *ModelManager) GetAvailableModels() map[string]interface{} {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	models := make(map[string]interface{})
	loras := make(map[string]interface{})
	
	for name, model := range mm.LoadedModels {
		models[name] = map[string]interface{}{
			"type":         model.Type,
			"memory_usage": model.MemoryUsage,
			"is_active":    model.IsActive,
			"config":       model.Config,
		}
	}
	
	for name, lora := range mm.LoRACache {
		loras[name] = map[string]interface{}{
			"memory_usage": lora.MemoryUsage,
			"is_loaded":    lora.IsLoaded,
			"tags":         lora.Tags,
			"description":  lora.Description,
		}
	}
	
	return map[string]interface{}{
		"models":         models,
		"loras":          loras,
		"memory_usage":   mm.currentMemory,
		"memory_limit":   mm.maxMemoryGB,
		"memory_free":    mm.maxMemoryGB - mm.currentMemory,
	}
}

// 辅助方法
func (mm *ModelManager) estimateModelMemory(modelType string) float64 {
	switch modelType {
	case "base":
		return 6.0 // 6GB
	case "turbo":
		return 3.5 // 3.5GB
	case "lightning":
		return 4.0 // 4GB
	case "lcm":
		return 3.0 // 3GB
	default:
		return 5.0
	}
}

func (mm *ModelManager) parseLoRATags(filename string) []string {
	// 从文件名解析标签
	tags := []string{}
	if contains(filename, "anime") {
		tags = append(tags, "anime")
	}
	if contains(filename, "realistic") {
		tags = append(tags, "realistic")
	}
	if contains(filename, "style") {
		tags = append(tags, "style")
	}
	return tags
}

func (mm *ModelManager) unloadLeastRecentlyUsed(requiredMemory float64) error {
	// 找到最久未使用的模型并卸载
	var oldestModel *LoadedModel
	var oldestTime time.Time = time.Now()
	
	for _, model := range mm.LoadedModels {
		if model.IsActive && model.LastUsed.Before(oldestTime) {
			oldestModel = model
			oldestTime = model.LastUsed
		}
	}
	
	if oldestModel != nil {
		return mm.UnloadModel(oldestModel.Name)
	}
	
	return fmt.Errorf("无法释放足够内存")
}

// 这些方法需要根据实际的推理引擎实现
func (mm *ModelManager) loadModelToGPU(model *LoadedModel) error {
	// 实际的模型加载逻辑
	time.Sleep(100 * time.Millisecond) // 模拟加载时间
	return nil
}

func (mm *ModelManager) unloadModelFromGPU(model *LoadedModel) error {
	// 实际的模型卸载逻辑
	return nil
}

func (mm *ModelManager) loadLoRAToGPU(lora *LoRAModel, strength float64) error {
	// 实际的LoRA加载逻辑
	return nil
}
