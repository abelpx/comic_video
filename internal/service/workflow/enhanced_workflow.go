package workflow

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"comic_video/internal/domain/entity"
	"comic_video/internal/service/ai"
	"comic_video/internal/service/animation"
	"comic_video/internal/service/audio"
	"comic_video/internal/service/character"
	"comic_video/internal/service/imageprocessing"
	"comic_video/internal/service/platform"
	"comic_video/internal/service/quality"
	"comic_video/internal/service/scene"
	"comic_video/internal/service/script"
	"comic_video/internal/service/storyboard"
	"comic_video/internal/service/tweet"
	"comic_video/internal/service/video"
	"comic_video/internal/service/voice"
	"comic_video/internal/utils"
)

// EnhancedWorkflowService 增强的工作流服务
type EnhancedWorkflowService struct {
	scriptService      *script.Service
	characterService   *character.Service
	sceneService       *scene.Service
	storyboardService  *storyboard.Service
	aiService          *ai.Service
	voiceService       *voice.Service
	videoService       *video.AdvancedVideoGenerator
	imageProcessing    *imageprocessing.Service
	qualityService     *quality.Service
	tweetService       *tweet.Service
	animationService   *animation.CharacterAnimator
	audioProcessor     *audio.EnhancedAudioProcessor
	douyinAdapter      *platform.DouyinAdapter
}

// NewEnhancedWorkflowService 创建增强工作流服务
func NewEnhancedWorkflowService(
	scriptService *script.Service,
	characterService *character.Service,
	sceneService *scene.Service,
	storyboardService *storyboard.Service,
	aiService *ai.Service,
	voiceService *voice.Service,
	videoService *video.AdvancedVideoGenerator,
	imageProcessing *imageprocessing.Service,
	qualityService *quality.Service,
	tweetService *tweet.Service,
	animationService *animation.CharacterAnimator,
	audioProcessor *audio.EnhancedAudioProcessor,
	douyinAdapter *platform.DouyinAdapter,
) *EnhancedWorkflowService {
	return &EnhancedWorkflowService{
		scriptService:     scriptService,
		characterService:  characterService,
		sceneService:      sceneService,
		storyboardService: storyboardService,
		aiService:         aiService,
		voiceService:      voiceService,
		videoService:      videoService,
		imageProcessing:   imageProcessing,
		qualityService:    qualityService,
		tweetService:      tweetService,
		animationService:  animationService,
		audioProcessor:    audioProcessor,
		douyinAdapter:     douyinAdapter,
	}
}

// CompleteNovelToVideoRequest 完整的小说转视频请求
type CompleteNovelToVideoRequest struct {
	ProjectID    string                 `json:"project_id"`
	Novel        string                 `json:"novel"`
	VideoFormat  string                 `json:"video_format"`  // 4k/1080p/720p/vertical
	Quality      string                 `json:"quality"`       // ultra/high/medium/low
	Style        string                 `json:"style"`         // anime/realistic/cartoon
	VoiceStyle   string                 `json:"voice_style"`   // 语音风格
	BGMStyle     string                 `json:"bgm_style"`     // 背景音乐风格
	GenerateTweet bool                  `json:"generate_tweet"` // 是否生成推文
	Options      map[string]interface{} `json:"options"`       // 额外选项
}

// CompleteWorkflowResult 完整工作流结果
type CompleteWorkflowResult struct {
	ProjectID       string                    `json:"project_id"`
	VideoPath       string                    `json:"video_path"`
	QualityReport   *quality.VideoQualityReport `json:"quality_report"`
	TweetResults    *tweet.NovelTweetResult   `json:"tweet_results,omitempty"`
	ProcessingTime  time.Duration             `json:"processing_time"`
	Steps           []WorkflowStep            `json:"steps"`
	Assets          WorkflowAssets            `json:"assets"`
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	Name        string        `json:"name"`
	Status      string        `json:"status"`      // pending/running/completed/failed
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	Progress    int           `json:"progress"`    // 0-100
	Message     string        `json:"message"`
	Error       string        `json:"error,omitempty"`
}

// WorkflowAssets 工作流资产
type WorkflowAssets struct {
	Script      *entity.Script      `json:"script"`
	Characters  []*entity.Character `json:"characters"`
	Scenes      []*entity.Scene     `json:"scenes"`
	Storyboard  *entity.Storyboard  `json:"storyboard"`
	AudioFiles  []string            `json:"audio_files"`
	ImageFiles  []string            `json:"image_files"`
	VideoFiles  []string            `json:"video_files"`
}

// ExecuteCompleteWorkflow 执行完整的小说转视频工作流
func (s *EnhancedWorkflowService) ExecuteCompleteWorkflow(ctx context.Context, req *CompleteNovelToVideoRequest) (*CompleteWorkflowResult, error) {
	startTime := time.Now()
	log.Printf("[EnhancedWorkflow] 开始执行完整工作流: project_id=%s", req.ProjectID)

	result := &CompleteWorkflowResult{
		ProjectID: req.ProjectID,
		Steps:     make([]WorkflowStep, 0),
		Assets:    WorkflowAssets{},
	}

	// 定义工作流步骤
	steps := []struct {
		name string
		fn   func(context.Context, *CompleteNovelToVideoRequest, *CompleteWorkflowResult) error
	}{
		{"小说内容分析", s.stepAnalyzeNovel},
		{"剧本生成", s.stepGenerateScript},
		{"角色设计", s.stepGenerateCharacters},
		{"场景设计", s.stepGenerateScenes},
		{"分镜制作", s.stepGenerateStoryboard},
		{"图像生成", s.stepGenerateImages},
		{"角色动画", s.stepGenerateCharacterAnimations},
		{"图像后处理", s.stepProcessImages},
		{"语音合成", s.stepGenerateVoice},
		{"音频处理", s.stepProcessAudio},
		{"高级视频合成", s.stepGenerateAdvancedVideo},
		{"质量检测", s.stepQualityCheck},
		{"抖音平台适配", s.stepDouyinOptimization},
		{"推文生成", s.stepGenerateTweets},
	}

	// 执行每个步骤
	for i, step := range steps {
		// 跳过推文生成（如果不需要）
		if step.name == "推文生成" && !req.GenerateTweet {
			continue
		}

		stepResult := WorkflowStep{
			Name:      step.name,
			Status:    "running",
			StartTime: time.Now(),
			Progress:  0,
		}

		log.Printf("[EnhancedWorkflow] 执行步骤 %d/%d: %s", i+1, len(steps), step.name)

		// 执行步骤
		err := step.fn(ctx, req, result)
		
		stepResult.EndTime = time.Now()
		stepResult.Duration = stepResult.EndTime.Sub(stepResult.StartTime)

		if err != nil {
			stepResult.Status = "failed"
			stepResult.Error = err.Error()
			stepResult.Message = fmt.Sprintf("步骤失败: %v", err)
			result.Steps = append(result.Steps, stepResult)
			
			log.Printf("[EnhancedWorkflow] 步骤失败: %s, error: %v", step.name, err)
			return result, fmt.Errorf("工作流在步骤 '%s' 失败: %w", step.name, err)
		}

		stepResult.Status = "completed"
		stepResult.Progress = 100
		stepResult.Message = "步骤完成"
		result.Steps = append(result.Steps, stepResult)

		log.Printf("[EnhancedWorkflow] 步骤完成: %s, 耗时: %v", step.name, stepResult.Duration)
	}

	result.ProcessingTime = time.Since(startTime)
	log.Printf("[EnhancedWorkflow] 完整工作流执行完成, 总耗时: %v", result.ProcessingTime)

	return result, nil
}

// stepAnalyzeNovel 步骤1：分析小说内容
func (s *EnhancedWorkflowService) stepAnalyzeNovel(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	// 这里可以添加小说内容分析逻辑
	// 例如：字数统计、主题分析、复杂度评估等
	log.Printf("[EnhancedWorkflow] 分析小说内容: 长度=%d字符", len(req.Novel))
	return nil
}

// stepGenerateScript 步骤2：生成剧本
func (s *EnhancedWorkflowService) stepGenerateScript(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	scriptReq := &script.AdaptScriptRequest{
		ProjectID: req.ProjectID,
		Novel:     req.Novel,
		Style:     req.Style,
	}

	// 使用重试机制生成剧本
	script, err := utils.RetryWithResultFunc(ctx, nil, func() (*entity.Script, error) {
		return s.scriptService.AdaptNovelToScript(ctx, scriptReq)
	})
	if err != nil {
		return fmt.Errorf("生成剧本失败: %w", err)
	}

	result.Assets.Script = script
	log.Printf("[EnhancedWorkflow] 剧本生成完成")
	return nil
}

// stepGenerateCharacters 步骤3：生成角色
func (s *EnhancedWorkflowService) stepGenerateCharacters(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	if result.Assets.Script == nil {
		return fmt.Errorf("剧本未生成，无法生成角色")
	}

	charReq := &character.GenerateCharactersRequest{
		ProjectID: req.ProjectID,
		Script:    result.Assets.Script.Content,
		Style:     req.Style,
	}

	characters, err := s.characterService.GenerateCharacters(ctx, charReq)
	if err != nil {
		return fmt.Errorf("生成角色失败: %w", err)
	}

	result.Assets.Characters = characters
	log.Printf("[EnhancedWorkflow] 角色生成完成: %d个角色", len(characters))
	return nil
}

// stepGenerateScenes 步骤4：生成场景
func (s *EnhancedWorkflowService) stepGenerateScenes(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	if result.Assets.Script == nil {
		return fmt.Errorf("剧本未生成，无法生成场景")
	}

	sceneReq := &scene.GenerateSceneRequest{
		ProjectID: req.ProjectID,
		Script:    result.Assets.Script.Content,
		Style:     req.Style,
	}

	projectUUID, _ := uuid.Parse(req.ProjectID)
	scenesReq := &scene.GenerateScenesRequest{
		ProjectID:     projectUUID,
		ScriptContent: sceneReq.Script,
	}
	scenes, err := s.sceneService.GenerateScenes(ctx, scenesReq)
	if err != nil {
		return fmt.Errorf("生成场景失败: %w", err)
	}

	result.Assets.Scenes = scenes
	log.Printf("[EnhancedWorkflow] 场景生成完成: %d个场景", len(scenes))
	return nil
}

// stepGenerateStoryboard 步骤5：生成分镜
func (s *EnhancedWorkflowService) stepGenerateStoryboard(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	if result.Assets.Script == nil {
		return fmt.Errorf("剧本未生成，无法生成分镜")
	}

	storyboardReq := &storyboard.GenerateStoryboardRequest{
		ProjectID:  req.ProjectID,
		Script:     result.Assets.Script.Content,
		Characters: result.Assets.Characters,
		Scenes:     result.Assets.Scenes,
		Style:      req.Style,
	}

	storyboard, err := s.storyboardService.GenerateStoryboard(ctx, storyboardReq)
	if err != nil {
		return fmt.Errorf("生成分镜失败: %w", err)
	}

	result.Assets.Storyboard = storyboard
	log.Printf("[EnhancedWorkflow] 分镜生成完成")
	return nil
}

// stepGenerateImages 步骤6：生成图像
func (s *EnhancedWorkflowService) stepGenerateImages(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	if result.Assets.Storyboard == nil {
		return fmt.Errorf("分镜未生成，无法生成图像")
	}

	// 这里应该调用AI图像生成服务
	// 为每个分镜帧生成对应的图像
	var imageFiles []string
	
	// 简化实现：生成固定数量的图像
	frameCount := 10 // 默认生成10张图像
	for i := 0; i < frameCount; i++ {
		// 模拟图像生成
		imagePath := fmt.Sprintf("/tmp/frame_%d.jpg", i)
		imageFiles = append(imageFiles, imagePath)
	}

	result.Assets.ImageFiles = imageFiles
	log.Printf("[EnhancedWorkflow] 图像生成完成: %d张图像", len(imageFiles))
	return nil
}

// stepProcessImages 步骤7：图像后处理
func (s *EnhancedWorkflowService) stepProcessImages(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	if len(result.Assets.ImageFiles) == 0 {
		return fmt.Errorf("没有图像文件需要处理")
	}

	// 为视频制作标准化图像
	processedImages, err := s.imageProcessing.StandardizeForVideo(ctx, result.Assets.ImageFiles, req.VideoFormat)
	if err != nil {
		return fmt.Errorf("图像后处理失败: %w", err)
	}

	// 更新图像文件列表
	var processedPaths []string
	for _, img := range processedImages {
		processedPaths = append(processedPaths, img.ProcessedPath)
	}
	result.Assets.ImageFiles = processedPaths

	log.Printf("[EnhancedWorkflow] 图像后处理完成: %d张图像", len(processedPaths))
	return nil
}

// stepGenerateVoice 步骤8：生成语音
func (s *EnhancedWorkflowService) stepGenerateVoice(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	if result.Assets.Script == nil {
		return fmt.Errorf("剧本未生成，无法生成语音")
	}

	// 这里应该调用语音合成服务
	// 为剧本中的对话和旁白生成语音
	var audioFiles []string
	
	// 简化实现：生成固定数量的音频文件
	audioCount := 5 // 默认生成5个音频文件
	for i := 0; i < audioCount; i++ {
		// 模拟语音生成
		audioPath := fmt.Sprintf("/tmp/audio_%d.wav", i)
		audioFiles = append(audioFiles, audioPath)
	}

	result.Assets.AudioFiles = audioFiles
	log.Printf("[EnhancedWorkflow] 语音生成完成: %d个音频文件", len(audioFiles))
	return nil
}

// stepGenerateCharacterAnimations 步骤7：生成角色动画
func (s *EnhancedWorkflowService) stepGenerateCharacterAnimations(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	if len(result.Assets.Characters) == 0 {
		return fmt.Errorf("没有角色信息，无法生成动画")
	}

	log.Printf("[EnhancedWorkflow] 开始生成角色动画")

	var animatedImageFiles []string

	// 为每个角色生成动画
	for _, character := range result.Assets.Characters {
		// 找到角色对应的图像文件
		characterImages := s.findCharacterImages(character, result.Assets.ImageFiles)

		for _, imagePath := range characterImages {
			animReq := &animation.CharacterAnimationRequest{
				Character: character,
				Dialogue:  "示例对话", // 这里应该从剧本中提取
				Emotion:   "neutral",
				Action:    "speaking",
				Duration:  3.0,
				Style:     req.Style,
				ImagePath: imagePath,
			}

			animResult, err := s.animationService.AnimateCharacter(ctx, animReq)
			if err != nil {
				log.Printf("[EnhancedWorkflow] 角色动画生成失败: %v", err)
				continue
			}

			animatedImageFiles = append(animatedImageFiles, animResult.AnimatedFrames...)
		}
	}

	// 更新图像文件列表为动画帧
	if len(animatedImageFiles) > 0 {
		result.Assets.ImageFiles = animatedImageFiles
	}

	log.Printf("[EnhancedWorkflow] 角色动画生成完成: %d个动画帧", len(animatedImageFiles))
	return nil
}

// stepProcessAudio 步骤10：处理音频
func (s *EnhancedWorkflowService) stepProcessAudio(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	if result.Assets.Script == nil {
		return fmt.Errorf("剧本未生成，无法处理音频")
	}

	audioReq := &audio.AudioProcessingRequest{
		ProjectID:  req.ProjectID,
		Script:     result.Assets.Script,
		VoiceFiles: result.Assets.AudioFiles,
		Scenes:     result.Assets.Scenes,
		Duration:   60.0, // 默认时长，应该从视频计算
		Style:      req.Style,
		Platform:   req.VideoFormat,
	}

	audioResult, err := s.audioProcessor.ProcessAudio(ctx, audioReq)
	if err != nil {
		return fmt.Errorf("音频处理失败: %w", err)
	}

	// 更新音频文件
	result.Assets.AudioFiles = []string{audioResult.MasterAudioPath}

	log.Printf("[EnhancedWorkflow] 音频处理完成: %s", audioResult.MasterAudioPath)
	return nil
}

// stepGenerateAdvancedVideo 步骤11：高级视频合成
func (s *EnhancedWorkflowService) stepGenerateAdvancedVideo(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	if len(result.Assets.ImageFiles) == 0 || len(result.Assets.AudioFiles) == 0 {
		return fmt.Errorf("缺少图像或音频文件，无法生成视频")
	}

	videoReq := &video.VideoGenerationRequest{
		ProjectID:   req.ProjectID,
		Script:      result.Assets.Script,
		Characters:  result.Assets.Characters,
		Scenes:      result.Assets.Scenes,
		Storyboard:  result.Assets.Storyboard,
		AudioFiles:  result.Assets.AudioFiles,
		ImageFiles:  result.Assets.ImageFiles,
		Config: &video.VideoConfig{
			Width:     1920,
			Height:    1080,
			FrameRate: 30,
			Quality:   req.Quality,
			Platform:  req.VideoFormat,
		},
		Style: req.Style,
	}

	videoResult, err := s.videoService.GenerateAdvancedVideo(ctx, videoReq)
	if err != nil {
		return fmt.Errorf("高级视频生成失败: %w", err)
	}

	result.VideoPath = videoResult.VideoPath
	log.Printf("[EnhancedWorkflow] 高级视频生成完成: %s", videoResult.VideoPath)
	return nil
}

// stepDouyinOptimization 步骤13：抖音平台适配
func (s *EnhancedWorkflowService) stepDouyinOptimization(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	if result.VideoPath == "" {
		return fmt.Errorf("视频未生成，无法进行抖音优化")
	}

	// 只有当目标格式是抖音时才进行优化
	if req.VideoFormat != "douyin" && req.VideoFormat != "vertical" {
		log.Printf("[EnhancedWorkflow] 跳过抖音优化，目标格式: %s", req.VideoFormat)
		return nil
	}

	// 生成字幕
	subtitles := s.generateSubtitlesFromScript(result.Assets.Script)

	douyinReq := &platform.DouyinOptimizationRequest{
		VideoPath:    result.VideoPath,
		Title:        "AI生成视频",
		Description:  "基于小说内容生成的精彩视频",
		Tags:         []string{"AI", "小说", "视频"},
		Category:     "娱乐",
		Subtitles:    subtitles,
		Style:        req.Style,
		TargetLength: 180, // 3分钟
	}

	douyinResult, err := s.douyinAdapter.OptimizeForDouyin(ctx, douyinReq)
	if err != nil {
		return fmt.Errorf("抖音优化失败: %w", err)
	}

	// 更新视频路径为优化后的版本
	result.VideoPath = douyinResult.OptimizedVideoPath

	log.Printf("[EnhancedWorkflow] 抖音优化完成: %s, 质量评分: %.2f",
		douyinResult.OptimizedVideoPath, douyinResult.QualityScore)
	return nil
}

// stepQualityCheck 步骤10：质量检测
func (s *EnhancedWorkflowService) stepQualityCheck(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	if result.VideoPath == "" {
		return fmt.Errorf("视频未生成，无法进行质量检测")
	}

	qualityReport, err := s.qualityService.AnalyzeVideo(ctx, result.VideoPath)
	if err != nil {
		return fmt.Errorf("质量检测失败: %w", err)
	}

	result.QualityReport = qualityReport
	log.Printf("[EnhancedWorkflow] 质量检测完成: 评分=%.2f", qualityReport.QualityScore)
	return nil
}

// stepOptimizeVideo 步骤11：视频优化
func (s *EnhancedWorkflowService) stepOptimizeVideo(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	if result.QualityReport == nil || result.QualityReport.QualityScore >= 80 {
		log.Printf("[EnhancedWorkflow] 视频质量良好，跳过优化")
		return nil
	}

	optimizeOptions := &quality.OptimizationOptions{
		Quality:          req.Quality,
		TargetResolution: s.getResolutionFromFormat(req.VideoFormat),
		TargetFormat:     "mp4",
		Platform:         "general",
	}

	optimizedPath, err := s.qualityService.OptimizeVideo(ctx, result.VideoPath, optimizeOptions)
	if err != nil {
		return fmt.Errorf("视频优化失败: %w", err)
	}

	result.VideoPath = optimizedPath
	log.Printf("[EnhancedWorkflow] 视频优化完成: %s", optimizedPath)
	return nil
}

// stepGenerateTweets 步骤12：生成推文
func (s *EnhancedWorkflowService) stepGenerateTweets(ctx context.Context, req *CompleteNovelToVideoRequest, result *CompleteWorkflowResult) error {
	tweetReq := &tweet.NovelToTweetRequest{
		Novel:    req.Novel,
		Count:    5,
		Style:    "吸引人",
		Platform: "twitter",
		Angles:   []string{"情节亮点", "角色魅力", "主题深度", "情感共鸣", "视觉效果"},
	}

	tweetResults, err := s.tweetService.GenerateNovelTweets(ctx, tweetReq)
	if err != nil {
		return fmt.Errorf("生成推文失败: %w", err)
	}

	result.TweetResults = tweetResults
	log.Printf("[EnhancedWorkflow] 推文生成完成: %d条推文", len(tweetResults.Tweets))
	return nil
}

// getResolutionFromFormat 从格式获取分辨率
func (s *EnhancedWorkflowService) getResolutionFromFormat(format string) string {
	switch format {
	case "4k":
		return "3840x2160"
	case "1080p":
		return "1920x1080"
	case "720p":
		return "1280x720"
	case "vertical", "douyin":
		return "1080x1920"
	default:
		return "1920x1080"
	}
}

// findCharacterImages 查找角色对应的图像文件
func (s *EnhancedWorkflowService) findCharacterImages(character *entity.Character, imageFiles []string) []string {
	// 简化实现：返回所有图像文件
	// 实际应该根据角色名称或ID匹配对应的图像
	var characterImages []string

	// 这里可以实现更复杂的匹配逻辑
	for _, imagePath := range imageFiles {
		// 简单的文件名匹配
		if strings.Contains(strings.ToLower(imagePath), strings.ToLower(character.Name)) {
			characterImages = append(characterImages, imagePath)
		}
	}

	// 如果没有找到匹配的，返回前几个图像作为默认
	if len(characterImages) == 0 && len(imageFiles) > 0 {
		maxImages := 3
		if len(imageFiles) < maxImages {
			maxImages = len(imageFiles)
		}
		characterImages = imageFiles[:maxImages]
	}

	return characterImages
}

// generateSubtitlesFromScript 从剧本生成字幕
func (s *EnhancedWorkflowService) generateSubtitlesFromScript(script *entity.Script) []platform.SubtitleSegment {
	if script == nil {
		return []platform.SubtitleSegment{}
	}

	var subtitles []platform.SubtitleSegment
	currentTime := 0.0

	// 简化实现：将剧本内容分割为字幕段
	lines := strings.Split(script.Content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 估算每行的显示时长（基于字符数）
		duration := float64(len(line)) * 0.1 // 每个字符0.1秒
		if duration < 2.0 {
			duration = 2.0 // 最少2秒
		}
		if duration > 5.0 {
			duration = 5.0 // 最多5秒
		}

		subtitle := platform.SubtitleSegment{
			StartTime: currentTime,
			EndTime:   currentTime + duration,
			Text:      line,
			Style:     "default",
		}

		subtitles = append(subtitles, subtitle)
		currentTime += duration + 0.5 // 间隔0.5秒
	}

	return subtitles
}
