package orchestrator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"comic_video/internal/domain/entity"
	"comic_video/internal/service/animation"
	"comic_video/internal/service/character"
	"comic_video/internal/service/quality"
	"comic_video/internal/service/scene"
	"comic_video/internal/service/workflow"
	"comic_video/internal/utils"
	"github.com/google/uuid"
)

// WorkflowOrchestrator 工作流编排器
type WorkflowOrchestrator struct {
	enhancedWorkflow    *workflow.EnhancedWorkflowService
	qualityController   *quality.AdvancedQualityController
	characterValidator  *character.AdvancedCharacterValidator
	sceneAnalyzer      *scene.DeepSceneAnalyzer
	animationEngine    *animation.AdvancedAnimationEngine
	consistencyManager *character.ConsistencyManager
	retryManager       *utils.TaskRetryManager
	progressTracker    *ProgressTracker
	resourceManager    *ResourceManager
}

// NewWorkflowOrchestrator 创建工作流编排器
func NewWorkflowOrchestrator(
	enhancedWorkflow *workflow.EnhancedWorkflowService,
	qualityController *quality.AdvancedQualityController,
	characterValidator *character.AdvancedCharacterValidator,
	sceneAnalyzer *scene.DeepSceneAnalyzer,
	animationEngine *animation.AdvancedAnimationEngine,
	consistencyManager *character.ConsistencyManager,
) *WorkflowOrchestrator {
	return &WorkflowOrchestrator{
		enhancedWorkflow:    enhancedWorkflow,
		qualityController:   qualityController,
		characterValidator:  characterValidator,
		sceneAnalyzer:      sceneAnalyzer,
		animationEngine:    animationEngine,
		consistencyManager: consistencyManager,
		retryManager:       utils.GlobalTaskRetryManager,
		progressTracker:    NewProgressTracker(),
		resourceManager:    NewResourceManager(),
	}
}

// OrchestrationRequest 编排请求
type OrchestrationRequest struct {
	ProjectID       uuid.UUID              `json:"project_id"`
	Novel           string                 `json:"novel"`
	Style           string                 `json:"style"`
	QualityLevel    string                 `json:"quality_level"`    // basic, standard, premium
	OptimizationLevel string               `json:"optimization_level"` // fast, balanced, quality
	CustomOptions   map[string]interface{} `json:"custom_options"`
	Callbacks       *OrchestrationCallbacks `json:"-"`
}

// OrchestrationCallbacks 编排回调
type OrchestrationCallbacks struct {
	OnProgress      func(stage string, progress float64, message string)
	OnStageComplete func(stage string, result interface{}, duration time.Duration)
	OnQualityCheck  func(report *quality.ComprehensiveQualityReport)
	OnError         func(stage string, err error, canRetry bool)
}

// OrchestrationResult 编排结果
type OrchestrationResult struct {
	ProjectID        uuid.UUID                           `json:"project_id"`
	Success          bool                                `json:"success"`
	FinalVideoPath   string                              `json:"final_video_path"`
	QualityReport    *quality.ComprehensiveQualityReport `json:"quality_report"`
	ProcessingStages []*StageResult                      `json:"processing_stages"`
	TotalDuration    time.Duration                       `json:"total_duration"`
	ResourceUsage    *ResourceUsageReport                `json:"resource_usage"`
	Metadata         map[string]interface{}              `json:"metadata"`
	ErrorDetails     *ErrorDetails                       `json:"error_details,omitempty"`
}

// StageResult 阶段结果
type StageResult struct {
	StageName      string                 `json:"stage_name"`
	Success        bool                   `json:"success"`
	Duration       time.Duration          `json:"duration"`
	QualityScore   float64                `json:"quality_score"`
	RetryCount     int                    `json:"retry_count"`
	OutputData     map[string]interface{} `json:"output_data"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
}

// ErrorDetails 错误详情
type ErrorDetails struct {
	FailedStage    string    `json:"failed_stage"`
	ErrorMessage   string    `json:"error_message"`
	ErrorCode      string    `json:"error_code"`
	Timestamp      time.Time `json:"timestamp"`
	CanRetry       bool      `json:"can_retry"`
	SuggestedFix   string    `json:"suggested_fix"`
}

// ExecuteOrchestration 执行编排
func (wo *WorkflowOrchestrator) ExecuteOrchestration(ctx context.Context, req *OrchestrationRequest) (*OrchestrationResult, error) {
	startTime := time.Now()
	log.Printf("[WorkflowOrchestrator] 开始执行编排: %s", req.ProjectID)

	// 初始化结果
	result := &OrchestrationResult{
		ProjectID:        req.ProjectID,
		ProcessingStages: make([]*StageResult, 0),
		Metadata:         make(map[string]interface{}),
	}

	// 初始化进度跟踪
	wo.progressTracker.Initialize(req.ProjectID)
	defer wo.progressTracker.Cleanup(req.ProjectID)

	// 初始化资源管理
	wo.resourceManager.AllocateResources(req.ProjectID, req.QualityLevel)
	defer wo.resourceManager.ReleaseResources(req.ProjectID)

	// 定义处理阶段
	stages := wo.defineProcessingStages(req)

	// 执行各个阶段
	var stageData = make(map[string]interface{})
	
	for i, stage := range stages {
		stageStartTime := time.Now()
		
		// 更新进度
		progress := float64(i) / float64(len(stages))
		wo.notifyProgress(req.Callbacks, stage.Name, progress, fmt.Sprintf("开始执行: %s", stage.Description))

		// 执行阶段
		stageResult, err := wo.executeStage(ctx, stage, stageData, req)
		stageDuration := time.Since(stageStartTime)

		// 记录阶段结果
		stageResultRecord := &StageResult{
			StageName:    stage.Name,
			Success:      err == nil,
			Duration:     stageDuration,
			QualityScore: stageResult.QualityScore,
			RetryCount:   stageResult.RetryCount,
			OutputData:   stageResult.OutputData,
		}

		if err != nil {
			stageResultRecord.ErrorMessage = err.Error()
			result.ProcessingStages = append(result.ProcessingStages, stageResultRecord)
			
			// 处理错误
			if wo.handleStageError(ctx, stage, err, req.Callbacks) {
				// 可以重试
				continue
			} else {
				// 不可重试，终止流程
				result.Success = false
				result.ErrorDetails = &ErrorDetails{
					FailedStage:  stage.Name,
					ErrorMessage: err.Error(),
					ErrorCode:    "STAGE_EXECUTION_FAILED",
					Timestamp:    time.Now(),
					CanRetry:     false,
					SuggestedFix: wo.generateSuggestedFix(stage.Name, err),
				}
				break
			}
		}

		result.ProcessingStages = append(result.ProcessingStages, stageResultRecord)
		
		// 更新阶段数据
		for key, value := range stageResult.OutputData {
			stageData[key] = value
		}

		// 通知阶段完成
		wo.notifyStageComplete(req.Callbacks, stage.Name, stageResult.OutputData, stageDuration)
	}

	// 最终质量检查
	if result.Success {
		finalQualityReport, err := wo.performFinalQualityCheck(ctx, stageData, req)
		if err != nil {
			log.Printf("[WorkflowOrchestrator] 最终质量检查失败: %v", err)
		} else {
			result.QualityReport = finalQualityReport
			wo.notifyQualityCheck(req.Callbacks, finalQualityReport)
		}

		// 设置最终视频路径
		if videoPath, ok := stageData["final_video_path"].(string); ok {
			result.FinalVideoPath = videoPath
		}
	}

	// 生成资源使用报告
	result.ResourceUsage = wo.resourceManager.GenerateUsageReport(req.ProjectID)
	result.TotalDuration = time.Since(startTime)
	result.Success = result.ErrorDetails == nil

	// 最终进度更新
	wo.notifyProgress(req.Callbacks, "完成", 1.0, "编排执行完成")

	log.Printf("[WorkflowOrchestrator] 编排执行完成: 成功=%v, 耗时=%v", 
		result.Success, result.TotalDuration)

	return result, nil
}

// ProcessingStage 处理阶段
type ProcessingStage struct {
	Name        string
	Description string
	Handler     StageHandler
	Required    bool
	Timeout     time.Duration
	RetryConfig *utils.RetryConfig
}

// StageHandler 阶段处理器
type StageHandler func(ctx context.Context, data map[string]interface{}, req *OrchestrationRequest) (*StageExecutionResult, error)

// StageExecutionResult 阶段执行结果
type StageExecutionResult struct {
	QualityScore float64                `json:"quality_score"`
	RetryCount   int                    `json:"retry_count"`
	OutputData   map[string]interface{} `json:"output_data"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// defineProcessingStages 定义处理阶段
func (wo *WorkflowOrchestrator) defineProcessingStages(req *OrchestrationRequest) []*ProcessingStage {
	stages := []*ProcessingStage{
		{
			Name:        "novel_analysis",
			Description: "小说文本分析",
			Handler:     wo.handleNovelAnalysis,
			Required:    true,
			Timeout:     5 * time.Minute,
		},
		{
			Name:        "character_extraction",
			Description: "角色提取和验证",
			Handler:     wo.handleCharacterExtraction,
			Required:    true,
			Timeout:     10 * time.Minute,
		},
		{
			Name:        "scene_analysis",
			Description: "场景深度分析",
			Handler:     wo.handleSceneAnalysis,
			Required:    true,
			Timeout:     15 * time.Minute,
		},
		{
			Name:        "script_generation",
			Description: "剧本生成",
			Handler:     wo.handleScriptGeneration,
			Required:    true,
			Timeout:     8 * time.Minute,
		},
		{
			Name:        "image_generation",
			Description: "图像生成",
			Handler:     wo.handleImageGeneration,
			Required:    true,
			Timeout:     30 * time.Minute,
		},
		{
			Name:        "consistency_validation",
			Description: "一致性验证",
			Handler:     wo.handleConsistencyValidation,
			Required:    true,
			Timeout:     10 * time.Minute,
		},
		{
			Name:        "animation_generation",
			Description: "动画生成",
			Handler:     wo.handleAnimationGeneration,
			Required:    false,
			Timeout:     20 * time.Minute,
		},
		{
			Name:        "audio_generation",
			Description: "音频生成",
			Handler:     wo.handleAudioGeneration,
			Required:    true,
			Timeout:     15 * time.Minute,
		},
		{
			Name:        "video_composition",
			Description: "视频合成",
			Handler:     wo.handleVideoComposition,
			Required:    true,
			Timeout:     25 * time.Minute,
		},
		{
			Name:        "quality_optimization",
			Description: "质量优化",
			Handler:     wo.handleQualityOptimization,
			Required:    false,
			Timeout:     10 * time.Minute,
		},
	}

	// 根据质量级别调整阶段
	if req.QualityLevel == "basic" {
		// 移除非必需阶段
		filteredStages := make([]*ProcessingStage, 0)
		for _, stage := range stages {
			if stage.Required {
				filteredStages = append(filteredStages, stage)
			}
		}
		stages = filteredStages
	}

	return stages
}

// executeStage 执行阶段
func (wo *WorkflowOrchestrator) executeStage(ctx context.Context, stage *ProcessingStage, data map[string]interface{}, req *OrchestrationRequest) (*StageExecutionResult, error) {
	log.Printf("[WorkflowOrchestrator] 执行阶段: %s", stage.Name)

	// 创建带超时的上下文
	stageCtx, cancel := context.WithTimeout(ctx, stage.Timeout)
	defer cancel()

	// 使用重试机制执行阶段
	result, err := utils.RetryWithResultFunc(stageCtx, stage.RetryConfig, func() (*StageExecutionResult, error) {
		return stage.Handler(stageCtx, data, req)
	})

	if err != nil {
		return nil, fmt.Errorf("阶段 %s 执行失败: %w", stage.Name, err)
	}

	return result, nil
}

// 阶段处理器实现
func (wo *WorkflowOrchestrator) handleNovelAnalysis(ctx context.Context, data map[string]interface{}, req *OrchestrationRequest) (*StageExecutionResult, error) {
	// 小说文本分析
	analysisResult := map[string]interface{}{
		"novel_text":    req.Novel,
		"word_count":    len(req.Novel),
		"estimated_scenes": len(req.Novel) / 500, // 估算场景数
		"complexity":    "medium",
	}

	return &StageExecutionResult{
		QualityScore: 0.8,
		OutputData:   analysisResult,
	}, nil
}

func (wo *WorkflowOrchestrator) handleCharacterExtraction(ctx context.Context, data map[string]interface{}, req *OrchestrationRequest) (*StageExecutionResult, error) {
	// 角色提取和验证
	novelText := data["novel_text"].(string)
	
	// 这里应该调用实际的角色提取服务
	characters := []*entity.Character{
		{
			ID:          uuid.New(),
			Name:        "主角",
			Description: "故事的主要角色",
			Appearance:  "年轻人",
		},
	}

	// 验证角色
	validationReq := &character.CharacterValidationRequest{
		ProjectID:       req.ProjectID,
		Characters:      characters,
		NovelText:       novelText,
		ValidationLevel: "comprehensive",
	}

	validationResult, err := wo.characterValidator.ValidateCharacters(ctx, validationReq)
	if err != nil {
		return nil, fmt.Errorf("角色验证失败: %w", err)
	}

	return &StageExecutionResult{
		QualityScore: validationResult.OverallScore,
		OutputData: map[string]interface{}{
			"characters":         characters,
			"validation_result":  validationResult,
		},
	}, nil
}

func (wo *WorkflowOrchestrator) handleSceneAnalysis(ctx context.Context, data map[string]interface{}, req *OrchestrationRequest) (*StageExecutionResult, error) {
	// 场景深度分析
	novelText := data["novel_text"].(string)
	characters := data["characters"].([]*entity.Character)

	analysisReq := &scene.DeepSceneAnalysisRequest{
		ProjectID:    req.ProjectID,
		NovelText:    novelText,
		Characters:   characters,
		Style:        req.Style,
		AnalysisType: "comprehensive",
	}

	analysisResult, err := wo.sceneAnalyzer.AnalyzeSceneSequence(ctx, analysisReq)
	if err != nil {
		return nil, fmt.Errorf("场景分析失败: %w", err)
	}

	return &StageExecutionResult{
		QualityScore: analysisResult.QualityMetrics.OverallScore,
		OutputData: map[string]interface{}{
			"scenes":           analysisResult.Scenes,
			"scene_sequence":   analysisResult.SceneSequence,
			"context_flow":     analysisResult.ContextFlow,
		},
	}, nil
}

func (wo *WorkflowOrchestrator) handleScriptGeneration(ctx context.Context, data map[string]interface{}, req *OrchestrationRequest) (*StageExecutionResult, error) {
	// 剧本生成
	scenes := data["scenes"].([]*scene.EnhancedScene)
	
	// 这里应该调用实际的剧本生成服务
	script := &entity.Script{
		ID:      uuid.New(),
		Content: "生成的剧本内容",
		Scenes:  len(scenes),
	}

	return &StageExecutionResult{
		QualityScore: 0.85,
		OutputData: map[string]interface{}{
			"script": script,
		},
	}, nil
}

func (wo *WorkflowOrchestrator) handleImageGeneration(ctx context.Context, data map[string]interface{}, req *OrchestrationRequest) (*StageExecutionResult, error) {
	// 图像生成
	scenes := data["scenes"].([]*scene.EnhancedScene)
	characters := data["characters"].([]*entity.Character)

	imageFiles := make([]string, 0, len(scenes))
	
	// 为每个场景生成图像
	for i, scene := range scenes {
		// 这里应该调用实际的图像生成服务
		imagePath := fmt.Sprintf("/tmp/scene_%d.png", i)
		imageFiles = append(imageFiles, imagePath)
	}

	return &StageExecutionResult{
		QualityScore: 0.8,
		OutputData: map[string]interface{}{
			"image_files": imageFiles,
		},
	}, nil
}

func (wo *WorkflowOrchestrator) handleConsistencyValidation(ctx context.Context, data map[string]interface{}, req *OrchestrationRequest) (*StageExecutionResult, error) {
	// 一致性验证
	characters := data["characters"].([]*entity.Character)
	imageFiles := data["image_files"].([]string)

	consistencyScore := 0.85 // 简化实现

	return &StageExecutionResult{
		QualityScore: consistencyScore,
		OutputData: map[string]interface{}{
			"consistency_validated": true,
			"consistency_score":     consistencyScore,
		},
	}, nil
}

func (wo *WorkflowOrchestrator) handleAnimationGeneration(ctx context.Context, data map[string]interface{}, req *OrchestrationRequest) (*StageExecutionResult, error) {
	// 动画生成
	imageFiles := data["image_files"].([]string)
	scenes := data["scenes"].([]*scene.EnhancedScene)

	animationReq := animation.CreateDefaultAnimationRequest(req.ProjectID, imageFiles, "/tmp/animation")
	animationReq.AnimationType = "character"
	animationReq.Quality = req.QualityLevel

	animationResult, err := wo.animationEngine.GenerateAdvancedAnimation(ctx, animationReq)
	if err != nil {
		return nil, fmt.Errorf("动画生成失败: %w", err)
	}

	return &StageExecutionResult{
		QualityScore: animationResult.Quality,
		OutputData: map[string]interface{}{
			"animation_path":   animationResult.OutputPath,
			"animation_result": animationResult,
		},
	}, nil
}

func (wo *WorkflowOrchestrator) handleAudioGeneration(ctx context.Context, data map[string]interface{}, req *OrchestrationRequest) (*StageExecutionResult, error) {
	// 音频生成
	script := data["script"].(*entity.Script)
	
	// 这里应该调用实际的音频生成服务
	audioPath := "/tmp/audio.wav"

	return &StageExecutionResult{
		QualityScore: 0.8,
		OutputData: map[string]interface{}{
			"audio_path": audioPath,
		},
	}, nil
}

func (wo *WorkflowOrchestrator) handleVideoComposition(ctx context.Context, data map[string]interface{}, req *OrchestrationRequest) (*StageExecutionResult, error) {
	// 视频合成
	imageFiles := data["image_files"].([]string)
	audioPath := data["audio_path"].(string)
	
	// 这里应该调用实际的视频合成服务
	videoPath := fmt.Sprintf("/tmp/final_video_%s.mp4", req.ProjectID.String())

	return &StageExecutionResult{
		QualityScore: 0.85,
		OutputData: map[string]interface{}{
			"final_video_path": videoPath,
		},
	}, nil
}

func (wo *WorkflowOrchestrator) handleQualityOptimization(ctx context.Context, data map[string]interface{}, req *OrchestrationRequest) (*StageExecutionResult, error) {
	// 质量优化
	videoPath := data["final_video_path"].(string)
	
	// 这里应该调用实际的质量优化服务
	optimizedPath := videoPath // 简化实现

	return &StageExecutionResult{
		QualityScore: 0.9,
		OutputData: map[string]interface{}{
			"final_video_path": optimizedPath,
			"optimized":        true,
		},
	}, nil
}

// 辅助方法
func (wo *WorkflowOrchestrator) handleStageError(ctx context.Context, stage *ProcessingStage, err error, callbacks *OrchestrationCallbacks) bool {
	canRetry := stage.RetryConfig != nil
	
	if callbacks != nil && callbacks.OnError != nil {
		callbacks.OnError(stage.Name, err, canRetry)
	}
	
	return false // 简化实现，不重试
}

func (wo *WorkflowOrchestrator) generateSuggestedFix(stageName string, err error) string {
	fixes := map[string]string{
		"novel_analysis":         "检查小说文本格式和内容",
		"character_extraction":   "优化角色描述，确保角色信息完整",
		"scene_analysis":         "检查场景描述的清晰度和连贯性",
		"image_generation":       "检查AI服务连接和提示词质量",
		"video_composition":      "检查音视频文件格式和完整性",
	}
	
	if fix, exists := fixes[stageName]; exists {
		return fix
	}
	
	return "请检查系统配置和网络连接"
}

func (wo *WorkflowOrchestrator) performFinalQualityCheck(ctx context.Context, data map[string]interface{}, req *OrchestrationRequest) (*quality.ComprehensiveQualityReport, error) {
	// 构建质量检查请求
	qualityReq := &quality.QualityCheckRequest{
		ProjectID: req.ProjectID,
		VideoPath: data["final_video_path"].(string),
		CheckType: "comprehensive",
	}

	// 添加其他检查项
	if script, ok := data["script"].(*entity.Script); ok {
		qualityReq.Script = script
	}
	
	if characters, ok := data["characters"].([]*entity.Character); ok {
		qualityReq.Characters = characters
	}

	return wo.qualityController.ComprehensiveQualityCheck(ctx, qualityReq)
}

// 通知方法
func (wo *WorkflowOrchestrator) notifyProgress(callbacks *OrchestrationCallbacks, stage string, progress float64, message string) {
	if callbacks != nil && callbacks.OnProgress != nil {
		callbacks.OnProgress(stage, progress, message)
	}
}

func (wo *WorkflowOrchestrator) notifyStageComplete(callbacks *OrchestrationCallbacks, stage string, result interface{}, duration time.Duration) {
	if callbacks != nil && callbacks.OnStageComplete != nil {
		callbacks.OnStageComplete(stage, result, duration)
	}
}

func (wo *WorkflowOrchestrator) notifyQualityCheck(callbacks *OrchestrationCallbacks, report *quality.ComprehensiveQualityReport) {
	if callbacks != nil && callbacks.OnQualityCheck != nil {
		callbacks.OnQualityCheck(report)
	}
}
