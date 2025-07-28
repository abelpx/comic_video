package animation

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

	"comic_video/internal/domain/entity"
	"comic_video/internal/service/ai"
)

// CharacterAnimator 角色动画器
type CharacterAnimator struct {
	aiService    *ai.Service
	ffmpegPath   string
	tempDir      string
	animationDB  *AnimationDatabase
}

// AnimationDatabase 动画数据库
type AnimationDatabase struct {
	Expressions map[string][]AnimationFrame `json:"expressions"`
	Gestures    map[string][]AnimationFrame `json:"gestures"`
	Movements   map[string][]AnimationFrame `json:"movements"`
}

// AnimationFrame 动画帧
type AnimationFrame struct {
	Type        string                 `json:"type"`        // expression/gesture/movement
	Parameters  map[string]interface{} `json:"parameters"`  // 动画参数
	Duration    float64                `json:"duration"`    // 持续时间
	Transition  string                 `json:"transition"`  // 过渡效果
}

// CharacterAnimationRequest 角色动画请求
type CharacterAnimationRequest struct {
	Character    *entity.Character `json:"character"`
	Dialogue     string           `json:"dialogue"`
	Emotion      string           `json:"emotion"`
	Action       string           `json:"action"`
	Duration     float64          `json:"duration"`
	Style        string           `json:"style"`
	ImagePath    string           `json:"image_path"`
}

// CharacterAnimationResult 角色动画结果
type CharacterAnimationResult struct {
	AnimatedFrames []string          `json:"animated_frames"`
	Duration       float64           `json:"duration"`
	FrameRate      int               `json:"frame_rate"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// NewCharacterAnimator 创建角色动画器
func NewCharacterAnimator(aiService *ai.Service, ffmpegPath, tempDir string) *CharacterAnimator {
	animator := &CharacterAnimator{
		aiService:  aiService,
		ffmpegPath: ffmpegPath,
		tempDir:    tempDir,
		animationDB: &AnimationDatabase{
			Expressions: make(map[string][]AnimationFrame),
			Gestures:    make(map[string][]AnimationFrame),
			Movements:   make(map[string][]AnimationFrame),
		},
	}
	
	// 初始化动画数据库
	animator.initializeAnimationDatabase()
	
	return animator
}

// AnimateCharacter 为角色添加动画
func (a *CharacterAnimator) AnimateCharacter(ctx context.Context, req *CharacterAnimationRequest) (*CharacterAnimationResult, error) {
	log.Printf("[CharacterAnimator] 开始为角色 %s 生成动画", req.Character.Name)
	
	// 1. 分析对话和情感
	animationPlan, err := a.analyzeDialogueForAnimation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("分析对话失败: %w", err)
	}
	
	// 2. 生成动画帧
	animatedFrames, err := a.generateAnimationFrames(ctx, req, animationPlan)
	if err != nil {
		return nil, fmt.Errorf("生成动画帧失败: %w", err)
	}
	
	// 3. 应用动画效果
	processedFrames, err := a.applyAnimationEffects(ctx, animatedFrames, req)
	if err != nil {
		return nil, fmt.Errorf("应用动画效果失败: %w", err)
	}
	
	result := &CharacterAnimationResult{
		AnimatedFrames: processedFrames,
		Duration:       req.Duration,
		FrameRate:      24,
		Metadata: map[string]interface{}{
			"character_name": req.Character.Name,
			"emotion":        req.Emotion,
			"action":         req.Action,
			"style":          req.Style,
		},
	}
	
	log.Printf("[CharacterAnimator] 角色动画生成完成: %d帧", len(processedFrames))
	return result, nil
}

// analyzeDialogueForAnimation 分析对话生成动画计划
func (a *CharacterAnimator) analyzeDialogueForAnimation(ctx context.Context, req *CharacterAnimationRequest) (*AnimationPlan, error) {
	prompt := fmt.Sprintf(`分析以下对话，生成角色动画计划：

角色：%s
对话："%s"
情感：%s
动作：%s
风格：%s

请分析对话中的：
1. 关键情感变化点
2. 需要的面部表情
3. 手势和身体动作
4. 语音同步点

输出JSON格式：
{
  "emotion_changes": [
    {"time": 0.0, "emotion": "neutral", "intensity": 0.5},
    {"time": 2.0, "emotion": "happy", "intensity": 0.8}
  ],
  "expressions": [
    {"time": 0.0, "type": "neutral", "duration": 1.0},
    {"time": 1.0, "type": "smile", "duration": 2.0}
  ],
  "gestures": [
    {"time": 1.5, "type": "hand_wave", "duration": 1.0}
  ],
  "lip_sync": [
    {"time": 0.5, "phoneme": "a", "duration": 0.2},
    {"time": 0.7, "phoneme": "o", "duration": 0.3}
  ]
}`, req.Character.Name, req.Dialogue, req.Emotion, req.Action, req.Style)

	response, err := a.aiService.GenerateText(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	var plan AnimationPlan
	cleanedResponse := a.cleanJSONResponse(response)
	if err := json.Unmarshal([]byte(cleanedResponse), &plan); err != nil {
		// 如果解析失败，返回默认计划
		return a.createDefaultAnimationPlan(req), nil
	}

	return &plan, nil
}

// AnimationPlan 动画计划
type AnimationPlan struct {
	EmotionChanges []EmotionChange `json:"emotion_changes"`
	Expressions    []Expression    `json:"expressions"`
	Gestures       []Gesture       `json:"gestures"`
	LipSync        []LipSyncPoint  `json:"lip_sync"`
}

// EmotionChange 情感变化
type EmotionChange struct {
	Time      float64 `json:"time"`
	Emotion   string  `json:"emotion"`
	Intensity float64 `json:"intensity"`
}

// Expression 表情
type Expression struct {
	Time     float64 `json:"time"`
	Type     string  `json:"type"`
	Duration float64 `json:"duration"`
}

// Gesture 手势
type Gesture struct {
	Time     float64 `json:"time"`
	Type     string  `json:"type"`
	Duration float64 `json:"duration"`
}

// LipSyncPoint 口型同步点
type LipSyncPoint struct {
	Time     float64 `json:"time"`
	Phoneme  string  `json:"phoneme"`
	Duration float64 `json:"duration"`
}

// generateAnimationFrames 生成动画帧
func (a *CharacterAnimator) generateAnimationFrames(ctx context.Context, req *CharacterAnimationRequest, plan *AnimationPlan) ([]string, error) {
	workDir := filepath.Join(a.tempDir, fmt.Sprintf("animation_%s_%d", req.Character.Name, time.Now().Unix()))
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, err
	}
	
	frameRate := 24
	totalFrames := int(req.Duration * float64(frameRate))
	var frames []string
	
	for i := 0; i < totalFrames; i++ {
		currentTime := float64(i) / float64(frameRate)
		
		// 确定当前时间的动画状态
		expression := a.getExpressionAtTime(plan.Expressions, currentTime)
		gesture := a.getGestureAtTime(plan.Gestures, currentTime)
		lipSync := a.getLipSyncAtTime(plan.LipSync, currentTime)
		
		// 生成当前帧
		framePath := filepath.Join(workDir, fmt.Sprintf("frame_%04d.jpg", i))
		if err := a.generateSingleFrame(ctx, req, expression, gesture, lipSync, framePath); err != nil {
			log.Printf("[CharacterAnimator] 生成帧失败: %v", err)
			// 使用原图作为备用
			if err := a.copyFile(req.ImagePath, framePath); err != nil {
				return nil, err
			}
		}
		
		frames = append(frames, framePath)
	}
	
	return frames, nil
}

// generateSingleFrame 生成单个动画帧
func (a *CharacterAnimator) generateSingleFrame(ctx context.Context, req *CharacterAnimationRequest, expression, gesture, lipSync string, outputPath string) error {
	// 使用FFmpeg应用动画变换
	filters := a.buildFrameFilters(expression, gesture, lipSync, req.Style)
	
	cmd := exec.CommandContext(ctx, a.ffmpegPath,
		"-i", req.ImagePath,
		"-vf", filters,
		"-q:v", "2",
		"-y", outputPath)
	
	return cmd.Run()
}

// buildFrameFilters 构建帧滤镜
func (a *CharacterAnimator) buildFrameFilters(expression, gesture, lipSync, style string) string {
	var filters []string
	
	// 基础滤镜
	filters = append(filters, "scale=1920:1080")
	
	// 表情滤镜
	switch expression {
	case "smile":
		filters = append(filters, "eq=brightness=0.05:saturation=1.1")
	case "sad":
		filters = append(filters, "eq=brightness=-0.05:saturation=0.9")
	case "angry":
		filters = append(filters, "eq=contrast=1.2:saturation=1.2")
	case "surprised":
		filters = append(filters, "eq=brightness=0.1:contrast=1.1")
	}
	
	// 手势效果（通过轻微的变换模拟）
	switch gesture {
	case "hand_wave":
		filters = append(filters, "rotate=0.05*sin(2*PI*t):fillcolor=none")
	case "nod":
		filters = append(filters, "crop=iw:ih*0.98:0:ih*0.01")
	case "shake_head":
		filters = append(filters, "crop=iw*0.98:ih:iw*0.01:0")
	}
	
	// 口型同步（简单的亮度变化模拟）
	if lipSync != "" {
		filters = append(filters, "eq=brightness=0.02")
	}
	
	return strings.Join(filters, ",")
}

// applyAnimationEffects 应用动画效果
func (a *CharacterAnimator) applyAnimationEffects(ctx context.Context, frames []string, req *CharacterAnimationRequest) ([]string, error) {
	var processedFrames []string
	
	for i, frame := range frames {
		outputPath := strings.Replace(frame, ".jpg", "_processed.jpg", 1)
		
		// 应用风格化效果
		if err := a.applyStyleEffects(ctx, frame, outputPath, req.Style); err != nil {
			log.Printf("[CharacterAnimator] 应用风格效果失败: %v", err)
			processedFrames = append(processedFrames, frame) // 使用原帧
		} else {
			processedFrames = append(processedFrames, outputPath)
		}
		
		if i%10 == 0 {
			log.Printf("[CharacterAnimator] 处理进度: %d/%d", i+1, len(frames))
		}
	}
	
	return processedFrames, nil
}

// applyStyleEffects 应用风格效果
func (a *CharacterAnimator) applyStyleEffects(ctx context.Context, inputPath, outputPath, style string) error {
	var filters string
	
	switch strings.ToLower(style) {
	case "anime":
		filters = "eq=saturation=1.3:contrast=1.2,unsharp=5:5:1.0:5:5:0.0"
	case "realistic":
		filters = "eq=gamma=1.1:contrast=1.05"
	case "cartoon":
		filters = "eq=saturation=1.4:brightness=0.05,unsharp=3:3:0.8:3:3:0.4"
	default:
		filters = "eq=contrast=1.05"
	}
	
	cmd := exec.CommandContext(ctx, a.ffmpegPath,
		"-i", inputPath,
		"-vf", filters,
		"-q:v", "2",
		"-y", outputPath)
	
	return cmd.Run()
}

// 辅助方法
func (a *CharacterAnimator) getExpressionAtTime(expressions []Expression, time float64) string {
	for _, expr := range expressions {
		if time >= expr.Time && time < expr.Time+expr.Duration {
			return expr.Type
		}
	}
	return "neutral"
}

func (a *CharacterAnimator) getGestureAtTime(gestures []Gesture, time float64) string {
	for _, gesture := range gestures {
		if time >= gesture.Time && time < gesture.Time+gesture.Duration {
			return gesture.Type
		}
	}
	return ""
}

func (a *CharacterAnimator) getLipSyncAtTime(lipSync []LipSyncPoint, time float64) string {
	for _, sync := range lipSync {
		if time >= sync.Time && time < sync.Time+sync.Duration {
			return sync.Phoneme
		}
	}
	return ""
}

func (a *CharacterAnimator) createDefaultAnimationPlan(req *CharacterAnimationRequest) *AnimationPlan {
	return &AnimationPlan{
		EmotionChanges: []EmotionChange{
			{Time: 0.0, Emotion: req.Emotion, Intensity: 0.7},
		},
		Expressions: []Expression{
			{Time: 0.0, Type: req.Emotion, Duration: req.Duration},
		},
		Gestures: []Gesture{},
		LipSync:  []LipSyncPoint{},
	}
}

func (a *CharacterAnimator) cleanJSONResponse(response string) string {
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

func (a *CharacterAnimator) copyFile(src, dst string) error {
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

func (a *CharacterAnimator) initializeAnimationDatabase() {
	// 初始化基础动画数据
	a.animationDB.Expressions["happy"] = []AnimationFrame{
		{Type: "expression", Parameters: map[string]interface{}{"brightness": 0.05, "saturation": 1.1}, Duration: 1.0},
	}
	a.animationDB.Expressions["sad"] = []AnimationFrame{
		{Type: "expression", Parameters: map[string]interface{}{"brightness": -0.05, "saturation": 0.9}, Duration: 1.0},
	}
	// 可以继续添加更多动画数据
}
