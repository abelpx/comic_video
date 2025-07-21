package ai

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// EnhancedCharacterProfile 增强版角色档案，确保角色一致性
type EnhancedCharacterProfile struct {
	// 基础信息
	Name         string `json:"name"`
	Description  string `json:"description"`
	Seed         int64  `json:"seed"`
	
	// 详细视觉特征
	FacialFeatures string `json:"facial_features"` // 面部特征：眼睛、鼻子、嘴巴等
	HairStyle      string `json:"hair_style"`      // 发型：长短、颜色、样式
	Clothing       string `json:"clothing"`        // 服装：风格、颜色、款式
	BodyType       string `json:"body_type"`       // 体型：高矮、胖瘦
	Age            string `json:"age"`             // 年龄段：青年、中年等
	Gender         string `json:"gender"`          // 性别
	
	// 风格控制
	ArtStyle       string   `json:"art_style"`       // 艺术风格：写实、动漫、卡通等
	VisualKeywords []string `json:"visual_keywords"` // 关键视觉词汇
	ColorScheme    string   `json:"color_scheme"`    // 角色专属色彩方案
	
	// 一致性控制参数
	ConsistencyPrompt string `json:"consistency_prompt"` // 一致性提示词
	ReferenceStyle    string `json:"reference_style"`    // 参考风格
}

// EnhancedSceneContext 增强版场景上下文，确保场景连贯性
type EnhancedSceneContext struct {
	// 基础场景信息
	Location     string `json:"location"`      // 具体地点
	TimeOfDay    string `json:"time_of_day"`   // 具体时间
	Weather      string `json:"weather"`       // 天气状况
	Season       string `json:"season"`        // 季节
	
	// 视觉风格
	ArtStyle     string `json:"art_style"`     // 整体艺术风格
	ColorPalette string `json:"color_palette"` // 主色调
	Lighting     string `json:"lighting"`      // 光照效果
	Atmosphere   string `json:"atmosphere"`    // 氛围感
	
	// 构图和视角
	CameraAngle  string `json:"camera_angle"`  // 镜头角度
	Composition  string `json:"composition"`   // 构图方式
	Perspective  string `json:"perspective"`   // 透视效果
	
	// 细节控制
	Background   string `json:"background"`    // 背景细节
	Foreground   string `json:"foreground"`    // 前景元素
	Props        string `json:"props"`         // 道具物品
	
	// 一致性种子
	SceneSeed    int64  `json:"scene_seed"`    // 场景一致性种子
}

// StoryConsistencyManager 故事一致性管理器
type StoryConsistencyManager struct {
	Characters    []EnhancedCharacterProfile `json:"characters"`
	SceneContext  EnhancedSceneContext       `json:"scene_context"`
	StoryStyle    string                     `json:"story_style"`    // 整体故事风格
	VisualTheme   string                     `json:"visual_theme"`   // 视觉主题
	ColorHarmony  string                     `json:"color_harmony"`  // 色彩和谐方案
	StoryMood     string                     `json:"story_mood"`     // 故事整体情绪
}

// GenerateCharacterSeed 为角色生成一致性种子
func GenerateCharacterSeed(name string) int64 {
	hash := md5.Sum([]byte(name))
	return int64(hash[0])<<24 | int64(hash[1])<<16 | int64(hash[2])<<8 | int64(hash[3])
}

// GenerateSceneSeed 为场景生成一致性种子
func GenerateSceneSeed(location, timeOfDay string) int64 {
	combined := location + "_" + timeOfDay
	hash := md5.Sum([]byte(combined))
	return int64(hash[0])<<24 | int64(hash[1])<<16 | int64(hash[2])<<8 | int64(hash[3])
}

// BuildCharacterPrompt 构建角色一致性提示词
func (c *EnhancedCharacterProfile) BuildCharacterPrompt() string {
	var parts []string
	
	// 基础描述
	if c.Name != "" {
		parts = append(parts, fmt.Sprintf("character named %s", c.Name))
	}
	
	// 年龄和性别
	if c.Age != "" && c.Gender != "" {
		parts = append(parts, fmt.Sprintf("%s %s", c.Age, c.Gender))
	}
	
	// 面部特征
	if c.FacialFeatures != "" {
		parts = append(parts, c.FacialFeatures)
	}
	
	// 发型
	if c.HairStyle != "" {
		parts = append(parts, c.HairStyle)
	}
	
	// 服装
	if c.Clothing != "" {
		parts = append(parts, c.Clothing)
	}
	
	// 体型
	if c.BodyType != "" {
		parts = append(parts, c.BodyType)
	}
	
	// 艺术风格
	if c.ArtStyle != "" {
		parts = append(parts, fmt.Sprintf("%s style", c.ArtStyle))
	}
	
	// 关键词
	if len(c.VisualKeywords) > 0 {
		parts = append(parts, strings.Join(c.VisualKeywords, ", "))
	}
	
	// 色彩方案
	if c.ColorScheme != "" {
		parts = append(parts, c.ColorScheme)
	}
	
	prompt := strings.Join(parts, ", ")
	
	// 添加一致性控制
	if c.Seed != 0 {
		prompt += fmt.Sprintf(", consistent character design, seed:%d", c.Seed)
	}
	
	return prompt
}

// BuildScenePrompt 构建场景一致性提示词
func (s *EnhancedSceneContext) BuildScenePrompt() string {
	var parts []string
	
	// 地点和时间
	if s.Location != "" {
		parts = append(parts, s.Location)
	}
	if s.TimeOfDay != "" {
		parts = append(parts, s.TimeOfDay)
	}
	
	// 天气和季节
	if s.Weather != "" {
		parts = append(parts, s.Weather)
	}
	if s.Season != "" {
		parts = append(parts, s.Season)
	}
	
	// 艺术风格
	if s.ArtStyle != "" {
		parts = append(parts, fmt.Sprintf("%s art style", s.ArtStyle))
	}
	
	// 色彩和光照
	if s.ColorPalette != "" {
		parts = append(parts, s.ColorPalette)
	}
	if s.Lighting != "" {
		parts = append(parts, s.Lighting)
	}
	
	// 氛围
	if s.Atmosphere != "" {
		parts = append(parts, s.Atmosphere)
	}
	
	// 镜头和构图
	if s.CameraAngle != "" {
		parts = append(parts, s.CameraAngle)
	}
	if s.Composition != "" {
		parts = append(parts, s.Composition)
	}
	
	// 背景和前景
	if s.Background != "" {
		parts = append(parts, fmt.Sprintf("background: %s", s.Background))
	}
	if s.Foreground != "" {
		parts = append(parts, fmt.Sprintf("foreground: %s", s.Foreground))
	}
	
	// 道具
	if s.Props != "" {
		parts = append(parts, fmt.Sprintf("props: %s", s.Props))
	}
	
	prompt := strings.Join(parts, ", ")
	
	// 添加一致性控制
	if s.SceneSeed != 0 {
		prompt += fmt.Sprintf(", consistent scene design, seed:%d", s.SceneSeed)
	}
	
	return prompt
}

// BuildConsistentPrompt 构建完整的一致性提示词
func (m *StoryConsistencyManager) BuildConsistentPrompt(panelDescription string, characterNames []string) string {
	var promptParts []string
	
	// 基础分镜描述
	promptParts = append(promptParts, panelDescription)
	
	// 角色一致性
	for _, char := range m.Characters {
		if strings.Contains(strings.ToLower(panelDescription), strings.ToLower(char.Name)) {
			charPrompt := char.BuildCharacterPrompt()
			if charPrompt != "" {
				promptParts = append(promptParts, fmt.Sprintf("(%s)", charPrompt))
			}
		}
	}
	
	// 场景一致性
	scenePrompt := m.SceneContext.BuildScenePrompt()
	if scenePrompt != "" {
		promptParts = append(promptParts, fmt.Sprintf("scene: (%s)", scenePrompt))
	}
	
	// 整体风格
	if m.StoryStyle != "" {
		promptParts = append(promptParts, fmt.Sprintf("overall style: %s", m.StoryStyle))
	}
	
	// 视觉主题
	if m.VisualTheme != "" {
		promptParts = append(promptParts, fmt.Sprintf("visual theme: %s", m.VisualTheme))
	}
	
	// 色彩和谐
	if m.ColorHarmony != "" {
		promptParts = append(promptParts, fmt.Sprintf("color harmony: %s", m.ColorHarmony))
	}
	
	// 故事情绪
	if m.StoryMood != "" {
		promptParts = append(promptParts, fmt.Sprintf("mood: %s", m.StoryMood))
	}
	
	// 质量控制
	promptParts = append(promptParts, "high quality, detailed, consistent art style, professional illustration")
	
	return strings.Join(promptParts, ", ")
}

// AnalyzeStoryStyle 分析故事整体风格
func AnalyzeStoryStyle(novel string) string {
	// 简化的风格分析逻辑
	novel = strings.ToLower(novel)
	
	if strings.Contains(novel, "科幻") || strings.Contains(novel, "未来") || strings.Contains(novel, "机器人") {
		return "sci-fi cyberpunk"
	}
	if strings.Contains(novel, "古代") || strings.Contains(novel, "武侠") || strings.Contains(novel, "仙侠") {
		return "traditional chinese fantasy"
	}
	if strings.Contains(novel, "现代") || strings.Contains(novel, "都市") {
		return "modern realistic"
	}
	if strings.Contains(novel, "魔法") || strings.Contains(novel, "奇幻") {
		return "fantasy magical"
	}
	if strings.Contains(novel, "恐怖") || strings.Contains(novel, "惊悚") {
		return "dark horror"
	}
	if strings.Contains(novel, "浪漫") || strings.Contains(novel, "爱情") {
		return "romantic soft"
	}
	
	return "anime manga style"
}

// GenerateColorHarmony 生成色彩和谐方案
func GenerateColorHarmony(storyStyle, mood string) string {
	switch {
	case strings.Contains(storyStyle, "sci-fi"):
		return "cool blue and cyan tones with neon accents"
	case strings.Contains(storyStyle, "fantasy"):
		return "warm magical colors with purple and gold highlights"
	case strings.Contains(storyStyle, "traditional"):
		return "classical chinese ink colors with red and gold"
	case strings.Contains(storyStyle, "horror"):
		return "dark monochrome with red accents"
	case strings.Contains(storyStyle, "romantic"):
		return "soft pastel colors with pink and warm tones"
	case strings.Contains(mood, "happy"):
		return "bright vibrant colors"
	case strings.Contains(mood, "sad"):
		return "muted blue and gray tones"
	case strings.Contains(mood, "tense"):
		return "high contrast with dramatic shadows"
	default:
		return "balanced natural color palette"
	}
}
