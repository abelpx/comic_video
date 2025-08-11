package character

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"comic_video/internal/domain/entity"
	"comic_video/internal/utils"
	"github.com/google/uuid"
)

// ConsistencyManager 角色一致性管理器
type ConsistencyManager struct {
	aiService       AIService
	cacheManager    *utils.CacheManager
	consistencyDB   map[string]*CharacterConsistencyProfile
	visualValidator *VisualConsistencyValidator
}

// NewConsistencyManager 创建角色一致性管理器
func NewConsistencyManager(aiService AIService, cacheManager *utils.CacheManager) *ConsistencyManager {
	return &ConsistencyManager{
		aiService:       aiService,
		cacheManager:    cacheManager,
		consistencyDB:   make(map[string]*CharacterConsistencyProfile),
		visualValidator: NewVisualConsistencyValidator(),
	}
}

// CharacterConsistencyProfile 角色一致性档案
type CharacterConsistencyProfile struct {
	CharacterID       uuid.UUID                    `json:"character_id"`
	Name              string                       `json:"name"`
	BasePrompt        string                       `json:"base_prompt"`        // 基础提示词
	VisualKeywords    []string                     `json:"visual_keywords"`    // 视觉关键词
	StyleModifiers    []string                     `json:"style_modifiers"`    // 风格修饰符
	ConsistencySeed   int64                        `json:"consistency_seed"`   // 一致性种子
	ReferenceImages   []string                     `json:"reference_images"`   // 参考图像
	VariationRules    *CharacterVariationRules     `json:"variation_rules"`    // 变化规则
	QualityMetrics    *ConsistencyQualityMetrics   `json:"quality_metrics"`    // 质量指标
	GenerationHistory []*GenerationRecord          `json:"generation_history"` // 生成历史
	LastUpdated       time.Time                    `json:"last_updated"`
}

// CharacterVariationRules 角色变化规则
type CharacterVariationRules struct {
	AllowedExpressions []string  `json:"allowed_expressions"` // 允许的表情
	AllowedPoses       []string  `json:"allowed_poses"`       // 允许的姿势
	AllowedAngles      []string  `json:"allowed_angles"`      // 允许的角度
	ClothingVariations []string  `json:"clothing_variations"` // 服装变化
	EmotionMapping     map[string]string `json:"emotion_mapping"` // 情感映射
	SceneAdaptations   map[string]string `json:"scene_adaptations"` // 场景适配
}

// ConsistencyQualityMetrics 一致性质量指标
type ConsistencyQualityMetrics struct {
	VisualSimilarity    float64 `json:"visual_similarity"`    // 视觉相似度
	FeatureConsistency  float64 `json:"feature_consistency"`  // 特征一致性
	StyleConsistency    float64 `json:"style_consistency"`    // 风格一致性
	OverallConsistency  float64 `json:"overall_consistency"`  // 整体一致性
	GenerationSuccess   float64 `json:"generation_success"`   // 生成成功率
	UserSatisfaction    float64 `json:"user_satisfaction"`    // 用户满意度
}

// GenerationRecord 生成记录
type GenerationRecord struct {
	Timestamp     time.Time `json:"timestamp"`
	Prompt        string    `json:"prompt"`
	ImagePath     string    `json:"image_path"`
	QualityScore  float64   `json:"quality_score"`
	ConsistencyScore float64 `json:"consistency_score"`
	Context       string    `json:"context"` // 生成上下文
}

// EnsureCharacterConsistency 确保角色一致性
func (cm *ConsistencyManager) EnsureCharacterConsistency(ctx context.Context, character *entity.Character, scene *entity.Scene, emotion string) (*CharacterGenerationRequest, error) {
	log.Printf("[ConsistencyManager] 确保角色一致性: %s", character.Name)

	// 1. 获取或创建一致性档案
	profile, err := cm.getOrCreateConsistencyProfile(ctx, character)
	if err != nil {
		return nil, fmt.Errorf("获取一致性档案失败: %w", err)
	}

	// 2. 根据场景和情感调整提示词
	adaptedPrompt, err := cm.adaptPromptForContext(ctx, profile, scene, emotion)
	if err != nil {
		return nil, fmt.Errorf("适配提示词失败: %w", err)
	}

	// 3. 生成一致性参数
	consistencyParams := cm.generateConsistencyParameters(profile, scene, emotion)

	// 4. 创建生成请求
	request := &CharacterGenerationRequest{
		Character:         character,
		Scene:            scene,
		Emotion:          emotion,
		BasePrompt:       adaptedPrompt,
		ConsistencyParams: consistencyParams,
		QualityThreshold: 0.8, // 质量阈值
		MaxRetries:       3,
	}

	return request, nil
}

// getOrCreateConsistencyProfile 获取或创建一致性档案
func (cm *ConsistencyManager) getOrCreateConsistencyProfile(ctx context.Context, character *entity.Character) (*CharacterConsistencyProfile, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("character_consistency:%s", character.ID.String())
	var profile CharacterConsistencyProfile
	
	err := cm.cacheManager.Get(ctx, cacheKey, &profile)
	if err == nil {
		return &profile, nil
	}

	// 创建新的一致性档案
	profile = CharacterConsistencyProfile{
		CharacterID:       character.ID,
		Name:              character.Name,
		BasePrompt:        cm.buildBasePrompt(character),
		VisualKeywords:    cm.extractVisualKeywords(character),
		StyleModifiers:    cm.generateStyleModifiers(character),
		ConsistencySeed:   cm.generateConsistencySeed(character.Name),
		ReferenceImages:   make([]string, 0),
		VariationRules:    cm.createVariationRules(character),
		QualityMetrics:    &ConsistencyQualityMetrics{},
		GenerationHistory: make([]*GenerationRecord, 0),
		LastUpdated:       time.Now(),
	}

	// 缓存档案
	err = cm.cacheManager.Set(ctx, cacheKey, &profile, 24*time.Hour)
	if err != nil {
		log.Printf("[ConsistencyManager] 缓存档案失败: %v", err)
	}

	cm.consistencyDB[character.ID.String()] = &profile
	return &profile, nil
}

// buildBasePrompt 构建基础提示词
func (cm *ConsistencyManager) buildBasePrompt(character *entity.Character) string {
	var promptParts []string

	// 基础描述
	if character.Description != "" {
		promptParts = append(promptParts, character.Description)
	}

	// 外观描述
	appearanceDesc := ""
	if character.FacialFeatures != "" {
		appearanceDesc += character.FacialFeatures
	}
	if character.HairStyle != "" {
		if appearanceDesc != "" {
			appearanceDesc += ", "
		}
		appearanceDesc += character.HairStyle
	}
	if character.Clothing != "" {
		if appearanceDesc != "" {
			appearanceDesc += ", "
		}
		appearanceDesc += character.Clothing
	}
	if appearanceDesc != "" {
		promptParts = append(promptParts, appearanceDesc)
	}

	// 添加质量修饰符
	qualityModifiers := []string{
		"high quality",
		"detailed",
		"professional artwork",
		"consistent character design",
		"anime style",
		"masterpiece",
	}
	promptParts = append(promptParts, strings.Join(qualityModifiers, ", "))

	return strings.Join(promptParts, ", ")
}

// extractVisualKeywords 提取视觉关键词
func (cm *ConsistencyManager) extractVisualKeywords(character *entity.Character) []string {
	keywords := make([]string, 0)

	// 从描述中提取关键词
	if character.Description != "" {
		keywords = append(keywords, cm.extractKeywordsFromText(character.Description)...)
	}

	// 从外观中提取关键词
	appearanceText := character.FacialFeatures + " " + character.HairStyle + " " + character.Clothing
	if appearanceText != "  " { // 检查不是空字符串
		keywords = append(keywords, cm.extractKeywordsFromText(appearanceText)...)
	}

	// 添加默认关键词
	defaultKeywords := []string{
		"consistent",
		"detailed",
		"high quality",
		"anime style",
	}
	keywords = append(keywords, defaultKeywords...)

	return cm.deduplicateKeywords(keywords)
}

// generateStyleModifiers 生成风格修饰符
func (cm *ConsistencyManager) generateStyleModifiers(character *entity.Character) []string {
	modifiers := []string{
		"consistent lighting",
		"consistent art style",
		"same character",
		"character reference sheet",
		"model sheet",
		"turnaround",
	}

	// 根据角色类型添加特定修饰符
	if strings.Contains(strings.ToLower(character.Description), "主角") {
		modifiers = append(modifiers, "protagonist", "main character", "hero")
	}

	return modifiers
}

// generateConsistencySeed 生成一致性种子
func (cm *ConsistencyManager) generateConsistencySeed(characterName string) int64 {
	hash := md5.Sum([]byte(characterName + "consistency"))
	seed := int64(hash[0])<<24 | int64(hash[1])<<16 | int64(hash[2])<<8 | int64(hash[3])
	return seed
}

// createVariationRules 创建变化规则
func (cm *ConsistencyManager) createVariationRules(character *entity.Character) *CharacterVariationRules {
	return &CharacterVariationRules{
		AllowedExpressions: []string{
			"neutral", "happy", "sad", "angry", "surprised", "worried", "determined",
		},
		AllowedPoses: []string{
			"standing", "sitting", "walking", "running", "pointing", "thinking",
		},
		AllowedAngles: []string{
			"front view", "side view", "three-quarter view", "back view",
		},
		ClothingVariations: []string{
			"same outfit", "casual clothes", "formal wear",
		},
		EmotionMapping: map[string]string{
			"happy":     "smiling, bright eyes, cheerful expression",
			"sad":       "downcast eyes, frowning, melancholic",
			"angry":     "furrowed brow, intense gaze, clenched jaw",
			"surprised": "wide eyes, open mouth, raised eyebrows",
		},
		SceneAdaptations: map[string]string{
			"indoor":  "indoor lighting, interior background",
			"outdoor": "natural lighting, outdoor background",
			"night":   "dim lighting, night scene",
			"day":     "bright lighting, daylight scene",
		},
	}
}

// adaptPromptForContext 根据上下文适配提示词
func (cm *ConsistencyManager) adaptPromptForContext(ctx context.Context, profile *CharacterConsistencyProfile, scene *entity.Scene, emotion string) (string, error) {
	basePrompt := profile.BasePrompt
	
	// 添加情感修饰
	if emotionModifier, exists := profile.VariationRules.EmotionMapping[emotion]; exists {
		basePrompt += ", " + emotionModifier
	}

	// 添加场景适配
	if scene != nil {
		if scene.Location != "" {
			if sceneModifier, exists := profile.VariationRules.SceneAdaptations[strings.ToLower(scene.Location)]; exists {
				basePrompt += ", " + sceneModifier
			}
		}
	}

	// 添加一致性修饰符
	basePrompt += ", " + strings.Join(profile.StyleModifiers, ", ")

	return basePrompt, nil
}

// generateConsistencyParameters 生成一致性参数
func (cm *ConsistencyManager) generateConsistencyParameters(profile *CharacterConsistencyProfile, scene *entity.Scene, emotion string) map[string]interface{} {
	params := map[string]interface{}{
		"seed":         profile.ConsistencySeed,
		"cfg_scale":    7.5,
		"steps":        20,
		"width":        512,
		"height":       768,
		"sampler_name": "DPM++ 2M Karras",
	}

	// 根据场景调整参数
	if scene != nil {
		if strings.Contains(strings.ToLower(scene.Description), "夜晚") {
			params["cfg_scale"] = 8.0 // 夜晚场景需要更强的引导
		}
	}

	return params
}

// ValidateConsistency 验证一致性
func (cm *ConsistencyManager) ValidateConsistency(ctx context.Context, character *entity.Character, generatedImagePath string, context string) (*ConsistencyValidationResult, error) {
	log.Printf("[ConsistencyManager] 验证角色一致性: %s", character.Name)

	profile, err := cm.getOrCreateConsistencyProfile(ctx, character)
	if err != nil {
		return nil, fmt.Errorf("获取一致性档案失败: %w", err)
	}

	// 1. 视觉一致性验证
	visualScore, err := cm.visualValidator.ValidateVisualConsistency(generatedImagePath, profile.ReferenceImages)
	if err != nil {
		log.Printf("[ConsistencyManager] 视觉一致性验证失败: %v", err)
		visualScore = 0.5 // 默认分数
	}

	// 2. 特征一致性验证
	featureScore := cm.validateFeatureConsistency(generatedImagePath, profile)

	// 3. 风格一致性验证
	styleScore := cm.validateStyleConsistency(generatedImagePath, profile)

	// 4. 计算综合分数
	overallScore := (visualScore + featureScore + styleScore) / 3.0

	// 5. 更新质量指标
	cm.updateQualityMetrics(profile, overallScore, visualScore, featureScore, styleScore)

	// 6. 记录生成历史
	record := &GenerationRecord{
		Timestamp:        time.Now(),
		ImagePath:        generatedImagePath,
		QualityScore:     overallScore,
		ConsistencyScore: visualScore,
		Context:          context,
	}
	profile.GenerationHistory = append(profile.GenerationHistory, record)

	// 7. 如果是高质量图像，添加为参考图像
	if overallScore > 0.8 {
		profile.ReferenceImages = append(profile.ReferenceImages, generatedImagePath)
		// 限制参考图像数量
		if len(profile.ReferenceImages) > 10 {
			profile.ReferenceImages = profile.ReferenceImages[1:]
		}
	}

	// 8. 更新缓存
	cacheKey := fmt.Sprintf("character_consistency:%s", character.ID.String())
	_ = cm.cacheManager.Set(ctx, cacheKey, profile, 24*time.Hour)

	result := &ConsistencyValidationResult{
		OverallScore:     overallScore,
		VisualScore:      visualScore,
		FeatureScore:     featureScore,
		StyleScore:       styleScore,
		IsConsistent:     overallScore > 0.7,
		Recommendations:  cm.generateRecommendations(overallScore, visualScore, featureScore, styleScore),
	}

	return result, nil
}

// 辅助方法
func (cm *ConsistencyManager) extractKeywordsFromText(text string) []string {
	// 简化实现，实际应该使用NLP工具
	words := strings.Fields(strings.ToLower(text))
	keywords := make([]string, 0)
	
	// 过滤常见词汇，保留描述性词汇
	stopWords := map[string]bool{
		"的": true, "是": true, "在": true, "有": true, "和": true,
		"一个": true, "这个": true, "那个": true,
	}
	
	for _, word := range words {
		if !stopWords[word] && len(word) > 1 {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

func (cm *ConsistencyManager) deduplicateKeywords(keywords []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	
	for _, keyword := range keywords {
		if !seen[keyword] {
			seen[keyword] = true
			result = append(result, keyword)
		}
	}
	
	return result
}

func (cm *ConsistencyManager) validateFeatureConsistency(imagePath string, profile *CharacterConsistencyProfile) float64 {
	// 简化实现，实际应该使用计算机视觉技术
	return 0.75 + rand.Float64()*0.2 // 0.75-0.95之间的随机分数
}

func (cm *ConsistencyManager) validateStyleConsistency(imagePath string, profile *CharacterConsistencyProfile) float64 {
	// 简化实现，实际应该分析图像风格
	return 0.8 + rand.Float64()*0.15 // 0.8-0.95之间的随机分数
}

func (cm *ConsistencyManager) updateQualityMetrics(profile *CharacterConsistencyProfile, overall, visual, feature, style float64) {
	metrics := profile.QualityMetrics
	
	// 使用移动平均更新指标
	alpha := 0.1 // 学习率
	metrics.OverallConsistency = metrics.OverallConsistency*(1-alpha) + overall*alpha
	metrics.VisualSimilarity = metrics.VisualSimilarity*(1-alpha) + visual*alpha
	metrics.FeatureConsistency = metrics.FeatureConsistency*(1-alpha) + feature*alpha
	metrics.StyleConsistency = metrics.StyleConsistency*(1-alpha) + style*alpha
}

func (cm *ConsistencyManager) generateRecommendations(overall, visual, feature, style float64) []string {
	recommendations := make([]string, 0)
	
	if visual < 0.7 {
		recommendations = append(recommendations, "建议增加更多参考图像以提高视觉一致性")
	}
	if feature < 0.7 {
		recommendations = append(recommendations, "建议优化角色特征描述")
	}
	if style < 0.7 {
		recommendations = append(recommendations, "建议统一艺术风格设定")
	}
	if overall < 0.6 {
		recommendations = append(recommendations, "建议重新生成角色图像")
	}
	
	return recommendations
}

// CharacterGenerationRequest 角色生成请求
type CharacterGenerationRequest struct {
	Character         *entity.Character      `json:"character"`
	Scene            *entity.Scene          `json:"scene"`
	Emotion          string                 `json:"emotion"`
	BasePrompt       string                 `json:"base_prompt"`
	ConsistencyParams map[string]interface{} `json:"consistency_params"`
	QualityThreshold float64                `json:"quality_threshold"`
	MaxRetries       int                    `json:"max_retries"`
}

// ConsistencyValidationResult 一致性验证结果
type ConsistencyValidationResult struct {
	OverallScore    float64  `json:"overall_score"`
	VisualScore     float64  `json:"visual_score"`
	FeatureScore    float64  `json:"feature_score"`
	StyleScore      float64  `json:"style_score"`
	IsConsistent    bool     `json:"is_consistent"`
	Recommendations []string `json:"recommendations"`
}
