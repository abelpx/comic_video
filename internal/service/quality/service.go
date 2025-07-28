package quality

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
)

// Service 质量检测服务
type Service struct {
	ffmpegPath string
	ffprobePath string
}

// NewService 创建质量检测服务
func NewService(ffmpegPath, ffprobePath string) *Service {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	if ffprobePath == "" {
		ffprobePath = "ffprobe"
	}
	
	return &Service{
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
	}
}

// VideoQualityReport 视频质量报告
type VideoQualityReport struct {
	FilePath        string                 `json:"file_path"`
	FileSize        int64                  `json:"file_size"`
	Duration        float64                `json:"duration"`
	Bitrate         int                    `json:"bitrate"`
	Resolution      Resolution             `json:"resolution"`
	FrameRate       float64                `json:"frame_rate"`
	AudioInfo       AudioInfo              `json:"audio_info"`
	QualityScore    float64                `json:"quality_score"`
	Issues          []QualityIssue         `json:"issues"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
}

// Resolution 分辨率信息
type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// AudioInfo 音频信息
type AudioInfo struct {
	Codec      string  `json:"codec"`
	Bitrate    int     `json:"bitrate"`
	SampleRate int     `json:"sample_rate"`
	Channels   int     `json:"channels"`
	Duration   float64 `json:"duration"`
}

// QualityIssue 质量问题
type QualityIssue struct {
	Type        string `json:"type"`        // 问题类型
	Severity    string `json:"severity"`    // 严重程度：low/medium/high/critical
	Description string `json:"description"` // 问题描述
	Solution    string `json:"solution"`    // 解决方案
}

// OptimizationOptions 优化选项
type OptimizationOptions struct {
	TargetBitrate    int    `json:"target_bitrate"`    // 目标比特率
	TargetResolution string `json:"target_resolution"` // 目标分辨率
	TargetFormat     string `json:"target_format"`     // 目标格式
	Quality          string `json:"quality"`           // 质量级别：low/medium/high/ultra
	Platform         string `json:"platform"`          // 目标平台
	MaxFileSize      int64  `json:"max_file_size"`     // 最大文件大小（字节）
}

// AnalyzeVideo 分析视频质量
func (s *Service) AnalyzeVideo(ctx context.Context, videoPath string) (*VideoQualityReport, error) {
	log.Printf("[QualityService] 开始分析视频质量: %s", videoPath)

	// 1. 获取基本信息
	metadata, err := s.getVideoMetadata(ctx, videoPath)
	if err != nil {
		return nil, fmt.Errorf("获取视频元数据失败: %w", err)
	}

	// 2. 获取文件信息
	fileInfo, err := os.Stat(videoPath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 3. 解析元数据
	report := &VideoQualityReport{
		FilePath:  videoPath,
		FileSize:  fileInfo.Size(),
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	s.parseVideoMetadata(report, metadata)

	// 4. 质量检测
	s.detectQualityIssues(report)

	// 5. 计算质量评分
	report.QualityScore = s.calculateQualityScore(report)

	// 6. 生成建议
	report.Recommendations = s.generateRecommendations(report)

	log.Printf("[QualityService] 视频质量分析完成: score=%.2f, issues=%d", 
		report.QualityScore, len(report.Issues))

	return report, nil
}

// OptimizeVideo 优化视频
func (s *Service) OptimizeVideo(ctx context.Context, inputPath string, options *OptimizationOptions) (string, error) {
	log.Printf("[QualityService] 开始优化视频: %s", inputPath)

	// 1. 生成输出路径
	outputPath := s.generateOptimizedPath(inputPath, options)

	// 2. 构建FFmpeg命令
	cmd := s.buildOptimizationCommand(inputPath, outputPath, options)

	// 3. 执行优化
	if err := s.executeFFmpegCommand(ctx, cmd); err != nil {
		return "", fmt.Errorf("视频优化失败: %w", err)
	}

	// 4. 验证输出
	if _, err := os.Stat(outputPath); err != nil {
		return "", fmt.Errorf("优化后的视频文件不存在: %w", err)
	}

	log.Printf("[QualityService] 视频优化完成: %s", outputPath)
	return outputPath, nil
}

// getVideoMetadata 获取视频元数据
func (s *Service) getVideoMetadata(ctx context.Context, videoPath string) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, s.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoPath)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(output, &metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

// parseVideoMetadata 解析视频元数据
func (s *Service) parseVideoMetadata(report *VideoQualityReport, metadata map[string]interface{}) {
	// 解析格式信息
	if format, ok := metadata["format"].(map[string]interface{}); ok {
		if duration, ok := format["duration"].(string); ok {
			if d, err := strconv.ParseFloat(duration, 64); err == nil {
				report.Duration = d
			}
		}
		if bitrate, ok := format["bit_rate"].(string); ok {
			if b, err := strconv.Atoi(bitrate); err == nil {
				report.Bitrate = b
			}
		}
	}

	// 解析流信息
	if streams, ok := metadata["streams"].([]interface{}); ok {
		for _, stream := range streams {
			if s, ok := stream.(map[string]interface{}); ok {
				codecType, _ := s["codec_type"].(string)
				
				switch codecType {
				case "video":
					parseVideoStream(report, s)
				case "audio":
					parseAudioStream(report, s)
				}
			}
		}
	}
}

// parseVideoStream 解析视频流
func parseVideoStream(report *VideoQualityReport, stream map[string]interface{}) {
	if width, ok := stream["width"].(float64); ok {
		report.Resolution.Width = int(width)
	}
	if height, ok := stream["height"].(float64); ok {
		report.Resolution.Height = int(height)
	}
	if frameRate, ok := stream["r_frame_rate"].(string); ok {
		if rate := parseFrameRate(frameRate); rate > 0 {
			report.FrameRate = rate
		}
	}
}

// parseAudioStream 解析音频流
func parseAudioStream(report *VideoQualityReport, stream map[string]interface{}) {
	if codec, ok := stream["codec_name"].(string); ok {
		report.AudioInfo.Codec = codec
	}
	if bitrate, ok := stream["bit_rate"].(string); ok {
		if b, err := strconv.Atoi(bitrate); err == nil {
			report.AudioInfo.Bitrate = b
		}
	}
	if sampleRate, ok := stream["sample_rate"].(string); ok {
		if sr, err := strconv.Atoi(sampleRate); err == nil {
			report.AudioInfo.SampleRate = sr
		}
	}
	if channels, ok := stream["channels"].(float64); ok {
		report.AudioInfo.Channels = int(channels)
	}
}

// parseFrameRate 解析帧率
func parseFrameRate(frameRateStr string) float64 {
	parts := strings.Split(frameRateStr, "/")
	if len(parts) == 2 {
		num, err1 := strconv.ParseFloat(parts[0], 64)
		den, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 == nil && err2 == nil && den != 0 {
			return num / den
		}
	}
	return 0
}

// detectQualityIssues 检测质量问题
func (s *Service) detectQualityIssues(report *VideoQualityReport) {
	var issues []QualityIssue

	// 检查分辨率
	if report.Resolution.Width < 720 || report.Resolution.Height < 480 {
		issues = append(issues, QualityIssue{
			Type:        "low_resolution",
			Severity:    "medium",
			Description: fmt.Sprintf("分辨率较低: %dx%d", report.Resolution.Width, report.Resolution.Height),
			Solution:    "建议使用至少720p分辨率",
		})
	}

	// 检查比特率
	if report.Bitrate < 1000000 { // 1Mbps
		issues = append(issues, QualityIssue{
			Type:        "low_bitrate",
			Severity:    "medium",
			Description: fmt.Sprintf("比特率较低: %d bps", report.Bitrate),
			Solution:    "建议提高比特率以改善画质",
		})
	}

	// 检查帧率
	if report.FrameRate < 24 {
		issues = append(issues, QualityIssue{
			Type:        "low_framerate",
			Severity:    "low",
			Description: fmt.Sprintf("帧率较低: %.2f fps", report.FrameRate),
			Solution:    "建议使用至少24fps",
		})
	}

	// 检查文件大小
	if report.FileSize > 100*1024*1024 { // 100MB
		issues = append(issues, QualityIssue{
			Type:        "large_file_size",
			Severity:    "low",
			Description: fmt.Sprintf("文件较大: %.2f MB", float64(report.FileSize)/(1024*1024)),
			Solution:    "考虑压缩以减小文件大小",
		})
	}

	// 检查音频
	if report.AudioInfo.Codec == "" {
		issues = append(issues, QualityIssue{
			Type:        "no_audio",
			Severity:    "high",
			Description: "视频没有音频轨道",
			Solution:    "添加音频轨道以提升观看体验",
		})
	}

	report.Issues = issues
}

// calculateQualityScore 计算质量评分
func (s *Service) calculateQualityScore(report *VideoQualityReport) float64 {
	score := 100.0

	// 分辨率评分
	if report.Resolution.Width >= 1920 && report.Resolution.Height >= 1080 {
		score += 0 // 满分
	} else if report.Resolution.Width >= 1280 && report.Resolution.Height >= 720 {
		score -= 10
	} else {
		score -= 25
	}

	// 比特率评分
	if report.Bitrate >= 5000000 { // 5Mbps
		score += 0
	} else if report.Bitrate >= 2000000 { // 2Mbps
		score -= 10
	} else {
		score -= 20
	}

	// 帧率评分
	if report.FrameRate >= 30 {
		score += 0
	} else if report.FrameRate >= 24 {
		score -= 5
	} else {
		score -= 15
	}

	// 音频评分
	if report.AudioInfo.Codec != "" {
		score += 0
	} else {
		score -= 30
	}

	// 问题扣分
	for _, issue := range report.Issues {
		switch issue.Severity {
		case "critical":
			score -= 25
		case "high":
			score -= 15
		case "medium":
			score -= 10
		case "low":
			score -= 5
		}
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
func (s *Service) generateRecommendations(report *VideoQualityReport) []string {
	var recommendations []string

	if report.QualityScore < 60 {
		recommendations = append(recommendations, "视频质量较低，建议重新制作")
	}

	if report.Resolution.Width < 1280 {
		recommendations = append(recommendations, "建议提升分辨率至至少720p")
	}

	if report.Bitrate < 2000000 {
		recommendations = append(recommendations, "建议提高比特率以改善画质")
	}

	if report.AudioInfo.Codec == "" {
		recommendations = append(recommendations, "建议添加音频轨道")
	}

	if report.FileSize > 50*1024*1024 {
		recommendations = append(recommendations, "考虑压缩文件以便于分享")
	}

	return recommendations
}

// generateOptimizedPath 生成优化后的文件路径
func (s *Service) generateOptimizedPath(inputPath string, options *OptimizationOptions) string {
	dir := filepath.Dir(inputPath)
	ext := filepath.Ext(inputPath)
	baseName := strings.TrimSuffix(filepath.Base(inputPath), ext)
	
	format := options.TargetFormat
	if format == "" {
		format = "mp4"
	}
	
	suffix := fmt.Sprintf("_optimized_%s", options.Quality)
	return filepath.Join(dir, fmt.Sprintf("%s%s.%s", baseName, suffix, format))
}

// buildOptimizationCommand 构建优化命令
func (s *Service) buildOptimizationCommand(inputPath, outputPath string, options *OptimizationOptions) *exec.Cmd {
	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", s.getQualityCRF(options.Quality),
	}

	// 分辨率设置
	if options.TargetResolution != "" {
		args = append(args, "-vf", fmt.Sprintf("scale=%s", options.TargetResolution))
	}

	// 比特率设置
	if options.TargetBitrate > 0 {
		args = append(args, "-b:v", fmt.Sprintf("%dk", options.TargetBitrate/1000))
	}

	// 音频设置
	args = append(args, "-c:a", "aac", "-b:a", "128k")

	// 输出文件
	args = append(args, "-y", outputPath)

	return exec.Command(s.ffmpegPath, args...)
}

// getQualityCRF 获取质量CRF值
func (s *Service) getQualityCRF(quality string) string {
	switch strings.ToLower(quality) {
	case "ultra":
		return "18"
	case "high":
		return "23"
	case "medium":
		return "28"
	case "low":
		return "32"
	default:
		return "23"
	}
}

// executeFFmpegCommand 执行FFmpeg命令
func (s *Service) executeFFmpegCommand(ctx context.Context, cmd *exec.Cmd) error {
	cmd.Dir = filepath.Dir(cmd.Args[len(cmd.Args)-1]) // 设置工作目录
	
	log.Printf("[QualityService] 执行命令: %s", strings.Join(cmd.Args, " "))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[QualityService] FFmpeg错误: %s", string(output))
		return fmt.Errorf("FFmpeg执行失败: %w", err)
	}
	
	return nil
}
