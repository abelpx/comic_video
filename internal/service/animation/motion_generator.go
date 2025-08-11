package animation

import (
	"fmt"
	"log"
	"math"
	"os/exec"
	"path/filepath"
)

// MotionGenerator 动作生成器
type MotionGenerator struct {
	// 可以集成AI动作生成模型
}

// NewMotionGenerator 创建动作生成器
func NewMotionGenerator() *MotionGenerator {
	return &MotionGenerator{}
}

// GenerateCharacterMotion 生成角色动作
func (mg *MotionGenerator) GenerateCharacterMotion(imagePath string, frameCount int, workDir string, sequenceIndex int) ([]string, error) {
	log.Printf("[MotionGenerator] 生成角色动作: %d 帧", frameCount)

	frames := make([]string, 0, frameCount)

	// 生成微妙的角色动作
	for i := 0; i < frameCount; i++ {
		framePath := filepath.Join(workDir, fmt.Sprintf("character_motion_%d_%d.png", sequenceIndex, i))
		
		// 计算动作参数
		progress := float64(i) / float64(frameCount-1)
		
		// 生成轻微的呼吸动作
		breathingScale := 1.0 + 0.02*math.Sin(progress*math.Pi*4) // 轻微缩放
		
		// 生成眨眼动作（随机时机）
		blinkAlpha := 1.0
		if i == frameCount/3 || i == frameCount*2/3 {
			blinkAlpha = 0.95 // 轻微眨眼效果
		}

		err := mg.applyCharacterMotion(imagePath, framePath, breathingScale, blinkAlpha)
		if err != nil {
			log.Printf("[MotionGenerator] 生成角色动作帧失败: %v", err)
			// 降级处理：复制原图
			err = mg.copyImage(imagePath, framePath)
			if err != nil {
				return nil, err
			}
		}

		frames = append(frames, framePath)
	}

	return frames, nil
}

// GenerateSceneMotion 生成场景动作
func (mg *MotionGenerator) GenerateSceneMotion(imagePath string, frameCount int, workDir string, sequenceIndex int) ([]string, error) {
	log.Printf("[MotionGenerator] 生成场景动作: %d 帧", frameCount)

	frames := make([]string, 0, frameCount)

	for i := 0; i < frameCount; i++ {
		framePath := filepath.Join(workDir, fmt.Sprintf("scene_motion_%d_%d.png", sequenceIndex, i))
		
		// 计算动作参数
		progress := float64(i) / float64(frameCount-1)
		
		// 生成轻微的摄像机移动
		panX := 0.01 * math.Sin(progress*math.Pi*2) // 水平摇摆
		panY := 0.005 * math.Cos(progress*math.Pi*3) // 垂直摇摆
		
		// 生成轻微的缩放变化
		zoomFactor := 1.0 + 0.01*math.Sin(progress*math.Pi*1.5)

		err := mg.applySceneMotion(imagePath, framePath, panX, panY, zoomFactor)
		if err != nil {
			log.Printf("[MotionGenerator] 生成场景动作帧失败: %v", err)
			// 降级处理：复制原图
			err = mg.copyImage(imagePath, framePath)
			if err != nil {
				return nil, err
			}
		}

		frames = append(frames, framePath)
	}

	return frames, nil
}

// applyCharacterMotion 应用角色动作
func (mg *MotionGenerator) applyCharacterMotion(inputPath, outputPath string, scale, alpha float64) error {
	// 使用FFmpeg应用角色动作效果
	scaleFilter := fmt.Sprintf("scale=iw*%.3f:ih*%.3f", scale, scale)
	alphaFilter := fmt.Sprintf("format=rgba,colorchannelmixer=aa=%.3f", alpha)
	
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath,
		"-vf", fmt.Sprintf("%s,%s", scaleFilter, alphaFilter),
		"-pix_fmt", "rgba", outputPath)
	
	return cmd.Run()
}

// applySceneMotion 应用场景动作
func (mg *MotionGenerator) applySceneMotion(inputPath, outputPath string, panX, panY, zoom float64) error {
	// 计算平移和缩放参数
	scaleFilter := fmt.Sprintf("scale=iw*%.3f:ih*%.3f", zoom, zoom)
	cropFilter := fmt.Sprintf("crop=iw/%.3f:ih/%.3f:iw*%.3f:ih*%.3f", zoom, zoom, panX, panY)
	
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath,
		"-vf", fmt.Sprintf("%s,%s,scale=1920:1080", scaleFilter, cropFilter),
		"-pix_fmt", "rgba", outputPath)
	
	return cmd.Run()
}

// copyImage 复制图像
func (mg *MotionGenerator) copyImage(srcPath, dstPath string) error {
	cmd := exec.Command("cp", srcPath, dstPath)
	return cmd.Run()
}

// FrameInterpolator 帧插值器
type FrameInterpolator struct {
	// 可以集成AI帧插值模型
}

// NewFrameInterpolator 创建帧插值器
func NewFrameInterpolator() *FrameInterpolator {
	return &FrameInterpolator{}
}

// InterpolateFrames 插值生成中间帧
func (fi *FrameInterpolator) InterpolateFrames(startImage, endImage string, frameCount int, workDir string, sequenceIndex int) ([]string, error) {
	log.Printf("[FrameInterpolator] 插值生成 %d 帧", frameCount)

	frames := make([]string, 0, frameCount)

	for i := 0; i < frameCount; i++ {
		framePath := filepath.Join(workDir, fmt.Sprintf("interpolated_%d_%d.png", sequenceIndex, i))
		
		// 计算插值权重
		weight := float64(i) / float64(frameCount-1)
		
		// 生成插值帧
		err := fi.blendImages(startImage, endImage, weight, framePath)
		if err != nil {
			log.Printf("[FrameInterpolator] 生成插值帧失败: %v", err)
			// 降级处理：使用简单的切换
			if weight < 0.5 {
				err = fi.copyImage(startImage, framePath)
			} else {
				err = fi.copyImage(endImage, framePath)
			}
			if err != nil {
				return nil, err
			}
		}

		frames = append(frames, framePath)
	}

	return frames, nil
}

// blendImages 混合两张图像
func (fi *FrameInterpolator) blendImages(image1, image2 string, weight float64, outputPath string) error {
	// 使用FFmpeg混合图像
	alpha1 := 1.0 - weight
	alpha2 := weight
	
	cmd := exec.Command("ffmpeg", "-y", "-i", image1, "-i", image2,
		"-filter_complex", fmt.Sprintf("[0]format=rgba,colorchannelmixer=aa=%.3f[a];[1]format=rgba,colorchannelmixer=aa=%.3f[b];[a][b]overlay", alpha1, alpha2),
		"-pix_fmt", "rgba", outputPath)
	
	return cmd.Run()
}

// copyImage 复制图像
func (fi *FrameInterpolator) copyImage(srcPath, dstPath string) error {
	cmd := exec.Command("cp", srcPath, dstPath)
	return cmd.Run()
}

// EffectsProcessor 效果处理器
type EffectsProcessor struct {
	// 可以集成各种视觉效果
}

// NewEffectsProcessor 创建效果处理器
func NewEffectsProcessor() *EffectsProcessor {
	return &EffectsProcessor{}
}

// ApplyEffect 应用效果
func (ep *EffectsProcessor) ApplyEffect(sequence []string, effect AnimationEffect, workDir string) ([]string, error) {
	log.Printf("[EffectsProcessor] 应用效果: %s", effect.Type)

	switch effect.Type {
	case "fade":
		return ep.applyFadeEffect(sequence, effect, workDir)
	case "zoom":
		return ep.applyZoomEffect(sequence, effect, workDir)
	case "pan":
		return ep.applyPanEffect(sequence, effect, workDir)
	case "rotate":
		return ep.applyRotateEffect(sequence, effect, workDir)
	default:
		log.Printf("[EffectsProcessor] 未知效果类型: %s", effect.Type)
		return sequence, nil
	}
}

// applyFadeEffect 应用淡入淡出效果
func (ep *EffectsProcessor) applyFadeEffect(sequence []string, effect AnimationEffect, workDir string) ([]string, error) {
	enhancedSequence := make([]string, len(sequence))
	
	for i, framePath := range sequence {
		outputPath := filepath.Join(workDir, fmt.Sprintf("fade_effect_%d.png", i))
		
		// 计算淡入淡出强度
		progress := float64(i) / float64(len(sequence)-1)
		alpha := ep.calculateFadeAlpha(progress, effect)
		
		err := ep.applyAlphaEffect(framePath, outputPath, alpha)
		if err != nil {
			log.Printf("[EffectsProcessor] 应用淡入淡出效果失败: %v", err)
			enhancedSequence[i] = framePath // 使用原图
		} else {
			enhancedSequence[i] = outputPath
		}
	}
	
	return enhancedSequence, nil
}

// applyZoomEffect 应用缩放效果
func (ep *EffectsProcessor) applyZoomEffect(sequence []string, effect AnimationEffect, workDir string) ([]string, error) {
	enhancedSequence := make([]string, len(sequence))
	
	zoomFactor, _ := effect.Parameters["zoom_factor"].(float64)
	if zoomFactor == 0 {
		zoomFactor = 1.1 // 默认缩放因子
	}
	
	for i, framePath := range sequence {
		outputPath := filepath.Join(workDir, fmt.Sprintf("zoom_effect_%d.png", i))
		
		// 计算缩放程度
		progress := float64(i) / float64(len(sequence)-1)
		currentZoom := 1.0 + (zoomFactor-1.0)*progress
		
		err := ep.applyZoomTransform(framePath, outputPath, currentZoom)
		if err != nil {
			log.Printf("[EffectsProcessor] 应用缩放效果失败: %v", err)
			enhancedSequence[i] = framePath // 使用原图
		} else {
			enhancedSequence[i] = outputPath
		}
	}
	
	return enhancedSequence, nil
}

// applyPanEffect 应用平移效果
func (ep *EffectsProcessor) applyPanEffect(sequence []string, effect AnimationEffect, workDir string) ([]string, error) {
	enhancedSequence := make([]string, len(sequence))
	
	direction, _ := effect.Parameters["direction"].(string)
	distance, _ := effect.Parameters["distance"].(float64)
	if distance == 0 {
		distance = 0.1 // 默认移动距离
	}
	
	for i, framePath := range sequence {
		outputPath := filepath.Join(workDir, fmt.Sprintf("pan_effect_%d.png", i))
		
		// 计算平移距离
		progress := float64(i) / float64(len(sequence)-1)
		panX, panY := ep.calculatePanOffset(direction, distance, progress)
		
		err := ep.applyPanTransform(framePath, outputPath, panX, panY)
		if err != nil {
			log.Printf("[EffectsProcessor] 应用平移效果失败: %v", err)
			enhancedSequence[i] = framePath // 使用原图
		} else {
			enhancedSequence[i] = outputPath
		}
	}
	
	return enhancedSequence, nil
}

// applyRotateEffect 应用旋转效果
func (ep *EffectsProcessor) applyRotateEffect(sequence []string, effect AnimationEffect, workDir string) ([]string, error) {
	enhancedSequence := make([]string, len(sequence))
	
	maxAngle, _ := effect.Parameters["max_angle"].(float64)
	if maxAngle == 0 {
		maxAngle = 5.0 // 默认最大旋转角度
	}
	
	for i, framePath := range sequence {
		outputPath := filepath.Join(workDir, fmt.Sprintf("rotate_effect_%d.png", i))
		
		// 计算旋转角度
		progress := float64(i) / float64(len(sequence)-1)
		angle := maxAngle * math.Sin(progress*math.Pi*2)
		
		err := ep.applyRotateTransform(framePath, outputPath, angle)
		if err != nil {
			log.Printf("[EffectsProcessor] 应用旋转效果失败: %v", err)
			enhancedSequence[i] = framePath // 使用原图
		} else {
			enhancedSequence[i] = outputPath
		}
	}
	
	return enhancedSequence, nil
}

// 辅助方法
func (ep *EffectsProcessor) calculateFadeAlpha(progress float64, effect AnimationEffect) float64 {
	fadeType, _ := effect.Parameters["fade_type"].(string)
	strength, _ := effect.Parameters["strength"].(float64)
	if strength == 0 {
		strength = 1.0
	}
	
	switch fadeType {
	case "in":
		return progress * strength
	case "out":
		return (1.0 - progress) * strength
	case "in_out":
		if progress < 0.5 {
			return progress * 2 * strength
		} else {
			return (2 - progress*2) * strength
		}
	default:
		return strength
	}
}

func (ep *EffectsProcessor) calculatePanOffset(direction string, distance, progress float64) (float64, float64) {
	switch direction {
	case "left":
		return -distance * progress, 0
	case "right":
		return distance * progress, 0
	case "up":
		return 0, -distance * progress
	case "down":
		return 0, distance * progress
	default:
		return 0, 0
	}
}

func (ep *EffectsProcessor) applyAlphaEffect(inputPath, outputPath string, alpha float64) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath,
		"-vf", fmt.Sprintf("format=rgba,colorchannelmixer=aa=%.3f", alpha),
		"-pix_fmt", "rgba", outputPath)
	return cmd.Run()
}

func (ep *EffectsProcessor) applyZoomTransform(inputPath, outputPath string, zoom float64) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath,
		"-vf", fmt.Sprintf("scale=iw*%.3f:ih*%.3f,crop=1920:1080", zoom, zoom),
		"-pix_fmt", "rgba", outputPath)
	return cmd.Run()
}

func (ep *EffectsProcessor) applyPanTransform(inputPath, outputPath string, panX, panY float64) error {
	offsetX := int(1920 * panX)
	offsetY := int(1080 * panY)
	
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath,
		"-vf", fmt.Sprintf("crop=1920:1080:%d:%d", offsetX, offsetY),
		"-pix_fmt", "rgba", outputPath)
	return cmd.Run()
}

func (ep *EffectsProcessor) applyRotateTransform(inputPath, outputPath string, angle float64) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath,
		"-vf", fmt.Sprintf("rotate=%.3f*PI/180", angle),
		"-pix_fmt", "rgba", outputPath)
	return cmd.Run()
}
