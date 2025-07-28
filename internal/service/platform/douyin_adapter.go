package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DouyinAdapter 抖音平台适配器
type DouyinAdapter struct {
	ffmpegPath string
	tempDir    string
	outputDir  string
	config     *DouyinConfig
}

// DouyinConfig 抖音配置
type DouyinConfig struct {
	Resolution    string  `json:"resolution"`     // 1080x1920
	AspectRatio   string  `json:"aspect_ratio"`   // 9:16
	MaxDuration   int     `json:"max_duration"`   // 1800秒 (30分钟)
	MinDuration   int     `json:"min_duration"`   // 60秒 (1分钟)
	FrameRate     int     `json:"frame_rate"`     // 30fps
	Bitrate       string  `json:"bitrate"`        // 2000-6000kbps
	AudioCodec    string  `json:"audio_codec"`    // AAC
	VideoCodec    string  `json:"video_codec"`    // H.264
	Quality       string  `json:"quality"`        // high
}

// DouyinOptimizationRequest 抖音优化请求
type DouyinOptimizationRequest struct {
	VideoPath     string            `json:"video_path"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Tags          []string          `json:"tags"`
	Category      string            `json:"category"`
	CoverImage    string            `json:"cover_image"`
	Subtitles     []SubtitleSegment `json:"subtitles"`
	Style         string            `json:"style"`
	TargetLength  int               `json:"target_length"` // 目标时长(秒)
}

// SubtitleSegment 字幕片段
type SubtitleSegment struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Text      string  `json:"text"`
	Style     string  `json:"style"`
}

// DouyinOptimizationResult 抖音优化结果
type DouyinOptimizationResult struct {
	OptimizedVideoPath string                 `json:"optimized_video_path"`
	ThumbnailPath      string                 `json:"thumbnail_path"`
	Duration           float64                `json:"duration"`
	FileSize           int64                  `json:"file_size"`
	Resolution         string                 `json:"resolution"`
	Bitrate            string                 `json:"bitrate"`
	QualityScore       float64                `json:"quality_score"`
	Metadata           map[string]interface{} `json:"metadata"`
	Recommendations    []string               `json:"recommendations"`
}

// NewDouyinAdapter 创建抖音适配器
func NewDouyinAdapter(ffmpegPath, tempDir, outputDir string) *DouyinAdapter {
	return &DouyinAdapter{
		ffmpegPath: ffmpegPath,
		tempDir:    tempDir,
		outputDir:  outputDir,
		config: &DouyinConfig{
			Resolution:  "1080x1920",
			AspectRatio: "9:16",
			MaxDuration: 1800,
			MinDuration: 60,
			FrameRate:   30,
			Bitrate:     "4000k",
			AudioCodec:  "aac",
			VideoCodec:  "libx264",
			Quality:     "high",
		},
	}
}

// OptimizeForDouyin 为抖音平台优化视频
func (d *DouyinAdapter) OptimizeForDouyin(ctx context.Context, req *DouyinOptimizationRequest) (*DouyinOptimizationResult, error) {
	log.Printf("[DouyinAdapter] 开始为抖音优化视频: %s", req.VideoPath)
	
	// 1. 创建工作目录
	workDir := filepath.Join(d.tempDir, fmt.Sprintf("douyin_opt_%d", time.Now().Unix()))
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("创建工作目录失败: %w", err)
	}
	defer os.RemoveAll(workDir)
	
	// 2. 分析原视频
	videoInfo, err := d.analyzeVideo(ctx, req.VideoPath)
	if err != nil {
		return nil, fmt.Errorf("分析视频失败: %w", err)
	}
	
	// 3. 转换为竖屏格式
	verticalVideo, err := d.convertToVertical(ctx, req.VideoPath, workDir)
	if err != nil {
		return nil, fmt.Errorf("转换竖屏失败: %w", err)
	}
	
	// 4. 添加抖音风格元素
	styledVideo, err := d.addDouyinStyleElements(ctx, verticalVideo, req, workDir)
	if err != nil {
		return nil, fmt.Errorf("添加风格元素失败: %w", err)
	}
	
	// 5. 添加字幕
	subtitledVideo, err := d.addAnimatedSubtitles(ctx, styledVideo, req.Subtitles, workDir)
	if err != nil {
		return nil, fmt.Errorf("添加字幕失败: %w", err)
	}
	
	// 6. 优化音频
	audioOptimizedVideo, err := d.optimizeAudio(ctx, subtitledVideo, workDir)
	if err != nil {
		return nil, fmt.Errorf("优化音频失败: %w", err)
	}
	
	// 7. 最终编码优化
	finalVideo, err := d.finalEncode(ctx, audioOptimizedVideo, req, workDir)
	if err != nil {
		return nil, fmt.Errorf("最终编码失败: %w", err)
	}
	
	// 8. 生成缩略图
	thumbnail, err := d.generateDouyinThumbnail(ctx, finalVideo, req, workDir)
	if err != nil {
		log.Printf("[DouyinAdapter] 生成缩略图失败: %v", err)
	}
	
	// 9. 移动到输出目录
	outputPath := filepath.Join(d.outputDir, fmt.Sprintf("douyin_%d.mp4", time.Now().Unix()))
	if err := d.moveFile(finalVideo, outputPath); err != nil {
		return nil, fmt.Errorf("移动输出文件失败: %w", err)
	}
	
	// 10. 生成结果
	result, err := d.generateResult(outputPath, thumbnail, videoInfo, req)
	if err != nil {
		return nil, fmt.Errorf("生成结果失败: %w", err)
	}
	
	log.Printf("[DouyinAdapter] 抖音优化完成: %s", outputPath)
	return result, nil
}

// analyzeVideo 分析视频信息
func (d *DouyinAdapter) analyzeVideo(ctx context.Context, videoPath string) (*VideoInfo, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoPath)
	
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var probeResult map[string]interface{}
	if err := json.Unmarshal(output, &probeResult); err != nil {
		return nil, err
	}
	
	return d.parseVideoInfo(probeResult), nil
}

// VideoInfo 视频信息
type VideoInfo struct {
	Duration   float64 `json:"duration"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	FrameRate  float64 `json:"frame_rate"`
	Bitrate    int     `json:"bitrate"`
	AudioCodec string  `json:"audio_codec"`
	VideoCodec string  `json:"video_codec"`
}

// parseVideoInfo 解析视频信息
func (d *DouyinAdapter) parseVideoInfo(probeResult map[string]interface{}) *VideoInfo {
	info := &VideoInfo{}
	
	// 解析格式信息
	if format, ok := probeResult["format"].(map[string]interface{}); ok {
		if duration, ok := format["duration"].(string); ok {
			fmt.Sscanf(duration, "%f", &info.Duration)
		}
		if bitrate, ok := format["bit_rate"].(string); ok {
			fmt.Sscanf(bitrate, "%d", &info.Bitrate)
		}
	}
	
	// 解析流信息
	if streams, ok := probeResult["streams"].([]interface{}); ok {
		for _, stream := range streams {
			if s, ok := stream.(map[string]interface{}); ok {
				codecType, _ := s["codec_type"].(string)
				if codecType == "video" {
					if width, ok := s["width"].(float64); ok {
						info.Width = int(width)
					}
					if height, ok := s["height"].(float64); ok {
						info.Height = int(height)
					}
					if frameRate, ok := s["r_frame_rate"].(string); ok {
						fmt.Sscanf(frameRate, "%f/%f", &info.FrameRate, new(float64))
					}
					if codec, ok := s["codec_name"].(string); ok {
						info.VideoCodec = codec
					}
				} else if codecType == "audio" {
					if codec, ok := s["codec_name"].(string); ok {
						info.AudioCodec = codec
					}
				}
			}
		}
	}
	
	return info
}

// convertToVertical 转换为竖屏格式
func (d *DouyinAdapter) convertToVertical(ctx context.Context, inputPath, workDir string) (string, error) {
	outputPath := filepath.Join(workDir, "vertical.mp4")
	
	// 构建竖屏转换滤镜
	filter := d.buildVerticalFilter()
	
	cmd := exec.CommandContext(ctx, d.ffmpegPath,
		"-i", inputPath,
		"-vf", filter,
		"-c:v", d.config.VideoCodec,
		"-preset", "medium",
		"-crf", "20",
		"-r", fmt.Sprintf("%d", d.config.FrameRate),
		"-c:a", d.config.AudioCodec,
		"-b:a", "128k",
		"-y", outputPath)
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("转换竖屏失败: %w", err)
	}
	
	return outputPath, nil
}

// buildVerticalFilter 构建竖屏滤镜
func (d *DouyinAdapter) buildVerticalFilter() string {
	// 智能竖屏转换：保持主要内容在中心，添加模糊背景
	return fmt.Sprintf(`
		[0:v]scale=1080:1920:force_original_aspect_ratio=decrease,
		pad=1080:1920:(ow-iw)/2:(oh-ih)/2:black[main];
		[0:v]scale=1080:1920,boxblur=20:1[bg];
		[bg][main]overlay=0:0
	`)
}

// addDouyinStyleElements 添加抖音风格元素
func (d *DouyinAdapter) addDouyinStyleElements(ctx context.Context, inputPath string, req *DouyinOptimizationRequest, workDir string) (string, error) {
	outputPath := filepath.Join(workDir, "styled.mp4")
	
	// 构建风格滤镜
	styleFilter := d.buildDouyinStyleFilter(req.Style)
	
	cmd := exec.CommandContext(ctx, d.ffmpegPath,
		"-i", inputPath,
		"-vf", styleFilter,
		"-c:v", d.config.VideoCodec,
		"-preset", "medium",
		"-crf", "20",
		"-c:a", "copy",
		"-y", outputPath)
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("添加风格元素失败: %w", err)
	}
	
	return outputPath, nil
}

// buildDouyinStyleFilter 构建抖音风格滤镜
func (d *DouyinAdapter) buildDouyinStyleFilter(style string) string {
	var filters []string
	
	// 基础优化
	filters = append(filters, "eq=contrast=1.1:saturation=1.1:brightness=0.02")
	
	// 根据风格添加特效
	switch strings.ToLower(style) {
	case "trendy", "时尚":
		filters = append(filters, "colorchannelmixer=.393:.769:.189:0:.349:.686:.168:0:.272:.534:.131")
	case "warm", "温暖":
		filters = append(filters, "colortemperature=4000")
	case "cool", "冷色":
		filters = append(filters, "colortemperature=7000")
	case "vintage", "复古":
		filters = append(filters, "curves=vintage")
	}
	
	// 添加锐化
	filters = append(filters, "unsharp=5:5:0.8:3:3:0.4")
	
	return strings.Join(filters, ",")
}

// addAnimatedSubtitles 添加动画字幕
func (d *DouyinAdapter) addAnimatedSubtitles(ctx context.Context, inputPath string, subtitles []SubtitleSegment, workDir string) (string, error) {
	if len(subtitles) == 0 {
		return inputPath, nil // 没有字幕，直接返回
	}
	
	outputPath := filepath.Join(workDir, "subtitled.mp4")
	
	// 生成字幕文件
	subtitleFile, err := d.generateSubtitleFile(subtitles, workDir)
	if err != nil {
		return "", fmt.Errorf("生成字幕文件失败: %w", err)
	}
	
	// 添加字幕
	cmd := exec.CommandContext(ctx, d.ffmpegPath,
		"-i", inputPath,
		"-vf", fmt.Sprintf("subtitles=%s:force_style='FontSize=24,PrimaryColour=&Hffffff,OutlineColour=&H000000,Outline=2'", subtitleFile),
		"-c:v", d.config.VideoCodec,
		"-preset", "medium",
		"-crf", "20",
		"-c:a", "copy",
		"-y", outputPath)
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("添加字幕失败: %w", err)
	}
	
	return outputPath, nil
}

// generateSubtitleFile 生成字幕文件
func (d *DouyinAdapter) generateSubtitleFile(subtitles []SubtitleSegment, workDir string) (string, error) {
	subtitlePath := filepath.Join(workDir, "subtitles.srt")
	
	file, err := os.Create(subtitlePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	for i, subtitle := range subtitles {
		startTime := d.formatTime(subtitle.StartTime)
		endTime := d.formatTime(subtitle.EndTime)
		
		_, err := file.WriteString(fmt.Sprintf("%d\n%s --> %s\n%s\n\n", 
			i+1, startTime, endTime, subtitle.Text))
		if err != nil {
			return "", err
		}
	}
	
	return subtitlePath, nil
}

// formatTime 格式化时间
func (d *DouyinAdapter) formatTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int((seconds - float64(int(seconds))) * 1000)
	
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, millis)
}

// optimizeAudio 优化音频
func (d *DouyinAdapter) optimizeAudio(ctx context.Context, inputPath, workDir string) (string, error) {
	outputPath := filepath.Join(workDir, "audio_optimized.mp4")
	
	cmd := exec.CommandContext(ctx, d.ffmpegPath,
		"-i", inputPath,
		"-c:v", "copy",
		"-c:a", d.config.AudioCodec,
		"-b:a", "128k",
		"-ar", "48000",
		"-ac", "2",
		"-af", "loudnorm=I=-16:TP=-1.5:LRA=11",
		"-y", outputPath)
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("优化音频失败: %w", err)
	}
	
	return outputPath, nil
}

// finalEncode 最终编码
func (d *DouyinAdapter) finalEncode(ctx context.Context, inputPath string, req *DouyinOptimizationRequest, workDir string) (string, error) {
	outputPath := filepath.Join(workDir, "final.mp4")
	
	cmd := exec.CommandContext(ctx, d.ffmpegPath,
		"-i", inputPath,
		"-c:v", d.config.VideoCodec,
		"-preset", "medium",
		"-crf", "20",
		"-maxrate", d.config.Bitrate,
		"-bufsize", "8000k",
		"-r", fmt.Sprintf("%d", d.config.FrameRate),
		"-s", d.config.Resolution,
		"-c:a", d.config.AudioCodec,
		"-b:a", "128k",
		"-movflags", "+faststart",
		"-y", outputPath)
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("最终编码失败: %w", err)
	}
	
	return outputPath, nil
}

// generateDouyinThumbnail 生成抖音缩略图
func (d *DouyinAdapter) generateDouyinThumbnail(ctx context.Context, videoPath string, req *DouyinOptimizationRequest, workDir string) (string, error) {
	thumbnailPath := filepath.Join(workDir, "thumbnail.jpg")

	// 从视频的1/3处提取缩略图
	cmd := exec.CommandContext(ctx, d.ffmpegPath,
		"-i", videoPath,
		"-ss", "00:00:02",
		"-vframes", "1",
		"-vf", "scale=1080:1920:force_original_aspect_ratio=decrease,pad=1080:1920:(ow-iw)/2:(oh-ih)/2:black",
		"-q:v", "2",
		"-y", thumbnailPath)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("生成缩略图失败: %w", err)
	}

	return thumbnailPath, nil
}

// moveFile 移动文件
func (d *DouyinAdapter) moveFile(src, dst string) error {
	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// 复制文件
	return d.copyFile(src, dst)
}

// copyFile 复制文件
func (d *DouyinAdapter) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

// generateResult 生成结果
func (d *DouyinAdapter) generateResult(videoPath, thumbnailPath string, videoInfo *VideoInfo, req *DouyinOptimizationRequest) (*DouyinOptimizationResult, error) {
	// 获取文件信息
	fileInfo, err := os.Stat(videoPath)
	if err != nil {
		return nil, err
	}

	// 计算质量评分
	qualityScore := d.calculateQualityScore(videoInfo, req)

	// 生成建议
	recommendations := d.generateRecommendations(videoInfo, qualityScore)

	result := &DouyinOptimizationResult{
		OptimizedVideoPath: videoPath,
		ThumbnailPath:      thumbnailPath,
		Duration:           videoInfo.Duration,
		FileSize:           fileInfo.Size(),
		Resolution:         d.config.Resolution,
		Bitrate:            d.config.Bitrate,
		QualityScore:       qualityScore,
		Metadata: map[string]interface{}{
			"original_resolution": fmt.Sprintf("%dx%d", videoInfo.Width, videoInfo.Height),
			"original_bitrate":    videoInfo.Bitrate,
			"frame_rate":          d.config.FrameRate,
			"audio_codec":         d.config.AudioCodec,
			"video_codec":         d.config.VideoCodec,
			"platform":            "douyin",
			"optimization_time":   time.Now().Format(time.RFC3339),
		},
		Recommendations: recommendations,
	}

	return result, nil
}

// calculateQualityScore 计算质量评分
func (d *DouyinAdapter) calculateQualityScore(videoInfo *VideoInfo, req *DouyinOptimizationRequest) float64 {
	score := 100.0

	// 时长评分
	if videoInfo.Duration < float64(d.config.MinDuration) {
		score -= 20 // 时长过短
	} else if videoInfo.Duration > float64(d.config.MaxDuration) {
		score -= 10 // 时长过长
	}

	// 分辨率评分
	if videoInfo.Width < 1080 || videoInfo.Height < 1920 {
		score -= 15 // 分辨率不足
	}

	// 帧率评分
	if videoInfo.FrameRate < 24 {
		score -= 10 // 帧率过低
	}

	// 比特率评分
	if videoInfo.Bitrate < 2000000 { // 2Mbps
		score -= 10 // 比特率过低
	}

	// 字幕评分
	if len(req.Subtitles) > 0 {
		score += 5 // 有字幕加分
	}

	// 标题和描述评分
	if req.Title != "" && len(req.Title) > 10 {
		score += 3 // 有详细标题
	}
	if req.Description != "" && len(req.Description) > 20 {
		score += 2 // 有详细描述
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// generateRecommendations 生成建议
func (d *DouyinAdapter) generateRecommendations(videoInfo *VideoInfo, qualityScore float64) []string {
	var recommendations []string

	if qualityScore < 70 {
		recommendations = append(recommendations, "视频质量较低，建议重新优化")
	}

	if videoInfo.Duration < 60 {
		recommendations = append(recommendations, "视频时长过短，建议增加内容至少1分钟")
	}

	if videoInfo.Duration > 1800 {
		recommendations = append(recommendations, "视频时长过长，建议分割为多个短视频")
	}

	if videoInfo.Width < 1080 || videoInfo.Height < 1920 {
		recommendations = append(recommendations, "建议使用1080x1920分辨率以获得最佳效果")
	}

	if videoInfo.FrameRate < 30 {
		recommendations = append(recommendations, "建议使用30fps以获得更流畅的播放效果")
	}

	recommendations = append(recommendations, "添加吸引人的封面图片")
	recommendations = append(recommendations, "使用热门话题标签增加曝光")
	recommendations = append(recommendations, "在前3秒内展示最精彩的内容")

	return recommendations
}

// GetDouyinSpecs 获取抖音规格要求
func (d *DouyinAdapter) GetDouyinSpecs() *DouyinConfig {
	return d.config
}

// ValidateForDouyin 验证视频是否符合抖音要求
func (d *DouyinAdapter) ValidateForDouyin(ctx context.Context, videoPath string) (*ValidationResult, error) {
	videoInfo, err := d.analyzeVideo(ctx, videoPath)
	if err != nil {
		return nil, err
	}

	result := &ValidationResult{
		IsValid: true,
		Issues:  []string{},
		Warnings: []string{},
	}

	// 检查时长
	if videoInfo.Duration < float64(d.config.MinDuration) {
		result.IsValid = false
		result.Issues = append(result.Issues, fmt.Sprintf("视频时长过短: %.1f秒，最少需要%d秒", videoInfo.Duration, d.config.MinDuration))
	}

	if videoInfo.Duration > float64(d.config.MaxDuration) {
		result.IsValid = false
		result.Issues = append(result.Issues, fmt.Sprintf("视频时长过长: %.1f秒，最多允许%d秒", videoInfo.Duration, d.config.MaxDuration))
	}

	// 检查分辨率
	if videoInfo.Width != 1080 || videoInfo.Height != 1920 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("分辨率不是最佳: %dx%d，推荐1080x1920", videoInfo.Width, videoInfo.Height))
	}

	// 检查帧率
	if videoInfo.FrameRate < 24 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("帧率较低: %.1ffps，推荐30fps", videoInfo.FrameRate))
	}

	return result, nil
}

// ValidationResult 验证结果
type ValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Issues   []string `json:"issues"`
	Warnings []string `json:"warnings"`
}
