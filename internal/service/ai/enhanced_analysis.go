package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// EnhancedCharacterExtractor 增强版角色提取器
type EnhancedCharacterExtractor struct {
	ollama *OllamaClient
}

// NewEnhancedCharacterExtractor 创建增强版角色提取器
func NewEnhancedCharacterExtractor(ollama *OllamaClient) *EnhancedCharacterExtractor {
	return &EnhancedCharacterExtractor{ollama: ollama}
}

// ExtractDetailedCharacters 提取详细的角色信息
func (e *EnhancedCharacterExtractor) ExtractDetailedCharacters(ctx context.Context, novel string) ([]EnhancedCharacterProfile, error) {
	prompt := fmt.Sprintf(`请仔细分析以下小说内容，提取主要角色的详细信息。

要求：
1. 识别小说中的主要角色（最多5个）
2. 为每个角色提供详细的外观描述
3. 分析角色的性格特征和风格
4. 输出JSON格式，包含以下字段：
   - name: 角色名字
   - age: 年龄段（如"青年"、"中年"）
   - gender: 性别
   - facial_features: 面部特征（眼睛、鼻子等）
   - hair_style: 发型描述
   - clothing: 服装风格
   - body_type: 体型描述
   - personality: 性格特征

输出格式：
[
  {
    "name": "角色名",
    "age": "年龄段",
    "gender": "性别",
    "facial_features": "面部特征描述",
    "hair_style": "发型描述",
    "clothing": "服装风格",
    "body_type": "体型描述",
    "personality": "性格特征"
  }
]

小说内容：
%s

请输出详细的角色信息JSON：`, novel)

	response, err := e.ollama.Generate(prompt, nil)
	if err != nil {
		log.Printf("[AI] 角色提取失败: %v", err)
		return e.generateDefaultCharacters(novel), nil
	}

	// 清理和解析JSON
	cleanedResponse := cleanAndExtractJSON(response)
	log.Printf("[AI] 角色提取原始响应: %s", response[:min(200, len(response))])
	log.Printf("[AI] 角色提取清理后: %s", cleanedResponse)

	// 使用灵活的解析方式处理各种可能的返回格式
	rawCharacters, err := parseFlexibleCharacterJSON(cleanedResponse)
	if err != nil {
		log.Printf("[AI] 详细角色信息解析失败: %v", err)
		log.Printf("[AI] 降级到简单格式")
		return e.extractSimpleCharacters(ctx, novel)
	}

	// 转换为增强角色档案
	var characters []EnhancedCharacterProfile
	for _, raw := range rawCharacters {
		if strings.TrimSpace(raw.Name) == "" {
			continue
		}

		char := EnhancedCharacterProfile{
			Name:           raw.Name,
			Description:    raw.Personality,
			Seed:           GenerateCharacterSeed(raw.Name),
			FacialFeatures: raw.FacialFeatures,
			HairStyle:      raw.HairStyle,
			Clothing:       raw.Clothing,
			BodyType:       raw.BodyType,
			Age:            raw.Age,
			Gender:         raw.Gender,
			ArtStyle:       "anime manga style",
			VisualKeywords: e.generateVisualKeywords(raw),
			ColorScheme:    e.generateCharacterColors(raw),
		}

		// 构建一致性提示词
		char.ConsistencyPrompt = char.BuildCharacterPrompt()
		characters = append(characters, char)
	}

	if len(characters) == 0 {
		return e.generateDefaultCharacters(novel), nil
	}

	log.Printf("[AI] 成功提取到%d个详细角色", len(characters))
	return characters, nil
}

// parseFlatCharacterData 解析扁平化的角色数据
func (e *EnhancedCharacterExtractor) parseFlatCharacterData(flatData []string) []EnhancedCharacterProfile {
	var characters []EnhancedCharacterProfile

	// 创建一个临时的角色对象
	var currentChar struct {
		Name           string
		Age            string
		Gender         string
		FacialFeatures string
		HairStyle      string
		Clothing       string
		BodyType       string
		Personality    string
	}

	// 解析扁平化数据
	for i := 0; i < len(flatData); i++ {
		field := strings.ToLower(strings.TrimSpace(flatData[i]))

		// 检查是否是字段名
		switch field {
		case "name":
			// 如果已经有角色数据，先保存
			if currentChar.Name != "" {
				char := e.createEnhancedCharacter(currentChar)
				if char.Name != "" {
					characters = append(characters, char)
				}
				// 重置当前角色
				currentChar = struct {
					Name           string
					Age            string
					Gender         string
					FacialFeatures string
					HairStyle      string
					Clothing       string
					BodyType       string
					Personality    string
				}{}
			}
			// 获取下一个值作为名字
			if i+1 < len(flatData) {
				i++
				currentChar.Name = flatData[i]
			}
		case "age":
			if i+1 < len(flatData) {
				i++
				currentChar.Age = flatData[i]
			}
		case "gender":
			if i+1 < len(flatData) {
				i++
				currentChar.Gender = flatData[i]
			}
		case "facial_features":
			if i+1 < len(flatData) {
				i++
				currentChar.FacialFeatures = flatData[i]
			}
		case "hair_style":
			if i+1 < len(flatData) {
				i++
				currentChar.HairStyle = flatData[i]
			}
		case "clothing":
			if i+1 < len(flatData) {
				i++
				currentChar.Clothing = flatData[i]
			}
		case "body_type":
			if i+1 < len(flatData) {
				i++
				currentChar.BodyType = flatData[i]
			}
		case "personality":
			if i+1 < len(flatData) {
				i++
				currentChar.Personality = flatData[i]
			}
		}
	}

	// 保存最后一个角色
	if currentChar.Name != "" {
		char := e.createEnhancedCharacter(currentChar)
		if char.Name != "" {
			characters = append(characters, char)
		}
	}

	log.Printf("[AI] 从扁平化数据解析出%d个角色", len(characters))
	return characters
}

// createEnhancedCharacter 从原始数据创建增强角色
func (e *EnhancedCharacterExtractor) createEnhancedCharacter(raw struct {
	Name           string
	Age            string
	Gender         string
	FacialFeatures string
	HairStyle      string
	Clothing       string
	BodyType       string
	Personality    string
}) EnhancedCharacterProfile {
	// 过滤无效角色
	if strings.TrimSpace(raw.Name) == "" ||
	   strings.Contains(strings.ToLower(raw.Name), "暴徒") ||
	   strings.Contains(strings.ToLower(raw.Name), "医生") ||
	   strings.Contains(strings.ToLower(raw.Name), "叙述") {
		return EnhancedCharacterProfile{}
	}

	char := EnhancedCharacterProfile{
		Name:           raw.Name,
		Description:    raw.Personality,
		Seed:           GenerateCharacterSeed(raw.Name),
		FacialFeatures: raw.FacialFeatures,
		HairStyle:      raw.HairStyle,
		Clothing:       raw.Clothing,
		BodyType:       raw.BodyType,
		Age:            raw.Age,
		Gender:         raw.Gender,
		ArtStyle:       "anime manga style",
		VisualKeywords: e.generateVisualKeywords(raw),
		ColorScheme:    e.generateCharacterColors(raw),
	}

	// 构建一致性提示词
	char.ConsistencyPrompt = char.BuildCharacterPrompt()
	return char
}

// extractSimpleCharacters 提取简单角色信息的降级方案
func (e *EnhancedCharacterExtractor) extractSimpleCharacters(ctx context.Context, novel string) ([]EnhancedCharacterProfile, error) {
	prompt := fmt.Sprintf(`请分析以下小说，提取主要角色名字：

小说内容：
%s

请输出角色名字的JSON数组，如：["张三", "李四"]`, novel)

	response, err := e.ollama.Generate(prompt, nil)
	if err != nil {
		return e.generateDefaultCharacters(novel), nil
	}

	cleanedResponse := cleanAndExtractJSON(response)
	var characterNames []string
	if err := json.Unmarshal([]byte(cleanedResponse), &characterNames); err != nil {
		return e.generateDefaultCharacters(novel), nil
	}

	var characters []EnhancedCharacterProfile
	for i, name := range characterNames {
		if strings.TrimSpace(name) == "" {
			continue
		}

		char := EnhancedCharacterProfile{
			Name:           name,
			Description:    "主要角色",
			Seed:           GenerateCharacterSeed(name),
			FacialFeatures: e.getDefaultFacialFeatures(i),
			HairStyle:      e.getDefaultHairStyle(i),
			Clothing:       e.getDefaultClothing(i),
			BodyType:       "标准体型",
			Age:            "青年",
			Gender:         e.getDefaultGender(i),
			ArtStyle:       "anime manga style",
			VisualKeywords: []string{"anime character", "detailed"},
			ColorScheme:    e.getDefaultColorScheme(i),
		}

		char.ConsistencyPrompt = char.BuildCharacterPrompt()
		characters = append(characters, char)
	}

	return characters, nil
}

// generateDefaultCharacters 生成默认角色
func (e *EnhancedCharacterExtractor) generateDefaultCharacters(novel string) []EnhancedCharacterProfile {
	// 分析小说类型来生成合适的默认角色
	storyStyle := AnalyzeStoryStyle(novel)
	
	var characters []EnhancedCharacterProfile
	
	// 主角
	mainChar := EnhancedCharacterProfile{
		Name:           "主角",
		Description:    "故事主人公，勇敢正义",
		Seed:           GenerateCharacterSeed("主角"),
		FacialFeatures: "清秀的面容，明亮的眼睛",
		HairStyle:      "整齐的短发",
		Clothing:       e.getStyleAppropriateClothing(storyStyle, "main"),
		BodyType:       "标准体型",
		Age:            "青年",
		Gender:         "男性",
		ArtStyle:       "anime manga style",
		VisualKeywords: []string{"protagonist", "heroic", "detailed"},
		ColorScheme:    "warm colors with blue accents",
	}
	mainChar.ConsistencyPrompt = mainChar.BuildCharacterPrompt()
	characters = append(characters, mainChar)

	// 配角
	supportChar := EnhancedCharacterProfile{
		Name:           "重要配角",
		Description:    "重要的支持角色，聪明善良",
		Seed:           GenerateCharacterSeed("重要配角"),
		FacialFeatures: "温和的面容，智慧的眼神",
		HairStyle:      "优雅的长发",
		Clothing:       e.getStyleAppropriateClothing(storyStyle, "support"),
		BodyType:       "苗条体型",
		Age:            "青年",
		Gender:         "女性",
		ArtStyle:       "anime manga style",
		VisualKeywords: []string{"supporting character", "elegant", "detailed"},
		ColorScheme:    "soft colors with purple accents",
	}
	supportChar.ConsistencyPrompt = supportChar.BuildCharacterPrompt()
	characters = append(characters, supportChar)

	return characters
}

// generateVisualKeywords 生成视觉关键词
func (e *EnhancedCharacterExtractor) generateVisualKeywords(char interface{}) []string {
	// 通用的关键词处理，支持多种结构体类型
	var gender, age, personality string

	// 使用类型断言或反射来获取字段值
	switch v := char.(type) {
	case CharacterData:
		gender = v.Gender
		age = v.Age
		personality = v.Personality
	case struct {
		Name           string `json:"name"`
		Age            string `json:"age"`
		Gender         string `json:"gender"`
		FacialFeatures string `json:"facial_features"`
		HairStyle      string `json:"hair_style"`
		Clothing       string `json:"clothing"`
		BodyType       string `json:"body_type"`
		Personality    string `json:"personality"`
	}:
		gender = v.Gender
		age = v.Age
		personality = v.Personality
	case struct {
		Name           string
		Age            string
		Gender         string
		FacialFeatures string
		HairStyle      string
		Clothing       string
		BodyType       string
		Personality    string
	}:
		gender = v.Gender
		age = v.Age
		personality = v.Personality
	default:
		// 默认关键词
		return []string{"anime character", "detailed", "high quality"}
	}
	keywords := []string{"anime character", "detailed", "high quality"}
	
	// 基于性别添加关键词
	if strings.Contains(strings.ToLower(gender), "女") {
		keywords = append(keywords, "female", "beautiful")
	} else {
		keywords = append(keywords, "male", "handsome")
	}

	// 基于年龄添加关键词
	if strings.Contains(age, "青年") {
		keywords = append(keywords, "young", "energetic")
	} else if strings.Contains(age, "中年") {
		keywords = append(keywords, "mature", "experienced")
	}

	// 基于性格添加关键词
	if strings.Contains(personality, "勇敢") {
		keywords = append(keywords, "brave", "heroic")
	}
	if strings.Contains(personality, "聪明") {
		keywords = append(keywords, "intelligent", "wise")
	}
	
	return keywords
}

// generateCharacterColors 生成角色色彩方案
func (e *EnhancedCharacterExtractor) generateCharacterColors(char interface{}) string {
	// 获取性格信息
	var personality string

	switch v := char.(type) {
	case CharacterData:
		personality = v.Personality
	case struct {
		Name           string `json:"name"`
		Age            string `json:"age"`
		Gender         string `json:"gender"`
		FacialFeatures string `json:"facial_features"`
		HairStyle      string `json:"hair_style"`
		Clothing       string `json:"clothing"`
		BodyType       string `json:"body_type"`
		Personality    string `json:"personality"`
	}:
		personality = v.Personality
	case struct {
		Name           string
		Age            string
		Gender         string
		FacialFeatures string
		HairStyle      string
		Clothing       string
		BodyType       string
		Personality    string
	}:
		personality = v.Personality
	default:
		return "balanced natural colors"
	}

	// 基于性格生成色彩方案
	if strings.Contains(personality, "勇敢") {
		return "warm colors with red and orange accents"
	}
	if strings.Contains(personality, "聪明") {
		return "cool colors with blue and purple accents"
	}
	if strings.Contains(personality, "温柔") {
		return "soft pastel colors with pink accents"
	}
	if strings.Contains(personality, "神秘") {
		return "dark colors with silver accents"
	}

	return "balanced natural colors"
}

// 辅助方法
func (e *EnhancedCharacterExtractor) getDefaultFacialFeatures(index int) string {
	features := []string{
		"清秀的面容，明亮的眼睛",
		"温和的面容，智慧的眼神",
		"英俊的面容，坚定的眼神",
		"美丽的面容，温柔的眼神",
		"成熟的面容，深邃的眼神",
	}
	return features[index%len(features)]
}

func (e *EnhancedCharacterExtractor) getDefaultHairStyle(index int) string {
	styles := []string{
		"整齐的短发",
		"优雅的长发",
		"时尚的中长发",
		"自然的卷发",
		"简洁的发型",
	}
	return styles[index%len(styles)]
}

func (e *EnhancedCharacterExtractor) getDefaultClothing(index int) string {
	clothes := []string{
		"休闲的现代服装",
		"优雅的正装",
		"时尚的街头服饰",
		"简约的日常装扮",
		"专业的工作服装",
	}
	return clothes[index%len(clothes)]
}

func (e *EnhancedCharacterExtractor) getDefaultGender(index int) string {
	if index%2 == 0 {
		return "男性"
	}
	return "女性"
}

func (e *EnhancedCharacterExtractor) getDefaultColorScheme(index int) string {
	schemes := []string{
		"warm colors with blue accents",
		"soft colors with purple accents",
		"cool colors with green accents",
		"bright colors with orange accents",
		"neutral colors with gold accents",
	}
	return schemes[index%len(schemes)]
}

func (e *EnhancedCharacterExtractor) getStyleAppropriateClothing(storyStyle, role string) string {
	switch {
	case strings.Contains(storyStyle, "sci-fi"):
		if role == "main" {
			return "futuristic tech suit with glowing elements"
		}
		return "sleek cyberpunk outfit with neon details"
	case strings.Contains(storyStyle, "fantasy"):
		if role == "main" {
			return "heroic fantasy armor with magical elements"
		}
		return "elegant magical robes with mystical symbols"
	case strings.Contains(storyStyle, "traditional"):
		if role == "main" {
			return "traditional chinese hanfu with elegant patterns"
		}
		return "classical chinese dress with silk fabric"
	default:
		if role == "main" {
			return "modern casual outfit with stylish details"
		}
		return "contemporary fashion with elegant design"
	}
}

// CharacterData 角色数据结构
type CharacterData struct {
	Name           string `json:"name"`
	Age            string `json:"age"`
	Gender         string `json:"gender"`
	FacialFeatures string `json:"facial_features"`
	HairStyle      string `json:"hair_style"`
	Clothing       string `json:"clothing"`
	BodyType       string `json:"body_type"`
	Personality    string `json:"personality"`
}

// parseFlexibleCharacterJSON 灵活解析角色JSON，处理各种格式
func parseFlexibleCharacterJSON(jsonStr string) ([]CharacterData, error) {
	// 首先尝试直接解析为目标结构数组
	var characters []CharacterData
	if err := json.Unmarshal([]byte(jsonStr), &characters); err == nil {
		return characters, nil
	}

	// 尝试解析为单个对象
	var singleChar CharacterData
	if err := json.Unmarshal([]byte(jsonStr), &singleChar); err == nil {
		return []CharacterData{singleChar}, nil
	}

	// 尝试解析为字符串数组（角色名字列表）
	var characterNames []string
	if err := json.Unmarshal([]byte(jsonStr), &characterNames); err == nil {
		var result []CharacterData
		for _, name := range characterNames {
			if strings.TrimSpace(name) != "" {
				result = append(result, CharacterData{
					Name:           name,
					Age:            "未知",
					Gender:         "未知",
					FacialFeatures: "普通面容",
					HairStyle:      "普通发型",
					Clothing:       "普通服装",
					BodyType:       "标准体型",
					Personality:    "角色",
				})
			}
		}
		return result, nil
	}

	// 尝试解析为map[string]interface{}处理混合类型
	var rawData interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
		return nil, fmt.Errorf("无法解析JSON: %v", err)
	}

	// 处理不同的数据结构
	switch data := rawData.(type) {
	case []interface{}:
		return parseCharacterArray(data)
	case map[string]interface{}:
		char := parseCharacterMap(data)
		if char.Name != "" {
			return []CharacterData{char}, nil
		}
	case string:
		// 如果是单个字符串，作为角色名处理
		if strings.TrimSpace(data) != "" {
			return []CharacterData{{
				Name:           data,
				Age:            "未知",
				Gender:         "未知",
				FacialFeatures: "普通面容",
				HairStyle:      "普通发型",
				Clothing:       "普通服装",
				BodyType:       "标准体型",
				Personality:    "角色",
			}}, nil
		}
	}

	return nil, fmt.Errorf("无法解析角色数据")
}

// parseCharacterArray 解析角色数组
func parseCharacterArray(data []interface{}) ([]CharacterData, error) {
	var characters []CharacterData

	for _, item := range data {
		switch v := item.(type) {
		case map[string]interface{}:
			char := parseCharacterMap(v)
			if char.Name != "" {
				characters = append(characters, char)
			}
		case string:
			if strings.TrimSpace(v) != "" {
				characters = append(characters, CharacterData{
					Name:           v,
					Age:            "未知",
					Gender:         "未知",
					FacialFeatures: "普通面容",
					HairStyle:      "普通发型",
					Clothing:       "普通服装",
					BodyType:       "标准体型",
					Personality:    "角色",
				})
			}
		}
	}

	return characters, nil
}

// parseCharacterMap 解析角色map
func parseCharacterMap(data map[string]interface{}) CharacterData {
	char := CharacterData{
		Name:           getStringValueFromMap(data, "name"),
		Age:            getStringValueFromMap(data, "age"),
		Gender:         getStringValueFromMap(data, "gender"),
		FacialFeatures: getStringValueFromMap(data, "facial_features"),
		HairStyle:      getStringValueFromMap(data, "hair_style"),
		Clothing:       getStringValueFromMap(data, "clothing"),
		BodyType:       getStringValueFromMap(data, "body_type"),
		Personality:    getStringValueFromMap(data, "personality"),
	}

	// 设置默认值
	if char.Age == "" {
		char.Age = "未知"
	}
	if char.Gender == "" {
		char.Gender = "未知"
	}
	if char.FacialFeatures == "" {
		char.FacialFeatures = "普通面容"
	}
	if char.HairStyle == "" {
		char.HairStyle = "普通发型"
	}
	if char.Clothing == "" {
		char.Clothing = "普通服装"
	}
	if char.BodyType == "" {
		char.BodyType = "标准体型"
	}
	if char.Personality == "" {
		char.Personality = "角色"
	}

	return char
}

// getStringValueFromMap 从map中获取字符串值
func getStringValueFromMap(data map[string]interface{}, key string) string {
	value, exists := data[key]
	if !exists {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case []interface{}:
		// 如果是数组，取第一个元素或连接
		if len(v) > 0 {
			if str, ok := v[0].(string); ok {
				return str
			}
		}
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
