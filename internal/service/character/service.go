package character

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/postgres"
	"comic_video/internal/service/ai"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service 角色服务
type Service struct {
	db           *gorm.DB
	characterRepo *postgres.CharacterRepository
	imageRepo     *postgres.CharacterImageRepository
	aiService     *ai.Service
}

// NewService 创建角色服务
func NewService(db *gorm.DB, aiService *ai.Service) *Service {
	return &Service{
		db:           db,
		characterRepo: postgres.NewCharacterRepository(db),
		imageRepo:     postgres.NewCharacterImageRepository(db),
		aiService:     aiService,
	}
}

// GenerateCharacters 生成角色
func (s *Service) GenerateCharacters(ctx context.Context, req *GenerateCharactersRequest) ([]*entity.Character, error) {
	log.Printf("[CharacterService] 开始生成角色: project=%s", req.ProjectID)

	// 1. 从剧本中提取角色信息
	characterProfiles, err := s.extractCharacterProfiles(ctx, req.ScriptContent)
	if err != nil {
		return nil, fmt.Errorf("提取角色信息失败: %w", err)
	}

	var characters []*entity.Character

	// 2. 为每个角色创建详细信息和生成图像
	for _, profile := range characterProfiles {
		projectUUID, _ := uuid.Parse(req.ProjectID)
		character, err := s.createCharacterFromProfile(ctx, projectUUID, profile)
		if err != nil {
			log.Printf("[CharacterService] 创建角色失败: %s, error: %v", profile.Name, err)
			continue
		}

		// 根据角色重要性决定生成图像的数量和质量
		if err := s.generateCharacterImagesWithPriority(ctx, character, profile.Importance); err != nil {
			log.Printf("[CharacterService] 生成角色图像失败: %s, error: %v", character.Name, err)
		}

		characters = append(characters, character)
	}

	log.Printf("[CharacterService] 角色生成完成: project=%s, count=%d", req.ProjectID, len(characters))
	return characters, nil
}

// extractCharacterProfiles 从剧本中提取角色信息
func (s *Service) extractCharacterProfiles(ctx context.Context, scriptContent string) ([]*CharacterProfile, error) {
	prompt := fmt.Sprintf(`分析以下剧本，提取所有重要角色的详细信息：

要求：
1. 识别所有重要角色（主角、配角、有台词或重要作用的次要角色）
2. 为每个角色提供详细的外观描述
3. 分析角色的性格特征和风格
4. 确保描述适合AI图像生成
5. 按重要性排序，主角在前，配角次之

输出JSON格式：
[
  {
    "name": "角色名",
    "age": "年龄段（如青年、中年）",
    "gender": "性别",
    "facial_features": "面部特征详细描述",
    "hair_style": "发型和发色描述",
    "clothing": "服装风格和颜色",
    "body_type": "体型描述",
    "personality": "性格特征",
    "role": "在故事中的作用",
    "importance": "角色重要性（主角/重要配角/次要角色/群众角色）",
    "screen_time": "预估出场频率（高/中/低）",
    "art_style": "艺术风格（如anime, realistic等）",
    "visual_keywords": "视觉关键词，用逗号分隔"
  }
]

剧本内容：
%s`, scriptContent)

	response, err := s.aiService.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI提取角色信息失败: %w", err)
	}

	// 解析JSON响应
	var profiles []*CharacterProfile
	cleanedResponse := s.cleanJSONResponse(response)
	if err := json.Unmarshal([]byte(cleanedResponse), &profiles); err != nil {
		return nil, fmt.Errorf("解析角色信息失败: %w", err)
	}

	return profiles, nil
}

// createCharacterFromProfile 从角色档案创建角色实体
func (s *Service) createCharacterFromProfile(ctx context.Context, projectID uuid.UUID, profile *CharacterProfile) (*entity.Character, error) {
	// 生成一致性提示词
	consistencyPrompt := s.buildConsistencyPrompt(profile)

	character := &entity.Character{
		ProjectID:         projectID,
		Name:              profile.Name,
		Description:       profile.Personality,
		Age:               profile.Age,
		Gender:            profile.Gender,
		FacialFeatures:    profile.FacialFeatures,
		HairStyle:         profile.HairStyle,
		Clothing:          profile.Clothing,
		BodyType:          profile.BodyType,
		Personality:       profile.Personality,
		Importance:        profile.Importance,    // 新增：角色重要性
		ScreenTime:        profile.ScreenTime,   // 新增：出场频率
		ArtStyle:          profile.ArtStyle,
		VisualKeywords:    profile.VisualKeywords,
		ColorScheme:       s.generateColorScheme(profile),
		ConsistencyPrompt: consistencyPrompt,
		Seed:              s.generateSeed(profile.Name),
	}

	if err := s.characterRepo.Create(ctx, character); err != nil {
		return nil, fmt.Errorf("保存角色失败: %w", err)
	}

	return character, nil
}

// generateCharacterImages 生成角色图像
func (s *Service) generateCharacterImages(ctx context.Context, character *entity.Character) error {
	imageTypes := []string{"portrait", "full_body", "expression_happy", "expression_sad"}

	for _, imageType := range imageTypes {
		prompt := s.buildImagePrompt(character, imageType)
		
		// 调用AI图像生成服务
		imageURL, err := s.aiService.GenerateImage(ctx, prompt, character.Seed)
		if err != nil {
			log.Printf("[CharacterService] 生成图像失败: %s, type: %s, error: %v", character.Name, imageType, err)
			continue
		}

		// 保存图像记录
		characterImage := &entity.CharacterImage{
			CharacterID: character.ID,
			ImageURL:    imageURL,
			ImageType:   imageType,
			Prompt:      prompt,
			Seed:        character.Seed,
			IsReference: imageType == "portrait", // 肖像作为参考图
		}

		if err := s.imageRepo.Create(ctx, characterImage); err != nil {
			log.Printf("[CharacterService] 保存图像记录失败: %v", err)
		}
	}

	return nil
}

// generateCharacterImagesWithPriority 根据角色重要性生成不同数量的图像
func (s *Service) generateCharacterImagesWithPriority(ctx context.Context, character *entity.Character, importance string) error {
	var imageTypes []string

	// 根据角色重要性决定生成的图像类型
	switch strings.ToLower(importance) {
	case "主角":
		// 主角生成最全面的图像集
		imageTypes = []string{
			"portrait", "full_body", "close_up",
			"expression_happy", "expression_sad", "expression_angry", "expression_surprised",
			"action_pose", "casual_outfit", "formal_outfit",
		}
	case "重要配角":
		// 重要配角生成较多图像
		imageTypes = []string{
			"portrait", "full_body",
			"expression_happy", "expression_sad", "expression_neutral",
			"action_pose",
		}
	case "次要角色":
		// 次要角色生成基础图像
		imageTypes = []string{
			"portrait", "full_body", "expression_neutral",
		}
	case "群众角色":
		// 群众角色只生成基本肖像
		imageTypes = []string{"portrait"}
	default:
		// 默认按配角处理
		imageTypes = []string{
			"portrait", "full_body", "expression_happy", "expression_sad",
		}
	}

	log.Printf("[CharacterService] 为角色 %s (%s) 生成 %d 种图像", character.Name, importance, len(imageTypes))

	for _, imageType := range imageTypes {
		prompt := s.buildImagePrompt(character, imageType)

		// 调用AI图像生成服务
		imageURL, err := s.aiService.GenerateImage(ctx, prompt, character.Seed)
		if err != nil {
			log.Printf("[CharacterService] 生成图像失败: %s, type: %s, error: %v", character.Name, imageType, err)
			continue
		}

		// 保存图像记录
		characterImage := &entity.CharacterImage{
			CharacterID: character.ID,
			ImageURL:    imageURL,
			ImageType:   imageType,
			Prompt:      prompt,
			Seed:        character.Seed,
			IsReference: imageType == "portrait", // 肖像作为参考图
		}

		if err := s.imageRepo.Create(ctx, characterImage); err != nil {
			log.Printf("[CharacterService] 保存图像记录失败: %v", err)
		}
	}

	return nil
}

// buildConsistencyPrompt 构建一致性提示词
func (s *Service) buildConsistencyPrompt(profile *CharacterProfile) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Character: %s", profile.Name))
	
	if profile.Age != "" {
		parts = append(parts, fmt.Sprintf("Age: %s", profile.Age))
	}
	
	if profile.Gender != "" {
		parts = append(parts, fmt.Sprintf("Gender: %s", profile.Gender))
	}
	
	if profile.FacialFeatures != "" {
		parts = append(parts, fmt.Sprintf("Face: %s", profile.FacialFeatures))
	}
	
	if profile.HairStyle != "" {
		parts = append(parts, fmt.Sprintf("Hair: %s", profile.HairStyle))
	}
	
	if profile.Clothing != "" {
		parts = append(parts, fmt.Sprintf("Clothing: %s", profile.Clothing))
	}
	
	if profile.BodyType != "" {
		parts = append(parts, fmt.Sprintf("Body: %s", profile.BodyType))
	}
	
	if profile.ArtStyle != "" {
		parts = append(parts, fmt.Sprintf("Style: %s", profile.ArtStyle))
	}
	
	if profile.VisualKeywords != "" {
		parts = append(parts, fmt.Sprintf("Keywords: %s", profile.VisualKeywords))
	}

	return strings.Join(parts, ", ")
}

// buildImagePrompt 构建图像生成提示词
func (s *Service) buildImagePrompt(character *entity.Character, imageType string) string {
	basePrompt := character.ConsistencyPrompt

	switch imageType {
	case "portrait":
		return fmt.Sprintf("%s, portrait, head and shoulders, detailed face, high quality", basePrompt)
	case "full_body":
		return fmt.Sprintf("%s, full body, standing pose, detailed, high quality", basePrompt)
	case "expression_happy":
		return fmt.Sprintf("%s, portrait, happy expression, smiling, detailed face", basePrompt)
	case "expression_sad":
		return fmt.Sprintf("%s, portrait, sad expression, detailed face", basePrompt)
	default:
		return basePrompt
	}
}

// generateColorScheme 生成配色方案
func (s *Service) generateColorScheme(profile *CharacterProfile) string {
	// 基于角色特征生成配色方案的简单逻辑
	schemes := map[string]string{
		"warm":   "#FF6B6B,#FFE66D,#FF8E53",
		"cool":   "#4ECDC4,#45B7D1,#96CEB4",
		"dark":   "#2C3E50,#34495E,#7F8C8D",
		"bright": "#E74C3C,#F39C12,#9B59B6",
	}

	// 根据角色性格选择配色
	personality := strings.ToLower(profile.Personality)
	if strings.Contains(personality, "开朗") || strings.Contains(personality, "活泼") {
		return schemes["bright"]
	} else if strings.Contains(personality, "冷静") || strings.Contains(personality, "理性") {
		return schemes["cool"]
	} else if strings.Contains(personality, "神秘") || strings.Contains(personality, "深沉") {
		return schemes["dark"]
	}

	return schemes["warm"] // 默认暖色调
}

// generateSeed 生成种子值
func (s *Service) generateSeed(name string) int64 {
	// 基于角色名生成一致的种子值
	hash := int64(0)
	for _, char := range name {
		hash = hash*31 + int64(char)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash % 1000000
}

// cleanJSONResponse 清理JSON响应
func (s *Service) cleanJSONResponse(response string) string {
	response = strings.ReplaceAll(response, "```json", "")
	response = strings.ReplaceAll(response, "```", "")
	response = strings.TrimSpace(response)

	start := strings.Index(response, "[")
	end := strings.LastIndex(response, "]")

	if start != -1 && end != -1 && end > start {
		return response[start : end+1]
	}

	return response
}

// GenerateCharactersRequest 生成角色请求
type GenerateCharactersRequest struct {
	ProjectID string `json:"project_id"`
	Script    string `json:"script"`
	ScriptContent string `json:"script_content"`
	Style     string `json:"style"`
}

// CharacterProfile 角色档案
type CharacterProfile struct {
	Name           string `json:"name"`
	Age            string `json:"age"`
	Gender         string `json:"gender"`
	FacialFeatures string `json:"facial_features"`
	HairStyle      string `json:"hair_style"`
	Clothing       string `json:"clothing"`
	BodyType       string `json:"body_type"`
	Personality    string `json:"personality"`
	Role           string `json:"role"`
	Importance     string `json:"importance"`     // 角色重要性
	ScreenTime     string `json:"screen_time"`    // 预估出场频率
	ArtStyle       string `json:"art_style"`
	VisualKeywords string `json:"visual_keywords"`
}
