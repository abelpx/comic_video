package storyboard

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

// Service 分镜服务
type Service struct {
	db            *gorm.DB
	storyboardRepo *postgres.StoryboardRepository
	frameRepo      *postgres.StoryboardFrameRepository
	aiService      *ai.Service
}

// NewService 创建分镜服务
func NewService(db *gorm.DB, aiService *ai.Service) *Service {
	return &Service{
		db:            db,
		storyboardRepo: postgres.NewStoryboardRepository(db),
		frameRepo:      postgres.NewStoryboardFrameRepository(db),
		aiService:      aiService,
	}
}

// GenerateStoryboard 生成分镜
func (s *Service) GenerateStoryboard(ctx context.Context, req *GenerateStoryboardRequest) (*entity.Storyboard, error) {
	log.Printf("[StoryboardService] 开始生成分镜: project=%s", req.ProjectID)

	// 1. 分析剧本生成分镜计划
	storyboardPlan, err := s.analyzeScriptForStoryboard(ctx, req.ScriptContent)
	if err != nil {
		return nil, fmt.Errorf("分析剧本生成分镜计划失败: %w", err)
	}

	// 2. 创建分镜实体
	projectUUID, _ := uuid.Parse(req.ProjectID)
	storyboard := &entity.Storyboard{
		ProjectID:   projectUUID,
		ScriptID:    req.ScriptID,
		Title:       req.Title,
		Description: storyboardPlan.Description,
		TotalFrames: len(storyboardPlan.Frames),
		Duration:    storyboardPlan.TotalDuration,
		FrameRate:   24,
		Resolution:  "1920x1080",
	}

	if err := s.storyboardRepo.Create(ctx, storyboard); err != nil {
		return nil, fmt.Errorf("保存分镜失败: %w", err)
	}

	// 3. 创建分镜帧
	for i, framePlan := range storyboardPlan.Frames {
		frame, err := s.createStoryboardFrame(ctx, storyboard.ID, framePlan, i+1)
		if err != nil {
			log.Printf("[StoryboardService] 创建分镜帧失败: frame=%d, error: %v", i+1, err)
			continue
		}

		// 生成分镜帧图像
		if err := s.generateFrameImage(ctx, frame); err != nil {
			log.Printf("[StoryboardService] 生成分镜帧图像失败: frame=%s, error: %v", frame.ID, err)
		}
	}

	log.Printf("[StoryboardService] 分镜生成完成: storyboard=%s, frames=%d", storyboard.ID, len(storyboardPlan.Frames))
	return storyboard, nil
}

// analyzeScriptForStoryboard 分析剧本生成分镜计划
func (s *Service) analyzeScriptForStoryboard(ctx context.Context, scriptContent string) (*StoryboardPlan, error) {
	prompt := fmt.Sprintf(`分析以下剧本，生成详细的分镜计划：

要求：
1. 将剧本分解为具体的镜头
2. 为每个镜头指定拍摄角度、景别、运动
3. 估算每个镜头的时长
4. 提供详细的视觉描述
5. 包含角色动作和对话

输出JSON格式：
{
  "description": "分镜总体描述",
  "total_duration": 120,
  "frames": [
    {
      "title": "镜头标题",
      "description": "镜头详细描述",
      "dialogue": "对话内容",
      "action": "动作描述",
      "camera_angle": "镜头角度（如俯视、仰视、平视）",
      "camera_move": "镜头运动（如推拉摇移）",
      "shot_type": "景别（如特写、中景、全景）",
      "duration": 5.0,
      "transition": "转场效果",
      "characters": ["角色1", "角色2"],
      "props": ["道具1", "道具2"],
      "scene_description": "场景描述"
    }
  ]
}

剧本内容：
%s`, scriptContent)

	response, err := s.aiService.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI生成分镜计划失败: %w", err)
	}

	// 解析JSON响应
	var plan StoryboardPlan
	cleanedResponse := s.cleanJSONResponse(response)
	if err := json.Unmarshal([]byte(cleanedResponse), &plan); err != nil {
		return nil, fmt.Errorf("解析分镜计划失败: %w", err)
	}

	return &plan, nil
}

// createStoryboardFrame 创建分镜帧
func (s *Service) createStoryboardFrame(ctx context.Context, storyboardID uuid.UUID, framePlan *FramePlan, frameNumber int) (*entity.StoryboardFrame, error) {
	// 将角色和道具数组转换为JSON字符串
	charactersJSON, _ := json.Marshal(framePlan.Characters)
	propsJSON, _ := json.Marshal(framePlan.Props)

	frame := &entity.StoryboardFrame{
		StoryboardID: storyboardID,
		FrameNumber:  frameNumber,
		Title:        framePlan.Title,
		Description:  framePlan.Description,
		Dialogue:     framePlan.Dialogue,
		Action:       framePlan.Action,
		CameraAngle:  framePlan.CameraAngle,
		CameraMove:   framePlan.CameraMove,
		ShotType:     framePlan.ShotType,
		Duration:     framePlan.Duration,
		Transition:   framePlan.Transition,
		Characters:   string(charactersJSON),
		Props:        string(propsJSON),
		Prompt:       s.buildFramePrompt(framePlan),
	}

	if err := s.frameRepo.Create(ctx, frame); err != nil {
		return nil, fmt.Errorf("保存分镜帧失败: %w", err)
	}

	return frame, nil
}

// generateFrameImage 生成分镜帧图像
func (s *Service) generateFrameImage(ctx context.Context, frame *entity.StoryboardFrame) error {
	prompt := frame.Prompt

	// 调用AI图像生成服务
	imageURL, err := s.aiService.GenerateImage(ctx, prompt, 0)
	if err != nil {
		return fmt.Errorf("生成分镜帧图像失败: %w", err)
	}

	// 更新分镜帧的图像URL
	frame.ImageURL = imageURL
	return s.frameRepo.Update(ctx, frame)
}

// buildFramePrompt 构建分镜帧提示词
func (s *Service) buildFramePrompt(framePlan *FramePlan) string {
	var parts []string

	// 基础场景描述
	if framePlan.SceneDescription != "" {
		parts = append(parts, framePlan.SceneDescription)
	}

	// 角色信息
	if len(framePlan.Characters) > 0 {
		parts = append(parts, fmt.Sprintf("characters: %s", strings.Join(framePlan.Characters, ", ")))
	}

	// 镜头信息
	if framePlan.ShotType != "" {
		parts = append(parts, framePlan.ShotType)
	}

	if framePlan.CameraAngle != "" {
		parts = append(parts, framePlan.CameraAngle)
	}

	// 动作描述
	if framePlan.Action != "" {
		parts = append(parts, framePlan.Action)
	}

	// 道具信息
	if len(framePlan.Props) > 0 {
		parts = append(parts, fmt.Sprintf("props: %s", strings.Join(framePlan.Props, ", ")))
	}

	// 添加质量标签
	parts = append(parts, "cinematic, high quality, detailed, professional lighting")

	return strings.Join(parts, ", ")
}

// cleanJSONResponse 清理JSON响应
func (s *Service) cleanJSONResponse(response string) string {
	response = strings.ReplaceAll(response, "```json", "")
	response = strings.ReplaceAll(response, "```", "")
	response = strings.TrimSpace(response)

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start != -1 && end != -1 && end > start {
		return response[start : end+1]
	}

	return response
}

// GenerateStoryboardRequest 生成分镜请求
type GenerateStoryboardRequest struct {
	ProjectID     string                 `json:"project_id"`
	ScriptID      uuid.UUID              `json:"script_id"`
	Script        string                 `json:"script"`
	Title         string                 `json:"title"`
	ScriptContent string                 `json:"script_content"`
	Characters    []*entity.Character    `json:"characters"`
	Scenes        []*entity.Scene        `json:"scenes"`
	Style         string                 `json:"style"`
}

// StoryboardPlan 分镜计划
type StoryboardPlan struct {
	Description   string       `json:"description"`
	TotalDuration int          `json:"total_duration"`
	Frames        []*FramePlan `json:"frames"`
}

// FramePlan 分镜帧计划
type FramePlan struct {
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Dialogue         string   `json:"dialogue"`
	Action           string   `json:"action"`
	CameraAngle      string   `json:"camera_angle"`
	CameraMove       string   `json:"camera_move"`
	ShotType         string   `json:"shot_type"`
	Duration         float64  `json:"duration"`
	Transition       string   `json:"transition"`
	Characters       []string `json:"characters"`
	Props            []string `json:"props"`
	SceneDescription string   `json:"scene_description"`
}
