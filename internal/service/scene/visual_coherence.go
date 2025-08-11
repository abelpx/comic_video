package scene

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"
)

// VisualAnalyzer 视觉分析器
type VisualAnalyzer struct {
	// 可以集成计算机视觉模型
}

// NewVisualAnalyzer 创建视觉分析器
func NewVisualAnalyzer() *VisualAnalyzer {
	return &VisualAnalyzer{}
}

// AnalyzeVisualElements 分析视觉元素
func (va *VisualAnalyzer) AnalyzeVisualElements(ctx context.Context, description string) (*VisualElements, error) {
	log.Printf("[VisualAnalyzer] 分析视觉元素")

	elements := &VisualElements{
		Composition:   va.analyzeComposition(description),
		ColorPalette:  va.analyzeColorPalette(description),
		LightingStyle: va.analyzeLighting(description),
		CameraAngle:   va.analyzeCameraAngle(description),
		DepthOfField:  va.analyzeDepthOfField(description),
		VisualEffects: va.analyzeVisualEffects(description),
		ArtisticStyle: va.analyzeArtisticStyle(description),
		DetailLevel:   va.analyzeDetailLevel(description),
	}

	return elements, nil
}

// analyzeComposition 分析构图
func (va *VisualAnalyzer) analyzeComposition(description string) string {
	lowerDesc := strings.ToLower(description)
	
	// 构图关键词映射
	compositionKeywords := map[string]string{
		"中心":     "centered",
		"对称":     "symmetrical",
		"三分法":    "rule_of_thirds",
		"对角线":    "diagonal",
		"前景":     "foreground_focus",
		"背景":     "background_focus",
		"全景":     "wide_shot",
		"特写":     "close_up",
		"中景":     "medium_shot",
		"远景":     "long_shot",
	}
	
	for keyword, composition := range compositionKeywords {
		if strings.Contains(lowerDesc, keyword) {
			return composition
		}
	}
	
	// 默认构图
	return "balanced"
}

// analyzeColorPalette 分析色彩调色板
func (va *VisualAnalyzer) analyzeColorPalette(description string) []string {
	palette := make([]string, 0)
	lowerDesc := strings.ToLower(description)
	
	// 颜色关键词
	colorKeywords := map[string]string{
		"红色":   "#FF0000",
		"蓝色":   "#0000FF",
		"绿色":   "#00FF00",
		"黄色":   "#FFFF00",
		"紫色":   "#800080",
		"橙色":   "#FFA500",
		"粉色":   "#FFC0CB",
		"黑色":   "#000000",
		"白色":   "#FFFFFF",
		"灰色":   "#808080",
		"金色":   "#FFD700",
		"银色":   "#C0C0C0",
		"棕色":   "#A52A2A",
		"青色":   "#00FFFF",
		"暖色":   "warm_tones",
		"冷色":   "cool_tones",
		"明亮":   "bright_colors",
		"暗淡":   "muted_colors",
		"鲜艳":   "vibrant_colors",
		"柔和":   "soft_colors",
	}
	
	for keyword, color := range colorKeywords {
		if strings.Contains(lowerDesc, keyword) {
			palette = append(palette, color)
		}
	}
	
	// 如果没有找到颜色，使用默认调色板
	if len(palette) == 0 {
		palette = []string{"natural_tones"}
	}
	
	return palette
}

// analyzeLighting 分析光照
func (va *VisualAnalyzer) analyzeLighting(description string) string {
	lowerDesc := strings.ToLower(description)
	
	// 光照关键词映射
	lightingKeywords := map[string]string{
		"阳光":     "natural_sunlight",
		"月光":     "moonlight",
		"烛光":     "candlelight",
		"灯光":     "artificial_light",
		"明亮":     "bright_lighting",
		"昏暗":     "dim_lighting",
		"阴影":     "dramatic_shadows",
		"柔光":     "soft_lighting",
		"强光":     "harsh_lighting",
		"背光":     "backlighting",
		"侧光":     "side_lighting",
		"顶光":     "top_lighting",
		"黄昏":     "golden_hour",
		"黎明":     "dawn_lighting",
		"夜晚":     "night_lighting",
		"室内":     "indoor_lighting",
		"室外":     "outdoor_lighting",
	}
	
	for keyword, lighting := range lightingKeywords {
		if strings.Contains(lowerDesc, keyword) {
			return lighting
		}
	}
	
	return "natural_lighting"
}

// analyzeCameraAngle 分析摄像机角度
func (va *VisualAnalyzer) analyzeCameraAngle(description string) string {
	lowerDesc := strings.ToLower(description)
	
	// 摄像机角度关键词
	angleKeywords := map[string]string{
		"俯视":     "bird_eye_view",
		"仰视":     "worm_eye_view",
		"平视":     "eye_level",
		"侧面":     "side_view",
		"正面":     "front_view",
		"背面":     "back_view",
		"斜角":     "three_quarter_view",
		"低角度":    "low_angle",
		"高角度":    "high_angle",
		"倾斜":     "dutch_angle",
	}
	
	for keyword, angle := range angleKeywords {
		if strings.Contains(lowerDesc, keyword) {
			return angle
		}
	}
	
	return "eye_level"
}

// analyzeDepthOfField 分析景深
func (va *VisualAnalyzer) analyzeDepthOfField(description string) string {
	lowerDesc := strings.ToLower(description)
	
	if strings.Contains(lowerDesc, "模糊") || strings.Contains(lowerDesc, "虚化") {
		return "shallow_depth"
	} else if strings.Contains(lowerDesc, "清晰") || strings.Contains(lowerDesc, "锐利") {
		return "deep_depth"
	} else if strings.Contains(lowerDesc, "焦点") {
		return "selective_focus"
	}
	
	return "normal_depth"
}

// analyzeVisualEffects 分析视觉效果
func (va *VisualAnalyzer) analyzeVisualEffects(description string) []string {
	effects := make([]string, 0)
	lowerDesc := strings.ToLower(description)
	
	// 视觉效果关键词
	effectKeywords := []string{
		"光晕", "镜头光晕", "bloom", "glow",
		"粒子", "火花", "烟雾", "雾气",
		"反射", "折射", "透明", "半透明",
		"动态模糊", "运动模糊", "径向模糊",
		"色彩分离", "色差", "噪点", "颗粒",
		"HDR", "高动态范围", "过曝", "欠曝",
		"景深", "虚化", "焦外", "散景",
	}
	
	for _, effect := range effectKeywords {
		if strings.Contains(lowerDesc, effect) {
			effects = append(effects, effect)
		}
	}
	
	return effects
}

// analyzeArtisticStyle 分析艺术风格
func (va *VisualAnalyzer) analyzeArtisticStyle(description string) string {
	lowerDesc := strings.ToLower(description)
	
	// 艺术风格关键词
	styleKeywords := map[string]string{
		"动漫":     "anime",
		"卡通":     "cartoon",
		"写实":     "realistic",
		"油画":     "oil_painting",
		"水彩":     "watercolor",
		"素描":     "sketch",
		"漫画":     "manga",
		"像素":     "pixel_art",
		"矢量":     "vector_art",
		"3D":      "3d_render",
		"手绘":     "hand_drawn",
		"数字绘画":   "digital_painting",
		"概念艺术":   "concept_art",
		"插画":     "illustration",
		"摄影":     "photographic",
		"电影":     "cinematic",
		"游戏":     "game_art",
		"传统":     "traditional_art",
		"现代":     "modern_art",
		"抽象":     "abstract",
		"超现实":    "surreal",
	}
	
	for keyword, style := range styleKeywords {
		if strings.Contains(lowerDesc, keyword) {
			return style
		}
	}
	
	return "anime" // 默认动漫风格
}

// analyzeDetailLevel 分析细节层次
func (va *VisualAnalyzer) analyzeDetailLevel(description string) string {
	lowerDesc := strings.ToLower(description)
	
	if strings.Contains(lowerDesc, "精细") || strings.Contains(lowerDesc, "详细") || strings.Contains(lowerDesc, "复杂") {
		return "high_detail"
	} else if strings.Contains(lowerDesc, "简单") || strings.Contains(lowerDesc, "简洁") || strings.Contains(lowerDesc, "极简") {
		return "low_detail"
	} else if strings.Contains(lowerDesc, "适中") || strings.Contains(lowerDesc, "平衡") {
		return "medium_detail"
	}
	
	return "medium_detail"
}

// CoherenceChecker 连贯性检查器
type CoherenceChecker struct {
	// 可以集成连贯性分析模型
}

// NewCoherenceChecker 创建连贯性检查器
func NewCoherenceChecker() *CoherenceChecker {
	return &CoherenceChecker{}
}

// CoherenceReport 连贯性报告
type CoherenceReport struct {
	OverallScore      float64                    `json:"overall_score"`
	VisualCoherence   float64                    `json:"visual_coherence"`
	NarrativeCoherence float64                   `json:"narrative_coherence"`
	CharacterCoherence float64                   `json:"character_coherence"`
	ThematicCoherence float64                    `json:"thematic_coherence"`
	TemporalCoherence float64                    `json:"temporal_coherence"`
	Issues            []*CoherenceIssue          `json:"issues"`
	Suggestions       []*CoherenceSuggestion     `json:"suggestions"`
	DetailedAnalysis  map[string]interface{}     `json:"detailed_analysis"`
}

// CoherenceIssue 连贯性问题
type CoherenceIssue struct {
	Type        string  `json:"type"`        // visual, narrative, character, thematic, temporal
	Severity    string  `json:"severity"`    // low, medium, high, critical
	Description string  `json:"description"`
	SceneIndex  int     `json:"scene_index"`
	Impact      float64 `json:"impact"`      // 对整体连贯性的影响 (0-1)
}

// CoherenceSuggestion 连贯性建议
type CoherenceSuggestion struct {
	Type            string  `json:"type"`
	Priority        string  `json:"priority"`
	Description     string  `json:"description"`
	Action          string  `json:"action"`
	ExpectedImprovement float64 `json:"expected_improvement"`
}

// CheckCoherence 检查连贯性
func (cc *CoherenceChecker) CheckCoherence(ctx context.Context, scenes []*EnhancedScene) (*CoherenceReport, error) {
	log.Printf("[CoherenceChecker] 检查连贯性: %d个场景", len(scenes))

	report := &CoherenceReport{
		Issues:           make([]*CoherenceIssue, 0),
		Suggestions:      make([]*CoherenceSuggestion, 0),
		DetailedAnalysis: make(map[string]interface{}),
	}

	// 1. 视觉连贯性检查
	visualScore, visualIssues := cc.checkVisualCoherence(scenes)
	report.VisualCoherence = visualScore
	report.Issues = append(report.Issues, visualIssues...)

	// 2. 叙事连贯性检查
	narrativeScore, narrativeIssues := cc.checkNarrativeCoherence(scenes)
	report.NarrativeCoherence = narrativeScore
	report.Issues = append(report.Issues, narrativeIssues...)

	// 3. 角色连贯性检查
	characterScore, characterIssues := cc.checkCharacterCoherence(scenes)
	report.CharacterCoherence = characterScore
	report.Issues = append(report.Issues, characterIssues...)

	// 4. 主题连贯性检查
	thematicScore, thematicIssues := cc.checkThematicCoherence(scenes)
	report.ThematicCoherence = thematicScore
	report.Issues = append(report.Issues, thematicIssues...)

	// 5. 时间连贯性检查
	temporalScore, temporalIssues := cc.checkTemporalCoherence(scenes)
	report.TemporalCoherence = temporalScore
	report.Issues = append(report.Issues, temporalIssues...)

	// 6. 计算总体分数
	report.OverallScore = cc.calculateOverallScore(
		visualScore, narrativeScore, characterScore, thematicScore, temporalScore)

	// 7. 生成改进建议
	report.Suggestions = cc.generateSuggestions(report)

	log.Printf("[CoherenceChecker] 连贯性检查完成: 总分=%.2f", report.OverallScore)
	return report, nil
}

// checkVisualCoherence 检查视觉连贯性
func (cc *CoherenceChecker) checkVisualCoherence(scenes []*EnhancedScene) (float64, []*CoherenceIssue) {
	issues := make([]*CoherenceIssue, 0)
	score := 1.0

	if len(scenes) < 2 {
		return score, issues
	}

	// 检查艺术风格一致性
	baseStyle := scenes[0].VisualElements.ArtisticStyle
	for i, scene := range scenes[1:] {
		if scene.VisualElements.ArtisticStyle != baseStyle {
			issues = append(issues, &CoherenceIssue{
				Type:        "visual",
				Severity:    "medium",
				Description: fmt.Sprintf("场景%d的艺术风格与基准不一致", i+1),
				SceneIndex:  i + 1,
				Impact:      0.1,
			})
			score -= 0.1
		}
	}

	// 检查色彩一致性
	score -= cc.checkColorConsistency(scenes, &issues)

	// 检查光照一致性
	score -= cc.checkLightingConsistency(scenes, &issues)

	if score < 0 {
		score = 0
	}

	return score, issues
}

// checkNarrativeCoherence 检查叙事连贯性
func (cc *CoherenceChecker) checkNarrativeCoherence(scenes []*EnhancedScene) (float64, []*CoherenceIssue) {
	issues := make([]*CoherenceIssue, 0)
	score := 1.0

	// 检查场景逻辑连接
	for i := 0; i < len(scenes)-1; i++ {
		if !cc.isLogicalTransition(scenes[i], scenes[i+1]) {
			issues = append(issues, &CoherenceIssue{
				Type:        "narrative",
				Severity:    "high",
				Description: fmt.Sprintf("场景%d到场景%d的过渡缺乏逻辑性", i, i+1),
				SceneIndex:  i,
				Impact:      0.15,
			})
			score -= 0.15
		}
	}

	// 检查情节发展
	if !cc.hasProgressiveNarrative(scenes) {
		issues = append(issues, &CoherenceIssue{
			Type:        "narrative",
			Severity:    "medium",
			Description: "整体情节发展缺乏递进性",
			SceneIndex:  -1,
			Impact:      0.2,
		})
		score -= 0.2
	}

	if score < 0 {
		score = 0
	}

	return score, issues
}

// checkCharacterCoherence 检查角色连贯性
func (cc *CoherenceChecker) checkCharacterCoherence(scenes []*EnhancedScene) (float64, []*CoherenceIssue) {
	issues := make([]*CoherenceIssue, 0)
	score := 1.0

	// 检查角色出现的连续性
	characterAppearances := cc.trackCharacterAppearances(scenes)
	
	for character, appearances := range characterAppearances {
		if cc.hasInconsistentAppearances(appearances) {
			issues = append(issues, &CoherenceIssue{
				Type:        "character",
				Severity:    "medium",
				Description: fmt.Sprintf("角色%s的出现缺乏连续性", character),
				SceneIndex:  -1,
				Impact:      0.1,
			})
			score -= 0.1
		}
	}

	if score < 0 {
		score = 0
	}

	return score, issues
}

// checkThematicCoherence 检查主题连贯性
func (cc *CoherenceChecker) checkThematicCoherence(scenes []*EnhancedScene) (float64, []*CoherenceIssue) {
	issues := make([]*CoherenceIssue, 0)
	score := 1.0

	// 检查主题一致性
	allThemes := cc.collectAllThemes(scenes)
	if len(allThemes) == 0 {
		issues = append(issues, &CoherenceIssue{
			Type:        "thematic",
			Severity:    "low",
			Description: "缺乏明确的主题",
			SceneIndex:  -1,
			Impact:      0.1,
		})
		score -= 0.1
	}

	// 检查主题发展
	if !cc.hasThematicProgression(scenes) {
		issues = append(issues, &CoherenceIssue{
			Type:        "thematic",
			Severity:    "medium",
			Description: "主题发展缺乏层次性",
			SceneIndex:  -1,
			Impact:      0.15,
		})
		score -= 0.15
	}

	if score < 0 {
		score = 0
	}

	return score, issues
}

// checkTemporalCoherence 检查时间连贯性
func (cc *CoherenceChecker) checkTemporalCoherence(scenes []*EnhancedScene) (float64, []*CoherenceIssue) {
	issues := make([]*CoherenceIssue, 0)
	score := 1.0

	// 检查时间顺序
	for i := 0; i < len(scenes)-1; i++ {
		if !cc.isValidTimeProgression(scenes[i], scenes[i+1]) {
			issues = append(issues, &CoherenceIssue{
				Type:        "temporal",
				Severity:    "medium",
				Description: fmt.Sprintf("场景%d到场景%d的时间进展不合理", i, i+1),
				SceneIndex:  i,
				Impact:      0.1,
			})
			score -= 0.1
		}
	}

	if score < 0 {
		score = 0
	}

	return score, issues
}

// 辅助方法
func (cc *CoherenceChecker) checkColorConsistency(scenes []*EnhancedScene, issues *[]*CoherenceIssue) float64 {
	// 简化的色彩一致性检查
	penalty := 0.0
	
	if len(scenes) < 2 {
		return penalty
	}
	
	// 检查相邻场景的色彩差异
	for i := 0; i < len(scenes)-1; i++ {
		if cc.hasSignificantColorDifference(scenes[i].VisualElements, scenes[i+1].VisualElements) {
			*issues = append(*issues, &CoherenceIssue{
				Type:        "visual",
				Severity:    "low",
				Description: fmt.Sprintf("场景%d和场景%d的色彩差异较大", i, i+1),
				SceneIndex:  i,
				Impact:      0.05,
			})
			penalty += 0.05
		}
	}
	
	return penalty
}

func (cc *CoherenceChecker) checkLightingConsistency(scenes []*EnhancedScene, issues *[]*CoherenceIssue) float64 {
	// 简化的光照一致性检查
	penalty := 0.0
	
	for i := 0; i < len(scenes)-1; i++ {
		if cc.hasInconsistentLighting(scenes[i].VisualElements, scenes[i+1].VisualElements) {
			*issues = append(*issues, &CoherenceIssue{
				Type:        "visual",
				Severity:    "low",
				Description: fmt.Sprintf("场景%d和场景%d的光照不一致", i, i+1),
				SceneIndex:  i,
				Impact:      0.05,
			})
			penalty += 0.05
		}
	}
	
	return penalty
}

func (cc *CoherenceChecker) isLogicalTransition(scene1, scene2 *EnhancedScene) bool {
	// 简化的逻辑过渡检查
	// 检查是否有共同角色或相关主题
	return cc.hasCommonElements(scene1, scene2)
}

func (cc *CoherenceChecker) hasProgressiveNarrative(scenes []*EnhancedScene) bool {
	// 检查情感强度是否有变化
	if len(scenes) < 3 {
		return true
	}
	
	intensities := make([]float64, len(scenes))
	for i, scene := range scenes {
		intensities[i] = scene.EmotionalTone.Intensity
	}
	
	// 检查是否有情感起伏
	hasVariation := false
	for i := 1; i < len(intensities); i++ {
		if math.Abs(intensities[i]-intensities[i-1]) > 0.2 {
			hasVariation = true
			break
		}
	}
	
	return hasVariation
}

func (cc *CoherenceChecker) trackCharacterAppearances(scenes []*EnhancedScene) map[string][]int {
	appearances := make(map[string][]int)
	
	for i, scene := range scenes {
		for _, character := range scene.Characters {
			if _, exists := appearances[character]; !exists {
				appearances[character] = make([]int, 0)
			}
			appearances[character] = append(appearances[character], i)
		}
	}
	
	return appearances
}

func (cc *CoherenceChecker) hasInconsistentAppearances(appearances []int) bool {
	// 检查角色出现是否有大的间隔
	if len(appearances) < 2 {
		return false
	}
	
	for i := 1; i < len(appearances); i++ {
		if appearances[i]-appearances[i-1] > 3 { // 间隔超过3个场景
			return true
		}
	}
	
	return false
}

func (cc *CoherenceChecker) collectAllThemes(scenes []*EnhancedScene) []string {
	themeSet := make(map[string]bool)
	
	for _, scene := range scenes {
		for _, theme := range scene.SemanticContext.Themes {
			themeSet[theme] = true
		}
	}
	
	themes := make([]string, 0, len(themeSet))
	for theme := range themeSet {
		themes = append(themes, theme)
	}
	
	return themes
}

func (cc *CoherenceChecker) hasThematicProgression(scenes []*EnhancedScene) bool {
	// 简化的主题发展检查
	return len(scenes) > 0 && len(scenes[0].SemanticContext.Themes) > 0
}

func (cc *CoherenceChecker) isValidTimeProgression(scene1, scene2 *EnhancedScene) bool {
	// 简化的时间进展检查
	// 如果时间信息不明确，认为是有效的
	if scene1.TimeOfDay == "未指定" || scene2.TimeOfDay == "未指定" {
		return true
	}
	
	// 检查时间顺序是否合理
	return true // 简化实现
}

func (cc *CoherenceChecker) hasSignificantColorDifference(visual1, visual2 *VisualElements) bool {
	// 简化的色彩差异检查
	return len(visual1.ColorPalette) != len(visual2.ColorPalette)
}

func (cc *CoherenceChecker) hasInconsistentLighting(visual1, visual2 *VisualElements) bool {
	// 简化的光照一致性检查
	return visual1.LightingStyle != visual2.LightingStyle
}

func (cc *CoherenceChecker) hasCommonElements(scene1, scene2 *EnhancedScene) bool {
	// 检查是否有共同角色
	for _, char1 := range scene1.Characters {
		for _, char2 := range scene2.Characters {
			if char1 == char2 {
				return true
			}
		}
	}
	
	// 检查是否有共同主题
	for _, theme1 := range scene1.SemanticContext.Themes {
		for _, theme2 := range scene2.SemanticContext.Themes {
			if theme1 == theme2 {
				return true
			}
		}
	}
	
	return false
}

func (cc *CoherenceChecker) calculateOverallScore(visual, narrative, character, thematic, temporal float64) float64 {
	// 加权平均
	weights := map[string]float64{
		"visual":    0.25,
		"narrative": 0.30,
		"character": 0.20,
		"thematic":  0.15,
		"temporal":  0.10,
	}
	
	return visual*weights["visual"] +
		narrative*weights["narrative"] +
		character*weights["character"] +
		thematic*weights["thematic"] +
		temporal*weights["temporal"]
}

func (cc *CoherenceChecker) generateSuggestions(report *CoherenceReport) []*CoherenceSuggestion {
	suggestions := make([]*CoherenceSuggestion, 0)
	
	// 根据问题生成建议
	for _, issue := range report.Issues {
		switch issue.Type {
		case "visual":
			suggestions = append(suggestions, &CoherenceSuggestion{
				Type:        "visual_improvement",
				Priority:    issue.Severity,
				Description: "改善视觉一致性",
				Action:      "统一艺术风格和色彩方案",
				ExpectedImprovement: issue.Impact,
			})
		case "narrative":
			suggestions = append(suggestions, &CoherenceSuggestion{
				Type:        "narrative_improvement",
				Priority:    issue.Severity,
				Description: "改善叙事连贯性",
				Action:      "优化场景过渡和情节发展",
				ExpectedImprovement: issue.Impact,
			})
		}
	}
	
	return suggestions
}

// SceneSequence 场景序列
type SceneSequence struct {
	Scenes       []*EnhancedScene `json:"scenes"`
	TotalDuration float64         `json:"total_duration"`
	SceneDurations []float64      `json:"scene_durations"`
	Transitions   []string        `json:"transitions"`
	Pacing        string          `json:"pacing"`
}

// VisualConsistencyReport 视觉一致性报告
type VisualConsistencyReport struct {
	OverallConsistency float64            `json:"overall_consistency"`
	StyleConsistency   float64            `json:"style_consistency"`
	ColorConsistency   float64            `json:"color_consistency"`
	LightingConsistency float64           `json:"lighting_consistency"`
	Issues             []string           `json:"issues"`
	Recommendations    []string           `json:"recommendations"`
}

// SceneQualityMetrics 场景质量指标
type SceneQualityMetrics struct {
	OverallScore       float64 `json:"overall_score"`
	AverageSceneScore  float64 `json:"average_scene_score"`
	SceneCount         int     `json:"scene_count"`
	CoherenceScore     float64 `json:"coherence_score"`
	VisualQuality      float64 `json:"visual_quality"`
	NarrativeQuality   float64 `json:"narrative_quality"`
	CharacterDevelopment float64 `json:"character_development"`
	ThematicDepth      float64 `json:"thematic_depth"`
}

// generateSceneSequence 生成场景序列
func (dsa *DeepSceneAnalyzer) generateSceneSequence(scenes []*EnhancedScene, contextFlow *ContextFlow) *SceneSequence {
	durations := make([]float64, len(scenes))
	transitions := make([]string, len(scenes)-1)
	
	// 计算场景时长
	for i, scene := range scenes {
		// 基于情感强度和内容复杂度计算时长
		baseDuration := 3.0 // 基础3秒
		intensityFactor := scene.EmotionalTone.Intensity * 2.0 // 情感强度影响
		complexityFactor := float64(len(scene.Description)) / 100.0 // 描述长度影响
		
		durations[i] = baseDuration + intensityFactor + complexityFactor
		if durations[i] > 8.0 {
			durations[i] = 8.0 // 最大8秒
		}
	}
	
	// 确定过渡类型
	for i := 0; i < len(scenes)-1; i++ {
		transitions[i] = scenes[i].TransitionHints.TransitionType
	}
	
	// 计算总时长
	totalDuration := 0.0
	for _, duration := range durations {
		totalDuration += duration
	}
	
	// 确定节奏
	pacing := "normal"
	if totalDuration < float64(len(scenes))*2 {
		pacing = "fast"
	} else if totalDuration > float64(len(scenes))*5 {
		pacing = "slow"
	}
	
	return &SceneSequence{
		Scenes:         scenes,
		TotalDuration:  totalDuration,
		SceneDurations: durations,
		Transitions:    transitions,
		Pacing:         pacing,
	}
}

// analyzeVisualConsistency 分析视觉一致性
func (dsa *DeepSceneAnalyzer) analyzeVisualConsistency(scenes []*EnhancedScene) *VisualConsistencyReport {
	if len(scenes) == 0 {
		return &VisualConsistencyReport{
			OverallConsistency: 1.0,
		}
	}
	
	// 计算各种一致性分数
	styleScore := dsa.calculateStyleConsistency(scenes)
	colorScore := dsa.calculateColorConsistency(scenes)
	lightingScore := dsa.calculateLightingConsistency(scenes)
	
	overallScore := (styleScore + colorScore + lightingScore) / 3.0
	
	issues := make([]string, 0)
	recommendations := make([]string, 0)
	
	if styleScore < 0.8 {
		issues = append(issues, "艺术风格不够一致")
		recommendations = append(recommendations, "统一艺术风格设定")
	}
	
	if colorScore < 0.7 {
		issues = append(issues, "色彩方案变化过大")
		recommendations = append(recommendations, "建立统一的色彩调色板")
	}
	
	if lightingScore < 0.7 {
		issues = append(issues, "光照风格不一致")
		recommendations = append(recommendations, "保持光照风格的连贯性")
	}
	
	return &VisualConsistencyReport{
		OverallConsistency:  overallScore,
		StyleConsistency:    styleScore,
		ColorConsistency:    colorScore,
		LightingConsistency: lightingScore,
		Issues:              issues,
		Recommendations:     recommendations,
	}
}

// calculateQualityMetrics 计算质量指标
func (dsa *DeepSceneAnalyzer) calculateQualityMetrics(scenes []*EnhancedScene, coherenceReport *CoherenceReport) *SceneQualityMetrics {
	if len(scenes) == 0 {
		return &SceneQualityMetrics{}
	}
	
	// 计算平均场景分数
	totalScore := 0.0
	for _, scene := range scenes {
		totalScore += scene.QualityScore
	}
	averageScore := totalScore / float64(len(scenes))
	
	// 计算各维度质量
	visualQuality := dsa.calculateVisualQuality(scenes)
	narrativeQuality := dsa.calculateNarrativeQuality(scenes)
	characterDevelopment := dsa.calculateCharacterDevelopment(scenes)
	thematicDepth := dsa.calculateThematicDepth(scenes)
	
	// 计算总体分数
	overallScore := (averageScore + coherenceReport.OverallScore + visualQuality + narrativeQuality) / 4.0
	
	return &SceneQualityMetrics{
		OverallScore:         overallScore,
		AverageSceneScore:    averageScore,
		SceneCount:           len(scenes),
		CoherenceScore:       coherenceReport.OverallScore,
		VisualQuality:        visualQuality,
		NarrativeQuality:     narrativeQuality,
		CharacterDevelopment: characterDevelopment,
		ThematicDepth:        thematicDepth,
	}
}

// 辅助计算方法
func (dsa *DeepSceneAnalyzer) calculateStyleConsistency(scenes []*EnhancedScene) float64 {
	if len(scenes) <= 1 {
		return 1.0
	}
	
	baseStyle := scenes[0].VisualElements.ArtisticStyle
	consistent := 0
	
	for _, scene := range scenes {
		if scene.VisualElements.ArtisticStyle == baseStyle {
			consistent++
		}
	}
	
	return float64(consistent) / float64(len(scenes))
}

func (dsa *DeepSceneAnalyzer) calculateColorConsistency(scenes []*EnhancedScene) float64 {
	// 简化的色彩一致性计算
	return 0.8
}

func (dsa *DeepSceneAnalyzer) calculateLightingConsistency(scenes []*EnhancedScene) float64 {
	// 简化的光照一致性计算
	return 0.75
}

func (dsa *DeepSceneAnalyzer) calculateVisualQuality(scenes []*EnhancedScene) float64 {
	// 简化的视觉质量计算
	return 0.8
}

func (dsa *DeepSceneAnalyzer) calculateNarrativeQuality(scenes []*EnhancedScene) float64 {
	// 简化的叙事质量计算
	return 0.75
}

func (dsa *DeepSceneAnalyzer) calculateCharacterDevelopment(scenes []*EnhancedScene) float64 {
	// 简化的角色发展计算
	return 0.7
}

func (dsa *DeepSceneAnalyzer) calculateThematicDepth(scenes []*EnhancedScene) float64 {
	// 简化的主题深度计算
	return 0.65
}

// AIService 接口定义
type AIService interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
}
