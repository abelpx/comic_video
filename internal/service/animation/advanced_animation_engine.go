package animation

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"comic_video/internal/domain/entity"
	"comic_video/internal/utils"
	"github.com/google/uuid"
)

// AdvancedAnimationEngine 高级动画引擎
type AdvancedAnimationEngine struct {
	frameInterpolator *FrameInterpolator
	motionGenerator   *MotionGenerator
	effectsProcessor  *EffectsProcessor
	transitionEngine  *TransitionEngine
	qualityController *AnimationQualityController
}

// NewAdvancedAnimationEngine 创建高级动画引擎
func NewAdvancedAnimationEngine() *AdvancedAnimationEngine {
	return &AdvancedAnimationEngine{
		frameInterpolator: NewFrameInterpolator(),
		motionGenerator:   NewMotionGenerator(),
		effectsProcessor:  NewEffectsProcessor(),
		transitionEngine:  NewTransitionEngine(),
		qualityController: NewAnimationQualityController(),
	}
}

// AnimationRequest 动画生成请求
type AnimationRequest struct {
	ProjectID       uuid.UUID              `json:"project_id"`
	Scenes          []*entity.Scene        `json:"scenes"`
	Characters      []*entity.Character    `json:"characters"`
	StaticImages    []string               `json:"static_images"`
	AnimationType   string                 `json:"animation_type"`   // character, scene, transition
	AnimationStyle  string                 `json:"animation_style"`  // smooth, dramatic, subtle
	Duration        float64                `json:"duration"`         // 动画时长(秒)
	FrameRate       int                    `json:"frame_rate"`       // 帧率
	Quality         string                 `json:"quality"`          // high, medium, low
	Effects         []AnimationEffect      `json:"effects"`          // 动画效果
	OutputPath      string                 `json:"output_path"`      // 输出路径
}

// AnimationEffect 动画效果
type AnimationEffect struct {
	Type       string                 `json:"type"`       // fade, zoom, pan, rotate, morph
	StartTime  float64                `json:"start_time"` // 开始时间(秒)
	Duration   float64                `json:"duration"`   // 持续时间(秒)
	Parameters map[string]interface{} `json:"parameters"` // 效果参数
}

// AnimationResult 动画生成结果
type AnimationResult struct {
	ProjectID      uuid.UUID `json:"project_id"`
	OutputPath     string    `json:"output_path"`
	Duration       float64   `json:"duration"`
	FrameCount     int       `json:"frame_count"`
	Quality        float64   `json:"quality"`
	ProcessingTime time.Duration `json:"processing_time"`
	Effects        []string  `json:"effects"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// GenerateAdvancedAnimation 生成高级动画
func (aae *AdvancedAnimationEngine) GenerateAdvancedAnimation(ctx context.Context, req *AnimationRequest) (*AnimationResult, error) {
	startTime := time.Now()
	log.Printf("[AdvancedAnimationEngine] 开始生成高级动画: %s", req.ProjectID)

	// 1. 创建工作目录
	workDir := filepath.Join(req.OutputPath, "animation_work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("创建工作目录失败: %w", err)
	}
	defer os.RemoveAll(workDir) // 清理临时文件

	// 2. 预处理静态图像
	processedImages, err := aae.preprocessImages(ctx, req.StaticImages, workDir)
	if err != nil {
		return nil, fmt.Errorf("预处理图像失败: %w", err)
	}

	// 3. 生成动画序列
	animationSequence, err := aae.generateAnimationSequence(ctx, req, processedImages, workDir)
	if err != nil {
		return nil, fmt.Errorf("生成动画序列失败: %w", err)
	}

	// 4. 应用动画效果
	enhancedSequence, err := aae.applyAnimationEffects(ctx, animationSequence, req.Effects, workDir)
	if err != nil {
		return nil, fmt.Errorf("应用动画效果失败: %w", err)
	}

	// 5. 生成过渡效果
	finalSequence, err := aae.generateTransitions(ctx, enhancedSequence, req, workDir)
	if err != nil {
		return nil, fmt.Errorf("生成过渡效果失败: %w", err)
	}

	// 6. 合成最终动画
	outputPath, err := aae.composeAnimation(ctx, finalSequence, req, workDir)
	if err != nil {
		return nil, fmt.Errorf("合成动画失败: %w", err)
	}

	// 7. 质量评估
	quality, err := aae.qualityController.EvaluateAnimationQuality(outputPath)
	if err != nil {
		log.Printf("[AdvancedAnimationEngine] 质量评估失败: %v", err)
		quality = 0.7 // 默认质量分数
	}

	result := &AnimationResult{
		ProjectID:      req.ProjectID,
		OutputPath:     outputPath,
		Duration:       req.Duration,
		FrameCount:     len(finalSequence),
		Quality:        quality,
		ProcessingTime: time.Since(startTime),
		Effects:        aae.extractEffectNames(req.Effects),
		Metadata: map[string]interface{}{
			"frame_rate":      req.FrameRate,
			"animation_type":  req.AnimationType,
			"animation_style": req.AnimationStyle,
			"quality_level":   req.Quality,
		},
	}

	log.Printf("[AdvancedAnimationEngine] 动画生成完成: 时长=%.2fs, 帧数=%d, 质量=%.2f", 
		result.Duration, result.FrameCount, result.Quality)

	return result, nil
}

// preprocessImages 预处理图像
func (aae *AdvancedAnimationEngine) preprocessImages(ctx context.Context, images []string, workDir string) ([]string, error) {
	log.Printf("[AdvancedAnimationEngine] 预处理 %d 张图像", len(images))

	processedImages := make([]string, 0, len(images))

	for i, imagePath := range images {
		// 检查图像是否存在
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			log.Printf("[AdvancedAnimationEngine] 图像不存在，跳过: %s", imagePath)
			continue
		}

		// 标准化图像尺寸和格式
		processedPath := filepath.Join(workDir, fmt.Sprintf("processed_%d.png", i))
		err := aae.standardizeImage(imagePath, processedPath)
		if err != nil {
			log.Printf("[AdvancedAnimationEngine] 标准化图像失败: %s, 错误: %v", imagePath, err)
			continue
		}

		processedImages = append(processedImages, processedPath)
	}

	log.Printf("[AdvancedAnimationEngine] 图像预处理完成: %d/%d", len(processedImages), len(images))
	return processedImages, nil
}

// standardizeImage 标准化图像
func (aae *AdvancedAnimationEngine) standardizeImage(inputPath, outputPath string) error {
	// 使用FFmpeg标准化图像尺寸和格式
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath, 
		"-vf", "scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2",
		"-pix_fmt", "rgba", outputPath)
	
	return cmd.Run()
}

// generateAnimationSequence 生成动画序列
func (aae *AdvancedAnimationEngine) generateAnimationSequence(ctx context.Context, req *AnimationRequest, images []string, workDir string) ([]string, error) {
	log.Printf("[AdvancedAnimationEngine] 生成动画序列")

	var sequence []string

	switch req.AnimationType {
	case "character":
		seq, err := aae.generateCharacterAnimation(ctx, images, req, workDir)
		if err != nil {
			return nil, err
		}
		sequence = seq

	case "scene":
		seq, err := aae.generateSceneAnimation(ctx, images, req, workDir)
		if err != nil {
			return nil, err
		}
		sequence = seq

	case "transition":
		seq, err := aae.generateTransitionAnimation(ctx, images, req, workDir)
		if err != nil {
			return nil, err
		}
		sequence = seq

	default:
		// 默认生成基础动画序列
		seq, err := aae.generateBasicAnimation(ctx, images, req, workDir)
		if err != nil {
			return nil, err
		}
		sequence = seq
	}

	log.Printf("[AdvancedAnimationEngine] 动画序列生成完成: %d 帧", len(sequence))
	return sequence, nil
}

// generateCharacterAnimation 生成角色动画
func (aae *AdvancedAnimationEngine) generateCharacterAnimation(ctx context.Context, images []string, req *AnimationRequest, workDir string) ([]string, error) {
	log.Printf("[AdvancedAnimationEngine] 生成角色动画")

	sequence := make([]string, 0)
	framesPerImage := int(req.Duration * float64(req.FrameRate) / float64(len(images)))

	for i, imagePath := range images {
		// 为每张图像生成动画帧
		animatedFrames, err := aae.motionGenerator.GenerateCharacterMotion(imagePath, framesPerImage, workDir, i)
		if err != nil {
			log.Printf("[AdvancedAnimationEngine] 生成角色动作失败: %v", err)
			// 降级处理：复制静态图像
			for j := 0; j < framesPerImage; j++ {
				sequence = append(sequence, imagePath)
			}
		} else {
			sequence = append(sequence, animatedFrames...)
		}
	}

	return sequence, nil
}

// generateSceneAnimation 生成场景动画
func (aae *AdvancedAnimationEngine) generateSceneAnimation(ctx context.Context, images []string, req *AnimationRequest, workDir string) ([]string, error) {
	log.Printf("[AdvancedAnimationEngine] 生成场景动画")

	sequence := make([]string, 0)
	framesPerImage := int(req.Duration * float64(req.FrameRate) / float64(len(images)))

	for i, imagePath := range images {
		// 为每张图像生成场景动画
		animatedFrames, err := aae.motionGenerator.GenerateSceneMotion(imagePath, framesPerImage, workDir, i)
		if err != nil {
			log.Printf("[AdvancedAnimationEngine] 生成场景动画失败: %v", err)
			// 降级处理：复制静态图像
			for j := 0; j < framesPerImage; j++ {
				sequence = append(sequence, imagePath)
			}
		} else {
			sequence = append(sequence, animatedFrames...)
		}
	}

	return sequence, nil
}

// generateTransitionAnimation 生成过渡动画
func (aae *AdvancedAnimationEngine) generateTransitionAnimation(ctx context.Context, images []string, req *AnimationRequest, workDir string) ([]string, error) {
	log.Printf("[AdvancedAnimationEngine] 生成过渡动画")

	if len(images) < 2 {
		return images, nil
	}

	sequence := make([]string, 0)
	transitionFrames := int(float64(req.FrameRate) * 0.5) // 0.5秒过渡

	for i := 0; i < len(images)-1; i++ {
		// 添加当前图像
		staticFrames := int(req.Duration * float64(req.FrameRate) / float64(len(images)))
		for j := 0; j < staticFrames-transitionFrames; j++ {
			sequence = append(sequence, images[i])
		}

		// 生成过渡帧
		transitionSeq, err := aae.frameInterpolator.InterpolateFrames(images[i], images[i+1], transitionFrames, workDir, i)
		if err != nil {
			log.Printf("[AdvancedAnimationEngine] 生成过渡帧失败: %v", err)
			// 降级处理：直接切换
			for j := 0; j < transitionFrames; j++ {
				sequence = append(sequence, images[i])
			}
		} else {
			sequence = append(sequence, transitionSeq...)
		}
	}

	// 添加最后一张图像
	lastImageFrames := int(req.Duration * float64(req.FrameRate) / float64(len(images)))
	for j := 0; j < lastImageFrames; j++ {
		sequence = append(sequence, images[len(images)-1])
	}

	return sequence, nil
}

// generateBasicAnimation 生成基础动画
func (aae *AdvancedAnimationEngine) generateBasicAnimation(ctx context.Context, images []string, req *AnimationRequest, workDir string) ([]string, error) {
	log.Printf("[AdvancedAnimationEngine] 生成基础动画")

	sequence := make([]string, 0)
	framesPerImage := int(req.Duration * float64(req.FrameRate) / float64(len(images)))

	for _, imagePath := range images {
		// 为每张图像生成指定数量的帧
		for j := 0; j < framesPerImage; j++ {
			sequence = append(sequence, imagePath)
		}
	}

	return sequence, nil
}

// applyAnimationEffects 应用动画效果
func (aae *AdvancedAnimationEngine) applyAnimationEffects(ctx context.Context, sequence []string, effects []AnimationEffect, workDir string) ([]string, error) {
	if len(effects) == 0 {
		return sequence, nil
	}

	log.Printf("[AdvancedAnimationEngine] 应用 %d 个动画效果", len(effects))

	enhancedSequence := make([]string, len(sequence))
	copy(enhancedSequence, sequence)

	for _, effect := range effects {
		var err error
		enhancedSequence, err = aae.effectsProcessor.ApplyEffect(enhancedSequence, effect, workDir)
		if err != nil {
			log.Printf("[AdvancedAnimationEngine] 应用效果失败: %s, 错误: %v", effect.Type, err)
			continue
		}
	}

	return enhancedSequence, nil
}

// generateTransitions 生成过渡效果
func (aae *AdvancedAnimationEngine) generateTransitions(ctx context.Context, sequence []string, req *AnimationRequest, workDir string) ([]string, error) {
	log.Printf("[AdvancedAnimationEngine] 生成过渡效果")

	// 根据动画风格选择过渡类型
	var transitionType string
	switch req.AnimationStyle {
	case "smooth":
		transitionType = "fade"
	case "dramatic":
		transitionType = "zoom"
	case "subtle":
		transitionType = "dissolve"
	default:
		transitionType = "fade"
	}

	finalSequence, err := aae.transitionEngine.ApplyTransitions(sequence, transitionType, workDir)
	if err != nil {
		log.Printf("[AdvancedAnimationEngine] 应用过渡效果失败: %v", err)
		return sequence, nil // 返回原序列
	}

	return finalSequence, nil
}

// composeAnimation 合成动画
func (aae *AdvancedAnimationEngine) composeAnimation(ctx context.Context, sequence []string, req *AnimationRequest, workDir string) (string, error) {
	log.Printf("[AdvancedAnimationEngine] 合成动画: %d 帧", len(sequence))

	// 创建帧列表文件
	frameListPath := filepath.Join(workDir, "frame_list.txt")
	frameListContent := strings.Builder{}
	
	for _, framePath := range sequence {
		frameListContent.WriteString(fmt.Sprintf("file '%s'\n", framePath))
		frameListContent.WriteString("duration 0.033333\n") // 30fps
	}

	err := os.WriteFile(frameListPath, []byte(frameListContent.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("创建帧列表文件失败: %w", err)
	}

	// 输出路径
	outputPath := filepath.Join(req.OutputPath, fmt.Sprintf("animation_%s.mp4", req.ProjectID.String()))

	// 使用FFmpeg合成视频
	var cmd *exec.Cmd
	switch req.Quality {
	case "high":
		cmd = exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", frameListPath,
			"-c:v", "libx264", "-preset", "slow", "-crf", "18", "-pix_fmt", "yuv420p",
			"-r", fmt.Sprintf("%d", req.FrameRate), outputPath)
	case "medium":
		cmd = exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", frameListPath,
			"-c:v", "libx264", "-preset", "medium", "-crf", "23", "-pix_fmt", "yuv420p",
			"-r", fmt.Sprintf("%d", req.FrameRate), outputPath)
	default: // low
		cmd = exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", frameListPath,
			"-c:v", "libx264", "-preset", "fast", "-crf", "28", "-pix_fmt", "yuv420p",
			"-r", fmt.Sprintf("%d", req.FrameRate), outputPath)
	}

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("FFmpeg合成失败: %w", err)
	}

	return outputPath, nil
}

// 辅助方法
func (aae *AdvancedAnimationEngine) extractEffectNames(effects []AnimationEffect) []string {
	names := make([]string, len(effects))
	for i, effect := range effects {
		names[i] = effect.Type
	}
	return names
}

// CreateDefaultAnimationRequest 创建默认动画请求
func CreateDefaultAnimationRequest(projectID uuid.UUID, images []string, outputPath string) *AnimationRequest {
	return &AnimationRequest{
		ProjectID:      projectID,
		StaticImages:   images,
		AnimationType:  "character",
		AnimationStyle: "smooth",
		Duration:       float64(len(images)) * 2.0, // 每张图2秒
		FrameRate:      30,
		Quality:        "medium",
		Effects:        []AnimationEffect{},
		OutputPath:     outputPath,
	}
}

// AddFadeEffect 添加淡入淡出效果
func (req *AnimationRequest) AddFadeEffect(startTime, duration float64) {
	effect := AnimationEffect{
		Type:      "fade",
		StartTime: startTime,
		Duration:  duration,
		Parameters: map[string]interface{}{
			"fade_type": "in_out",
			"strength":  0.8,
		},
	}
	req.Effects = append(req.Effects, effect)
}

// AddZoomEffect 添加缩放效果
func (req *AnimationRequest) AddZoomEffect(startTime, duration float64, zoomFactor float64) {
	effect := AnimationEffect{
		Type:      "zoom",
		StartTime: startTime,
		Duration:  duration,
		Parameters: map[string]interface{}{
			"zoom_factor": zoomFactor,
			"zoom_center": []float64{0.5, 0.5}, // 中心缩放
		},
	}
	req.Effects = append(req.Effects, effect)
}

// AddPanEffect 添加平移效果
func (req *AnimationRequest) AddPanEffect(startTime, duration float64, direction string) {
	effect := AnimationEffect{
		Type:      "pan",
		StartTime: startTime,
		Duration:  duration,
		Parameters: map[string]interface{}{
			"direction": direction, // left, right, up, down
			"distance":  0.1,       // 移动距离（相对于图像尺寸）
		},
	}
	req.Effects = append(req.Effects, effect)
}
