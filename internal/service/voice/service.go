package voice

import (
	"context"
	"fmt"
	"log"
	"strings"

	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/postgres"
	"comic_video/internal/service/ai"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service 语音服务
type Service struct {
	db               *gorm.DB
	voiceGenRepo     *postgres.VoiceGenerationRepository
	audioRepo        *postgres.AudioRepository
	aiService        *ai.Service
}

// NewService 创建语音服务
func NewService(db *gorm.DB, aiService *ai.Service) *Service {
	return &Service{
		db:           db,
		voiceGenRepo: postgres.NewVoiceGenerationRepository(db),
		audioRepo:    postgres.NewAudioRepository(db),
		aiService:    aiService,
	}
}

// GenerateVoices 生成语音
func (s *Service) GenerateVoices(ctx context.Context, req *GenerateVoicesRequest) ([]*entity.VoiceGeneration, error) {
	log.Printf("[VoiceService] 开始生成语音: project=%s", req.ProjectID)

	var voiceGenerations []*entity.VoiceGeneration

	// 为每个语音片段生成音频
	for _, voiceReq := range req.VoiceRequests {
		voiceGen, err := s.generateSingleVoice(ctx, req.ProjectID, voiceReq)
		if err != nil {
			log.Printf("[VoiceService] 生成语音失败: text=%s, error: %v", voiceReq.Text[:50], err)
			continue
		}

		voiceGenerations = append(voiceGenerations, voiceGen)
	}

	log.Printf("[VoiceService] 语音生成完成: project=%s, count=%d", req.ProjectID, len(voiceGenerations))
	return voiceGenerations, nil
}

// generateSingleVoice 生成单个语音
func (s *Service) generateSingleVoice(ctx context.Context, projectID uuid.UUID, voiceReq *VoiceRequest) (*entity.VoiceGeneration, error) {
	// 创建语音生成记录
	voiceGen := &entity.VoiceGeneration{
		ProjectID:   projectID,
		CharacterID: voiceReq.CharacterID,
		Text:        voiceReq.Text,
		VoiceModel:  voiceReq.VoiceModel,
		Language:    voiceReq.Language,
		Speed:       voiceReq.Speed,
		Pitch:       voiceReq.Pitch,
		Volume:      voiceReq.Volume,
		Emotion:     voiceReq.Emotion,
		Status:      "pending",
	}

	if err := s.voiceGenRepo.Create(ctx, voiceGen); err != nil {
		return nil, fmt.Errorf("创建语音生成记录失败: %w", err)
	}

	// 更新状态为处理中
	voiceGen.Status = "processing"
	s.voiceGenRepo.Update(ctx, voiceGen)

	// 调用AI语音生成服务
	audioURL, duration, err := s.aiService.GenerateVoice(ctx, &ai.VoiceGenerationRequest{
		Text:       voiceReq.Text,
		VoiceModel: voiceReq.VoiceModel,
		Language:   voiceReq.Language,
		Speed:      voiceReq.Speed,
		Pitch:      voiceReq.Pitch,
		Volume:     voiceReq.Volume,
		Emotion:    voiceReq.Emotion,
	})

	if err != nil {
		// 更新错误状态
		voiceGen.Status = "failed"
		voiceGen.Error = err.Error()
		s.voiceGenRepo.Update(ctx, voiceGen)
		return nil, fmt.Errorf("AI语音生成失败: %w", err)
	}

	// 创建音频记录
	audio := &entity.Audio{
		ProjectID:   projectID,
		Name:        s.generateAudioName(voiceReq),
		Type:        entity.AudioTypeVoice,
		Description: fmt.Sprintf("Generated voice for: %s", s.truncateText(voiceReq.Text, 50)),
		FileURL:     audioURL,
		Duration:    duration,
		Format:      "wav",
	}

	if err := s.audioRepo.Create(ctx, audio); err != nil {
		log.Printf("[VoiceService] 保存音频记录失败: %v", err)
	} else {
		voiceGen.AudioID = &audio.ID
	}

	// 更新完成状态
	voiceGen.Status = "completed"
	if err := s.voiceGenRepo.Update(ctx, voiceGen); err != nil {
		log.Printf("[VoiceService] 更新语音生成状态失败: %v", err)
	}

	return voiceGen, nil
}

// GenerateNarration 生成旁白
func (s *Service) GenerateNarration(ctx context.Context, req *GenerateNarrationRequest) (*entity.VoiceGeneration, error) {
	log.Printf("[VoiceService] 开始生成旁白: project=%s", req.ProjectID)

	voiceReq := &VoiceRequest{
		Text:       req.NarrationText,
		VoiceModel: req.NarratorVoice,
		Language:   "zh-CN",
		Speed:      1.0,
		Pitch:      1.0,
		Volume:     1.0,
		Emotion:    "neutral",
	}

	return s.generateSingleVoice(ctx, req.ProjectID, voiceReq)
}

// AssignVoicesToCharacters 为角色分配语音模型
func (s *Service) AssignVoicesToCharacters(ctx context.Context, req *AssignVoicesRequest) error {
	log.Printf("[VoiceService] 开始为角色分配语音: project=%s", req.ProjectID)

	for _, assignment := range req.VoiceAssignments {
		// 这里可以保存角色语音分配信息到数据库
		// 或者更新角色实体的语音模型字段
		log.Printf("[VoiceService] 角色 %s 分配语音模型: %s", assignment.CharacterID, assignment.VoiceModel)
	}

	return nil
}

// GetAvailableVoiceModels 获取可用的语音模型
func (s *Service) GetAvailableVoiceModels(ctx context.Context) ([]*VoiceModel, error) {
	// 这里返回预定义的语音模型列表
	// 实际实现中可能需要从AI服务获取
	models := []*VoiceModel{
		{
			ID:          "female_young_1",
			Name:        "年轻女声1",
			Gender:      "female",
			Age:         "young",
			Language:    "zh-CN",
			Description: "清甜的年轻女性声音",
		},
		{
			ID:          "female_young_2",
			Name:        "年轻女声2",
			Gender:      "female",
			Age:         "young",
			Language:    "zh-CN",
			Description: "活泼的年轻女性声音",
		},
		{
			ID:          "male_young_1",
			Name:        "年轻男声1",
			Gender:      "male",
			Age:         "young",
			Language:    "zh-CN",
			Description: "清朗的年轻男性声音",
		},
		{
			ID:          "male_mature_1",
			Name:        "成熟男声1",
			Gender:      "male",
			Age:         "mature",
			Language:    "zh-CN",
			Description: "沉稳的成熟男性声音",
		},
		{
			ID:          "narrator_1",
			Name:        "旁白声1",
			Gender:      "neutral",
			Age:         "mature",
			Language:    "zh-CN",
			Description: "专业的旁白声音",
		},
	}

	return models, nil
}

// generateAudioName 生成音频名称
func (s *Service) generateAudioName(voiceReq *VoiceRequest) string {
	textPreview := s.truncateText(voiceReq.Text, 20)
	return fmt.Sprintf("voice_%s_%s", voiceReq.VoiceModel, textPreview)
}

// truncateText 截断文本
func (s *Service) truncateText(text string, maxLen int) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.TrimSpace(text)
	
	if len(text) <= maxLen {
		return text
	}
	
	return text[:maxLen] + "..."
}

// GenerateVoicesRequest 生成语音请求
type GenerateVoicesRequest struct {
	ProjectID     uuid.UUID       `json:"project_id"`
	VoiceRequests []*VoiceRequest `json:"voice_requests"`
}

// VoiceRequest 单个语音请求
type VoiceRequest struct {
	CharacterID *uuid.UUID `json:"character_id,omitempty"`
	Text        string     `json:"text"`
	VoiceModel  string     `json:"voice_model"`
	Language    string     `json:"language"`
	Speed       float64    `json:"speed"`
	Pitch       float64    `json:"pitch"`
	Volume      float64    `json:"volume"`
	Emotion     string     `json:"emotion"`
}

// GenerateNarrationRequest 生成旁白请求
type GenerateNarrationRequest struct {
	ProjectID     uuid.UUID `json:"project_id"`
	NarrationText string    `json:"narration_text"`
	NarratorVoice string    `json:"narrator_voice"`
}

// AssignVoicesRequest 分配语音请求
type AssignVoicesRequest struct {
	ProjectID        uuid.UUID           `json:"project_id"`
	VoiceAssignments []*VoiceAssignment  `json:"voice_assignments"`
}

// VoiceAssignment 语音分配
type VoiceAssignment struct {
	CharacterID uuid.UUID `json:"character_id"`
	VoiceModel  string    `json:"voice_model"`
}

// VoiceModel 语音模型
type VoiceModel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Gender      string `json:"gender"`
	Age         string `json:"age"`
	Language    string `json:"language"`
	Description string `json:"description"`
}
