package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
)

// EnhancedSceneAnalyzer 增强版场景分析器
type EnhancedSceneAnalyzer struct {
	ollama *OllamaClient
}

// NewEnhancedSceneAnalyzer 创建增强版场景分析器
func NewEnhancedSceneAnalyzer(ollama *OllamaClient) *EnhancedSceneAnalyzer {
	return &EnhancedSceneAnalyzer{ollama: ollama}
}

// AnalyzeDetailedScene 分析详细的场景信息
func (e *EnhancedSceneAnalyzer) AnalyzeDetailedScene(ctx context.Context, novel string, panels []string) (EnhancedSceneContext, error) {
	// 合并所有分镜内容进行整体分析
	allContent := novel + "\n\n分镜内容：\n" + strings.Join(panels, "\n")

	prompt := fmt.Sprintf(`分析以下内容，输出JSON格式的场景信息。

严格要求：
1. 必须输出JSON对象，不是数组
2. 包含所有必需字段
3. 不要添加任何解释文字

输出格式（必须完全按照此格式）：
{
  "location": "具体地点",
  "time_of_day": "具体时间",
  "weather": "天气状况",
  "season": "季节",
  "art_style": "艺术风格",
  "color_palette": "色彩方案",
  "lighting": "光照效果",
  "atmosphere": "氛围感",
  "camera_angle": "镜头角度",
  "composition": "构图方式",
  "background": "背景描述",
  "foreground": "前景描述",
  "props": "道具物品"
}

内容：
%s

JSON输出：`, allContent)

	response, err := e.ollama.Generate(prompt, nil)
	if err != nil {
		log.Printf("[AI] 场景分析失败: %v", err)
		return e.generateDefaultScene(novel), nil
	}

	// 清理和解析JSON
	cleanedResponse := cleanSceneAnalysisJSON(response)
	log.Printf("[AI] 场景分析原始响应: %s", response[:min(200, len(response))])
	log.Printf("[AI] 场景分析清理后: %s", cleanedResponse)

	// 使用灵活的解析方式处理可能的数组字段
	sceneData, err := parseFlexibleSceneJSON(cleanedResponse)
	if err != nil {
		log.Printf("[AI] 详细场景信息解析失败: %v", err)
		log.Printf("[AI] 响应内容: %s", cleanedResponse)

		// 尝试使用默认场景
		sceneData = getDefaultSceneData()
		log.Printf("[AI] 使用默认场景信息")
	}

	// 构建增强场景上下文
	scene := EnhancedSceneContext{
		Location:     e.enhanceLocation(sceneData.Location),
		TimeOfDay:    e.enhanceTimeOfDay(sceneData.TimeOfDay),
		Weather:      e.enhanceWeather(sceneData.Weather),
		Season:       e.enhanceSeason(sceneData.Season),
		ArtStyle:     e.enhanceArtStyle(sceneData.ArtStyle, novel),
		ColorPalette: e.enhanceColorPalette(sceneData.ColorPalette, novel),
		Lighting:     e.enhanceLighting(sceneData.Lighting, sceneData.TimeOfDay),
		Atmosphere:   e.enhanceAtmosphere(sceneData.Atmosphere, novel),
		CameraAngle:  e.enhanceCameraAngle(sceneData.CameraAngle),
		Composition:  e.enhanceComposition(sceneData.Composition),
		Perspective:  e.generatePerspective(sceneData.CameraAngle),
		Background:   e.enhanceBackground(sceneData.Background, sceneData.Location),
		Foreground:   e.enhanceForeground(sceneData.Foreground),
		Props:        e.enhanceProps(sceneData.Props),
		SceneSeed:    GenerateSceneSeed(sceneData.Location, sceneData.TimeOfDay),
	}

	log.Printf("[AI] 成功分析详细场景信息")
	return scene, nil
}

// buildSceneFromColors 基于色彩信息构建场景
func (e *EnhancedSceneAnalyzer) buildSceneFromColors(novel string, panels []string, colors []string) EnhancedSceneContext {
	log.Printf("[AI] 基于色彩信息构建场景: %v", colors)

	// 分析色彩特征
	var isDark, isViolent, isNight bool
	colorPalette := strings.Join(colors, ", ")

	for _, color := range colors {
		colorLower := strings.ToLower(color)
		if strings.Contains(colorLower, "深") || strings.Contains(colorLower, "黑") || strings.Contains(colorLower, "暗") {
			isDark = true
		}
		if strings.Contains(colorLower, "血") || strings.Contains(colorLower, "红") {
			isViolent = true
		}
		if strings.Contains(colorLower, "夜") || strings.Contains(colorLower, "月") {
			isNight = true
		}
	}

	// 基于色彩和内容推断场景
	var location, timeOfDay, atmosphere, lighting string

	if isDark && isViolent {
		location = "危险的城市暗巷，破败的建筑环境"
		atmosphere = "紧张危险的氛围，充满威胁感"
		lighting = "昏暗的街灯，阴影重重"
	} else if isNight {
		location = "夜晚的城市街道，灯火阑珊"
		atmosphere = "神秘宁静的夜晚氛围"
		lighting = "柔和的月光和街灯照明"
	} else {
		location = e.inferLocation(novel)
		atmosphere = e.inferAtmosphere(novel)
		lighting = "自然光照，明暗适中"
	}

	if isNight {
		timeOfDay = "深沉的夜晚，夜色浓重"
	} else {
		timeOfDay = e.inferTimeOfDay(novel)
	}

	// 构建完整场景
	scene := EnhancedSceneContext{
		Location:     location,
		TimeOfDay:    timeOfDay,
		Weather:      "适宜的天气条件",
		Season:       "当前季节",
		ArtStyle:     "anime manga style, cinematic",
		ColorPalette: colorPalette,
		Lighting:     lighting,
		Atmosphere:   atmosphere,
		CameraAngle:  "戏剧性角度，增强视觉冲击",
		Composition:  "动态构图，突出重点",
		Perspective:  "透视感强，层次分明",
		Background:   "详细的环境背景，符合故事情节",
		Foreground:   "清晰的前景元素，突出主体",
		Props:        e.inferProps(novel),
		SceneSeed:    GenerateSceneSeed(location, timeOfDay),
	}

	log.Printf("[AI] 基于色彩构建场景完成: 地点=%s 时间=%s", scene.Location, scene.TimeOfDay)
	return scene
}

// analyzeSimpleScene 简单场景分析的降级方案
func (e *EnhancedSceneAnalyzer) analyzeSimpleScene(ctx context.Context, novel string, panels []string) (EnhancedSceneContext, error) {
	// 基于小说内容进行简单分析
	storyStyle := AnalyzeStoryStyle(novel)

	scene := EnhancedSceneContext{
		Location:     e.inferLocation(novel),
		TimeOfDay:    e.inferTimeOfDay(novel),
		Weather:      "晴朗",
		Season:       "春季",
		ArtStyle:     storyStyle,
		ColorPalette: e.getStyleColorPalette(storyStyle),
		Lighting:     "自然光照",
		Atmosphere:   e.inferAtmosphere(novel),
		CameraAngle:  "平视角度",
		Composition:  "居中构图",
		Perspective:  "正常透视",
		Background:   e.getStyleBackground(storyStyle),
		Foreground:   "清晰前景",
		Props:        e.inferProps(novel),
		SceneSeed:    GenerateSceneSeed(e.inferLocation(novel), e.inferTimeOfDay(novel)),
	}

	return scene, nil
}

// generateDefaultScene 生成默认场景
func (e *EnhancedSceneAnalyzer) generateDefaultScene(novel string) EnhancedSceneContext {
	storyStyle := AnalyzeStoryStyle(novel)

	return EnhancedSceneContext{
		Location:     "现代室内环境",
		TimeOfDay:    "白天",
		Weather:      "晴朗",
		Season:       "春季",
		ArtStyle:     storyStyle,
		ColorPalette: "自然色调",
		Lighting:     "柔和自然光",
		Atmosphere:   "平静祥和",
		CameraAngle:  "平视角度",
		Composition:  "三分法构图",
		Perspective:  "正常透视",
		Background:   "简洁背景",
		Foreground:   "清晰前景",
		Props:        "日常物品",
		SceneSeed:    GenerateSceneSeed("现代室内环境", "白天"),
	}
}

// 增强方法
func (e *EnhancedSceneAnalyzer) enhanceLocation(location string) string {
	if location == "" {
		return "现代室内环境，整洁明亮的空间"
	}

	// 添加更多细节
	enhancements := map[string]string{
		"室内":   "现代室内环境，整洁明亮的空间，温馨的家居装饰",
		"户外":   "开阔的户外环境，自然景观优美，空气清新",
		"城市":   "繁华的现代都市，高楼林立，车水马龙",
		"乡村":   "宁静的乡村环境，田园风光，自然和谐",
		"学校":   "现代化的校园环境，书香气息浓厚",
		"办公室": "专业的办公环境，现代化设施齐全",
	}

	for key, enhancement := range enhancements {
		if strings.Contains(location, key) {
			return enhancement
		}
	}

	return location + "，环境细节丰富"
}

func (e *EnhancedSceneAnalyzer) enhanceWeather(weather string) string {
	if weather == "" {
		return "晴朗，天气宜人"
	}

	enhancements := map[string]string{
		"晴": "晴朗明媚，阳光充足，微风轻拂",
		"雨": "细雨绵绵，空气清新，雨滴晶莹",
		"雪": "雪花飞舞，银装素裹，纯净美丽",
		"云": "云层密布，光影变幻，氛围神秘",
		"风": "微风习习，空气流动，自然清新",
	}

	for key, enhancement := range enhancements {
		if strings.Contains(weather, key) {
			return enhancement
		}
	}

	return weather + "，天气状况良好"
}

func (e *EnhancedSceneAnalyzer) enhanceSeason(season string) string {
	if season == "" {
		return "春季，万物复苏"
	}

	enhancements := map[string]string{
		"春": "春季，万物复苏，生机盎然，花开满园",
		"夏": "夏季，绿意盎然，阳光明媚，活力四射",
		"秋": "秋季，金桂飘香，层林尽染，收获满满",
		"冬": "冬季，雪花纷飞，银装素裹，宁静致远",
	}

	for key, enhancement := range enhancements {
		if strings.Contains(season, key) {
			return enhancement
		}
	}

	return season + "，季节特色鲜明"
}

func (e *EnhancedSceneAnalyzer) enhanceTimeOfDay(timeOfDay string) string {
	if timeOfDay == "" {
		return "明亮的白天，阳光充足"
	}

	enhancements := map[string]string{
		"白天": "明亮的白天，阳光充足，光线柔和",
		"早晨": "清新的早晨，朝阳初升，金色光辉",
		"中午": "明亮的中午，阳光直射，光影分明",
		"下午": "温暖的下午，斜阳西照，光线温柔",
		"傍晚": "美丽的傍晚，夕阳西下，天空绚烂",
		"夜晚": "宁静的夜晚，月光皎洁，星光点点",
		"深夜": "深沉的夜晚，万籁俱寂，神秘氛围",
	}

	for key, enhancement := range enhancements {
		if strings.Contains(timeOfDay, key) {
			return enhancement
		}
	}

	return timeOfDay
}

func (e *EnhancedSceneAnalyzer) enhanceArtStyle(artStyle, novel string) string {
	if artStyle == "" {
		return AnalyzeStoryStyle(novel)
	}

	// 添加质量修饰符
	qualityModifiers := []string{
		"high quality",
		"detailed illustration",
		"professional artwork",
		"masterpiece quality",
	}

	return artStyle + ", " + strings.Join(qualityModifiers, ", ")
}

func (e *EnhancedSceneAnalyzer) enhanceColorPalette(colorPalette, novel string) string {
	if colorPalette == "" {
		storyStyle := AnalyzeStoryStyle(novel)
		return GenerateColorHarmony(storyStyle, "neutral")
	}

	return colorPalette + ", harmonious color scheme, professional color grading"
}

func (e *EnhancedSceneAnalyzer) enhanceLighting(lighting, timeOfDay string) string {
	if lighting == "" {
		// 基于时间推断光照
		if strings.Contains(timeOfDay, "早晨") {
			return "温暖的晨光，金色调，柔和阴影"
		} else if strings.Contains(timeOfDay, "中午") {
			return "明亮的直射光，清晰阴影，高对比度"
		} else if strings.Contains(timeOfDay, "傍晚") {
			return "温暖的夕阳光，橙红色调，长阴影"
		} else if strings.Contains(timeOfDay, "夜晚") {
			return "柔和的月光，蓝色调，神秘阴影"
		}
		return "自然光照，平衡明暗"
	}

	return lighting + ", cinematic lighting, professional illumination"
}

func (e *EnhancedSceneAnalyzer) enhanceAtmosphere(atmosphere, novel string) string {
	if atmosphere == "" {
		// 基于小说内容推断氛围
		novel = strings.ToLower(novel)
		if strings.Contains(novel, "紧张") || strings.Contains(novel, "危险") {
			return "紧张刺激的氛围，戏剧性效果"
		} else if strings.Contains(novel, "浪漫") || strings.Contains(novel, "温馨") {
			return "温馨浪漫的氛围，温暖感人"
		} else if strings.Contains(novel, "神秘") || strings.Contains(novel, "悬疑") {
			return "神秘悬疑的氛围，引人入胜"
		}
		return "平静和谐的氛围"
	}

	return atmosphere + ", immersive atmosphere, emotional depth"
}

// 推断方法
func (e *EnhancedSceneAnalyzer) inferLocation(novel string) string {
	novel = strings.ToLower(novel)
	if strings.Contains(novel, "学校") || strings.Contains(novel, "教室") {
		return "现代化校园环境，书香气息浓厚"
	} else if strings.Contains(novel, "办公") || strings.Contains(novel, "公司") {
		return "专业办公环境，现代化设施齐全"
	} else if strings.Contains(novel, "家") || strings.Contains(novel, "房间") {
		return "温馨的家居环境，舒适宜人"
	} else if strings.Contains(novel, "街道") || strings.Contains(novel, "城市") {
		return "繁华都市街道，现代化建筑"
	}
	return "现代室内环境，整洁明亮"
}

func (e *EnhancedSceneAnalyzer) inferTimeOfDay(novel string) string {
	novel = strings.ToLower(novel)
	if strings.Contains(novel, "早晨") || strings.Contains(novel, "早上") {
		return "清新的早晨，朝阳初升"
	} else if strings.Contains(novel, "中午") {
		return "明亮的中午，阳光直射"
	} else if strings.Contains(novel, "下午") {
		return "温暖的下午，斜阳西照"
	} else if strings.Contains(novel, "傍晚") || strings.Contains(novel, "黄昏") {
		return "美丽的傍晚，夕阳西下"
	} else if strings.Contains(novel, "夜晚") || strings.Contains(novel, "晚上") {
		return "宁静的夜晚，月光皎洁"
	}
	return "明亮的白天，阳光充足"
}

func (e *EnhancedSceneAnalyzer) inferAtmosphere(novel string) string {
	novel = strings.ToLower(novel)
	if strings.Contains(novel, "紧张") || strings.Contains(novel, "激烈") {
		return "紧张刺激的氛围"
	} else if strings.Contains(novel, "温馨") || strings.Contains(novel, "温暖") {
		return "温馨和谐的氛围"
	} else if strings.Contains(novel, "神秘") || strings.Contains(novel, "奇怪") {
		return "神秘悬疑的氛围"
	} else if strings.Contains(novel, "浪漫") || strings.Contains(novel, "甜蜜") {
		return "浪漫温馨的氛围"
	}
	return "平静祥和的氛围"
}

func (e *EnhancedSceneAnalyzer) getStyleColorPalette(storyStyle string) string {
	return GenerateColorHarmony(storyStyle, "neutral")
}

func (e *EnhancedSceneAnalyzer) getStyleBackground(storyStyle string) string {
	switch {
	case strings.Contains(storyStyle, "sci-fi"):
		return "futuristic cityscape with neon lights and high-tech buildings"
	case strings.Contains(storyStyle, "fantasy"):
		return "magical landscape with mystical elements and enchanted atmosphere"
	case strings.Contains(storyStyle, "traditional"):
		return "classical chinese architecture with traditional decorations"
	default:
		return "modern clean background with subtle details"
	}
}

// 其他增强方法...
func (e *EnhancedSceneAnalyzer) enhanceCameraAngle(angle string) string {
	if angle == "" {
		return "平视角度，自然视角"
	}
	return angle + ", cinematic framing"
}

func (e *EnhancedSceneAnalyzer) enhanceComposition(composition string) string {
	if composition == "" {
		return "三分法构图，平衡美观"
	}
	return composition + ", professional composition"
}

func (e *EnhancedSceneAnalyzer) generatePerspective(cameraAngle string) string {
	if strings.Contains(cameraAngle, "俯视") {
		return "俯视透视，空间感强"
	} else if strings.Contains(cameraAngle, "仰视") {
		return "仰视透视，气势宏伟"
	}
	return "正常透视，自然视角"
}

func (e *EnhancedSceneAnalyzer) enhanceBackground(background, location string) string {
	if background == "" {
		return location + "的精美背景，细节丰富"
	}
	return background + ", detailed background, depth of field"
}

func (e *EnhancedSceneAnalyzer) enhanceForeground(foreground string) string {
	if foreground == "" {
		return "清晰的前景元素，层次分明"
	}
	return foreground + ", sharp foreground, clear details"
}

func (e *EnhancedSceneAnalyzer) enhanceProps(props string) string {
	if props == "" {
		return "相关道具物品，增强真实感"
	}
	return props + ", realistic props, detailed objects"
}

func (e *EnhancedSceneAnalyzer) inferProps(novel string) string {
	novel = strings.ToLower(novel)
	if strings.Contains(novel, "书") || strings.Contains(novel, "学习") {
		return "书籍文具，学习用品"
	} else if strings.Contains(novel, "咖啡") || strings.Contains(novel, "茶") {
		return "饮品器具，休闲物品"
	} else if strings.Contains(novel, "电脑") || strings.Contains(novel, "手机") {
		return "电子设备，现代科技"
	}
	return "日常生活用品，环境装饰"
}

// cleanSceneAnalysisJSON 专门清理场景分析的JSON响应
func cleanSceneAnalysisJSON(rawResponse string) string {
	// 移除思考标签
	cleaned := regexp.MustCompile(`(?s)<think>.*?</think>`).ReplaceAllString(rawResponse, "")

	// 移除常见的前缀
	cleaned = regexp.MustCompile(`(?i)^(场景信息|scene|分析结果|result)[:：]\s*`).ReplaceAllString(cleaned, "")

	// 尝试提取JSON对象
	jsonObjectRegex := regexp.MustCompile(`(?s)\{[^{}]*"location"[^{}]*\}`)
	matches := jsonObjectRegex.FindAllString(cleaned, -1)

	if len(matches) > 0 {
		// 找到完整的JSON对象
		return strings.TrimSpace(matches[0])
	}

	// 如果没有找到完整的JSON对象，尝试构建一个
	return buildDefaultSceneJSON()
}

// buildDefaultSceneJSON 构建默认的场景JSON
func buildDefaultSceneJSON() string {
	defaultScene := map[string]string{
		"location":      "城市街道",
		"time_of_day":   "深夜",
		"weather":       "阴雨",
		"season":        "秋季",
		"art_style":     "现代写实",
		"color_palette": "冷色调",
		"lighting":      "昏暗灯光",
		"atmosphere":    "紧张",
		"camera_angle":  "平视",
		"composition":   "中心构图",
		"background":    "城市建筑",
		"foreground":    "人物",
		"props":         "手机",
	}

	jsonBytes, _ := json.Marshal(defaultScene)
	return string(jsonBytes)
}

// SceneData 场景数据结构
type SceneData struct {
	Location     string `json:"location"`
	TimeOfDay    string `json:"time_of_day"`
	Weather      string `json:"weather"`
	Season       string `json:"season"`
	ArtStyle     string `json:"art_style"`
	ColorPalette string `json:"color_palette"`
	Lighting     string `json:"lighting"`
	Atmosphere   string `json:"atmosphere"`
	CameraAngle  string `json:"camera_angle"`
	Composition  string `json:"composition"`
	Background   string `json:"background"`
	Foreground   string `json:"foreground"`
	Props        string `json:"props"`
}

// parseFlexibleSceneJSON 灵活解析场景JSON，处理数组字段
func parseFlexibleSceneJSON(jsonStr string) (SceneData, error) {
	// 首先尝试直接解析为目标结构
	var sceneData SceneData
	if err := json.Unmarshal([]byte(jsonStr), &sceneData); err == nil {
		return sceneData, nil
	}

	// 如果直接解析失败，尝试解析为map[string]interface{}
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
		return SceneData{}, fmt.Errorf("无法解析JSON: %v", err)
	}

	// 手动转换每个字段，处理可能的数组
	sceneData = SceneData{
		Location:     getStringValue(rawData, "location"),
		TimeOfDay:    getStringValue(rawData, "time_of_day"),
		Weather:      getStringValue(rawData, "weather"),
		Season:       getStringValue(rawData, "season"),
		ArtStyle:     getStringValue(rawData, "art_style"),
		ColorPalette: getStringValue(rawData, "color_palette"),
		Lighting:     getStringValue(rawData, "lighting"),
		Atmosphere:   getStringValue(rawData, "atmosphere"),
		CameraAngle:  getStringValue(rawData, "camera_angle"),
		Composition:  getStringValue(rawData, "composition"),
		Background:   getStringValue(rawData, "background"),
		Foreground:   getStringValue(rawData, "foreground"),
		Props:        getStringValue(rawData, "props"),
	}

	return sceneData, nil
}

// getStringValue 从map中获取字符串值，处理数组情况
func getStringValue(data map[string]interface{}, key string) string {
	value, exists := data[key]
	if !exists {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case []interface{}:
		// 如果是数组，转换为逗号分隔的字符串
		var items []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				items = append(items, str)
			}
		}
		return strings.Join(items, "、")
	case []string:
		// 如果是字符串数组
		return strings.Join(v, "、")
	default:
		// 其他类型转换为字符串
		return fmt.Sprintf("%v", v)
	}
}

// getDefaultSceneData 获取默认场景数据
func getDefaultSceneData() SceneData {
	return SceneData{
		Location:     "城市街道",
		TimeOfDay:    "深夜",
		Weather:      "阴雨",
		Season:       "秋季",
		ArtStyle:     "现代写实",
		ColorPalette: "冷色调",
		Lighting:     "昏暗灯光",
		Atmosphere:   "紧张",
		CameraAngle:  "平视",
		Composition:  "中心构图",
		Background:   "城市建筑",
		Foreground:   "人物",
		Props:        "手机",
	}
}
