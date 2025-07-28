package video

import (
	"context"
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

// AdvancedVideoGenerator 高级视频生成器
type AdvancedVideoGenerator struct {
	aiService       *ai.Service
	ffmpegPath      string
	tempDir         string
	outputDir       string
	videoConfig     *VideoConfig
}

// VideoConfig 视频配置
type VideoConfig struct {
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	FrameRate   int     `json:"frame_rate"`
	Bitrate     string  `json:"bitrate"`
	Duration    float64 `json:"duration"`
	Format      string  `json:"format"`
	Quality     string  `json:"quality"`
	Platform    string  `json:"platform"`
}

// VideoGenerationRequest 视频生成请求
type VideoGenerationRequest struct {
	ProjectID    string                 `json:"project_id"`
	Script       *entity.Script         `json:"script"`
	Characters   []*entity.Character    `json:"characters"`
	Scenes       []*entity.Scene        `json:"scenes"`
	Storyboard   *entity.Storyboard     `json:"storyboard"`
	AudioFiles   []string               `json:"audio_files"`
	ImageFiles   []string               `json:"image_files"`
	Config       *VideoConfig           `json:"config"`
	Style        string                 `json:"style"`
}

// VideoGenerationResult 视频生成结果
type VideoGenerationResult struct {
	VideoPath      string                 `json:"video_path"`
	ThumbnailPath  string                 `json:"thumbnail_path"`
	Duration       float64                `json:"duration"`
	FileSize       int64                  `json:"file_size"`
	Resolution     string                 `json:"resolution"`
	Quality        float64                `json:"quality"`
	Metadata       map[string]interface{} `json:"metadata"`
	ProcessingTime time.Duration          `json:"processing_time"`
}

// NewAdvancedVideoGenerator 创建高级视频生成器
func NewAdvancedVideoGenerator(aiService *ai.Service, ffmpegPath, tempDir, outputDir string) *AdvancedVideoGenerator {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	
	return &AdvancedVideoGenerator{
		aiService:  aiService,
		ffmpegPath: ffmpegPath,
		tempDir:    tempDir,
		outputDir:  outputDir,
		videoConfig: &VideoConfig{
			Width:     1920,
			Height:    1080,
			FrameRate: 30,
			Bitrate:   "4000k",
			Format:    "mp4",
			Quality:   "high",
			Platform:  "general",
		},
	}
}

// GenerateAdvancedVideo 生成高级视频
func (g *AdvancedVideoGenerator) GenerateAdvancedVideo(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	startTime := time.Now()
	log.Printf("[AdvancedVideoGenerator] 开始生成高级视频: project_id=%s", req.ProjectID)

	// 1. 准备工作目录
	workDir := filepath.Join(g.tempDir, req.ProjectID)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("创建工作目录失败: %w", err)
	}
	defer os.RemoveAll(workDir)

	// 2. 应用平台特定配置
	config := g.getPlatformConfig(req.Config, req.Config.Platform)

	// 3. 生成动态图像序列
	imageSequence, err := g.generateDynamicImageSequence(ctx, req, workDir)
	if err != nil {
		return nil, fmt.Errorf("生成动态图像序列失败: %w", err)
	}

	// 4. 处理音频
	processedAudio, err := g.processAudioTracks(ctx, req.AudioFiles, workDir)
	if err != nil {
		return nil, fmt.Errorf("处理音频失败: %w", err)
	}

	// 5. 生成视频
	videoPath, err := g.createVideoFromSequence(ctx, imageSequence, processedAudio, config, workDir)
	if err != nil {
		return nil, fmt.Errorf("创建视频失败: %w", err)
	}

	// 6. 后期处理
	finalVideoPath, err := g.postProcessVideo(ctx, videoPath, req, config, workDir)
	if err != nil {
		return nil, fmt.Errorf("后期处理失败: %w", err)
	}

	// 7. 生成缩略图
	thumbnailPath, err := g.generateThumbnail(ctx, finalVideoPath, workDir)
	if err != nil {
		log.Printf("[AdvancedVideoGenerator] 生成缩略图失败: %v", err)
	}

	// 8. 移动到输出目录
	outputVideoPath := filepath.Join(g.outputDir, fmt.Sprintf("%s_final.%s", req.ProjectID, config.Format))
	if err := g.moveToOutput(finalVideoPath, outputVideoPath); err != nil {
		return nil, fmt.Errorf("移动输出文件失败: %w", err)
	}

	// 9. 获取视频信息
	duration, fileSize, err := g.getVideoInfo(outputVideoPath)
	if err != nil {
		log.Printf("[AdvancedVideoGenerator] 获取视频信息失败: %v", err)
	}

	result := &VideoGenerationResult{
		VideoPath:      outputVideoPath,
		ThumbnailPath:  thumbnailPath,
		Duration:       duration,
		FileSize:       fileSize,
		Resolution:     fmt.Sprintf("%dx%d", config.Width, config.Height),
		Quality:        8.5, // 默认质量评分
		ProcessingTime: time.Since(startTime),
		Metadata: map[string]interface{}{
			"platform":    config.Platform,
			"style":       req.Style,
			"frame_rate":  config.FrameRate,
			"bitrate":     config.Bitrate,
		},
	}

	log.Printf("[AdvancedVideoGenerator] 视频生成完成: %s, 耗时: %v", outputVideoPath, result.ProcessingTime)
	return result, nil
}

// getPlatformConfig 获取平台特定配置
func (g *AdvancedVideoGenerator) getPlatformConfig(baseConfig *VideoConfig, platform string) *VideoConfig {
	config := *baseConfig // 复制基础配置
	
	switch strings.ToLower(platform) {
	case "douyin", "tiktok":
		// 抖音竖屏配置
		config.Width = 1080
		config.Height = 1920
		config.FrameRate = 30
		config.Bitrate = "3000k"
		config.Quality = "high"
	case "bilibili":
		// B站横屏配置
		config.Width = 1920
		config.Height = 1080
		config.FrameRate = 60
		config.Bitrate = "6000k"
		config.Quality = "ultra"
	case "youtube":
		// YouTube配置
		config.Width = 1920
		config.Height = 1080
		config.FrameRate = 30
		config.Bitrate = "5000k"
		config.Quality = "high"
	case "weibo":
		// 微博配置
		config.Width = 1280
		config.Height = 720
		config.FrameRate = 25
		config.Bitrate = "2000k"
		config.Quality = "medium"
	}
	
	return &config
}

// generateDynamicImageSequence 生成动态图像序列
func (g *AdvancedVideoGenerator) generateDynamicImageSequence(ctx context.Context, req *VideoGenerationRequest, workDir string) ([]string, error) {
	log.Printf("[AdvancedVideoGenerator] 生成动态图像序列")
	
	var imageSequence []string
	
	// 为每个分镜帧生成多帧图像以创建动画效果
	for i, imagePath := range req.ImageFiles {
		// 生成帧间插值图像
		interpolatedFrames, err := g.generateInterpolatedFrames(ctx, imagePath, 8, workDir) // 每张图生成8帧
		if err != nil {
			log.Printf("[AdvancedVideoGenerator] 生成插值帧失败: %v", err)
			// 如果插值失败，使用原图重复
			for j := 0; j < 8; j++ {
				imageSequence = append(imageSequence, imagePath)
			}
		} else {
			imageSequence = append(imageSequence, interpolatedFrames...)
		}
		
		log.Printf("[AdvancedVideoGenerator] 处理图像 %d/%d", i+1, len(req.ImageFiles))
	}
	
	return imageSequence, nil
}

// generateInterpolatedFrames 生成插值帧
func (g *AdvancedVideoGenerator) generateInterpolatedFrames(ctx context.Context, imagePath string, frameCount int, workDir string) ([]string, error) {
	// 这里可以集成RIFE或其他帧插值算法
	// 目前使用简单的图像变换模拟动画效果
	
	var frames []string
	baseName := filepath.Base(imagePath)
	nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	
	for i := 0; i < frameCount; i++ {
		framePath := filepath.Join(workDir, fmt.Sprintf("%s_frame_%03d.jpg", nameWithoutExt, i))
		
		// 使用FFmpeg创建轻微的变换效果
		cmd := exec.CommandContext(ctx, g.ffmpegPath,
			"-i", imagePath,
			"-vf", fmt.Sprintf("scale=iw*%f:ih*%f,crop=iw:ih", 1.0+float64(i)*0.001, 1.0+float64(i)*0.001),
			"-q:v", "2",
			"-y", framePath)
		
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("生成插值帧失败: %w", err)
		}
		
		frames = append(frames, framePath)
	}
	
	return frames, nil
}

// processAudioTracks 处理音频轨道
func (g *AdvancedVideoGenerator) processAudioTracks(ctx context.Context, audioFiles []string, workDir string) (string, error) {
	if len(audioFiles) == 0 {
		return "", fmt.Errorf("没有音频文件")
	}
	
	outputPath := filepath.Join(workDir, "processed_audio.wav")
	
	if len(audioFiles) == 1 {
		// 单个音频文件，直接处理
		cmd := exec.CommandContext(ctx, g.ffmpegPath,
			"-i", audioFiles[0],
			"-ar", "48000",
			"-ac", "2",
			"-c:a", "pcm_s16le",
			"-y", outputPath)
		
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("处理音频失败: %w", err)
		}
	} else {
		// 多个音频文件，需要混合
		return g.mixAudioFiles(ctx, audioFiles, workDir)
	}
	
	return outputPath, nil
}

// mixAudioFiles 混合多个音频文件
func (g *AdvancedVideoGenerator) mixAudioFiles(ctx context.Context, audioFiles []string, workDir string) (string, error) {
	outputPath := filepath.Join(workDir, "mixed_audio.wav")
	
	// 构建FFmpeg命令来混合音频
	args := []string{}
	
	// 添加输入文件
	for _, audioFile := range audioFiles {
		args = append(args, "-i", audioFile)
	}
	
	// 添加混合滤镜
	filterComplex := fmt.Sprintf("amix=inputs=%d:duration=longest", len(audioFiles))
	args = append(args, "-filter_complex", filterComplex)
	
	// 添加输出参数
	args = append(args, "-ar", "48000", "-ac", "2", "-c:a", "pcm_s16le", "-y", outputPath)
	
	cmd := exec.CommandContext(ctx, g.ffmpegPath, args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("混合音频失败: %w", err)
	}
	
	return outputPath, nil
}

// createVideoFromSequence 从图像序列创建视频
func (g *AdvancedVideoGenerator) createVideoFromSequence(ctx context.Context, imageSequence []string, audioPath string, config *VideoConfig, workDir string) (string, error) {
	log.Printf("[AdvancedVideoGenerator] 从图像序列创建视频")
	
	// 创建图像列表文件
	listFile := filepath.Join(workDir, "image_list.txt")
	if err := g.createImageListFile(imageSequence, listFile); err != nil {
		return "", fmt.Errorf("创建图像列表失败: %w", err)
	}
	
	outputPath := filepath.Join(workDir, "raw_video.mp4")
	
	// 构建FFmpeg命令
	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", listFile,
		"-i", audioPath,
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", g.getQualityCRF(config.Quality),
		"-r", strconv.Itoa(config.FrameRate),
		"-s", fmt.Sprintf("%dx%d", config.Width, config.Height),
		"-c:a", "aac",
		"-b:a", "128k",
		"-shortest",
		"-y", outputPath,
	}
	
	cmd := exec.CommandContext(ctx, g.ffmpegPath, args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("创建视频失败: %w", err)
	}
	
	return outputPath, nil
}

// createImageListFile 创建图像列表文件
func (g *AdvancedVideoGenerator) createImageListFile(imageSequence []string, listFile string) error {
	file, err := os.Create(listFile)
	if err != nil {
		return err
	}
	defer file.Close()
	
	for _, imagePath := range imageSequence {
		// 每张图片显示0.125秒 (8fps)
		_, err := file.WriteString(fmt.Sprintf("file '%s'\nduration 0.125\n", imagePath))
		if err != nil {
			return err
		}
	}
	
	// 最后一张图片
	if len(imageSequence) > 0 {
		_, err := file.WriteString(fmt.Sprintf("file '%s'\n", imageSequence[len(imageSequence)-1]))
		if err != nil {
			return err
		}
	}
	
	return nil
}

// getQualityCRF 获取质量CRF值
func (g *AdvancedVideoGenerator) getQualityCRF(quality string) string {
	switch strings.ToLower(quality) {
	case "ultra":
		return "16"
	case "high":
		return "20"
	case "medium":
		return "24"
	case "low":
		return "28"
	default:
		return "20"
	}
}

// postProcessVideo 后期处理视频
func (g *AdvancedVideoGenerator) postProcessVideo(ctx context.Context, videoPath string, req *VideoGenerationRequest, config *VideoConfig, workDir string) (string, error) {
	log.Printf("[AdvancedVideoGenerator] 开始后期处理")

	outputPath := filepath.Join(workDir, "post_processed_video.mp4")

	// 构建后期处理滤镜链
	filters := g.buildPostProcessingFilters(req, config)

	args := []string{
		"-i", videoPath,
		"-vf", filters,
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", g.getQualityCRF(config.Quality),
		"-c:a", "copy",
		"-y", outputPath,
	}

	cmd := exec.CommandContext(ctx, g.ffmpegPath, args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("后期处理失败: %w", err)
	}

	return outputPath, nil
}

// buildPostProcessingFilters 构建后期处理滤镜
func (g *AdvancedVideoGenerator) buildPostProcessingFilters(req *VideoGenerationRequest, config *VideoConfig) string {
	var filters []string

	// 基础滤镜
	filters = append(filters, fmt.Sprintf("scale=%d:%d", config.Width, config.Height))

	// 根据风格添加滤镜
	switch strings.ToLower(req.Style) {
	case "anime", "cartoon":
		// 动漫风格：增强饱和度和对比度
		filters = append(filters, "eq=saturation=1.2:contrast=1.1")
		filters = append(filters, "unsharp=5:5:1.0:5:5:0.0")
	case "realistic", "photographic":
		// 写实风格：自然色彩
		filters = append(filters, "eq=brightness=0.02:contrast=1.05")
	case "vintage", "retro":
		// 复古风格
		filters = append(filters, "colorchannelmixer=.393:.769:.189:0:.349:.686:.168:0:.272:.534:.131")
	case "dark", "noir":
		// 暗黑风格
		filters = append(filters, "eq=brightness=-0.1:contrast=1.3:saturation=0.8")
	}

	// 平台特定优化
	switch strings.ToLower(config.Platform) {
	case "douyin", "tiktok":
		// 抖音：增强对比度，适合手机观看
		filters = append(filters, "eq=contrast=1.1:saturation=1.1")
	case "bilibili":
		// B站：保持原色彩
		filters = append(filters, "eq=gamma=1.0")
	}

	// 添加锐化
	filters = append(filters, "unsharp=5:5:0.8:3:3:0.4")

	return strings.Join(filters, ",")
}

// generateThumbnail 生成缩略图
func (g *AdvancedVideoGenerator) generateThumbnail(ctx context.Context, videoPath, workDir string) (string, error) {
	thumbnailPath := filepath.Join(workDir, "thumbnail.jpg")

	cmd := exec.CommandContext(ctx, g.ffmpegPath,
		"-i", videoPath,
		"-ss", "00:00:03", // 第3秒的帧
		"-vframes", "1",
		"-q:v", "2",
		"-y", thumbnailPath)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("生成缩略图失败: %w", err)
	}

	return thumbnailPath, nil
}

// moveToOutput 移动文件到输出目录
func (g *AdvancedVideoGenerator) moveToOutput(srcPath, dstPath string) error {
	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}

	// 复制文件
	return g.copyFile(srcPath, dstPath)
}

// copyFile 复制文件
func (g *AdvancedVideoGenerator) copyFile(src, dst string) error {
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

// getVideoInfo 获取视频信息
func (g *AdvancedVideoGenerator) getVideoInfo(videoPath string) (duration float64, fileSize int64, err error) {
	// 获取文件大小
	fileInfo, err := os.Stat(videoPath)
	if err != nil {
		return 0, 0, err
	}
	fileSize = fileInfo.Size()

	// 使用ffprobe获取视频时长
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		videoPath)

	output, err := cmd.Output()
	if err != nil {
		return 0, fileSize, err
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err = strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fileSize, err
	}

	return duration, fileSize, nil
}
