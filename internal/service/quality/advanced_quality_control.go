package quality

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"comic_video/internal/domain/entity"
	"comic_video/internal/utils"
)

// AdvancedQualityController 高级质量控制器
type AdvancedQualityController struct {
	aiService      AIService
	imageAnalyzer  *ImageQualityAnalyzer
	audioAnalyzer  *AudioQualityAnalyzer
	videoAnalyzer  *VideoQualityAnalyzer
	textAnalyzer   *TextQualityAnalyzer
	qualityRules   *QualityRuleEngine
}

// NewAdvancedQualityController 创建高级质量控制器
func NewAdvancedQualityController(aiService AIService) *AdvancedQualityController {
	return &AdvancedQualityController{
		aiService:     aiService,
		imageAnalyzer: NewImageQualityAnalyzer(),
		audioAnalyzer: NewAudioQualityAnalyzer(),
		videoAnalyzer: NewVideoQualityAnalyzer(),
		textAnalyzer:  NewTextQualityAnalyzer(),
		qualityRules:  NewQualityRuleEngine(),
	}
}

// ComprehensiveQualityCheck 综合质量检查
func (aqc *AdvancedQualityController) ComprehensiveQualityCheck(ctx context.Context, req *QualityCheckRequest) (*ComprehensiveQualityReport, error) {
	log.Printf("[AdvancedQualityController] 开始综合质量检查")
	
	report := &ComprehensiveQualityReport{
		ProjectID:   req.ProjectID,
		CheckTime:   time.Now(),
		CheckType:   "comprehensive",
		Components:  make(map[string]*ComponentQualityReport),
		Issues:      make([]*QualityIssue, 0),
		Suggestions: make([]*QualityImprovement, 0),
	}

	// 1. 文本质量检查
	if req.Script != nil {
		textReport, err := aqc.checkTextQuality(ctx, req.Script)
		if err != nil {
			log.Printf("[AdvancedQualityController] 文本质量检查失败: %v", err)
		} else {
			report.Components["text"] = textReport
			report.Issues = append(report.Issues, textReport.Issues...)
			report.Suggestions = append(report.Suggestions, textReport.Suggestions...)
		}
	}

	// 2. 角色质量检查
	if len(req.Characters) > 0 {
		characterReport, err := aqc.checkCharacterQuality(ctx, req.Characters)
		if err != nil {
			log.Printf("[AdvancedQualityController] 角色质量检查失败: %v", err)
		} else {
			report.Components["characters"] = characterReport
			report.Issues = append(report.Issues, characterReport.Issues...)
			report.Suggestions = append(report.Suggestions, characterReport.Suggestions...)
		}
	}

	// 3. 场景质量检查
	if len(req.Scenes) > 0 {
		sceneReport, err := aqc.checkSceneQuality(ctx, req.Scenes)
		if err != nil {
			log.Printf("[AdvancedQualityController] 场景质量检查失败: %v", err)
		} else {
			report.Components["scenes"] = sceneReport
			report.Issues = append(report.Issues, sceneReport.Issues...)
			report.Suggestions = append(report.Suggestions, sceneReport.Suggestions...)
		}
	}

	// 4. 图像质量检查
	if len(req.ImageFiles) > 0 {
		imageReport, err := aqc.checkImageQuality(ctx, req.ImageFiles)
		if err != nil {
			log.Printf("[AdvancedQualityController] 图像质量检查失败: %v", err)
		} else {
			report.Components["images"] = imageReport
			report.Issues = append(report.Issues, imageReport.Issues...)
			report.Suggestions = append(report.Suggestions, imageReport.Suggestions...)
		}
	}

	// 5. 音频质量检查
	if len(req.AudioFiles) > 0 {
		audioReport, err := aqc.checkAudioQuality(ctx, req.AudioFiles)
		if err != nil {
			log.Printf("[AdvancedQualityController] 音频质量检查失败: %v", err)
		} else {
			report.Components["audio"] = audioReport
			report.Issues = append(report.Issues, audioReport.Issues...)
			report.Suggestions = append(report.Suggestions, audioReport.Suggestions...)
		}
	}

	// 6. 视频质量检查
	if req.VideoPath != "" {
		videoReport, err := aqc.checkVideoQuality(ctx, req.VideoPath)
		if err != nil {
			log.Printf("[AdvancedQualityController] 视频质量检查失败: %v", err)
		} else {
			report.Components["video"] = videoReport
			report.Issues = append(report.Issues, videoReport.Issues...)
			report.Suggestions = append(report.Suggestions, videoReport.Suggestions...)
		}
	}

	// 7. 计算综合质量分数
	report.OverallScore = aqc.calculateOverallScore(report.Components)
	report.QualityLevel = aqc.determineQualityLevel(report.OverallScore)

	// 8. 生成质量改进建议
	improvementPlan := aqc.generateImprovementPlan(ctx, report)
	report.ImprovementPlan = improvementPlan

	log.Printf("[AdvancedQualityController] 综合质量检查完成: 总分=%.2f, 等级=%s", 
		report.OverallScore, report.QualityLevel)

	return report, nil
}

// checkTextQuality 检查文本质量
func (aqc *AdvancedQualityController) checkTextQuality(ctx context.Context, script *entity.Script) (*ComponentQualityReport, error) {
	report := &ComponentQualityReport{
		ComponentType: "text",
		Score:         0,
		Issues:        make([]*QualityIssue, 0),
		Suggestions:   make([]*QualityImprovement, 0),
		Metrics:       make(map[string]float64),
	}

	// 1. 基础文本分析
	textMetrics := aqc.textAnalyzer.AnalyzeText(script.Content)
	report.Metrics = textMetrics

	// 2. 语言质量检查
	languageScore, languageIssues := aqc.checkLanguageQuality(script.Content)
	report.Issues = append(report.Issues, languageIssues...)

	// 3. 结构质量检查
	structureScore, structureIssues := aqc.checkTextStructure(script.Content)
	report.Issues = append(report.Issues, structureIssues...)

	// 4. 内容连贯性检查
	coherenceScore, coherenceIssues := aqc.checkContentCoherence(ctx, script.Content)
	report.Issues = append(report.Issues, coherenceIssues...)

	// 5. 计算综合分数
	report.Score = (languageScore + structureScore + coherenceScore) / 3.0

	// 6. 生成改进建议
	if report.Score < 80 {
		report.Suggestions = append(report.Suggestions, &QualityImprovement{
			Type:        "text_improvement",
			Priority:    "high",
			Description: "文本质量需要改进",
			Action:      "重新生成或人工编辑文本内容",
			ExpectedImprovement: 15.0,
		})
	}

	return report, nil
}

// checkCharacterQuality 检查角色质量
func (aqc *AdvancedQualityController) checkCharacterQuality(ctx context.Context, characters []*entity.Character) (*ComponentQualityReport, error) {
	report := &ComponentQualityReport{
		ComponentType: "characters",
		Score:         0,
		Issues:        make([]*QualityIssue, 0),
		Suggestions:   make([]*QualityImprovement, 0),
		Metrics:       make(map[string]float64),
	}

	if len(characters) == 0 {
		report.Issues = append(report.Issues, &QualityIssue{
			Type:        "missing_characters",
			Severity:    "critical",
			Description: "没有检测到角色信息",
			Location:    "character_extraction",
		})
		return report, nil
	}

	var totalScore float64
	characterCount := len(characters)

	for i, character := range characters {
		// 1. 角色描述完整性检查
		descriptionScore := aqc.checkCharacterDescription(character)
		
		// 2. 角色一致性检查
		consistencyScore := aqc.checkCharacterConsistency(character)
		
		// 3. 角色独特性检查
		uniquenessScore := aqc.checkCharacterUniqueness(character, characters)

		characterScore := (descriptionScore + consistencyScore + uniquenessScore) / 3.0
		totalScore += characterScore

		// 记录低质量角色
		if characterScore < 70 {
			report.Issues = append(report.Issues, &QualityIssue{
				Type:        "low_quality_character",
				Severity:    "medium",
				Description: fmt.Sprintf("角色 '%s' 质量较低 (%.1f分)", character.Name, characterScore),
				Location:    fmt.Sprintf("character_%d", i),
			})
		}
	}

	report.Score = totalScore / float64(characterCount)
	report.Metrics["character_count"] = float64(characterCount)
	report.Metrics["average_character_score"] = report.Score

	// 生成改进建议
	if report.Score < 75 {
		report.Suggestions = append(report.Suggestions, &QualityImprovement{
			Type:        "character_enhancement",
			Priority:    "medium",
			Description: "角色质量需要提升",
			Action:      "重新生成角色描述或增加角色细节",
			ExpectedImprovement: 20.0,
		})
	}

	return report, nil
}

// checkSceneQuality 检查场景质量
func (aqc *AdvancedQualityController) checkSceneQuality(ctx context.Context, scenes []*entity.Scene) (*ComponentQualityReport, error) {
	report := &ComponentQualityReport{
		ComponentType: "scenes",
		Score:         0,
		Issues:        make([]*QualityIssue, 0),
		Suggestions:   make([]*QualityImprovement, 0),
		Metrics:       make(map[string]float64),
	}

	if len(scenes) == 0 {
		report.Issues = append(report.Issues, &QualityIssue{
			Type:        "missing_scenes",
			Severity:    "critical",
			Description: "没有检测到场景信息",
			Location:    "scene_extraction",
		})
		return report, nil
	}

	var totalScore float64
	sceneCount := len(scenes)

	for i, scene := range scenes {
		// 1. 场景描述完整性检查
		descriptionScore := aqc.checkSceneDescription(scene)
		
		// 2. 场景连贯性检查
		coherenceScore := aqc.checkSceneCoherence(scene, scenes)
		
		// 3. 场景视觉化程度检查
		visualScore := aqc.checkSceneVisualization(scene)

		sceneScore := (descriptionScore + coherenceScore + visualScore) / 3.0
		totalScore += sceneScore

		// 记录低质量场景
		if sceneScore < 70 {
			report.Issues = append(report.Issues, &QualityIssue{
				Type:        "low_quality_scene",
				Severity:    "medium",
				Description: fmt.Sprintf("场景 %d 质量较低 (%.1f分)", i+1, sceneScore),
				Location:    fmt.Sprintf("scene_%d", i),
			})
		}
	}

	report.Score = totalScore / float64(sceneCount)
	report.Metrics["scene_count"] = float64(sceneCount)
	report.Metrics["average_scene_score"] = report.Score

	// 生成改进建议
	if report.Score < 75 {
		report.Suggestions = append(report.Suggestions, &QualityImprovement{
			Type:        "scene_enhancement",
			Priority:    "medium",
			Description: "场景质量需要提升",
			Action:      "重新生成场景描述或增加场景细节",
			ExpectedImprovement: 18.0,
		})
	}

	return report, nil
}

// checkImageQuality 检查图像质量
func (aqc *AdvancedQualityController) checkImageQuality(ctx context.Context, imageFiles []string) (*ComponentQualityReport, error) {
	report := &ComponentQualityReport{
		ComponentType: "images",
		Score:         0,
		Issues:        make([]*QualityIssue, 0),
		Suggestions:   make([]*QualityImprovement, 0),
		Metrics:       make(map[string]float64),
	}

	if len(imageFiles) == 0 {
		report.Issues = append(report.Issues, &QualityIssue{
			Type:        "missing_images",
			Severity:    "critical",
			Description: "没有生成图像文件",
			Location:    "image_generation",
		})
		return report, nil
	}

	var totalScore float64
	validImageCount := 0

	for i, imagePath := range imageFiles {
		imageScore, imageIssues := aqc.imageAnalyzer.AnalyzeImage(imagePath)
		if imageScore > 0 {
			totalScore += imageScore
			validImageCount++
		}

		// 添加图像特定问题
		for _, issue := range imageIssues {
			issue.Location = fmt.Sprintf("image_%d", i)
			report.Issues = append(report.Issues, issue)
		}
	}

	if validImageCount > 0 {
		report.Score = totalScore / float64(validImageCount)
	}

	report.Metrics["total_images"] = float64(len(imageFiles))
	report.Metrics["valid_images"] = float64(validImageCount)
	report.Metrics["average_image_score"] = report.Score

	// 生成改进建议
	if report.Score < 70 {
		report.Suggestions = append(report.Suggestions, &QualityImprovement{
			Type:        "image_regeneration",
			Priority:    "high",
			Description: "图像质量不达标",
			Action:      "重新生成低质量图像",
			ExpectedImprovement: 25.0,
		})
	}

	return report, nil
}

// checkAudioQuality 检查音频质量
func (aqc *AdvancedQualityController) checkAudioQuality(ctx context.Context, audioFiles []string) (*ComponentQualityReport, error) {
	report := &ComponentQualityReport{
		ComponentType: "audio",
		Score:         0,
		Issues:        make([]*QualityIssue, 0),
		Suggestions:   make([]*QualityImprovement, 0),
		Metrics:       make(map[string]float64),
	}

	if len(audioFiles) == 0 {
		report.Issues = append(report.Issues, &QualityIssue{
			Type:        "missing_audio",
			Severity:    "critical",
			Description: "没有生成音频文件",
			Location:    "audio_generation",
		})
		return report, nil
	}

	var totalScore float64
	validAudioCount := 0

	for i, audioPath := range audioFiles {
		audioScore, audioIssues := aqc.audioAnalyzer.AnalyzeAudio(audioPath)
		if audioScore > 0 {
			totalScore += audioScore
			validAudioCount++
		}

		// 添加音频特定问题
		for _, issue := range audioIssues {
			issue.Location = fmt.Sprintf("audio_%d", i)
			report.Issues = append(report.Issues, issue)
		}
	}

	if validAudioCount > 0 {
		report.Score = totalScore / float64(validAudioCount)
	}

	report.Metrics["total_audio_files"] = float64(len(audioFiles))
	report.Metrics["valid_audio_files"] = float64(validAudioCount)
	report.Metrics["average_audio_score"] = report.Score

	// 生成改进建议
	if report.Score < 75 {
		report.Suggestions = append(report.Suggestions, &QualityImprovement{
			Type:        "audio_enhancement",
			Priority:    "medium",
			Description: "音频质量需要改进",
			Action:      "调整音频参数或重新生成",
			ExpectedImprovement: 20.0,
		})
	}

	return report, nil
}

// checkVideoQuality 检查视频质量
func (aqc *AdvancedQualityController) checkVideoQuality(ctx context.Context, videoPath string) (*ComponentQualityReport, error) {
	report := &ComponentQualityReport{
		ComponentType: "video",
		Score:         0,
		Issues:        make([]*QualityIssue, 0),
		Suggestions:   make([]*QualityImprovement, 0),
		Metrics:       make(map[string]float64),
	}

	if videoPath == "" {
		report.Issues = append(report.Issues, &QualityIssue{
			Type:        "missing_video",
			Severity:    "critical",
			Description: "没有生成视频文件",
			Location:    "video_generation",
		})
		return report, nil
	}

	// 分析视频质量
	videoScore, videoIssues := aqc.videoAnalyzer.AnalyzeVideo(videoPath)
	report.Score = videoScore
	report.Issues = append(report.Issues, videoIssues...)

	// 获取视频指标
	videoMetrics := aqc.videoAnalyzer.GetVideoMetrics(videoPath)
	report.Metrics = videoMetrics

	// 生成改进建议
	if report.Score < 80 {
		report.Suggestions = append(report.Suggestions, &QualityImprovement{
			Type:        "video_optimization",
			Priority:    "high",
			Description: "视频质量需要优化",
			Action:      "调整视频编码参数或重新合成",
			ExpectedImprovement: 15.0,
		})
	}

	return report, nil
}

// calculateOverallScore 计算综合质量分数
func (aqc *AdvancedQualityController) calculateOverallScore(components map[string]*ComponentQualityReport) float64 {
	if len(components) == 0 {
		return 0
	}

	// 权重配置
	weights := map[string]float64{
		"text":       0.15,
		"characters": 0.20,
		"scenes":     0.15,
		"images":     0.25,
		"audio":      0.15,
		"video":      0.10,
	}

	var weightedSum float64
	var totalWeight float64

	for componentType, report := range components {
		if weight, exists := weights[componentType]; exists {
			weightedSum += report.Score * weight
			totalWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	return weightedSum / totalWeight
}

// determineQualityLevel 确定质量等级
func (aqc *AdvancedQualityController) determineQualityLevel(score float64) string {
	switch {
	case score >= 90:
		return "excellent"
	case score >= 80:
		return "good"
	case score >= 70:
		return "acceptable"
	case score >= 60:
		return "poor"
	default:
		return "unacceptable"
	}
}

// generateImprovementPlan 生成改进计划
func (aqc *AdvancedQualityController) generateImprovementPlan(ctx context.Context, report *ComprehensiveQualityReport) *QualityImprovementPlan {
	plan := &QualityImprovementPlan{
		CurrentScore:    report.OverallScore,
		TargetScore:     math.Min(report.OverallScore+20, 95), // 目标提升20分，最高95分
		Actions:         make([]*ImprovementAction, 0),
		EstimatedTime:   0,
		EstimatedCost:   0,
		Priority:        "medium",
	}

	// 根据问题严重程度排序建议
	criticalIssues := 0
	highPriorityActions := 0

	for _, suggestion := range report.Suggestions {
		action := &ImprovementAction{
			Type:                suggestion.Type,
			Description:         suggestion.Description,
			Action:              suggestion.Action,
			Priority:            suggestion.Priority,
			ExpectedImprovement: suggestion.ExpectedImprovement,
			EstimatedTime:       aqc.estimateActionTime(suggestion.Type),
			EstimatedCost:       aqc.estimateActionCost(suggestion.Type),
		}

		plan.Actions = append(plan.Actions, action)
		plan.EstimatedTime += action.EstimatedTime
		plan.EstimatedCost += action.EstimatedCost

		if suggestion.Priority == "high" {
			highPriorityActions++
		}
	}

	for _, issue := range report.Issues {
		if issue.Severity == "critical" {
			criticalIssues++
		}
	}

	// 确定整体优先级
	if criticalIssues > 0 || highPriorityActions > 2 {
		plan.Priority = "high"
	} else if highPriorityActions > 0 {
		plan.Priority = "medium"
	} else {
		plan.Priority = "low"
	}

	return plan
}

// 辅助方法
func (aqc *AdvancedQualityController) checkLanguageQuality(text string) (float64, []*QualityIssue) {
	// 简化实现，实际应该使用NLP工具
	score := 80.0
	issues := make([]*QualityIssue, 0)

	// 检查文本长度
	if len(text) < 100 {
		score -= 20
		issues = append(issues, &QualityIssue{
			Type:        "text_too_short",
			Severity:    "medium",
			Description: "文本内容过短",
			Location:    "text_analysis",
		})
	}

	return score, issues
}

func (aqc *AdvancedQualityController) checkTextStructure(text string) (float64, []*QualityIssue) {
	score := 85.0
	issues := make([]*QualityIssue, 0)

	// 检查段落结构
	paragraphs := strings.Split(text, "\n\n")
	if len(paragraphs) < 3 {
		score -= 15
		issues = append(issues, &QualityIssue{
			Type:        "poor_structure",
			Severity:    "medium",
			Description: "文本结构不够清晰",
			Location:    "structure_analysis",
		})
	}

	return score, issues
}

func (aqc *AdvancedQualityController) checkContentCoherence(ctx context.Context, text string) (float64, []*QualityIssue) {
	// 这里应该使用AI模型进行连贯性分析
	score := 75.0
	issues := make([]*QualityIssue, 0)

	return score, issues
}

func (aqc *AdvancedQualityController) checkCharacterDescription(character *entity.Character) float64 {
	score := 70.0
	
	if character.Description != "" {
		score += 15
	}
	if character.Appearance != "" {
		score += 15
	}
	
	return math.Min(score, 100)
}

func (aqc *AdvancedQualityController) checkCharacterConsistency(character *entity.Character) float64 {
	// 简化实现，实际应该检查角色在不同场景中的一致性
	return 80.0
}

func (aqc *AdvancedQualityController) checkCharacterUniqueness(character *entity.Character, allCharacters []*entity.Character) float64 {
	// 简化实现，实际应该检查角色的独特性
	return 75.0
}

func (aqc *AdvancedQualityController) checkSceneDescription(scene *entity.Scene) float64 {
	score := 70.0
	
	if scene.Description != "" {
		score += 20
	}
	if scene.Location != "" {
		score += 10
	}
	
	return math.Min(score, 100)
}

func (aqc *AdvancedQualityController) checkSceneCoherence(scene *entity.Scene, allScenes []*entity.Scene) float64 {
	// 简化实现，实际应该检查场景之间的连贯性
	return 80.0
}

func (aqc *AdvancedQualityController) checkSceneVisualization(scene *entity.Scene) float64 {
	// 简化实现，实际应该检查场景的可视化程度
	return 75.0
}

func (aqc *AdvancedQualityController) estimateActionTime(actionType string) int {
	timeMap := map[string]int{
		"text_improvement":     30,  // 30分钟
		"character_enhancement": 45,  // 45分钟
		"scene_enhancement":    40,  // 40分钟
		"image_regeneration":   60,  // 60分钟
		"audio_enhancement":    25,  // 25分钟
		"video_optimization":   90,  // 90分钟
	}
	
	if time, exists := timeMap[actionType]; exists {
		return time
	}
	return 30 // 默认30分钟
}

func (aqc *AdvancedQualityController) estimateActionCost(actionType string) float64 {
	costMap := map[string]float64{
		"text_improvement":     0.5,  // 0.5元
		"character_enhancement": 1.0,  // 1元
		"scene_enhancement":    0.8,  // 0.8元
		"image_regeneration":   2.0,  // 2元
		"audio_enhancement":    1.5,  // 1.5元
		"video_optimization":   3.0,  // 3元
	}
	
	if cost, exists := costMap[actionType]; exists {
		return cost
	}
	return 1.0 // 默认1元
}
