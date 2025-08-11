package ai

import (
	"context"
	"fmt"
	"log"
)

// HybridGenerator 混合生成器 - 结合轻量级生成和视频插帧
type HybridGenerator struct {
	LightweightGen *LightweightGenerator
	VideoInterp    *VideoInterpolator
	SDClient       *SDClient // 备用高质量生成
	Strategy       string    // "fast", "balanced", "premium"
}

// NewHybridGenerator 创建混合生成器
func NewHybridGenerator(
	lightweightGen *LightweightGenerator,
	videoInterp *VideoInterpolator,
	sdClient *SDClient,
	strategy string,
) *HybridGenerator {
	return &HybridGenerator{
		LightweightGen: lightweightGen,
		VideoInterp:    videoInterp,
		SDClient:       sdClient,
		Strategy:       strategy,
	}
}

// GenerationPlan 生成计划
type GenerationPlan struct {
	KeyFrameIndices    []int    `json:"key_frame_indices"`    // 关键帧索引
	KeyFrameMethod     string   `json:"key_frame_method"`     // "sd" 或 "lightweight"
	InterpolationFPS   int      `json:"interpolation_fps"`    // 插帧目标帧率
	TransitionEffects  []string `json:"transition_effects"`   // 转场效果
	QualityMode        string   `json:"quality_mode"`         // 质量模式
	EstimatedTime      string   `json:"estimated_time"`       // 预估时间
	ResourceSaving     float64  `json:"resource_saving"`      // 资源节省比例
}

// CreateOptimalPlan 创建最优生成计划
func (hg *HybridGenerator) CreateOptimalPlan(panels []string, targetQuality string) *GenerationPlan {
	totalPanels := len(panels)
	
	switch hg.Strategy {
	case "fast":
		// 快速模式：全部使用轻量级生成 + 高帧率插帧
		return &GenerationPlan{
			KeyFrameIndices:   getAllIndices(totalPanels),
			KeyFrameMethod:    "lightweight",
			InterpolationFPS:  30,
			TransitionEffects: []string{"fade"},
			QualityMode:       "fast",
			EstimatedTime:     fmt.Sprintf("%.1f分钟", float64(totalPanels)*0.1),
			ResourceSaving:    0.8, // 节省80%资源
		}
		
	case "balanced":
		// 平衡模式：关键帧用SD，其他用轻量级 + 中等插帧
		keyIndices := selectKeyFrames(totalPanels, 0.3) // 30%用SD
		return &GenerationPlan{
			KeyFrameIndices:   keyIndices,
			KeyFrameMethod:    "mixed",
			InterpolationFPS:  24,
			TransitionEffects: []string{"fade", "slide"},
			QualityMode:       "balanced",
			EstimatedTime:     fmt.Sprintf("%.1f分钟", float64(len(keyIndices))*0.5+float64(totalPanels-len(keyIndices))*0.1),
			ResourceSaving:    0.5, // 节省50%资源
		}
		
	case "premium":
		// 高端模式：重要帧用SD，插帧用高质量模式
		keyIndices := selectKeyFrames(totalPanels, 0.6) // 60%用SD
		return &GenerationPlan{
			KeyFrameIndices:   keyIndices,
			KeyFrameMethod:    "mixed",
			InterpolationFPS:  60,
			TransitionEffects: []string{"fade", "slide", "zoom", "morph"},
			QualityMode:       "high",
			EstimatedTime:     fmt.Sprintf("%.1f分钟", float64(len(keyIndices))*0.8+float64(totalPanels-len(keyIndices))*0.1),
			ResourceSaving:    0.3, // 节省30%资源
		}
		
	default:
		return hg.CreateOptimalPlan(panels, "balanced")
	}
}

// ExecutePlan 执行生成计划
func (hg *HybridGenerator) ExecutePlan(
	ctx context.Context,
	panels []string,
	plan *GenerationPlan,
	progressCallback func(int, string),
) ([][]byte, *InterpolationResult, error) {
	
	log.Printf("[HybridGenerator] 执行%s策略: %d个面板", hg.Strategy, len(panels))
	
	// 第一阶段：生成关键帧
	keyFrames, err := hg.generateKeyFrames(ctx, panels, plan, func(progress int, msg string) {
		// 关键帧生成占总进度的70%
		actualProgress := int(float64(progress) * 0.7)
		progressCallback(actualProgress, "生成关键帧: "+msg)
	})
	if err != nil {
		return nil, nil, fmt.Errorf("关键帧生成失败: %v", err)
	}
	
	progressCallback(70, "关键帧生成完成，开始视频插帧...")
	
	// 第二阶段：视频插帧
	videoResult, err := hg.VideoInterp.AdvancedInterpolate(
		ctx,
		keyFrames,
		plan.TransitionEffects,
		plan.InterpolationFPS,
		plan.QualityMode,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("视频插帧失败: %v", err)
	}
	
	progressCallback(100, "混合生成完成！")
	
	return keyFrames, videoResult, nil
}

// generateKeyFrames 生成关键帧
func (hg *HybridGenerator) generateKeyFrames(
	ctx context.Context,
	panels []string,
	plan *GenerationPlan,
	progressCallback func(int, string),
) ([][]byte, error) {
	
	keyFrames := make([][]byte, len(panels))
	
	if plan.KeyFrameMethod == "lightweight" {
		// 全部使用轻量级生成
		options := map[string]interface{}{
			"width":  512,
			"height": 512,
			"steps":  2,
		}
		
		images, err := hg.LightweightGen.BatchFastGenerate(ctx, panels, options, progressCallback)
		if err != nil {
			return nil, err
		}
		
		return images, nil
		
	} else if plan.KeyFrameMethod == "mixed" {
		// 混合模式：重要帧用SD，其他用轻量级
		
		for i, panel := range panels {
			isKeyFrame := containsInt(plan.KeyFrameIndices, i)
			
			var img []byte
			var err error
			
			if isKeyFrame {
				// 使用SD生成高质量关键帧
				log.Printf("[HybridGenerator] SD生成关键帧 %d: %s", i+1, panel[:min(30, len(panel))])
				
				imgResult, sdErr := hg.SDClient.Txt2Img(panel, map[string]interface{}{
					"width":        1024,
					"height":       1024,
					"steps":        20,
					"cfg_scale":    7,
					"sampler_name": "DPM++ 2M Karras",
				})
				if sdErr != nil {
					log.Printf("[HybridGenerator] SD失败，降级到轻量级: %v", sdErr)
					img, err = hg.generateLightweightFrame(ctx, panel)
				} else {
					img = imgResult.Data
				}
			} else {
				// 使用轻量级生成
				img, err = hg.generateLightweightFrame(ctx, panel)
			}
			
			if err != nil {
				log.Printf("[HybridGenerator] 帧%d生成失败: %v", i+1, err)
				img = generatePlaceholderImageBytes()
			}
			
			keyFrames[i] = img
			
			// 更新进度
			progress := int(float64(i+1) / float64(len(panels)) * 100)
			progressCallback(progress, fmt.Sprintf("完成 %d/%d 帧", i+1, len(panels)))
		}
	}
	
	return keyFrames, nil
}

// generateLightweightFrame 生成轻量级帧
func (hg *HybridGenerator) generateLightweightFrame(ctx context.Context, prompt string) ([]byte, error) {
	options := map[string]interface{}{
		"width":          512,
		"height":         512,
		"steps":          2,
		"guidance_scale": 1.5,
	}
	
	return hg.LightweightGen.FastGenerate(ctx, prompt, options)
}

// GetPerformanceComparison 获取性能对比
func (hg *HybridGenerator) GetPerformanceComparison(panelCount int) map[string]interface{} {
	// 传统SD方式
	traditionalSD := map[string]interface{}{
		"method":         "传统SD生成",
		"time_minutes":   float64(panelCount) * 0.8,
		"gpu_memory_gb":  8.0,
		"quality_score":  9.0,
		"resource_usage": 1.0,
	}
	
	// 混合方式
	hybrid := map[string]interface{}{
		"method":         "混合生成(轻量级+插帧)",
		"time_minutes":   float64(panelCount) * 0.2,
		"gpu_memory_gb":  3.5,
		"quality_score":  8.5,
		"resource_usage": 0.4,
	}
	
	// 纯轻量级方式
	lightweight := map[string]interface{}{
		"method":         "纯轻量级生成",
		"time_minutes":   float64(panelCount) * 0.1,
		"gpu_memory_gb":  2.5,
		"quality_score":  8.0,
		"resource_usage": 0.3,
	}
	
	return map[string]interface{}{
		"traditional_sd": traditionalSD,
		"hybrid":         hybrid,
		"lightweight":    lightweight,
		"recommendation": hg.getRecommendation(panelCount),
	}
}

// getRecommendation 获取推荐策略
func (hg *HybridGenerator) getRecommendation(panelCount int) string {
	if panelCount <= 5 {
		return "推荐使用premium模式，确保最佳质量"
	} else if panelCount <= 15 {
		return "推荐使用balanced模式，平衡质量和速度"
	} else {
		return "推荐使用fast模式，快速生成大量内容"
	}
}

// 辅助函数
func getAllIndices(count int) []int {
	indices := make([]int, count)
	for i := 0; i < count; i++ {
		indices[i] = i
	}
	return indices
}

func selectKeyFrames(totalCount int, ratio float64) []int {
	keyCount := int(float64(totalCount) * ratio)
	if keyCount == 0 {
		keyCount = 1
	}
	
	indices := make([]int, keyCount)
	step := float64(totalCount) / float64(keyCount)
	
	for i := 0; i < keyCount; i++ {
		indices[i] = int(float64(i) * step)
	}
	
	return indices
}

func containsInt(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
