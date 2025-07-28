package editing

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/postgres"
	"comic_video/internal/service/ai"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service 视频编辑服务
type Service struct {
	db        *gorm.DB
	videoRepo *postgres.VideoRepository
	aiService *ai.Service
}

// NewService 创建视频编辑服务
func NewService(db *gorm.DB, aiService *ai.Service) *Service {
	return &Service{
		db:        db,
		videoRepo: postgres.NewVideoRepository(db),
		aiService: aiService,
	}
}

// ComposeVideo 合成视频
func (s *Service) ComposeVideo(ctx context.Context, req *ComposeVideoRequest) (*entity.Video, error) {
	log.Printf("[EditingService] 开始合成视频: project=%s", req.ProjectID)

	// 1. 准备视频素材
	videoAssets, err := s.prepareVideoAssets(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("准备视频素材失败: %w", err)
	}

	// 2. 创建视频记录
	video := &entity.Video{
		ProjectID:   &req.ProjectID,
		UserID:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "processing",
		Duration:    float64(req.EstimatedDuration),
		Resolution:  req.Resolution,
		FrameRate:   req.FrameRate,
		Format:      "mp4",
	}

	if err := s.videoRepo.Create(ctx, video); err != nil {
		return nil, fmt.Errorf("创建视频记录失败: %w", err)
	}

	// 3. 异步执行视频合成
	go s.executeVideoComposition(context.Background(), video, videoAssets, req)

	log.Printf("[EditingService] 视频合成任务已启动: video=%s", video.ID)
	return video, nil
}

// executeVideoComposition 执行视频合成
func (s *Service) executeVideoComposition(ctx context.Context, video *entity.Video, assets *VideoAssets, req *ComposeVideoRequest) {
	log.Printf("[EditingService] 开始执行视频合成: video=%s", video.ID)

	// 1. 生成视频序列
	videoSequence, err := s.generateVideoSequence(ctx, assets, req)
	if err != nil {
		s.markVideoFailed(ctx, video, fmt.Errorf("生成视频序列失败: %w", err))
		return
	}

	// 2. 添加音频轨道
	if err := s.addAudioTracks(ctx, videoSequence, assets); err != nil {
		s.markVideoFailed(ctx, video, fmt.Errorf("添加音频轨道失败: %w", err))
		return
	}

	// 3. 添加特效和转场
	if err := s.addEffectsAndTransitions(ctx, videoSequence, req.Effects); err != nil {
		s.markVideoFailed(ctx, video, fmt.Errorf("添加特效失败: %w", err))
		return
	}

	// 4. 渲染最终视频
	outputURL, err := s.renderFinalVideo(ctx, videoSequence, req)
	if err != nil {
		s.markVideoFailed(ctx, video, fmt.Errorf("渲染视频失败: %w", err))
		return
	}

	// 5. 更新视频记录
	video.FileURL = outputURL
	video.Status = "completed"
	video.ProcessedAt = &time.Time{}
	*video.ProcessedAt = time.Now()

	if err := s.videoRepo.Update(ctx, video); err != nil {
		log.Printf("[EditingService] 更新视频记录失败: %v", err)
	}

	log.Printf("[EditingService] 视频合成完成: video=%s, url=%s", video.ID, outputURL)
}

// prepareVideoAssets 准备视频素材
func (s *Service) prepareVideoAssets(ctx context.Context, req *ComposeVideoRequest) (*VideoAssets, error) {
	assets := &VideoAssets{
		StoryboardFrames: make([]*FrameAsset, 0),
		VoiceAudios:      make([]*AudioAsset, 0),
		BackgroundMusic:  make([]*AudioAsset, 0),
	}

	// 收集分镜帧素材
	for _, frameReq := range req.StoryboardFrames {
		frameAsset := &FrameAsset{
			ID:       frameReq.FrameID,
			ImageURL: frameReq.ImageURL,
			Duration: frameReq.Duration,
			Order:    frameReq.Order,
		}
		assets.StoryboardFrames = append(assets.StoryboardFrames, frameAsset)
	}

	// 收集语音素材
	for _, voiceReq := range req.VoiceAudios {
		audioAsset := &AudioAsset{
			ID:       voiceReq.AudioID,
			FileURL:  voiceReq.FileURL,
			Duration: voiceReq.Duration,
			Type:     "voice",
			Order:    voiceReq.Order,
		}
		assets.VoiceAudios = append(assets.VoiceAudios, audioAsset)
	}

	// 收集背景音乐素材
	for _, musicReq := range req.BackgroundMusic {
		audioAsset := &AudioAsset{
			ID:       musicReq.AudioID,
			FileURL:  musicReq.FileURL,
			Duration: musicReq.Duration,
			Type:     "music",
			Volume:   musicReq.Volume,
		}
		assets.BackgroundMusic = append(assets.BackgroundMusic, audioAsset)
	}

	return assets, nil
}

// generateVideoSequence 生成视频序列
func (s *Service) generateVideoSequence(ctx context.Context, assets *VideoAssets, req *ComposeVideoRequest) (*VideoSequence, error) {
	sequence := &VideoSequence{
		Clips:      make([]*VideoClip, 0),
		AudioTracks: make([]*AudioTrack, 0),
		Effects:    make([]*Effect, 0),
	}

	// 为每个分镜帧创建视频片段
	for _, frame := range assets.StoryboardFrames {
		clip := &VideoClip{
			ID:        frame.ID,
			ImageURL:  frame.ImageURL,
			Duration:  frame.Duration,
			StartTime: s.calculateStartTime(sequence.Clips),
			Order:     frame.Order,
		}
		sequence.Clips = append(sequence.Clips, clip)
	}

	return sequence, nil
}

// addAudioTracks 添加音频轨道
func (s *Service) addAudioTracks(ctx context.Context, sequence *VideoSequence, assets *VideoAssets) error {
	// 添加语音轨道
	for _, voice := range assets.VoiceAudios {
		track := &AudioTrack{
			ID:        voice.ID,
			FileURL:   voice.FileURL,
			Type:      "voice",
			StartTime: s.calculateAudioStartTime(voice.Order, assets.VoiceAudios),
			Duration:  voice.Duration,
			Volume:    1.0,
		}
		sequence.AudioTracks = append(sequence.AudioTracks, track)
	}

	// 添加背景音乐轨道
	for _, music := range assets.BackgroundMusic {
		track := &AudioTrack{
			ID:        music.ID,
			FileURL:   music.FileURL,
			Type:      "music",
			StartTime: 0, // 背景音乐从头开始
			Duration:  music.Duration,
			Volume:    music.Volume,
		}
		sequence.AudioTracks = append(sequence.AudioTracks, track)
	}

	return nil
}

// addEffectsAndTransitions 添加特效和转场
func (s *Service) addEffectsAndTransitions(ctx context.Context, sequence *VideoSequence, effects []*EffectRequest) error {
	for _, effectReq := range effects {
		effect := &Effect{
			Type:      effectReq.Type,
			StartTime: effectReq.StartTime,
			Duration:  effectReq.Duration,
			Parameters: effectReq.Parameters,
		}
		sequence.Effects = append(sequence.Effects, effect)
	}

	return nil
}

// renderFinalVideo 渲染最终视频
func (s *Service) renderFinalVideo(ctx context.Context, sequence *VideoSequence, req *ComposeVideoRequest) (string, error) {
	// 这里调用实际的视频渲染服务
	// 可能是FFmpeg、云端渲染服务等
	
	outputFilename := fmt.Sprintf("video_%s_%d.mp4", req.ProjectID.String()[:8], time.Now().Unix())
	outputPath := filepath.Join("output", outputFilename)

	// 模拟渲染过程
	log.Printf("[EditingService] 开始渲染视频: clips=%d, audio_tracks=%d", len(sequence.Clips), len(sequence.AudioTracks))
	
	// 实际实现中这里会调用视频渲染引擎
	// 例如：FFmpeg命令行工具或者云端API
	
	return outputPath, nil
}

// calculateStartTime 计算开始时间
func (s *Service) calculateStartTime(clips []*VideoClip) float64 {
	totalTime := 0.0
	for _, clip := range clips {
		totalTime += clip.Duration
	}
	return totalTime
}

// calculateAudioStartTime 计算音频开始时间
func (s *Service) calculateAudioStartTime(order int, audios []*AudioAsset) float64 {
	startTime := 0.0
	for i, audio := range audios {
		if i >= order {
			break
		}
		startTime += audio.Duration
	}
	return startTime
}

// markVideoFailed 标记视频失败
func (s *Service) markVideoFailed(ctx context.Context, video *entity.Video, err error) {
	video.Status = "failed"
	video.Error = err.Error()
	if updateErr := s.videoRepo.Update(ctx, video); updateErr != nil {
		log.Printf("[EditingService] 更新视频失败状态失败: %v", updateErr)
	}
	log.Printf("[EditingService] 视频合成失败: video=%s, error=%v", video.ID, err)
}

// ComposeVideoRequest 合成视频请求
type ComposeVideoRequest struct {
	ProjectID         uuid.UUID           `json:"project_id"`
	UserID            uuid.UUID           `json:"user_id"`
	Title             string              `json:"title"`
	Description       string              `json:"description"`
	Resolution        string              `json:"resolution"`
	FrameRate         int                 `json:"frame_rate"`
	EstimatedDuration int                 `json:"estimated_duration"`
	StoryboardFrames  []*FrameRequest     `json:"storyboard_frames"`
	VoiceAudios       []*AudioRequest     `json:"voice_audios"`
	BackgroundMusic   []*MusicRequest     `json:"background_music"`
	Effects           []*EffectRequest    `json:"effects"`
}

// FrameRequest 分镜帧请求
type FrameRequest struct {
	FrameID  uuid.UUID `json:"frame_id"`
	ImageURL string    `json:"image_url"`
	Duration float64   `json:"duration"`
	Order    int       `json:"order"`
}

// AudioRequest 音频请求
type AudioRequest struct {
	AudioID  uuid.UUID `json:"audio_id"`
	FileURL  string    `json:"file_url"`
	Duration float64   `json:"duration"`
	Order    int       `json:"order"`
}

// MusicRequest 音乐请求
type MusicRequest struct {
	AudioID  uuid.UUID `json:"audio_id"`
	FileURL  string    `json:"file_url"`
	Duration float64   `json:"duration"`
	Volume   float64   `json:"volume"`
}

// EffectRequest 特效请求
type EffectRequest struct {
	Type       string                 `json:"type"`
	StartTime  float64                `json:"start_time"`
	Duration   float64                `json:"duration"`
	Parameters map[string]interface{} `json:"parameters"`
}

// VideoAssets 视频素材
type VideoAssets struct {
	StoryboardFrames []*FrameAsset  `json:"storyboard_frames"`
	VoiceAudios      []*AudioAsset  `json:"voice_audios"`
	BackgroundMusic  []*AudioAsset  `json:"background_music"`
}

// FrameAsset 帧素材
type FrameAsset struct {
	ID       uuid.UUID `json:"id"`
	ImageURL string    `json:"image_url"`
	Duration float64   `json:"duration"`
	Order    int       `json:"order"`
}

// AudioAsset 音频素材
type AudioAsset struct {
	ID       uuid.UUID `json:"id"`
	FileURL  string    `json:"file_url"`
	Duration float64   `json:"duration"`
	Type     string    `json:"type"`
	Volume   float64   `json:"volume"`
	Order    int       `json:"order"`
}

// VideoSequence 视频序列
type VideoSequence struct {
	Clips       []*VideoClip   `json:"clips"`
	AudioTracks []*AudioTrack  `json:"audio_tracks"`
	Effects     []*Effect      `json:"effects"`
}

// VideoClip 视频片段
type VideoClip struct {
	ID        uuid.UUID `json:"id"`
	ImageURL  string    `json:"image_url"`
	Duration  float64   `json:"duration"`
	StartTime float64   `json:"start_time"`
	Order     int       `json:"order"`
}

// AudioTrack 音频轨道
type AudioTrack struct {
	ID        uuid.UUID `json:"id"`
	FileURL   string    `json:"file_url"`
	Type      string    `json:"type"`
	StartTime float64   `json:"start_time"`
	Duration  float64   `json:"duration"`
	Volume    float64   `json:"volume"`
}

// Effect 特效
type Effect struct {
	Type       string                 `json:"type"`
	StartTime  float64                `json:"start_time"`
	Duration   float64                `json:"duration"`
	Parameters map[string]interface{} `json:"parameters"`
}
