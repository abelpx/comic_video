package audio

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"comic_video/internal/domain/entity"
	"comic_video/internal/service/ai"
)

// EnhancedAudioProcessor 增强音频处理器
type EnhancedAudioProcessor struct {
	aiService    *ai.Service
	ffmpegPath   string
	tempDir      string
	musicLibrary *MusicLibrary
	sfxLibrary   *SFXLibrary
}

// MusicLibrary 音乐库
type MusicLibrary struct {
	Tracks map[string][]MusicTrack `json:"tracks"` // 按情绪分类
}

// MusicTrack 音乐轨道
type MusicTrack struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Path     string  `json:"path"`
	Duration float64 `json:"duration"`
	Mood     string  `json:"mood"`
	Genre    string  `json:"genre"`
	BPM      int     `json:"bpm"`
	Key      string  `json:"key"`
}

// SFXLibrary 音效库
type SFXLibrary struct {
	Effects map[string][]SoundEffect `json:"effects"` // 按类型分类
}

// SoundEffect 音效
type SoundEffect struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Path     string  `json:"path"`
	Duration float64 `json:"duration"`
	Category string  `json:"category"`
	Tags     []string `json:"tags"`
}

// AudioProcessingRequest 音频处理请求
type AudioProcessingRequest struct {
	ProjectID     string                 `json:"project_id"`
	Script        *entity.Script         `json:"script"`
	VoiceFiles    []string               `json:"voice_files"`
	Scenes        []*entity.Scene        `json:"scenes"`
	Duration      float64                `json:"duration"`
	Style         string                 `json:"style"`
	Platform      string                 `json:"platform"`
	Options       map[string]interface{} `json:"options"`
}

// AudioProcessingResult 音频处理结果
type AudioProcessingResult struct {
	MasterAudioPath string                 `json:"master_audio_path"`
	VoiceTrackPath  string                 `json:"voice_track_path"`
	MusicTrackPath  string                 `json:"music_track_path"`
	SFXTrackPath    string                 `json:"sfx_track_path"`
	Duration        float64                `json:"duration"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// NewEnhancedAudioProcessor 创建增强音频处理器
func NewEnhancedAudioProcessor(aiService *ai.Service, ffmpegPath, tempDir string) *EnhancedAudioProcessor {
	processor := &EnhancedAudioProcessor{
		aiService:  aiService,
		ffmpegPath: ffmpegPath,
		tempDir:    tempDir,
		musicLibrary: &MusicLibrary{
			Tracks: make(map[string][]MusicTrack),
		},
		sfxLibrary: &SFXLibrary{
			Effects: make(map[string][]SoundEffect),
		},
	}
	
	// 初始化音频库
	processor.initializeAudioLibraries()
	
	return processor
}

// ProcessAudio 处理音频
func (p *EnhancedAudioProcessor) ProcessAudio(ctx context.Context, req *AudioProcessingRequest) (*AudioProcessingResult, error) {
	log.Printf("[EnhancedAudioProcessor] 开始处理音频: project_id=%s", req.ProjectID)
	
	// 1. 创建工作目录
	workDir := filepath.Join(p.tempDir, fmt.Sprintf("audio_%s_%d", req.ProjectID, time.Now().Unix()))
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("创建工作目录失败: %w", err)
	}
	defer os.RemoveAll(workDir)
	
	// 2. 分析音频需求
	audioAnalysis, err := p.analyzeAudioRequirements(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("分析音频需求失败: %w", err)
	}
	
	// 3. 处理语音轨道
	voiceTrack, err := p.processVoiceTrack(ctx, req.VoiceFiles, audioAnalysis, workDir)
	if err != nil {
		return nil, fmt.Errorf("处理语音轨道失败: %w", err)
	}
	
	// 4. 生成背景音乐
	musicTrack, err := p.generateBackgroundMusic(ctx, audioAnalysis, req.Duration, workDir)
	if err != nil {
		return nil, fmt.Errorf("生成背景音乐失败: %w", err)
	}
	
	// 5. 添加音效
	sfxTrack, err := p.addSoundEffects(ctx, audioAnalysis, req.Duration, workDir)
	if err != nil {
		return nil, fmt.Errorf("添加音效失败: %w", err)
	}
	
	// 6. 混合所有音轨
	masterAudio, err := p.mixAudioTracks(ctx, voiceTrack, musicTrack, sfxTrack, req, workDir)
	if err != nil {
		return nil, fmt.Errorf("混合音轨失败: %w", err)
	}
	
	// 7. 最终优化
	finalAudio, err := p.finalizeAudio(ctx, masterAudio, req, workDir)
	if err != nil {
		return nil, fmt.Errorf("最终优化失败: %w", err)
	}
	
	result := &AudioProcessingResult{
		MasterAudioPath: finalAudio,
		VoiceTrackPath:  voiceTrack,
		MusicTrackPath:  musicTrack,
		SFXTrackPath:    sfxTrack,
		Duration:        req.Duration,
		Metadata: map[string]interface{}{
			"style":    req.Style,
			"platform": req.Platform,
			"scenes":   len(req.Scenes),
			"processing_time": time.Now().Format(time.RFC3339),
		},
	}
	
	log.Printf("[EnhancedAudioProcessor] 音频处理完成: %s", finalAudio)
	return result, nil
}

// AudioAnalysis 音频分析结果
type AudioAnalysis struct {
	Moods          []MoodSegment    `json:"moods"`
	MusicStyle     string           `json:"music_style"`
	SoundEffects   []SFXRequirement `json:"sound_effects"`
	VoiceSettings  VoiceSettings    `json:"voice_settings"`
	OverallTone    string           `json:"overall_tone"`
}

// MoodSegment 情绪片段
type MoodSegment struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Mood      string  `json:"mood"`
	Intensity float64 `json:"intensity"`
}

// SFXRequirement 音效需求
type SFXRequirement struct {
	Time     float64 `json:"time"`
	Type     string  `json:"type"`
	Duration float64 `json:"duration"`
	Volume   float64 `json:"volume"`
}

// VoiceSettings 语音设置
type VoiceSettings struct {
	EmotionCurve []EmotionPoint `json:"emotion_curve"`
	SpeedCurve   []SpeedPoint   `json:"speed_curve"`
	VolumeCurve  []VolumePoint  `json:"volume_curve"`
}

// EmotionPoint 情感点
type EmotionPoint struct {
	Time     float64 `json:"time"`
	Emotion  string  `json:"emotion"`
	Intensity float64 `json:"intensity"`
}

// SpeedPoint 语速点
type SpeedPoint struct {
	Time  float64 `json:"time"`
	Speed float64 `json:"speed"` // 1.0为正常速度
}

// VolumePoint 音量点
type VolumePoint struct {
	Time   float64 `json:"time"`
	Volume float64 `json:"volume"` // 0.0-1.0
}

// analyzeAudioRequirements 分析音频需求
func (p *EnhancedAudioProcessor) analyzeAudioRequirements(ctx context.Context, req *AudioProcessingRequest) (*AudioAnalysis, error) {
	prompt := fmt.Sprintf(`分析以下剧本的音频需求：

剧本内容：%s
风格：%s
平台：%s
时长：%.1f秒

请分析：
1. 不同时间段的情绪变化
2. 适合的背景音乐风格
3. 需要的音效类型和时机
4. 语音的情感表达要求

输出JSON格式：
{
  "moods": [
    {"start_time": 0.0, "end_time": 30.0, "mood": "peaceful", "intensity": 0.6},
    {"start_time": 30.0, "end_time": 60.0, "mood": "exciting", "intensity": 0.8}
  ],
  "music_style": "cinematic",
  "sound_effects": [
    {"time": 15.0, "type": "footsteps", "duration": 2.0, "volume": 0.3},
    {"time": 45.0, "type": "door_close", "duration": 1.0, "volume": 0.5}
  ],
  "voice_settings": {
    "emotion_curve": [
      {"time": 0.0, "emotion": "neutral", "intensity": 0.5},
      {"time": 30.0, "emotion": "excited", "intensity": 0.8}
    ],
    "speed_curve": [
      {"time": 0.0, "speed": 1.0},
      {"time": 30.0, "speed": 1.2}
    ],
    "volume_curve": [
      {"time": 0.0, "volume": 0.8},
      {"time": 30.0, "volume": 1.0}
    ]
  },
  "overall_tone": "dramatic"
}`, req.Script.Content, req.Style, req.Platform, req.Duration)

	response, err := p.aiService.GenerateText(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	var analysis AudioAnalysis
	cleanedResponse := p.cleanJSONResponse(response)
	if err := json.Unmarshal([]byte(cleanedResponse), &analysis); err != nil {
		// 如果解析失败，返回默认分析
		return p.createDefaultAudioAnalysis(req), nil
	}

	return &analysis, nil
}

// processVoiceTrack 处理语音轨道
func (p *EnhancedAudioProcessor) processVoiceTrack(ctx context.Context, voiceFiles []string, analysis *AudioAnalysis, workDir string) (string, error) {
	if len(voiceFiles) == 0 {
		return "", fmt.Errorf("没有语音文件")
	}
	
	outputPath := filepath.Join(workDir, "processed_voice.wav")
	
	// 如果只有一个文件，直接处理
	if len(voiceFiles) == 1 {
		return p.enhanceVoiceFile(ctx, voiceFiles[0], analysis.VoiceSettings, outputPath)
	}
	
	// 多个文件需要先合并
	mergedVoice := filepath.Join(workDir, "merged_voice.wav")
	if err := p.mergeVoiceFiles(ctx, voiceFiles, mergedVoice); err != nil {
		return "", err
	}
	
	return p.enhanceVoiceFile(ctx, mergedVoice, analysis.VoiceSettings, outputPath)
}

// enhanceVoiceFile 增强语音文件
func (p *EnhancedAudioProcessor) enhanceVoiceFile(ctx context.Context, inputPath string, settings VoiceSettings, outputPath string) (string, error) {
	// 构建音频增强滤镜
	filters := p.buildVoiceEnhancementFilters(settings)
	
	cmd := exec.CommandContext(ctx, p.ffmpegPath,
		"-i", inputPath,
		"-af", filters,
		"-ar", "48000",
		"-ac", "2",
		"-c:a", "pcm_s16le",
		"-y", outputPath)
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("增强语音失败: %w", err)
	}
	
	return outputPath, nil
}

// buildVoiceEnhancementFilters 构建语音增强滤镜
func (p *EnhancedAudioProcessor) buildVoiceEnhancementFilters(settings VoiceSettings) string {
	var filters []string
	
	// 基础音频清理
	filters = append(filters, "highpass=f=80") // 高通滤波器去除低频噪音
	filters = append(filters, "lowpass=f=8000") // 低通滤波器去除高频噪音
	
	// 动态范围压缩
	filters = append(filters, "compand=0.1|0.1:1|1:-90/-90|-70/-70|-30/-9|0/-3:6:0:0:0")
	
	// 音量标准化
	filters = append(filters, "loudnorm=I=-16:TP=-1.5:LRA=11")
	
	// 根据情感曲线调整音调（简化实现）
	if len(settings.EmotionCurve) > 0 {
		// 这里可以根据情感添加音调调整
		for _, emotion := range settings.EmotionCurve {
			switch emotion.Emotion {
			case "excited", "happy":
				filters = append(filters, "asetrate=48000*1.05,aresample=48000") // 轻微提高音调
			case "sad", "melancholy":
				filters = append(filters, "asetrate=48000*0.95,aresample=48000") // 轻微降低音调
			}
		}
	}
	
	return strings.Join(filters, ",")
}

// mergeVoiceFiles 合并语音文件
func (p *EnhancedAudioProcessor) mergeVoiceFiles(ctx context.Context, voiceFiles []string, outputPath string) error {
	// 创建文件列表
	listFile := filepath.Join(filepath.Dir(outputPath), "voice_list.txt")
	file, err := os.Create(listFile)
	if err != nil {
		return err
	}
	defer file.Close()
	
	for _, voiceFile := range voiceFiles {
		_, err := file.WriteString(fmt.Sprintf("file '%s'\n", voiceFile))
		if err != nil {
			return err
		}
	}
	
	// 使用FFmpeg合并
	cmd := exec.CommandContext(ctx, p.ffmpegPath,
		"-f", "concat",
		"-safe", "0",
		"-i", listFile,
		"-c", "copy",
		"-y", outputPath)
	
	return cmd.Run()
}

// generateBackgroundMusic 生成背景音乐
func (p *EnhancedAudioProcessor) generateBackgroundMusic(ctx context.Context, analysis *AudioAnalysis, duration float64, workDir string) (string, error) {
	outputPath := filepath.Join(workDir, "background_music.wav")
	
	// 根据分析结果选择合适的音乐
	musicTrack := p.selectMusicTrack(analysis.MusicStyle, analysis.OverallTone)
	if musicTrack == nil {
		// 如果没有合适的音乐，生成静音轨道
		return p.generateSilentTrack(ctx, duration, outputPath)
	}
	
	// 调整音乐长度和音量
	return p.adaptMusicTrack(ctx, musicTrack.Path, duration, analysis.Moods, outputPath)
}

// selectMusicTrack 选择音乐轨道
func (p *EnhancedAudioProcessor) selectMusicTrack(style, tone string) *MusicTrack {
	// 根据风格和基调选择音乐
	if tracks, exists := p.musicLibrary.Tracks[tone]; exists {
		for _, track := range tracks {
			if strings.Contains(strings.ToLower(track.Genre), strings.ToLower(style)) {
				return &track
			}
		}
		// 如果没有完全匹配的，返回第一个
		if len(tracks) > 0 {
			return &tracks[0]
		}
	}
	
	return nil
}

// adaptMusicTrack 调整音乐轨道
func (p *EnhancedAudioProcessor) adaptMusicTrack(ctx context.Context, musicPath string, targetDuration float64, moods []MoodSegment, outputPath string) (string, error) {
	// 构建音乐调整滤镜
	filters := p.buildMusicAdaptationFilters(moods, targetDuration)
	
	cmd := exec.CommandContext(ctx, p.ffmpegPath,
		"-i", musicPath,
		"-af", filters,
		"-t", strconv.FormatFloat(targetDuration, 'f', 1, 64),
		"-c:a", "pcm_s16le",
		"-y", outputPath)
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("调整音乐轨道失败: %w", err)
	}
	
	return outputPath, nil
}

// buildMusicAdaptationFilters 构建音乐调整滤镜
func (p *EnhancedAudioProcessor) buildMusicAdaptationFilters(moods []MoodSegment, duration float64) string {
	var filters []string
	
	// 基础音量调整（背景音乐应该比语音低）
	filters = append(filters, "volume=0.3")
	
	// 根据情绪调整音乐
	if len(moods) > 0 {
		// 简化实现：根据整体情绪调整
		overallIntensity := 0.0
		for _, mood := range moods {
			overallIntensity += mood.Intensity
		}
		overallIntensity /= float64(len(moods))
		
		// 根据强度调整音量
		volumeAdjust := 0.2 + (overallIntensity * 0.3) // 0.2-0.5之间
		filters = append(filters, fmt.Sprintf("volume=%.2f", volumeAdjust))
	}
	
	// 淡入淡出
	filters = append(filters, "afade=t=in:ss=0:d=2,afade=t=out:st=" + strconv.FormatFloat(duration-2, 'f', 1, 64) + ":d=2")
	
	return strings.Join(filters, ",")
}

// addSoundEffects 添加音效
func (p *EnhancedAudioProcessor) addSoundEffects(ctx context.Context, analysis *AudioAnalysis, duration float64, workDir string) (string, error) {
	outputPath := filepath.Join(workDir, "sound_effects.wav")

	if len(analysis.SoundEffects) == 0 {
		// 没有音效，生成静音轨道
		return p.generateSilentTrack(ctx, duration, outputPath)
	}

	// 创建音效轨道
	return p.createSFXTrack(ctx, analysis.SoundEffects, duration, outputPath)
}

// createSFXTrack 创建音效轨道
func (p *EnhancedAudioProcessor) createSFXTrack(ctx context.Context, sfxRequirements []SFXRequirement, duration float64, outputPath string) (string, error) {
	// 首先创建静音基础轨道
	baseTrack := filepath.Join(filepath.Dir(outputPath), "sfx_base.wav")
	baseTrack, err := p.generateSilentTrack(ctx, duration, baseTrack)
	if err != nil {
		return "", err
	}

	currentTrack := baseTrack

	// 逐个添加音效
	for i, sfx := range sfxRequirements {
		effect := p.selectSoundEffect(sfx.Type)
		if effect == nil {
			continue // 跳过找不到的音效
		}

		tempOutput := filepath.Join(filepath.Dir(outputPath), fmt.Sprintf("sfx_temp_%d.wav", i))
		if err := p.addSingleSFX(ctx, currentTrack, effect.Path, sfx, tempOutput); err != nil {
			log.Printf("[EnhancedAudioProcessor] 添加音效失败: %v", err)
			continue
		}

		currentTrack = tempOutput
	}

	// 重命名最终文件
	return currentTrack, os.Rename(currentTrack, outputPath)
}

// selectSoundEffect 选择音效
func (p *EnhancedAudioProcessor) selectSoundEffect(effectType string) *SoundEffect {
	if effects, exists := p.sfxLibrary.Effects[effectType]; exists && len(effects) > 0 {
		return &effects[0] // 返回第一个匹配的音效
	}
	return nil
}

// addSingleSFX 添加单个音效
func (p *EnhancedAudioProcessor) addSingleSFX(ctx context.Context, baseTrack, sfxPath string, sfx SFXRequirement, outputPath string) error {
	cmd := exec.CommandContext(ctx, p.ffmpegPath,
		"-i", baseTrack,
		"-i", sfxPath,
		"-filter_complex", fmt.Sprintf("[1:a]volume=%.2f,adelay=%d|%d[sfx];[0:a][sfx]amix=inputs=2:duration=longest",
			sfx.Volume, int(sfx.Time*1000), int(sfx.Time*1000)),
		"-c:a", "pcm_s16le",
		"-y", outputPath)

	return cmd.Run()
}

// generateSilentTrack 生成静音轨道
func (p *EnhancedAudioProcessor) generateSilentTrack(ctx context.Context, duration float64, outputPath string) (string, error) {
	cmd := exec.CommandContext(ctx, p.ffmpegPath,
		"-f", "lavfi",
		"-i", fmt.Sprintf("anullsrc=channel_layout=stereo:sample_rate=48000:duration=%.1f", duration),
		"-c:a", "pcm_s16le",
		"-y", outputPath)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("生成静音轨道失败: %w", err)
	}

	return outputPath, nil
}

// mixAudioTracks 混合音轨
func (p *EnhancedAudioProcessor) mixAudioTracks(ctx context.Context, voiceTrack, musicTrack, sfxTrack string, req *AudioProcessingRequest, workDir string) (string, error) {
	outputPath := filepath.Join(workDir, "mixed_audio.wav")

	// 构建混音命令
	cmd := exec.CommandContext(ctx, p.ffmpegPath,
		"-i", voiceTrack,
		"-i", musicTrack,
		"-i", sfxTrack,
		"-filter_complex", "[0:a][1:a][2:a]amix=inputs=3:duration=longest:weights=1.0 0.4 0.6",
		"-c:a", "pcm_s16le",
		"-y", outputPath)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("混合音轨失败: %w", err)
	}

	return outputPath, nil
}

// finalizeAudio 最终优化音频
func (p *EnhancedAudioProcessor) finalizeAudio(ctx context.Context, inputPath string, req *AudioProcessingRequest, workDir string) (string, error) {
	outputPath := filepath.Join(workDir, "final_audio.wav")

	// 根据平台优化音频
	filters := p.buildPlatformOptimizationFilters(req.Platform)

	cmd := exec.CommandContext(ctx, p.ffmpegPath,
		"-i", inputPath,
		"-af", filters,
		"-ar", "48000",
		"-ac", "2",
		"-c:a", "pcm_s16le",
		"-y", outputPath)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("最终优化失败: %w", err)
	}

	return outputPath, nil
}

// buildPlatformOptimizationFilters 构建平台优化滤镜
func (p *EnhancedAudioProcessor) buildPlatformOptimizationFilters(platform string) string {
	var filters []string

	switch strings.ToLower(platform) {
	case "douyin", "tiktok":
		// 抖音：增强低频，适合手机扬声器
		filters = append(filters, "bass=g=3")
		filters = append(filters, "loudnorm=I=-14:TP=-1:LRA=7") // 更高的响度
	case "bilibili":
		// B站：平衡的音频
		filters = append(filters, "loudnorm=I=-16:TP=-1.5:LRA=11")
	case "youtube":
		// YouTube：标准化
		filters = append(filters, "loudnorm=I=-16:TP=-1.5:LRA=11")
	default:
		// 通用优化
		filters = append(filters, "loudnorm=I=-16:TP=-1.5:LRA=11")
	}

	// 最终限制器防止削波
	filters = append(filters, "alimiter=level_in=1:level_out=0.9:limit=0.9")

	return strings.Join(filters, ",")
}

// 辅助方法
func (p *EnhancedAudioProcessor) createDefaultAudioAnalysis(req *AudioProcessingRequest) *AudioAnalysis {
	return &AudioAnalysis{
		Moods: []MoodSegment{
			{StartTime: 0.0, EndTime: req.Duration, Mood: "neutral", Intensity: 0.5},
		},
		MusicStyle: "ambient",
		SoundEffects: []SFXRequirement{},
		VoiceSettings: VoiceSettings{
			EmotionCurve: []EmotionPoint{
				{Time: 0.0, Emotion: "neutral", Intensity: 0.5},
			},
			SpeedCurve: []SpeedPoint{
				{Time: 0.0, Speed: 1.0},
			},
			VolumeCurve: []VolumePoint{
				{Time: 0.0, Volume: 0.8},
			},
		},
		OverallTone: "neutral",
	}
}

func (p *EnhancedAudioProcessor) cleanJSONResponse(response string) string {
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

func (p *EnhancedAudioProcessor) initializeAudioLibraries() {
	// 初始化音乐库
	p.musicLibrary.Tracks["peaceful"] = []MusicTrack{
		{ID: "peaceful_01", Name: "Gentle Breeze", Mood: "peaceful", Genre: "ambient", BPM: 60},
	}
	p.musicLibrary.Tracks["exciting"] = []MusicTrack{
		{ID: "exciting_01", Name: "Rising Action", Mood: "exciting", Genre: "cinematic", BPM: 120},
	}
	p.musicLibrary.Tracks["dramatic"] = []MusicTrack{
		{ID: "dramatic_01", Name: "Epic Moment", Mood: "dramatic", Genre: "orchestral", BPM: 90},
	}

	// 初始化音效库
	p.sfxLibrary.Effects["footsteps"] = []SoundEffect{
		{ID: "footsteps_01", Name: "Walking on Wood", Category: "movement", Tags: []string{"walk", "wood"}},
	}
	p.sfxLibrary.Effects["door"] = []SoundEffect{
		{ID: "door_01", Name: "Door Close", Category: "interaction", Tags: []string{"door", "close"}},
	}
	p.sfxLibrary.Effects["nature"] = []SoundEffect{
		{ID: "nature_01", Name: "Wind in Trees", Category: "ambient", Tags: []string{"wind", "trees"}},
	}
}
