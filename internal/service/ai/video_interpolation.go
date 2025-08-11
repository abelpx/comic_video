package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// VideoInterpolator 视频插帧器
type VideoInterpolator struct {
	Endpoint string
	Timeout  time.Duration
	Model    string // "rife", "film", "flavr"
}

// NewVideoInterpolator 创建视频插帧器
func NewVideoInterpolator(endpoint string, model string) *VideoInterpolator {
	return &VideoInterpolator{
		Endpoint: endpoint,
		Timeout:  5 * time.Minute,
		Model:    model,
	}
}

// InterpolationRequest 插帧请求
type InterpolationRequest struct {
	Images        []string `json:"images"`         // base64编码的图片序列
	TargetFPS     int      `json:"target_fps"`     // 目标帧率
	Model         string   `json:"model"`          // 插帧模型
	Quality       string   `json:"quality"`        // "fast", "balanced", "high"
	SmoothFactor  float64  `json:"smooth_factor"`  // 平滑因子
	MotionBlur    bool     `json:"motion_blur"`    // 是否添加运动模糊
	SceneDetect   bool     `json:"scene_detect"`   // 场景检测
}

// InterpolationResult 插帧结果
type InterpolationResult struct {
	VideoData    []byte        `json:"video_data"`
	FrameCount   int           `json:"frame_count"`
	Duration     time.Duration `json:"duration"`
	FPS          int           `json:"fps"`
	Resolution   string        `json:"resolution"`
	FileSize     int64         `json:"file_size"`
	ProcessTime  time.Duration `json:"process_time"`
}

// InterpolateFrames 图片序列插帧生成视频
func (vi *VideoInterpolator) InterpolateFrames(
	ctx context.Context,
	images [][]byte,
	targetFPS int,
	quality string,
) (*InterpolationResult, error) {
	
	startTime := time.Now()
	
	// 将图片转换为base64
	base64Images := make([]string, len(images))
	for i, img := range images {
		base64Images[i] = encodeBase64(img)
	}
	
	req := InterpolationRequest{
		Images:       base64Images,
		TargetFPS:    targetFPS,
		Model:        vi.Model,
		Quality:      quality,
		SmoothFactor: vi.getSmoothFactor(quality),
		MotionBlur:   quality != "fast",
		SceneDetect:  true,
	}
	
	log.Printf("[VideoInterpolator] 开始插帧: %d张图片 -> %dFPS", len(images), targetFPS)
	
	result, err := vi.processInterpolation(ctx, req)
	if err != nil {
		return nil, err
	}
	
	result.ProcessTime = time.Since(startTime)
	log.Printf("[VideoInterpolator] 插帧完成: %d帧, 耗时: %v", result.FrameCount, result.ProcessTime)
	
	return result, nil
}

// processInterpolation 处理插帧请求
func (vi *VideoInterpolator) processInterpolation(ctx context.Context, req InterpolationRequest) (*InterpolationResult, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", vi.Endpoint+"/interpolate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: vi.Timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("插帧失败 (状态码: %d): %s", resp.StatusCode, string(body))
	}
	
	var result InterpolationResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}
	
	return &result, nil
}

// AdvancedInterpolate 高级插帧 (支持场景转换)
func (vi *VideoInterpolator) AdvancedInterpolate(
	ctx context.Context,
	scenes [][]byte, // 每个场景的关键帧
	transitions []string, // 转场类型: "fade", "slide", "zoom", "morph"
	targetFPS int,
	quality string,
) (*InterpolationResult, error) {
	
	log.Printf("[VideoInterpolator] 高级插帧: %d个场景", len(scenes))
	
	// 为每个场景生成过渡帧
	allFrames := [][]byte{}
	
	for i, scene := range scenes {
		// 添加场景帧
		allFrames = append(allFrames, scene)
		
		// 如果不是最后一个场景，添加转场
		if i < len(scenes)-1 {
			transitionType := "fade"
			if i < len(transitions) {
				transitionType = transitions[i]
			}
			
			transitionFrames, err := vi.generateTransition(ctx, scene, scenes[i+1], transitionType)
			if err != nil {
				log.Printf("[VideoInterpolator] 转场生成失败: %v", err)
				// 使用简单的淡入淡出
				transitionFrames = vi.generateSimpleTransition(scene, scenes[i+1])
			}
			
			allFrames = append(allFrames, transitionFrames...)
		}
	}
	
	// 对所有帧进行插帧
	return vi.InterpolateFrames(ctx, allFrames, targetFPS, quality)
}

// generateTransition 生成转场效果
func (vi *VideoInterpolator) generateTransition(ctx context.Context, frame1, frame2 []byte, transitionType string) ([][]byte, error) {
	req := map[string]interface{}{
		"frame1":          encodeBase64(frame1),
		"frame2":          encodeBase64(frame2),
		"transition_type": transitionType,
		"frame_count":     8, // 生成8帧转场
	}
	
	jsonData, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", vi.Endpoint+"/transition", bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: vi.Timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Frames []string `json:"frames"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	frames := make([][]byte, len(result.Frames))
	for i, frameB64 := range result.Frames {
		frameData, _ := decodeBase64Image(frameB64)
		frames[i] = frameData
	}
	
	return frames, nil
}

// generateSimpleTransition 生成简单转场 (本地处理)
func (vi *VideoInterpolator) generateSimpleTransition(frame1, frame2 []byte) [][]byte {
	// 简单的线性插值转场
	// 这里可以实现基本的淡入淡出效果
	frames := make([][]byte, 4)
	
	// 简化实现：返回原始帧的副本
	for i := 0; i < 4; i++ {
		if i < 2 {
			frames[i] = make([]byte, len(frame1))
			copy(frames[i], frame1)
		} else {
			frames[i] = make([]byte, len(frame2))
			copy(frames[i], frame2)
		}
	}
	
	return frames
}

// getSmoothFactor 获取平滑因子
func (vi *VideoInterpolator) getSmoothFactor(quality string) float64 {
	switch quality {
	case "fast":
		return 0.3
	case "balanced":
		return 0.6
	case "high":
		return 0.9
	default:
		return 0.6
	}
}

// OptimizeForAnimation 针对动画优化
func (vi *VideoInterpolator) OptimizeForAnimation(
	ctx context.Context,
	keyFrames [][]byte,
	animationType string, // "smooth", "dramatic", "action"
	targetFPS int,
) (*InterpolationResult, error) {
	
	quality := "balanced"

	switch animationType {
	case "smooth":
		quality = "high"
		if targetFPS > 30 {
			targetFPS = 30 // 平滑动画不需要太高帧率
		}
	case "dramatic":
		quality = "balanced"
		if targetFPS > 24 {
			targetFPS = 24 // 电影感帧率
		}
	case "action":
		quality = "fast"
		targetFPS = min(targetFPS, 60) // 高帧率动作
	}
	
	log.Printf("[VideoInterpolator] 动画优化: %s模式, %dFPS", animationType, targetFPS)
	
	return vi.InterpolateFrames(ctx, keyFrames, targetFPS, quality)
}

// GetInterpolationCapabilities 获取插帧能力
func (vi *VideoInterpolator) GetInterpolationCapabilities() map[string]interface{} {
	capabilities := map[string]interface{}{
		"models": []string{"rife", "film", "flavr"},
		"max_fps": 60,
		"max_resolution": "1920x1080",
		"supported_transitions": []string{"fade", "slide", "zoom", "morph", "wipe"},
		"quality_modes": []string{"fast", "balanced", "high"},
		"max_input_frames": 100,
		"estimated_processing_time": map[string]string{
			"fast":     "0.5x实时",
			"balanced": "1.0x实时", 
			"high":     "2.0x实时",
		},
	}
	
	switch vi.Model {
	case "rife":
		capabilities["description"] = "RIFE - 实时插帧，速度快，质量好"
		capabilities["best_for"] = "实时应用，游戏录制"
	case "film":
		capabilities["description"] = "FILM - 电影级插帧，质量最高"
		capabilities["best_for"] = "电影制作，高质量动画"
	case "flavr":
		capabilities["description"] = "FLAVR - 平衡性能和质量"
		capabilities["best_for"] = "通用视频处理"
	}
	
	return capabilities
}

// ResourceUsageEstimate 资源使用估算
type ResourceUsageEstimate struct {
	GPUMemory    float64       `json:"gpu_memory_gb"`
	ProcessTime  time.Duration `json:"process_time"`
	CPUUsage     float64       `json:"cpu_usage_percent"`
	OutputSize   int64         `json:"output_size_mb"`
}

// EstimateResourceUsage 估算资源使用
func (vi *VideoInterpolator) EstimateResourceUsage(frameCount int, targetFPS int, quality string) ResourceUsageEstimate {
	baseMemory := 2.0 // 基础显存使用
	baseTime := time.Duration(frameCount) * 100 * time.Millisecond
	
	switch quality {
	case "fast":
		return ResourceUsageEstimate{
			GPUMemory:   baseMemory * 0.7,
			ProcessTime: baseTime * 1,
			CPUUsage:    30.0,
			OutputSize:  int64(float64(frameCount * targetFPS) * 0.5), // MB
		}
	case "balanced":
		return ResourceUsageEstimate{
			GPUMemory:   baseMemory * 1.0,
			ProcessTime: baseTime * 2,
			CPUUsage:    50.0,
			OutputSize:  int64(float64(frameCount * targetFPS) * 0.8),
		}
	case "high":
		return ResourceUsageEstimate{
			GPUMemory:   baseMemory * 1.5,
			ProcessTime: baseTime * 4,
			CPUUsage:    70.0,
			OutputSize:  int64(float64(frameCount * targetFPS) * 1.2),
		}
	default:
		return ResourceUsageEstimate{
			GPUMemory:   baseMemory,
			ProcessTime: baseTime * 2,
			CPUUsage:    50.0,
			OutputSize:  int64(float64(frameCount * targetFPS) * 0.8),
		}
	}
}
