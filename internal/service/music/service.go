package music

import (
	"context"
	"fmt"
	"log"

	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/postgres"
	"comic_video/internal/service/ai"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service 音乐服务
type Service struct {
	db              *gorm.DB
	musicGenRepo    *postgres.MusicGenerationRepository
	audioRepo       *postgres.AudioRepository
	aiService       *ai.Service
}

// NewService 创建音乐服务
func NewService(db *gorm.DB, aiService *ai.Service) *Service {
	return &Service{
		db:           db,
		musicGenRepo: postgres.NewMusicGenerationRepository(db),
		audioRepo:    postgres.NewAudioRepository(db),
		aiService:    aiService,
	}
}

// GenerateBackgroundMusic 生成背景音乐
func (s *Service) GenerateBackgroundMusic(ctx context.Context, req *GenerateMusicRequest) (*entity.MusicGeneration, error) {
	log.Printf("[MusicService] 开始生成背景音乐: project=%s", req.ProjectID)

	// 创建音乐生成记录
	musicGen := &entity.MusicGeneration{
		ProjectID: req.ProjectID,
		Prompt:    req.Prompt,
		Style:     req.Style,
		Mood:      req.Mood,
		Tempo:     req.Tempo,
		Duration:  req.Duration,
		Status:    "pending",
	}

	if err := s.musicGenRepo.Create(ctx, musicGen); err != nil {
		return nil, fmt.Errorf("创建音乐生成记录失败: %w", err)
	}

	// 更新状态为处理中
	musicGen.Status = "processing"
	s.musicGenRepo.Update(ctx, musicGen)

	// 调用AI音乐生成服务
	audioURL, duration, err := s.aiService.GenerateMusic(ctx, &ai.MusicGenerationRequest{
		Prompt:   req.Prompt,
		Style:    req.Style,
		Mood:     req.Mood,
		Tempo:    req.Tempo,
		Duration: req.Duration,
	})

	if err != nil {
		// 更新错误状态
		musicGen.Status = "failed"
		musicGen.Error = err.Error()
		s.musicGenRepo.Update(ctx, musicGen)
		return nil, fmt.Errorf("AI音乐生成失败: %w", err)
	}

	// 创建音频记录
	audio := &entity.Audio{
		ProjectID:   req.ProjectID,
		Name:        s.generateMusicName(req),
		Type:        entity.AudioTypeMusic,
		Description: fmt.Sprintf("Generated background music: %s", req.Style),
		FileURL:     audioURL,
		Duration:    duration,
		Format:      "wav",
	}

	if err := s.audioRepo.Create(ctx, audio); err != nil {
		log.Printf("[MusicService] 保存音频记录失败: %v", err)
	} else {
		musicGen.AudioID = &audio.ID
	}

	// 更新完成状态
	musicGen.Status = "completed"
	if err := s.musicGenRepo.Update(ctx, musicGen); err != nil {
		log.Printf("[MusicService] 更新音乐生成状态失败: %v", err)
	}

	log.Printf("[MusicService] 背景音乐生成完成: music=%s", musicGen.ID)
	return musicGen, nil
}

// GenerateThemeMusic 生成主题音乐
func (s *Service) GenerateThemeMusic(ctx context.Context, req *GenerateThemeMusicRequest) (*entity.MusicGeneration, error) {
	log.Printf("[MusicService] 开始生成主题音乐: project=%s", req.ProjectID)

	// 基于故事主题和情感生成音乐提示词
	prompt := s.buildThemeMusicPrompt(req)

	musicReq := &GenerateMusicRequest{
		ProjectID: req.ProjectID,
		Prompt:    prompt,
		Style:     req.Style,
		Mood:      req.Mood,
		Tempo:     req.Tempo,
		Duration:  req.Duration,
	}

	return s.GenerateBackgroundMusic(ctx, musicReq)
}

// GenerateSceneMusic 为特定场景生成音乐
func (s *Service) GenerateSceneMusic(ctx context.Context, req *GenerateSceneMusicRequest) ([]*entity.MusicGeneration, error) {
	log.Printf("[MusicService] 开始为场景生成音乐: project=%s, scenes=%d", req.ProjectID, len(req.SceneRequests))

	var musicGenerations []*entity.MusicGeneration

	for _, sceneReq := range req.SceneRequests {
		// 基于场景信息生成音乐
		prompt := s.buildSceneMusicPrompt(sceneReq)

		musicReq := &GenerateMusicRequest{
			ProjectID: req.ProjectID,
			Prompt:    prompt,
			Style:     sceneReq.Style,
			Mood:      sceneReq.Mood,
			Tempo:     sceneReq.Tempo,
			Duration:  sceneReq.Duration,
		}

		musicGen, err := s.GenerateBackgroundMusic(ctx, musicReq)
		if err != nil {
			log.Printf("[MusicService] 场景音乐生成失败: scene=%s, error: %v", sceneReq.SceneName, err)
			continue
		}

		musicGenerations = append(musicGenerations, musicGen)
	}

	log.Printf("[MusicService] 场景音乐生成完成: project=%s, count=%d", req.ProjectID, len(musicGenerations))
	return musicGenerations, nil
}

// GetMusicStyles 获取可用的音乐风格
func (s *Service) GetMusicStyles(ctx context.Context) ([]*MusicStyle, error) {
	styles := []*MusicStyle{
		{
			ID:          "orchestral",
			Name:        "管弦乐",
			Description: "宏大的管弦乐风格，适合史诗场景",
		},
		{
			ID:          "piano",
			Name:        "钢琴",
			Description: "优美的钢琴音乐，适合情感场景",
		},
		{
			ID:          "electronic",
			Name:        "电子音乐",
			Description: "现代电子音乐，适合科技场景",
		},
		{
			ID:          "folk",
			Name:        "民谣",
			Description: "温暖的民谣风格，适合日常场景",
		},
		{
			ID:          "ambient",
			Name:        "氛围音乐",
			Description: "空灵的氛围音乐，适合背景音乐",
		},
		{
			ID:          "cinematic",
			Name:        "电影配乐",
			Description: "专业的电影配乐风格",
		},
	}

	return styles, nil
}

// buildThemeMusicPrompt 构建主题音乐提示词
func (s *Service) buildThemeMusicPrompt(req *GenerateThemeMusicRequest) string {
	prompt := fmt.Sprintf("Theme music for story: %s", req.StoryTheme)
	
	if req.EmotionalTone != "" {
		prompt += fmt.Sprintf(", emotional tone: %s", req.EmotionalTone)
	}
	
	if req.Genre != "" {
		prompt += fmt.Sprintf(", genre: %s", req.Genre)
	}

	return prompt
}

// buildSceneMusicPrompt 构建场景音乐提示词
func (s *Service) buildSceneMusicPrompt(req *SceneMusicRequest) string {
	prompt := fmt.Sprintf("Scene music for: %s", req.SceneName)
	
	if req.SceneDescription != "" {
		prompt += fmt.Sprintf(", scene: %s", req.SceneDescription)
	}
	
	if req.Atmosphere != "" {
		prompt += fmt.Sprintf(", atmosphere: %s", req.Atmosphere)
	}

	return prompt
}

// generateMusicName 生成音乐名称
func (s *Service) generateMusicName(req *GenerateMusicRequest) string {
	if req.Style != "" {
		return fmt.Sprintf("bgm_%s_%s", req.Style, req.Mood)
	}
	return fmt.Sprintf("bgm_%s", req.Mood)
}

// GenerateMusicRequest 生成音乐请求
type GenerateMusicRequest struct {
	ProjectID uuid.UUID `json:"project_id"`
	Prompt    string    `json:"prompt"`
	Style     string    `json:"style"`
	Mood      string    `json:"mood"`
	Tempo     string    `json:"tempo"`
	Duration  int       `json:"duration"`
}

// GenerateThemeMusicRequest 生成主题音乐请求
type GenerateThemeMusicRequest struct {
	ProjectID     uuid.UUID `json:"project_id"`
	StoryTheme    string    `json:"story_theme"`
	EmotionalTone string    `json:"emotional_tone"`
	Genre         string    `json:"genre"`
	Style         string    `json:"style"`
	Mood          string    `json:"mood"`
	Tempo         string    `json:"tempo"`
	Duration      int       `json:"duration"`
}

// GenerateSceneMusicRequest 生成场景音乐请求
type GenerateSceneMusicRequest struct {
	ProjectID     uuid.UUID            `json:"project_id"`
	SceneRequests []*SceneMusicRequest `json:"scene_requests"`
}

// SceneMusicRequest 场景音乐请求
type SceneMusicRequest struct {
	SceneName        string `json:"scene_name"`
	SceneDescription string `json:"scene_description"`
	Atmosphere       string `json:"atmosphere"`
	Style            string `json:"style"`
	Mood             string `json:"mood"`
	Tempo            string `json:"tempo"`
	Duration         int    `json:"duration"`
}

// MusicStyle 音乐风格
type MusicStyle struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
