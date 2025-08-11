package ai

import (
	"context"
	"fmt"
	"log"
	"time"
)

// WorkflowManager 工作流管理器
type WorkflowManager struct {
	ModelManager    *ModelManager
	InferenceEngine *InferenceEngine
	Workflows       map[string]*Workflow
	PresetConfigs   map[string]*WorkflowConfig
}

// Workflow 工作流
type Workflow struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Config      *WorkflowConfig  `json:"config"`
	Steps       []*WorkflowStep  `json:"steps"`
	Status      string           `json:"status"`
	Progress    float64          `json:"progress"`
	StartTime   time.Time        `json:"start_time"`
	EndTime     time.Time        `json:"end_time"`
	Results     [][]byte         `json:"results,omitempty"`
	Errors      []string         `json:"errors,omitempty"`
}

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	Strategy        string                 `json:"strategy"`         // "speed", "balanced", "quality"
	BaseModel       string                 `json:"base_model"`       // 基础模型
	LoRAs           []LoRAConfig           `json:"loras"`            // LoRA配置
	Parameters      GenerationParameters   `json:"parameters"`       // 生成参数
	PostProcessing  PostProcessConfig      `json:"post_processing"`  // 后处理
	MemoryOptim     bool                   `json:"memory_optim"`     // 内存优化
	BatchSize       int                    `json:"batch_size"`       // 批次大小
	MaxRetries      int                    `json:"max_retries"`      // 最大重试次数
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`        // "generation", "postprocess", "validation"
	Status      string                 `json:"status"`      // "pending", "processing", "completed", "failed"
	Progress    float64                `json:"progress"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output"`
	Error       string                 `json:"error,omitempty"`
}

// PostProcessConfig 后处理配置
type PostProcessConfig struct {
	Upscale       bool    `json:"upscale"`
	UpscaleFactor float64 `json:"upscale_factor"`
	FaceRestore   bool    `json:"face_restore"`
	ColorCorrect  bool    `json:"color_correct"`
	Sharpen       bool    `json:"sharpen"`
	Denoise       bool    `json:"denoise"`
}

// NewWorkflowManager 创建工作流管理器
func NewWorkflowManager(modelManager *ModelManager, inferenceEngine *InferenceEngine) *WorkflowManager {
	wm := &WorkflowManager{
		ModelManager:    modelManager,
		InferenceEngine: inferenceEngine,
		Workflows:       make(map[string]*Workflow),
		PresetConfigs:   make(map[string]*WorkflowConfig),
	}
	
	wm.initializePresets()
	return wm
}

// initializePresets 初始化预设配置
func (wm *WorkflowManager) initializePresets() {
	// 速度优先预设
	wm.PresetConfigs["speed"] = &WorkflowConfig{
		Strategy:  "speed",
		BaseModel: "turbo",
		Parameters: GenerationParameters{
			Width:    512,
			Height:   512,
			Steps:    2,
			CFGScale: 1.5,
			Sampler:  "Euler a",
			BatchSize: 4,
		},
		PostProcessing: PostProcessConfig{
			Upscale:      false,
			FaceRestore:  false,
			ColorCorrect: true,
		},
		MemoryOptim: true,
		BatchSize:   4,
		MaxRetries:  2,
	}
	
	// 平衡预设
	wm.PresetConfigs["balanced"] = &WorkflowConfig{
		Strategy:  "balanced",
		BaseModel: "lightning",
		Parameters: GenerationParameters{
			Width:    768,
			Height:   768,
			Steps:    4,
			CFGScale: 2.0,
			Sampler:  "DPM++ 2M",
			BatchSize: 2,
		},
		PostProcessing: PostProcessConfig{
			Upscale:      true,
			UpscaleFactor: 1.5,
			FaceRestore:  true,
			ColorCorrect: true,
		},
		MemoryOptim: true,
		BatchSize:   2,
		MaxRetries:  3,
	}
	
	// 质量优先预设
	wm.PresetConfigs["quality"] = &WorkflowConfig{
		Strategy:  "quality",
		BaseModel: "base",
		Parameters: GenerationParameters{
			Width:    1024,
			Height:   1024,
			Steps:    20,
			CFGScale: 7.0,
			Sampler:  "DPM++ 2M Karras",
			BatchSize: 1,
		},
		PostProcessing: PostProcessConfig{
			Upscale:      true,
			UpscaleFactor: 2.0,
			FaceRestore:  true,
			ColorCorrect: true,
			Sharpen:     true,
			Denoise:     true,
		},
		MemoryOptim: false,
		BatchSize:   1,
		MaxRetries:  3,
	}
}

// CreateWorkflow 创建工作流
func (wm *WorkflowManager) CreateWorkflow(name, preset string, prompts []string, customConfig *WorkflowConfig) (*Workflow, error) {
	// 获取预设配置
	config := wm.PresetConfigs[preset]
	if config == nil {
		return nil, fmt.Errorf("未知预设: %s", preset)
	}
	
	// 应用自定义配置
	if customConfig != nil {
		config = wm.mergeConfigs(config, customConfig)
	}
	
	// 创建工作流
	workflow := &Workflow{
		ID:          generateWorkflowID(),
		Name:        name,
		Description: fmt.Sprintf("使用%s预设生成%d张图片", preset, len(prompts)),
		Config:      config,
		Status:      "created",
		Progress:    0.0,
	}
	
	// 创建工作流步骤
	workflow.Steps = wm.createWorkflowSteps(prompts, config)
	
	wm.Workflows[workflow.ID] = workflow
	
	log.Printf("[WorkflowManager] 创建工作流: %s (%s)", workflow.ID, preset)
	return workflow, nil
}

// ExecuteWorkflow 执行工作流
func (wm *WorkflowManager) ExecuteWorkflow(ctx context.Context, workflowID string, progressCallback func(float64, string)) error {
	workflow, exists := wm.Workflows[workflowID]
	if !exists {
		return fmt.Errorf("工作流不存在: %s", workflowID)
	}
	
	workflow.Status = "running"
	workflow.StartTime = time.Now()
	
	log.Printf("[WorkflowManager] 开始执行工作流: %s", workflowID)
	
	defer func() {
		workflow.EndTime = time.Now()
		if workflow.Status == "running" {
			workflow.Status = "completed"
		}
	}()
	
	totalSteps := len(workflow.Steps)
	
	for i, step := range workflow.Steps {
		select {
		case <-ctx.Done():
			workflow.Status = "cancelled"
			return ctx.Err()
		default:
		}
		
		step.Status = "processing"
		step.StartTime = time.Now()
		
		// 执行步骤
		if err := wm.executeStep(ctx, workflow, step); err != nil {
			step.Status = "failed"
			step.Error = err.Error()
			workflow.Status = "failed"
			workflow.Errors = append(workflow.Errors, fmt.Sprintf("步骤%s失败: %v", step.Name, err))
			return err
		}
		
		step.Status = "completed"
		step.EndTime = time.Now()
		step.Progress = 1.0
		
		// 更新总进度
		workflow.Progress = float64(i+1) / float64(totalSteps)
		
		if progressCallback != nil {
			progressCallback(workflow.Progress, fmt.Sprintf("完成步骤: %s", step.Name))
		}
		
		log.Printf("[WorkflowManager] 步骤完成: %s (耗时: %v)", step.Name, step.EndTime.Sub(step.StartTime))
	}
	
	workflow.Status = "completed"
	workflow.Progress = 1.0
	
	log.Printf("[WorkflowManager] 工作流完成: %s (总耗时: %v)", workflowID, workflow.EndTime.Sub(workflow.StartTime))
	return nil
}

// executeStep 执行单个步骤
func (wm *WorkflowManager) executeStep(ctx context.Context, workflow *Workflow, step *WorkflowStep) error {
	switch step.Type {
	case "generation":
		return wm.executeGenerationStep(ctx, workflow, step)
	case "postprocess":
		return wm.executePostProcessStep(ctx, workflow, step)
	case "validation":
		return wm.executeValidationStep(ctx, workflow, step)
	default:
		return fmt.Errorf("未知步骤类型: %s", step.Type)
	}
}

// executeGenerationStep 执行生成步骤
func (wm *WorkflowManager) executeGenerationStep(ctx context.Context, workflow *Workflow, step *WorkflowStep) error {
	prompts, ok := step.Input["prompts"].([]string)
	if !ok {
		return fmt.Errorf("无效的提示词输入")
	}
	
	// 创建推理任务
	jobs := make([]*InferenceJob, len(prompts))
	for i, prompt := range prompts {
		jobs[i] = &InferenceJob{
			ID:         fmt.Sprintf("%s-job-%d", step.Name, i),
			Prompt:     prompt,
			NegPrompt:  "blurry, low quality, distorted",
			Model:      workflow.Config.BaseModel,
			LoRAs:      workflow.Config.LoRAs,
			Parameters: workflow.Config.Parameters,
		}
	}
	
	// 批量生成
	results, err := wm.InferenceEngine.BatchGenerate(ctx, jobs)
	if err != nil {
		return err
	}
	
	// 保存结果
	step.Output = map[string]interface{}{
		"images": results,
		"count":  len(results),
	}
	
	workflow.Results = append(workflow.Results, results...)
	
	return nil
}

// executePostProcessStep 执行后处理步骤
func (wm *WorkflowManager) executePostProcessStep(ctx context.Context, workflow *Workflow, step *WorkflowStep) error {
	images, ok := step.Input["images"].([][]byte)
	if !ok {
		return fmt.Errorf("无效的图像输入")
	}
	
	config := workflow.Config.PostProcessing
	processedImages := make([][]byte, len(images))
	
	for i, img := range images {
		processed := img // 开始处理
		
		if config.Upscale {
			processed = wm.upscaleImage(processed, config.UpscaleFactor)
		}
		
		if config.FaceRestore {
			processed = wm.restoreFace(processed)
		}
		
		if config.ColorCorrect {
			processed = wm.correctColor(processed)
		}
		
		if config.Sharpen {
			processed = wm.sharpenImage(processed)
		}
		
		if config.Denoise {
			processed = wm.denoiseImage(processed)
		}
		
		processedImages[i] = processed
	}
	
	step.Output = map[string]interface{}{
		"images": processedImages,
		"count":  len(processedImages),
	}
	
	return nil
}

// executeValidationStep 执行验证步骤
func (wm *WorkflowManager) executeValidationStep(ctx context.Context, workflow *Workflow, step *WorkflowStep) error {
	images, ok := step.Input["images"].([][]byte)
	if !ok {
		return fmt.Errorf("无效的图像输入")
	}
	
	validImages := [][]byte{}
	invalidCount := 0
	
	for _, img := range images {
		if wm.validateImage(img) {
			validImages = append(validImages, img)
		} else {
			invalidCount++
		}
	}
	
	step.Output = map[string]interface{}{
		"valid_images":   validImages,
		"valid_count":    len(validImages),
		"invalid_count":  invalidCount,
		"success_rate":   float64(len(validImages)) / float64(len(images)),
	}
	
	return nil
}

// createWorkflowSteps 创建工作流步骤
func (wm *WorkflowManager) createWorkflowSteps(prompts []string, config *WorkflowConfig) []*WorkflowStep {
	steps := []*WorkflowStep{}
	
	// 生成步骤
	genStep := &WorkflowStep{
		Name:   "图片生成",
		Type:   "generation",
		Status: "pending",
		Input: map[string]interface{}{
			"prompts": prompts,
		},
	}
	steps = append(steps, genStep)
	
	// 后处理步骤
	if config.PostProcessing.Upscale || config.PostProcessing.FaceRestore {
		postStep := &WorkflowStep{
			Name:   "后处理",
			Type:   "postprocess",
			Status: "pending",
			Input: map[string]interface{}{
				"source_step": "图片生成",
			},
		}
		steps = append(steps, postStep)
	}
	
	// 验证步骤
	validStep := &WorkflowStep{
		Name:   "质量验证",
		Type:   "validation",
		Status: "pending",
		Input: map[string]interface{}{
			"source_step": "后处理",
		},
	}
	steps = append(steps, validStep)
	
	return steps
}

// GetWorkflowStatus 获取工作流状态
func (wm *WorkflowManager) GetWorkflowStatus(workflowID string) (*Workflow, error) {
	workflow, exists := wm.Workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("工作流不存在: %s", workflowID)
	}
	
	return workflow, nil
}

// GetAvailablePresets 获取可用预设
func (wm *WorkflowManager) GetAvailablePresets() map[string]interface{} {
	presets := make(map[string]interface{})
	
	for name, config := range wm.PresetConfigs {
		presets[name] = map[string]interface{}{
			"strategy":     config.Strategy,
			"base_model":   config.BaseModel,
			"parameters":   config.Parameters,
			"memory_optim": config.MemoryOptim,
			"description":  wm.getPresetDescription(name),
		}
	}
	
	return presets
}

// 辅助方法
func (wm *WorkflowManager) mergeConfigs(base, custom *WorkflowConfig) *WorkflowConfig {
	// 合并配置逻辑
	merged := *base
	if custom.BaseModel != "" {
		merged.BaseModel = custom.BaseModel
	}
	if len(custom.LoRAs) > 0 {
		merged.LoRAs = custom.LoRAs
	}
	return &merged
}

func (wm *WorkflowManager) getPresetDescription(preset string) string {
	descriptions := map[string]string{
		"speed":    "极速生成，2步出图，适合快速预览",
		"balanced": "平衡模式，4步出图，质量与速度兼顾",
		"quality":  "高质量模式，20步出图，最佳画质",
	}
	return descriptions[preset]
}

// 图像处理方法 (简化实现)
func (wm *WorkflowManager) upscaleImage(img []byte, factor float64) []byte {
	// 模拟放大处理
	return img
}

func (wm *WorkflowManager) restoreFace(img []byte) []byte {
	// 模拟面部修复
	return img
}

func (wm *WorkflowManager) correctColor(img []byte) []byte {
	// 模拟色彩校正
	return img
}

func (wm *WorkflowManager) sharpenImage(img []byte) []byte {
	// 模拟锐化
	return img
}

func (wm *WorkflowManager) denoiseImage(img []byte) []byte {
	// 模拟降噪
	return img
}

func (wm *WorkflowManager) validateImage(img []byte) bool {
	// 模拟图像验证
	return len(img) > 0
}

func generateWorkflowID() string {
	return fmt.Sprintf("workflow_%d", time.Now().UnixNano())
}
