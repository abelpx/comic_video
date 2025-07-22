package scene

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

// Service 场景服务
type Service struct {
	db        *gorm.DB
	sceneRepo *postgres.SceneRepository
	imageRepo *postgres.SceneImageRepository
	aiService *ai.Service
}

// NewService 创建场景服务
func NewService(db *gorm.DB, aiService *ai.Service) *Service {
	return &Service{
		db:        db,
		sceneRepo: postgres.NewSceneRepository(db),
		imageRepo: postgres.NewSceneImageRepository(db),
		aiService: aiService,
	}
}

// GenerateScenes 生成场景
func (s *Service) GenerateScenes(ctx context.Context, req *GenerateScenesRequest) ([]*entity.Scene, error) {
	log.Printf("[SceneService] 开始生成场景: project=%s", req.ProjectID)

	// 1. 从剧本中分析场景信息
	sceneAnalysis, err := s.analyzeScriptScenes(ctx, req.ScriptContent)
	if err != nil {
		return nil, fmt.Errorf("分析剧本场景失败: %w", err)
	}

	var scenes []*entity.Scene

	// 2. 为每个场景创建详细信息和生成图像
	for i, analysis := range sceneAnalysis {
		scene, err := s.createSceneFromAnalysis(ctx, req.ProjectID, analysis, i+1)
		if err != nil {
			log.Printf("[SceneService] 创建场景失败: %s, error: %v", analysis.Name, err)
			continue
		}

		// 生成场景图像
		if err := s.generateSceneImages(ctx, scene); err != nil {
			log.Printf("[SceneService] 生成场景图像失败: %s, error: %v", scene.Name, err)
		}

		scenes = append(scenes, scene)
	}

	log.Printf("[SceneService] 场景生成完成: project=%s, count=%d", req.ProjectID, len(scenes))
	return scenes, nil
}

// analyzeScriptScenes 分析剧本场景
func (s *Service) analyzeScriptScenes(ctx context.Context, scriptContent string) ([]*SceneAnalysis, error) {
	prompt := fmt.Sprintf(`分析以下剧本，提取所有场景的详细信息：

要求：
1. 识别所有不同的场景和地点
2. 为每个场景提供详细的视觉描述
3. 分析场景的氛围、光照、色彩等
4. 确保描述适合AI图像生成

输出JSON格式：
[
  {
    "name": "场景名称",
    "description": "场景详细描述",
    "location": "具体地点",
    "time_of_day": "时间（如白天、夜晚、黄昏）",
    "weather": "天气状况",
    "season": "季节",
    "art_style": "艺术风格",
    "color_palette": "主要色彩",
    "lighting": "光照效果",
    "atmosphere": "氛围感受",
    "camera_angle": "推荐镜头角度",
    "composition": "构图方式",
    "background": "背景描述",
    "foreground": "前景描述",
    "props": "重要道具"
  }
]

剧本内容：
%s`, scriptContent)

	response, err := s.aiService.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI分析场景失败: %w", err)
	}

	// 解析JSON响应
	var analyses []*SceneAnalysis
	cleanedResponse := s.cleanJSONResponse(response)
	if err := json.Unmarshal([]byte(cleanedResponse), &analyses); err != nil {
		return nil, fmt.Errorf("解析场景分析失败: %w", err)
	}

	return analyses, nil
}

// createSceneFromAnalysis 从分析结果创建场景实体
func (s *Service) createSceneFromAnalysis(ctx context.Context, projectID uuid.UUID, analysis *SceneAnalysis, order int) (*entity.Scene, error) {
	// 构建场景提示词
	prompt := s.buildScenePrompt(analysis)

	scene := &entity.Scene{
		ProjectID:    projectID,
		Name:         analysis.Name,
		Description:  analysis.Description,
		Location:     analysis.Location,
		TimeOfDay:    analysis.TimeOfDay,
		Weather:      analysis.Weather,
		Season:       analysis.Season,
		ArtStyle:     analysis.ArtStyle,
		ColorPalette: analysis.ColorPalette,
		Lighting:     analysis.Lighting,
		Atmosphere:   analysis.Atmosphere,
		CameraAngle:  analysis.CameraAngle,
		Composition:  analysis.Composition,
		Background:   analysis.Background,
		Foreground:   analysis.Foreground,
		Props:        analysis.Props,
		Prompt:       prompt,
	}

	if err := s.sceneRepo.Create(ctx, scene); err != nil {
		return nil, fmt.Errorf("保存场景失败: %w", err)
	}

	return scene, nil
}

// generateSceneImages 生成场景图像
func (s *Service) generateSceneImages(ctx context.Context, scene *entity.Scene) error {
	imageTypes := []string{"concept", "wide_shot", "detail", "atmosphere"}

	for _, imageType := range imageTypes {
		prompt := s.buildImagePrompt(scene, imageType)
		
		// 调用AI图像生成服务
		imageURL, err := s.aiService.GenerateImage(ctx, prompt, 0) // 场景不需要固定种子
		if err != nil {
			log.Printf("[SceneService] 生成图像失败: %s, type: %s, error: %v", scene.Name, imageType, err)
			continue
		}

		// 保存图像记录
		sceneImage := &entity.SceneImage{
			SceneID:   scene.ID,
			ImageURL:  imageURL,
			ImageType: imageType,
			Prompt:    prompt,
		}

		if err := s.imageRepo.Create(ctx, sceneImage); err != nil {
			log.Printf("[SceneService] 保存图像记录失败: %v", err)
		}
	}

	return nil
}

// buildScenePrompt 构建场景提示词
func (s *Service) buildScenePrompt(analysis *SceneAnalysis) string {
	var parts []string

	if analysis.Location != "" {
		parts = append(parts, analysis.Location)
	}

	if analysis.TimeOfDay != "" {
		parts = append(parts, analysis.TimeOfDay)
	}

	if analysis.Weather != "" {
		parts = append(parts, analysis.Weather)
	}

	if analysis.Season != "" {
		parts = append(parts, analysis.Season)
	}

	if analysis.Atmosphere != "" {
		parts = append(parts, analysis.Atmosphere)
	}

	if analysis.Lighting != "" {
		parts = append(parts, analysis.Lighting)
	}

	if analysis.ColorPalette != "" {
		parts = append(parts, analysis.ColorPalette)
	}

	if analysis.ArtStyle != "" {
		parts = append(parts, analysis.ArtStyle)
	}

	if analysis.Background != "" {
		parts = append(parts, fmt.Sprintf("background: %s", analysis.Background))
	}

	if analysis.Foreground != "" {
		parts = append(parts, fmt.Sprintf("foreground: %s", analysis.Foreground))
	}

	if analysis.Props != "" {
		parts = append(parts, fmt.Sprintf("props: %s", analysis.Props))
	}

	parts = append(parts, "high quality, detailed, cinematic")

	return strings.Join(parts, ", ")
}

// buildImagePrompt 构建图像生成提示词
func (s *Service) buildImagePrompt(scene *entity.Scene, imageType string) string {
	basePrompt := scene.Prompt

	switch imageType {
	case "concept":
		return fmt.Sprintf("%s, concept art, detailed environment design", basePrompt)
	case "wide_shot":
		return fmt.Sprintf("%s, wide shot, establishing shot, panoramic view", basePrompt)
	case "detail":
		return fmt.Sprintf("%s, detailed view, close-up elements, texture details", basePrompt)
	case "atmosphere":
		return fmt.Sprintf("%s, atmospheric lighting, mood, cinematic composition", basePrompt)
	default:
		return basePrompt
	}
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

// GenerateScenesRequest 生成场景请求
type GenerateScenesRequest struct {
	ProjectID     uuid.UUID `json:"project_id"`
	ScriptContent string    `json:"script_content"`
}

// SceneAnalysis 场景分析结果
type SceneAnalysis struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
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
