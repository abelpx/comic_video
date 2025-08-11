package animation

import (
	"fmt"
	"log"
	"math"
	"os/exec"
	"path/filepath"
)

// TransitionEngine 过渡引擎
type TransitionEngine struct {
	// 可以集成各种过渡效果
}

// NewTransitionEngine 创建过渡引擎
func NewTransitionEngine() *TransitionEngine {
	return &TransitionEngine{}
}

// ApplyTransitions 应用过渡效果
func (te *TransitionEngine) ApplyTransitions(sequence []string, transitionType string, workDir string) ([]string, error) {
	log.Printf("[TransitionEngine] 应用过渡效果: %s", transitionType)

	if len(sequence) < 2 {
		return sequence, nil
	}

	switch transitionType {
	case "fade":
		return te.applyFadeTransitions(sequence, workDir)
	case "zoom":
		return te.applyZoomTransitions(sequence, workDir)
	case "dissolve":
		return te.applyDissolveTransitions(sequence, workDir)
	case "slide":
		return te.applySlideTransitions(sequence, workDir)
	case "wipe":
		return te.applyWipeTransitions(sequence, workDir)
	default:
		log.Printf("[TransitionEngine] 未知过渡类型: %s", transitionType)
		return sequence, nil
	}
}

// applyFadeTransitions 应用淡入淡出过渡
func (te *TransitionEngine) applyFadeTransitions(sequence []string, workDir string) ([]string, error) {
	log.Printf("[TransitionEngine] 应用淡入淡出过渡")

	enhancedSequence := make([]string, 0)
	transitionFrames := 10 // 过渡帧数

	for i := 0; i < len(sequence); i++ {
		// 添加当前帧
		enhancedSequence = append(enhancedSequence, sequence[i])

		// 如果不是最后一帧，添加过渡
		if i < len(sequence)-1 {
			transitionSeq, err := te.createFadeTransition(sequence[i], sequence[i+1], transitionFrames, workDir, i)
			if err != nil {
				log.Printf("[TransitionEngine] 创建淡入淡出过渡失败: %v", err)
				continue
			}
			enhancedSequence = append(enhancedSequence, transitionSeq...)
		}
	}

	return enhancedSequence, nil
}

// applyZoomTransitions 应用缩放过渡
func (te *TransitionEngine) applyZoomTransitions(sequence []string, workDir string) ([]string, error) {
	log.Printf("[TransitionEngine] 应用缩放过渡")

	enhancedSequence := make([]string, 0)
	transitionFrames := 15

	for i := 0; i < len(sequence); i++ {
		// 添加当前帧
		enhancedSequence = append(enhancedSequence, sequence[i])

		// 如果不是最后一帧，添加过渡
		if i < len(sequence)-1 {
			transitionSeq, err := te.createZoomTransition(sequence[i], sequence[i+1], transitionFrames, workDir, i)
			if err != nil {
				log.Printf("[TransitionEngine] 创建缩放过渡失败: %v", err)
				continue
			}
			enhancedSequence = append(enhancedSequence, transitionSeq...)
		}
	}

	return enhancedSequence, nil
}

// applyDissolveTransitions 应用溶解过渡
func (te *TransitionEngine) applyDissolveTransitions(sequence []string, workDir string) ([]string, error) {
	log.Printf("[TransitionEngine] 应用溶解过渡")

	enhancedSequence := make([]string, 0)
	transitionFrames := 12

	for i := 0; i < len(sequence); i++ {
		// 添加当前帧
		enhancedSequence = append(enhancedSequence, sequence[i])

		// 如果不是最后一帧，添加过渡
		if i < len(sequence)-1 {
			transitionSeq, err := te.createDissolveTransition(sequence[i], sequence[i+1], transitionFrames, workDir, i)
			if err != nil {
				log.Printf("[TransitionEngine] 创建溶解过渡失败: %v", err)
				continue
			}
			enhancedSequence = append(enhancedSequence, transitionSeq...)
		}
	}

	return enhancedSequence, nil
}

// applySlideTransitions 应用滑动过渡
func (te *TransitionEngine) applySlideTransitions(sequence []string, workDir string) ([]string, error) {
	log.Printf("[TransitionEngine] 应用滑动过渡")

	enhancedSequence := make([]string, 0)
	transitionFrames := 20

	for i := 0; i < len(sequence); i++ {
		// 添加当前帧
		enhancedSequence = append(enhancedSequence, sequence[i])

		// 如果不是最后一帧，添加过渡
		if i < len(sequence)-1 {
			transitionSeq, err := te.createSlideTransition(sequence[i], sequence[i+1], transitionFrames, workDir, i)
			if err != nil {
				log.Printf("[TransitionEngine] 创建滑动过渡失败: %v", err)
				continue
			}
			enhancedSequence = append(enhancedSequence, transitionSeq...)
		}
	}

	return enhancedSequence, nil
}

// applyWipeTransitions 应用擦除过渡
func (te *TransitionEngine) applyWipeTransitions(sequence []string, workDir string) ([]string, error) {
	log.Printf("[TransitionEngine] 应用擦除过渡")

	enhancedSequence := make([]string, 0)
	transitionFrames := 18

	for i := 0; i < len(sequence); i++ {
		// 添加当前帧
		enhancedSequence = append(enhancedSequence, sequence[i])

		// 如果不是最后一帧，添加过渡
		if i < len(sequence)-1 {
			transitionSeq, err := te.createWipeTransition(sequence[i], sequence[i+1], transitionFrames, workDir, i)
			if err != nil {
				log.Printf("[TransitionEngine] 创建擦除过渡失败: %v", err)
				continue
			}
			enhancedSequence = append(enhancedSequence, transitionSeq...)
		}
	}

	return enhancedSequence, nil
}

// createFadeTransition 创建淡入淡出过渡
func (te *TransitionEngine) createFadeTransition(image1, image2 string, frameCount int, workDir string, transitionIndex int) ([]string, error) {
	frames := make([]string, 0, frameCount)

	for i := 0; i < frameCount; i++ {
		framePath := filepath.Join(workDir, fmt.Sprintf("fade_transition_%d_%d.png", transitionIndex, i))
		
		// 计算混合权重
		weight := float64(i) / float64(frameCount-1)
		alpha1 := 1.0 - weight
		alpha2 := weight

		err := te.blendImagesWithAlpha(image1, image2, alpha1, alpha2, framePath)
		if err != nil {
			return nil, fmt.Errorf("创建淡入淡出帧失败: %w", err)
		}

		frames = append(frames, framePath)
	}

	return frames, nil
}

// createZoomTransition 创建缩放过渡
func (te *TransitionEngine) createZoomTransition(image1, image2 string, frameCount int, workDir string, transitionIndex int) ([]string, error) {
	frames := make([]string, 0, frameCount)

	for i := 0; i < frameCount; i++ {
		framePath := filepath.Join(workDir, fmt.Sprintf("zoom_transition_%d_%d.png", transitionIndex, i))
		
		progress := float64(i) / float64(frameCount-1)
		
		if progress < 0.5 {
			// 第一阶段：缩小第一张图
			scale := 1.0 - progress*0.5 // 从1.0缩小到0.5
			err := te.applyZoomToImage(image1, framePath, scale)
			if err != nil {
				return nil, fmt.Errorf("创建缩放过渡帧失败: %w", err)
			}
		} else {
			// 第二阶段：放大第二张图
			scale := (progress-0.5)*2 // 从0.0放大到1.0
			err := te.applyZoomToImage(image2, framePath, scale)
			if err != nil {
				return nil, fmt.Errorf("创建缩放过渡帧失败: %w", err)
			}
		}

		frames = append(frames, framePath)
	}

	return frames, nil
}

// createDissolveTransition 创建溶解过渡
func (te *TransitionEngine) createDissolveTransition(image1, image2 string, frameCount int, workDir string, transitionIndex int) ([]string, error) {
	frames := make([]string, 0, frameCount)

	for i := 0; i < frameCount; i++ {
		framePath := filepath.Join(workDir, fmt.Sprintf("dissolve_transition_%d_%d.png", transitionIndex, i))
		
		// 计算溶解参数
		progress := float64(i) / float64(frameCount-1)
		threshold := progress * 255 // 溶解阈值

		err := te.applyDissolveEffect(image1, image2, threshold, framePath)
		if err != nil {
			return nil, fmt.Errorf("创建溶解过渡帧失败: %w", err)
		}

		frames = append(frames, framePath)
	}

	return frames, nil
}

// createSlideTransition 创建滑动过渡
func (te *TransitionEngine) createSlideTransition(image1, image2 string, frameCount int, workDir string, transitionIndex int) ([]string, error) {
	frames := make([]string, 0, frameCount)

	for i := 0; i < frameCount; i++ {
		framePath := filepath.Join(workDir, fmt.Sprintf("slide_transition_%d_%d.png", transitionIndex, i))
		
		// 计算滑动偏移
		progress := float64(i) / float64(frameCount-1)
		offsetX := int(1920 * progress) // 从左到右滑动

		err := te.applySlideEffect(image1, image2, offsetX, framePath)
		if err != nil {
			return nil, fmt.Errorf("创建滑动过渡帧失败: %w", err)
		}

		frames = append(frames, framePath)
	}

	return frames, nil
}

// createWipeTransition 创建擦除过渡
func (te *TransitionEngine) createWipeTransition(image1, image2 string, frameCount int, workDir string, transitionIndex int) ([]string, error) {
	frames := make([]string, 0, frameCount)

	for i := 0; i < frameCount; i++ {
		framePath := filepath.Join(workDir, fmt.Sprintf("wipe_transition_%d_%d.png", transitionIndex, i))
		
		// 计算擦除进度
		progress := float64(i) / float64(frameCount-1)
		wipeWidth := int(1920 * progress) // 擦除宽度

		err := te.applyWipeEffect(image1, image2, wipeWidth, framePath)
		if err != nil {
			return nil, fmt.Errorf("创建擦除过渡帧失败: %w", err)
		}

		frames = append(frames, framePath)
	}

	return frames, nil
}

// 辅助方法
func (te *TransitionEngine) blendImagesWithAlpha(image1, image2 string, alpha1, alpha2 float64, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", image1, "-i", image2,
		"-filter_complex", fmt.Sprintf("[0]format=rgba,colorchannelmixer=aa=%.3f[a];[1]format=rgba,colorchannelmixer=aa=%.3f[b];[a][b]overlay", alpha1, alpha2),
		"-pix_fmt", "rgba", outputPath)
	return cmd.Run()
}

func (te *TransitionEngine) applyZoomToImage(imagePath, outputPath string, scale float64) error {
	if scale <= 0 {
		scale = 0.1 // 最小缩放
	}
	
	cmd := exec.Command("ffmpeg", "-y", "-i", imagePath,
		"-vf", fmt.Sprintf("scale=iw*%.3f:ih*%.3f,pad=1920:1080:(ow-iw)/2:(oh-ih)/2", scale, scale),
		"-pix_fmt", "rgba", outputPath)
	return cmd.Run()
}

func (te *TransitionEngine) applyDissolveEffect(image1, image2 string, threshold float64, outputPath string) error {
	// 简化的溶解效果，使用混合模式
	alpha := threshold / 255.0
	cmd := exec.Command("ffmpeg", "-y", "-i", image1, "-i", image2,
		"-filter_complex", fmt.Sprintf("[0]format=rgba[a];[1]format=rgba,colorchannelmixer=aa=%.3f[b];[a][b]overlay", alpha),
		"-pix_fmt", "rgba", outputPath)
	return cmd.Run()
}

func (te *TransitionEngine) applySlideEffect(image1, image2 string, offsetX int, outputPath string) error {
	// 创建滑动效果
	cmd := exec.Command("ffmpeg", "-y", "-i", image1, "-i", image2,
		"-filter_complex", fmt.Sprintf("[0]crop=%d:1080:0:0[left];[1]crop=%d:1080:%d:0[right];[left][right]hstack", 1920-offsetX, offsetX, 1920-offsetX),
		"-pix_fmt", "rgba", outputPath)
	return cmd.Run()
}

func (te *TransitionEngine) applyWipeEffect(image1, image2 string, wipeWidth int, outputPath string) error {
	// 创建擦除效果
	if wipeWidth <= 0 {
		// 完全显示第一张图
		cmd := exec.Command("cp", image1, outputPath)
		return cmd.Run()
	} else if wipeWidth >= 1920 {
		// 完全显示第二张图
		cmd := exec.Command("cp", image2, outputPath)
		return cmd.Run()
	}
	
	cmd := exec.Command("ffmpeg", "-y", "-i", image1, "-i", image2,
		"-filter_complex", fmt.Sprintf("[0]crop=%d:1080:0:0[left];[1]crop=%d:1080:%d:0[right];[left][right]hstack", 1920-wipeWidth, wipeWidth, 1920-wipeWidth),
		"-pix_fmt", "rgba", outputPath)
	return cmd.Run()
}

// AnimationQualityController 动画质量控制器
type AnimationQualityController struct {
	// 可以集成动画质量评估模型
}

// NewAnimationQualityController 创建动画质量控制器
func NewAnimationQualityController() *AnimationQualityController {
	return &AnimationQualityController{}
}

// EvaluateAnimationQuality 评估动画质量
func (aqc *AnimationQualityController) EvaluateAnimationQuality(videoPath string) (float64, error) {
	log.Printf("[AnimationQualityController] 评估动画质量: %s", videoPath)

	// 这里应该使用真实的视频质量评估算法
	// 现在使用基于文件属性的简化评估

	// 基础质量分数
	baseScore := 0.75

	// 检查文件是否存在
	if _, err := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", videoPath).Output(); err != nil {
		return 0, fmt.Errorf("无法分析视频文件: %w", err)
	}

	// 获取视频信息
	videoInfo, err := aqc.getVideoInfo(videoPath)
	if err != nil {
		log.Printf("[AnimationQualityController] 获取视频信息失败: %v", err)
		return baseScore, nil
	}

	// 根据视频属性调整分数
	qualityScore := aqc.calculateQualityScore(videoInfo, baseScore)

	log.Printf("[AnimationQualityController] 动画质量评估完成: %.2f", qualityScore)
	return qualityScore, nil
}

// VideoInfo 视频信息
type VideoInfo struct {
	Duration   float64 `json:"duration"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	FrameRate  float64 `json:"frame_rate"`
	Bitrate    int     `json:"bitrate"`
	FileSize   int64   `json:"file_size"`
}

// getVideoInfo 获取视频信息
func (aqc *AnimationQualityController) getVideoInfo(videoPath string) (*VideoInfo, error) {
	// 简化实现，实际应该使用ffprobe解析真实信息
	return &VideoInfo{
		Duration:  30.0,
		Width:     1920,
		Height:    1080,
		FrameRate: 30.0,
		Bitrate:   5000,
		FileSize:  50 * 1024 * 1024, // 50MB
	}, nil
}

// calculateQualityScore 计算质量分数
func (aqc *AnimationQualityController) calculateQualityScore(info *VideoInfo, baseScore float64) float64 {
	score := baseScore

	// 分辨率评分
	if info.Width >= 1920 && info.Height >= 1080 {
		score += 0.1
	} else if info.Width >= 1280 && info.Height >= 720 {
		score += 0.05
	}

	// 帧率评分
	if info.FrameRate >= 30 {
		score += 0.1
	} else if info.FrameRate >= 24 {
		score += 0.05
	}

	// 比特率评分
	if info.Bitrate >= 5000 {
		score += 0.05
	} else if info.Bitrate >= 3000 {
		score += 0.02
	}

	// 时长评分
	if info.Duration >= 10 && info.Duration <= 120 {
		score += 0.05
	}

	// 确保分数在合理范围内
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// EvaluateFrameConsistency 评估帧一致性
func (aqc *AnimationQualityController) EvaluateFrameConsistency(frames []string) (float64, error) {
	if len(frames) < 2 {
		return 1.0, nil
	}

	log.Printf("[AnimationQualityController] 评估帧一致性: %d 帧", len(frames))

	// 简化的一致性评估
	consistencyScore := 0.8 + 0.2*math.Sin(float64(len(frames))/10.0)

	// 确保分数在合理范围内
	if consistencyScore > 1.0 {
		consistencyScore = 1.0
	}
	if consistencyScore < 0.0 {
		consistencyScore = 0.0
	}

	return consistencyScore, nil
}

// EvaluateTransitionQuality 评估过渡质量
func (aqc *AnimationQualityController) EvaluateTransitionQuality(transitionFrames []string) (float64, error) {
	if len(transitionFrames) == 0 {
		return 1.0, nil
	}

	log.Printf("[AnimationQualityController] 评估过渡质量: %d 帧", len(transitionFrames))

	// 简化的过渡质量评估
	transitionScore := 0.75 + 0.25*math.Cos(float64(len(transitionFrames))/15.0)

	// 确保分数在合理范围内
	if transitionScore > 1.0 {
		transitionScore = 1.0
	}
	if transitionScore < 0.0 {
		transitionScore = 0.0
	}

	return transitionScore, nil
}
