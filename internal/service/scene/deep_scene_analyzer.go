package scene

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"comic_video/internal/domain/entity"
	"comic_video/internal/utils"
	"github.com/google/uuid"
)

// DeepSceneAnalyzer 深度场景分析器
type DeepSceneAnalyzer struct {
	aiService         AIService
	semanticAnalyzer  *SemanticAnalyzer
	contextTracker    *ContextTracker
	visualAnalyzer    *VisualAnalyzer
	coherenceChecker  *CoherenceChecker
	cacheManager      *utils.CacheManager
}

// NewDeepSceneAnalyzer 创建深度场景分析器
func NewDeepSceneAnalyzer(aiService AIService, cacheManager *utils.CacheManager) *DeepSceneAnalyzer {
	return &DeepSceneAnalyzer{
		aiService:        aiService,
		semanticAnalyzer: NewSemanticAnalyzer(aiService),
		contextTracker:   NewContextTracker(),
		visualAnalyzer:   NewVisualAnalyzer(),
		coherenceChecker: NewCoherenceChecker(),
		cacheManager:     cacheManager,
	}
}

// DeepSceneAnalysisRequest 深度场景分析请求
type DeepSceneAnalysisRequest struct {
	ProjectID    uuid.UUID `json:"project_id"`
	NovelText    string    `json:"novel_text"`
	Characters   []*entity.Character `json:"characters"`
	Style        string    `json:"style"`
	AnalysisType string    `json:"analysis_type"` // comprehensive, quick, focused
	Options      map[string]interface{} `json:"options"`
}

// DeepSceneAnalysisResult 深度场景分析结果
type DeepSceneAnalysisResult struct {
	ProjectID        uuid.UUID                `json:"project_id"`
	Scenes           []*EnhancedScene         `json:"scenes"`
	SceneSequence    *SceneSequence           `json:"scene_sequence"`
	ContextFlow      *ContextFlow             `json:"context_flow"`
	VisualConsistency *VisualConsistencyReport `json:"visual_consistency"`
	QualityMetrics   *SceneQualityMetrics     `json:"quality_metrics"`
	ProcessingTime   time.Duration            `json:"processing_time"`
}

// EnhancedScene 增强场景
type EnhancedScene struct {
	*entity.Scene
	SemanticContext   *SemanticContext   `json:"semantic_context"`
	VisualElements    *VisualElements    `json:"visual_elements"`
	EmotionalTone     *EmotionalTone     `json:"emotional_tone"`
	CinematicStyle    *CinematicStyle    `json:"cinematic_style"`
	TransitionHints   *TransitionHints   `json:"transition_hints"`
	QualityScore      float64            `json:"quality_score"`
	GenerationPrompt  string             `json:"generation_prompt"`
}

// SemanticContext 语义上下文
type SemanticContext struct {
	MainAction       string            `json:"main_action"`       // 主要动作
	SubActions       []string          `json:"sub_actions"`       // 次要动作
	ObjectsPresent   []string          `json:"objects_present"`   // 存在的物体
	SpatialRelations []string          `json:"spatial_relations"` // 空间关系
	TemporalMarkers  []string          `json:"temporal_markers"`  // 时间标记
	CausalLinks      []string          `json:"causal_links"`      // 因果关系
	Themes           []string          `json:"themes"`            // 主题
	Symbolism        []string          `json:"symbolism"`         // 象征意义
}

// VisualElements 视觉元素
type VisualElements struct {
	Composition      string   `json:"composition"`       // 构图
	ColorPalette     []string `json:"color_palette"`     // 色彩调色板
	LightingStyle    string   `json:"lighting_style"`    // 光照风格
	CameraAngle      string   `json:"camera_angle"`      // 摄像机角度
	DepthOfField     string   `json:"depth_of_field"`    // 景深
	VisualEffects    []string `json:"visual_effects"`    // 视觉效果
	ArtisticStyle    string   `json:"artistic_style"`    // 艺术风格
	DetailLevel      string   `json:"detail_level"`      // 细节层次
}

// EmotionalTone 情感基调
type EmotionalTone struct {
	PrimaryEmotion   string            `json:"primary_emotion"`   // 主要情感
	SecondaryEmotions []string         `json:"secondary_emotions"` // 次要情感
	Intensity        float64           `json:"intensity"`         // 强度 (0-1)
	Mood             string            `json:"mood"`              // 情绪
	Atmosphere       string            `json:"atmosphere"`        // 氛围
	TensionLevel     float64           `json:"tension_level"`     // 紧张程度
	EmotionalArc     string            `json:"emotional_arc"`     // 情感弧线
}

// CinematicStyle 电影风格
type CinematicStyle struct {
	ShotType         string   `json:"shot_type"`         // 镜头类型
	MovementStyle    string   `json:"movement_style"`    // 运动风格
	EditingRhythm    string   `json:"editing_rhythm"`    // 剪辑节奏
	VisualMetaphors  []string `json:"visual_metaphors"`  // 视觉隐喻
	GenreConventions []string `json:"genre_conventions"` // 类型惯例
	DirectorStyle    string   `json:"director_style"`    // 导演风格
}

// TransitionHints 过渡提示
type TransitionHints struct {
	TransitionType   string            `json:"transition_type"`   // 过渡类型
	TransitionSpeed  string            `json:"transition_speed"`  // 过渡速度
	VisualBridge     string            `json:"visual_bridge"`     // 视觉桥接
	AudioBridge      string            `json:"audio_bridge"`      // 音频桥接
	MotionContinuity string            `json:"motion_continuity"` // 动作连续性
	ThematicLink     string            `json:"thematic_link"`     // 主题联系
}

// AnalyzeSceneSequence 分析场景序列
func (dsa *DeepSceneAnalyzer) AnalyzeSceneSequence(ctx context.Context, req *DeepSceneAnalysisRequest) (*DeepSceneAnalysisResult, error) {
	startTime := time.Now()
	log.Printf("[DeepSceneAnalyzer] 开始深度场景分析: %s", req.ProjectID)

	// 1. 检查缓存
	cacheKey := fmt.Sprintf("deep_scene_analysis:%s", req.ProjectID.String())
	var cachedResult DeepSceneAnalysisResult
	if err := dsa.cacheManager.Get(ctx, cacheKey, &cachedResult); err == nil {
		log.Printf("[DeepSceneAnalyzer] 从缓存获取分析结果")
		return &cachedResult, nil
	}

	// 2. 基础场景提取
	basicScenes, err := dsa.extractBasicScenes(ctx, req.NovelText)
	if err != nil {
		return nil, fmt.Errorf("基础场景提取失败: %w", err)
	}

	// 3. 语义分析增强
	enhancedScenes, err := dsa.enhanceWithSemanticAnalysis(ctx, basicScenes, req)
	if err != nil {
		return nil, fmt.Errorf("语义分析增强失败: %w", err)
	}

	// 4. 视觉分析增强
	visuallyEnhancedScenes, err := dsa.enhanceWithVisualAnalysis(ctx, enhancedScenes, req)
	if err != nil {
		return nil, fmt.Errorf("视觉分析增强失败: %w", err)
	}

	// 5. 上下文跟踪
	contextFlow, err := dsa.contextTracker.TrackContextFlow(ctx, visuallyEnhancedScenes)
	if err != nil {
		return nil, fmt.Errorf("上下文跟踪失败: %w", err)
	}

	// 6. 连贯性检查
	coherenceReport, err := dsa.coherenceChecker.CheckCoherence(ctx, visuallyEnhancedScenes)
	if err != nil {
		log.Printf("[DeepSceneAnalyzer] 连贯性检查失败: %v", err)
		coherenceReport = &CoherenceReport{OverallScore: 0.7} // 默认分数
	}

	// 7. 生成场景序列
	sceneSequence := dsa.generateSceneSequence(visuallyEnhancedScenes, contextFlow)

	// 8. 视觉一致性分析
	visualConsistency := dsa.analyzeVisualConsistency(visuallyEnhancedScenes)

	// 9. 质量评估
	qualityMetrics := dsa.calculateQualityMetrics(visuallyEnhancedScenes, coherenceReport)

	// 10. 构建结果
	result := &DeepSceneAnalysisResult{
		ProjectID:         req.ProjectID,
		Scenes:           visuallyEnhancedScenes,
		SceneSequence:    sceneSequence,
		ContextFlow:      contextFlow,
		VisualConsistency: visualConsistency,
		QualityMetrics:   qualityMetrics,
		ProcessingTime:   time.Since(startTime),
	}

	// 11. 缓存结果
	_ = dsa.cacheManager.Set(ctx, cacheKey, result, 2*time.Hour)

	log.Printf("[DeepSceneAnalyzer] 深度场景分析完成: %d个场景, 质量=%.2f, 耗时=%v", 
		len(result.Scenes), result.QualityMetrics.OverallScore, result.ProcessingTime)

	return result, nil
}

// extractBasicScenes 提取基础场景
func (dsa *DeepSceneAnalyzer) extractBasicScenes(ctx context.Context, novelText string) ([]*entity.Scene, error) {
	log.Printf("[DeepSceneAnalyzer] 提取基础场景")

	// 使用AI服务提取场景
	prompt := fmt.Sprintf(`
请分析以下小说文本，提取出所有场景。对于每个场景，请提供：
1. 场景描述
2. 地点
3. 时间
4. 主要角色
5. 主要动作

小说文本：
%s

请以JSON格式返回场景列表。`, novelText)

	response, err := dsa.aiService.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI场景提取失败: %w", err)
	}

	// 解析AI响应
	scenes, err := dsa.parseSceneResponse(response)
	if err != nil {
		return nil, fmt.Errorf("解析场景响应失败: %w", err)
	}

	return scenes, nil
}

// enhanceWithSemanticAnalysis 语义分析增强
func (dsa *DeepSceneAnalyzer) enhanceWithSemanticAnalysis(ctx context.Context, scenes []*entity.Scene, req *DeepSceneAnalysisRequest) ([]*EnhancedScene, error) {
	log.Printf("[DeepSceneAnalyzer] 语义分析增强")

	enhancedScenes := make([]*EnhancedScene, 0, len(scenes))

	for i, scene := range scenes {
		// 语义分析
		semanticContext, err := dsa.semanticAnalyzer.AnalyzeSemantics(ctx, scene.Description)
		if err != nil {
			log.Printf("[DeepSceneAnalyzer] 场景%d语义分析失败: %v", i, err)
			semanticContext = &SemanticContext{} // 使用空上下文
		}

		// 情感分析
		emotionalTone, err := dsa.analyzeEmotionalTone(ctx, scene.Description)
		if err != nil {
			log.Printf("[DeepSceneAnalyzer] 场景%d情感分析失败: %v", i, err)
			emotionalTone = &EmotionalTone{PrimaryEmotion: "neutral", Intensity: 0.5}
		}

		enhancedScene := &EnhancedScene{
			Scene:           scene,
			SemanticContext: semanticContext,
			EmotionalTone:   emotionalTone,
		}

		enhancedScenes = append(enhancedScenes, enhancedScene)
	}

	return enhancedScenes, nil
}

// enhanceWithVisualAnalysis 视觉分析增强
func (dsa *DeepSceneAnalyzer) enhanceWithVisualAnalysis(ctx context.Context, scenes []*EnhancedScene, req *DeepSceneAnalysisRequest) ([]*EnhancedScene, error) {
	log.Printf("[DeepSceneAnalyzer] 视觉分析增强")

	for i, scene := range scenes {
		// 视觉元素分析
		visualElements, err := dsa.visualAnalyzer.AnalyzeVisualElements(ctx, scene.Description)
		if err != nil {
			log.Printf("[DeepSceneAnalyzer] 场景%d视觉分析失败: %v", i, err)
			visualElements = &VisualElements{} // 使用默认视觉元素
		}

		// 电影风格分析
		cinematicStyle, err := dsa.analyzeCinematicStyle(ctx, scene.Description, scene.EmotionalTone)
		if err != nil {
			log.Printf("[DeepSceneAnalyzer] 场景%d电影风格分析失败: %v", i, err)
			cinematicStyle = &CinematicStyle{} // 使用默认风格
		}

		// 过渡提示生成
		transitionHints := dsa.generateTransitionHints(scene, i, len(scenes))

		// 生成提示词
		generationPrompt := dsa.buildGenerationPrompt(scene, visualElements, cinematicStyle)

		// 质量评分
		qualityScore := dsa.calculateSceneQuality(scene, visualElements, cinematicStyle)

		// 更新场景
		scene.VisualElements = visualElements
		scene.CinematicStyle = cinematicStyle
		scene.TransitionHints = transitionHints
		scene.GenerationPrompt = generationPrompt
		scene.QualityScore = qualityScore
	}

	return scenes, nil
}

// 辅助方法
func (dsa *DeepSceneAnalyzer) parseSceneResponse(response string) ([]*entity.Scene, error) {
	// 简化的JSON解析，实际应该更健壮
	scenes := make([]*entity.Scene, 0)
	
	// 尝试提取JSON部分
	jsonStart := strings.Index(response, "[")
	jsonEnd := strings.LastIndex(response, "]")
	
	if jsonStart == -1 || jsonEnd == -1 {
		// 如果没有找到JSON，使用文本解析
		return dsa.parseSceneText(response), nil
	}
	
	jsonStr := response[jsonStart : jsonEnd+1]
	
	var rawScenes []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawScenes); err != nil {
		// JSON解析失败，使用文本解析
		return dsa.parseSceneText(response), nil
	}
	
	// 转换为Scene对象
	for i, rawScene := range rawScenes {
		scene := &entity.Scene{
			ID:          uuid.New(),
			Description: getString(rawScene, "description"),
			Location:    getString(rawScene, "location"),
			TimeOfDay:   getString(rawScene, "time"),
			Characters:  getStringSlice(rawScene, "characters"),
			Actions:     getStringSlice(rawScene, "actions"),
			Order:       i + 1,
		}
		scenes = append(scenes, scene)
	}
	
	return scenes, nil
}

func (dsa *DeepSceneAnalyzer) parseSceneText(text string) []*entity.Scene {
	// 简化的文本解析
	scenes := make([]*entity.Scene, 0)
	
	// 按段落分割
	paragraphs := strings.Split(text, "\n\n")
	
	for i, paragraph := range paragraphs {
		if len(strings.TrimSpace(paragraph)) > 50 { // 过滤太短的段落
			scene := &entity.Scene{
				ID:          uuid.New(),
				Description: strings.TrimSpace(paragraph),
				Location:    "未指定",
				TimeOfDay:   "未指定",
				Order:       i + 1,
			}
			scenes = append(scenes, scene)
		}
	}
	
	return scenes
}

func (dsa *DeepSceneAnalyzer) analyzeEmotionalTone(ctx context.Context, description string) (*EmotionalTone, error) {
	// 简化的情感分析
	tone := &EmotionalTone{
		PrimaryEmotion: "neutral",
		Intensity:      0.5,
		Mood:          "平静",
		Atmosphere:    "中性",
		TensionLevel:  0.3,
	}
	
	// 基于关键词的简单情感识别
	lowerDesc := strings.ToLower(description)
	
	if strings.Contains(lowerDesc, "愤怒") || strings.Contains(lowerDesc, "生气") {
		tone.PrimaryEmotion = "angry"
		tone.Intensity = 0.8
		tone.Mood = "愤怒"
		tone.TensionLevel = 0.9
	} else if strings.Contains(lowerDesc, "高兴") || strings.Contains(lowerDesc, "快乐") {
		tone.PrimaryEmotion = "happy"
		tone.Intensity = 0.7
		tone.Mood = "愉快"
		tone.TensionLevel = 0.2
	} else if strings.Contains(lowerDesc, "悲伤") || strings.Contains(lowerDesc, "难过") {
		tone.PrimaryEmotion = "sad"
		tone.Intensity = 0.6
		tone.Mood = "忧郁"
		tone.TensionLevel = 0.4
	}
	
	return tone, nil
}

func (dsa *DeepSceneAnalyzer) analyzeCinematicStyle(ctx context.Context, description string, emotion *EmotionalTone) (*CinematicStyle, error) {
	style := &CinematicStyle{
		ShotType:      "medium_shot",
		MovementStyle: "static",
		EditingRhythm: "normal",
	}
	
	// 根据情感调整电影风格
	switch emotion.PrimaryEmotion {
	case "angry":
		style.ShotType = "close_up"
		style.MovementStyle = "dynamic"
		style.EditingRhythm = "fast"
	case "sad":
		style.ShotType = "wide_shot"
		style.MovementStyle = "slow"
		style.EditingRhythm = "slow"
	case "happy":
		style.ShotType = "medium_shot"
		style.MovementStyle = "smooth"
		style.EditingRhythm = "upbeat"
	}
	
	return style, nil
}

func (dsa *DeepSceneAnalyzer) generateTransitionHints(scene *EnhancedScene, index, total int) *TransitionHints {
	hints := &TransitionHints{
		TransitionType:  "fade",
		TransitionSpeed: "normal",
	}
	
	// 根据场景位置和情感调整过渡
	if index == 0 {
		hints.TransitionType = "fade_in"
	} else if index == total-1 {
		hints.TransitionType = "fade_out"
	} else {
		// 根据情感强度选择过渡类型
		if scene.EmotionalTone.Intensity > 0.7 {
			hints.TransitionType = "cut"
			hints.TransitionSpeed = "fast"
		} else {
			hints.TransitionType = "dissolve"
			hints.TransitionSpeed = "slow"
		}
	}
	
	return hints
}

func (dsa *DeepSceneAnalyzer) buildGenerationPrompt(scene *EnhancedScene, visual *VisualElements, cinematic *CinematicStyle) string {
	prompt := scene.Description
	
	// 添加视觉元素
	if visual.ArtisticStyle != "" {
		prompt += ", " + visual.ArtisticStyle
	}
	if visual.LightingStyle != "" {
		prompt += ", " + visual.LightingStyle
	}
	
	// 添加电影风格
	if cinematic.ShotType != "" {
		prompt += ", " + cinematic.ShotType
	}
	
	// 添加质量修饰符
	prompt += ", high quality, detailed, masterpiece, anime style"
	
	return prompt
}

func (dsa *DeepSceneAnalyzer) calculateSceneQuality(scene *EnhancedScene, visual *VisualElements, cinematic *CinematicStyle) float64 {
	score := 0.7 // 基础分数
	
	// 描述长度加分
	if len(scene.Description) > 100 {
		score += 0.1
	}
	
	// 位置信息加分
	if scene.Location != "" && scene.Location != "未指定" {
		score += 0.1
	}
	
	// 情感强度加分
	if scene.EmotionalTone.Intensity > 0.5 {
		score += 0.1
	}
	
	return score
}

// 辅助函数
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key]; ok {
		if slice, ok := v.([]interface{}); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return []string{}
}
