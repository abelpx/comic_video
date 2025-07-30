package ai

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/minio"
	"comic_video/internal/repository/redis"
)

// Service AI服务
type Service struct {
	sdClient      *SDClient
	ollamaClient  *OllamaClient
	ttsClient     *TTSClient
	whisperClient *WhisperClient
	minioClient   minio.MinioClient
	redisClient   *redis.Client
}

// NewService 创建AI服务
func NewService(
	sdClient *SDClient,
	ollamaClient *OllamaClient,
	ttsClient *TTSClient,
	whisperClient *WhisperClient,
	minioClient minio.MinioClient,
	redisClient *redis.Client,
) *Service {
	return &Service{
		sdClient:      sdClient,
		ollamaClient:  ollamaClient,
		ttsClient:     ttsClient,
		whisperClient: whisperClient,
		minioClient:   minioClient,
		redisClient:   redisClient,
	}
}

// GetTTSClient 获取TTS客户端
func (s *Service) GetTTSClient() *TTSClient {
	return s.ttsClient
}

// GenerateText 生成文本
func (s *Service) GenerateText(ctx context.Context, prompt string) (string, error) {
	return s.ollamaClient.Generate(prompt, nil)
}

// GenerateImage 生成图像
func (s *Service) GenerateImage(ctx context.Context, prompt string, seed int64) (string, error) {
	// 调用SD API生成图像
	imageData, err := s.sdClient.GenerateImage(prompt, seed)
	if err != nil {
		return "", fmt.Errorf("生成图像失败: %w", err)
	}

	// 上传到MinIO
	filename := fmt.Sprintf("generated_%d_%d.png", time.Now().Unix(), seed)
	imageURL, err := s.minioClient.UploadFromBytes(ctx, filename, imageData, "image/png")
	if err != nil {
		return "", fmt.Errorf("上传图像失败: %w", err)
	}

	return imageURL, nil
}

// VoiceGenerationRequest 语音生成请求
type VoiceGenerationRequest struct {
	Text       string  `json:"text"`
	VoiceModel string  `json:"voice_model"`
	Language   string  `json:"language"`
	Speed      float64 `json:"speed"`
	Pitch      float64 `json:"pitch"`
	Volume     float64 `json:"volume"`
	Emotion    string  `json:"emotion"`
}

// GenerateVoice 生成语音
func (s *Service) GenerateVoice(ctx context.Context, req *VoiceGenerationRequest) (string, float64, error) {
	// 调用TTS API生成语音
	audioData, duration, err := s.ttsClient.GenerateVoice(req.Text, req.VoiceModel, req.Language, req.Speed, req.Pitch, req.Volume, req.Emotion)
	if err != nil {
		return "", 0, fmt.Errorf("生成语音失败: %w", err)
	}

	// 上传到MinIO
	filename := fmt.Sprintf("voice_%d.wav", time.Now().Unix())
	audioURL, err := s.minioClient.UploadFromBytes(ctx, filename, audioData, "audio/wav")
	if err != nil {
		return "", 0, fmt.Errorf("上传语音失败: %w", err)
	}

	return audioURL, duration, nil
}

// MusicGenerationRequest 音乐生成请求
type MusicGenerationRequest struct {
	Prompt   string `json:"prompt"`
	Style    string `json:"style"`
	Mood     string `json:"mood"`
	Tempo    string `json:"tempo"`
	Duration int    `json:"duration"`
}

// GenerateMusic 生成音乐
func (s *Service) GenerateMusic(ctx context.Context, req *MusicGenerationRequest) (string, float64, error) {
	// 这里应该调用音乐生成API，目前使用模拟实现
	// 实际实现中可能需要集成MusicGen、Jukebox等模型

	// 模拟音乐生成
	log.Printf("[AI] 模拟生成音乐: prompt=%s, style=%s, duration=%d", req.Prompt, req.Style, req.Duration)

	// 返回模拟的音乐URL和时长
	musicURL := fmt.Sprintf("https://example.com/music_%d.wav", time.Now().Unix())
	duration := float64(req.Duration)

	return musicURL, duration, nil
}

// CharacterProfile 角色档案，用于保持角色一致性
type CharacterProfile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Seed        int64  `json:"seed"`       // 固定种子确保一致性
	Appearance  string `json:"appearance"` // 外观描述
}

// SceneContext 场景上下文，用于保持场景连贯性
type SceneContext struct {
	Location     string `json:"location"`      // 地点
	TimeOfDay    string `json:"time_of_day"`   // 时间
	Weather      string `json:"weather"`       // 天气
	Style        string `json:"style"`         // 画风
	ColorPalette string `json:"color_palette"` // 色调
	Mood         string `json:"mood"`          // 氛围
}

// ProcessNovelToVideo: 小说转动漫视频一键生成主流程
func ProcessNovelToVideo(ctx context.Context, task *entity.Task, redisClient *redis.Client, sd *SDClient, ollama *OllamaClient, tts *TTSClient, minioBucket string) error {
	task.Status = entity.TaskStatusProcessing
	task.Progress = 5
	task.UpdatedAt = time.Now()
	_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)

	log.Printf("[AI] Ollama模型: %s", ollama.Model)
	log.Printf("[AI] Ollama状态: endpoint=%s apikey=%s", ollama.Endpoint, ollama.ApiKey)

	// 初始化提示词翻译器
	InitPromptTranslator(ollama)

	log.Printf("[AI] 开始分镜生成: task=%v", task.ID)
	var req struct {
		Novel string `json:"novel"`
	}
	_ = json.Unmarshal([]byte(task.Params), &req)
	if req.Novel == "" {
		task.Status = entity.TaskStatusFailed
		task.Error = "小说内容为空"
		task.UpdatedAt = time.Now()
		_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)
		log.Printf("[AI] 任务失败: 小说内容为空 task=%v", task.ID)
		return fmt.Errorf("novel empty")
	}

	// 重新设计prompt，确保输出标准JSON格式
	cnPrompt := fmt.Sprintf(`你是专业的分镜师，请将以下小说内容转换为5-8个简洁的分镜描述。

要求：
1. 每个分镜描述要简洁明了，30-60字
2. 描述要包含场景、人物、动作等关键视觉元素
3. 输出格式必须是标准JSON数组
4. 不要包含复杂的嵌套结构，只要简单的字符串数组

输出格式示例：
["张三在雨夜街头紧张地查看手机，霓虹灯在湿润地面反射", "李四在废弃仓库中小心穿行，灯光摇晃投下阴影", "主角被困在铁笼中，月光透过栅栏洒在地面"]

注意：
- 直接输出JSON数组，不要任何解释
- 不要使用<think>标签
- 确保是有效的JSON格式

小说内容：
%s

请输出分镜JSON数组：`, req.Novel)

	enPrompt := fmt.Sprintf(`You are a professional storyboard artist. Please convert the following novel content into 5-8 simple panel descriptions.

Requirements:
1. Each panel description should be concise, 30-60 words
2. Include key visual elements: scene, characters, actions
3. Output must be a standard JSON array format
4. Simple string array, no complex nested structures

Output format example:
["Character A nervously checks phone on rainy street with neon reflections", "Character B carefully walks through abandoned warehouse with flickering lights", "Protagonist trapped in iron cage with moonlight streaming through bars"]

Note:
- Output JSON array directly, no explanations
- No <think> tags
- Ensure valid JSON format

Novel content:
%s

Please output panel JSON array:`, req.Novel)

	var script string
	var err error
	maxRetry := 5 // 增加重试次数
	usedEn := false
	// 先用中文prompt
	for retry := 1; retry <= maxRetry; retry++ {
		script, err = ollama.Generate(cnPrompt, nil)
		if err != nil {
			log.Printf("[AI] 分镜生成失败(中文): %v task=%v 第%d次", err, task.ID, retry)
			continue
		}

		// 清理和提取JSON
		cleanedScript := cleanAndExtractJSON(script)
		log.Printf("[AI] 原始输出(中文)第%d次: %s", retry, script[:min(100, len(script))])
		log.Printf("[AI] 清理后输出(中文)第%d次: %s", retry, cleanedScript)

		var panelsTest []string
		if json.Unmarshal([]byte(cleanedScript), &panelsTest) == nil && len(panelsTest) > 0 {
			// 验证分镜内容质量
			if isValidPanelContent(panelsTest) {
				script = cleanedScript // 使用清理后的JSON
				log.Printf("[AI] 分镜生成成功(中文): task=%v panels=%d 第%d次", task.ID, len(panelsTest), retry)
				break
			} else {
				log.Printf("[AI] 分镜内容质量不合格(中文)，第%d次: %v", retry, panelsTest)
			}
		}
		log.Printf("[AI] 分镜输出不合法(中文)，第%d次: %s", retry, cleanedScript)
		time.Sleep(2 * time.Second)
		if retry == maxRetry {
			// 中文失败，切换英文prompt
			log.Printf("[AI] 中文prompt分镜生成失败，自动切换英文prompt重试 task=%v", task.ID)
			usedEn = true
			for enRetry := 1; enRetry <= maxRetry; enRetry++ {
				script, err = ollama.Generate(enPrompt, nil)
				if err != nil {
					log.Printf("[AI] 分镜生成失败(英文): %v task=%v 第%d次", err, task.ID, enRetry)
					continue
				}

				// 清理和提取JSON
				cleanedScript = cleanAndExtractJSON(script)
				log.Printf("[AI] 原始输出(英文)第%d次: %s", enRetry, script[:min(100, len(script))])
				log.Printf("[AI] 清理后输出(英文)第%d次: %s", enRetry, cleanedScript)

				if json.Unmarshal([]byte(cleanedScript), &panelsTest) == nil && len(panelsTest) > 0 {
					// 验证分镜内容质量
					if isValidPanelContent(panelsTest) {
						script = cleanedScript // 使用清理后的JSON
						log.Printf("[AI] 分镜生成成功(英文): task=%v panels=%d 第%d次", task.ID, len(panelsTest), enRetry)
						break
					} else {
						log.Printf("[AI] 分镜内容质量不合格(英文)，第%d次: %v", enRetry, panelsTest)
					}
				}
				log.Printf("[AI] 分镜输出不合法(英文)，第%d次: %s", enRetry, cleanedScript)
				time.Sleep(2 * time.Second)
				if enRetry == maxRetry {
					log.Printf("[AI] 英文prompt分镜生成也失败，尝试补充分镜 task=%v", task.ID)
					// 如果有部分分镜，尝试补充
					if len(panelsTest) > 0 && len(panelsTest) < 10 {
						panelsTest = supplementPanels(panelsTest, req.Novel, 10-len(panelsTest))
						// 重新序列化为JSON
						if supplementedJSON, err := json.Marshal(panelsTest); err == nil {
							script = string(supplementedJSON)
							log.Printf("[AI] 分镜补充成功: task=%v 最终数量=%d", task.ID, len(panelsTest))
						}
					}
				}
			}
			break
		}
	}
	if err != nil {
		task.Status = entity.TaskStatusFailed
		task.Error = "分镜生成失败: " + err.Error()
		task.UpdatedAt = time.Now()
		_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)
		log.Printf("[AI] 分镜生成失败: %v task=%v", err, task.ID)
		return err
	}
	task.Progress = 20
	_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)
	log.Printf("[AI] 分镜生成完成(%s): task=%v script=%s", func() string {
		if usedEn {
			return "英文"
		} else {
			return "中文"
		}
	}(), task.ID, script)

	// 2. 解析分镜 - 再次清理确保JSON格式正确
	var panels []string
	finalScript := cleanAndExtractJSON(script)
	log.Printf("[AI] 最终脚本内容: %s", finalScript)

	if err := json.Unmarshal([]byte(finalScript), &panels); err != nil || len(panels) == 0 {
		log.Printf("[AI] JSON解析失败，尝试其他方式: %v", err)
		// 容错：若不是JSON数组，按换行分割
		panels = []string{}
		for _, line := range splitLines(finalScript) {
			if line != "" {
				panels = append(panels, line)
			}
		}
		if len(panels) == 0 {
			task.Status = entity.TaskStatusFailed
			task.Error = "分镜解析失败: 无法从AI输出中提取有效的分镜内容"
			task.UpdatedAt = time.Now()
			_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)
			log.Printf("[AI] 分镜解析失败: task=%v, 原始输出: %s", task.ID, script)
			return fmt.Errorf("panel parse error: unable to extract valid panels from AI output")
		}
	}
	log.Printf("[AI] 分镜解析完成: task=%v panels=%d", task.ID, len(panels))

	// panels内容清洗和自动翻译
	for i, p := range panels {
		// 去除首尾空格
		p = strings.TrimSpace(p)
		// 去除镜头编号（如“镜头1：”、“Panel 1:”等）
		p = regexp.MustCompile(`^(镜头|Panel|panel)?\s*\d+[:：\.]?`).ReplaceAllString(p, "")
		// 去除多余标点和空格
		p = strings.Trim(p, " \\t\n\r,.;:：")
		// 跳过无效内容
		if p == "" || strings.ToLower(p) == "think" {
			panels[i] = ""
			continue
		}
		panels[i] = p
	}
	// 过滤空内容
	cleanPanels := make([]string, 0, len(panels))
	for _, p := range panels {
		if p != "" {
			cleanPanels = append(cleanPanels, p)
		}
	}
	panels = cleanPanels
	log.Printf("[AI] panels清洗后剩余%d条", len(panels))

	// 若用英文prompt且内容为英文，自动翻译为中文（简单调用ollama或留接口）
	if usedEn && len(panels) > 0 {
		log.Printf("[AI] 尝试自动翻译panels为中文")
		for i, p := range panels {
			// 简单调用ollama英文转中文（可换为更优API）
			transPrompt := "请将下列内容翻译为中文：" + p
			trans, err := ollama.Generate(transPrompt, nil)
			if err == nil && trans != "" {
				panels[i] = strings.TrimSpace(trans)
				log.Printf("[AI] 翻译panel %d: %s => %s", i+1, p, panels[i])
				continue
			}
			log.Printf("[AI] 翻译失败，保留原文: %s", p)
		}
	}

	// 3. 提取增强角色信息，确保一致性
	var characters []CharacterProfile
	enhancedExtractor := NewEnhancedCharacterExtractor(ollama)
	enhancedCharacters, err := enhancedExtractor.ExtractDetailedCharacters(ctx, req.Novel)
	if err != nil {
		log.Printf("[AI] 增强角色提取失败，使用简化版本: %v", err)
		// 降级到原始方法
		characters = extractCharacters(req.Novel, ollama)
		log.Printf("[AI] 使用简化角色信息: %d个角色", len(characters))
	} else {
		log.Printf("[AI] 提取到增强角色信息: %d个角色", len(enhancedCharacters))
		// 转换为兼容格式
		for _, enhanced := range enhancedCharacters {
			characters = append(characters, CharacterProfile{
				Name:        enhanced.Name,
				Description: enhanced.Description,
				Seed:        enhanced.Seed,
				Appearance:  enhanced.ConsistencyPrompt, // 使用增强的一致性提示词
			})
		}
	}

	// 分析增强场景信息，确保连贯性
	var sceneContext SceneContext
	enhancedSceneAnalyzer := NewEnhancedSceneAnalyzer(ollama)
	enhancedSceneContext, err := enhancedSceneAnalyzer.AnalyzeDetailedScene(ctx, req.Novel, panels)
	if err != nil {
		log.Printf("[AI] 增强场景分析失败，使用简化版本: %v", err)
		// 降级到原始方法
		sceneContext = analyzeSceneContext(req.Novel, ollama)
		log.Printf("[AI] 使用简化场景信息: 地点=%s 时间=%s 风格=%s", sceneContext.Location, sceneContext.TimeOfDay, sceneContext.Style)
	} else {
		log.Printf("[AI] 增强场景分析完成: 地点=%s 时间=%s 风格=%s", enhancedSceneContext.Location, enhancedSceneContext.TimeOfDay, enhancedSceneContext.ArtStyle)
		// 转换为兼容格式
		sceneContext = SceneContext{
			Location:     enhancedSceneContext.Location,
			TimeOfDay:    enhancedSceneContext.TimeOfDay,
			Weather:      enhancedSceneContext.Weather,
			Style:        enhancedSceneContext.ArtStyle,
			ColorPalette: enhancedSceneContext.ColorPalette,
			Mood:         enhancedSceneContext.Atmosphere,
		}
	}

	// 生成每格图片（SD）- 添加角色和场景一致性
	images := make([]string, len(panels))
	for i, panel := range panels {
		log.Printf("[AI] 开始生成第%d格图片: %s", i+1, panel)

		// 添加超时控制
		imgCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		var img ImageResult
		var err error

		// 根据配置决定是否翻译提示词
		finalPrompt := processPromptForSD(panel, characters, sceneContext)

		log.Printf("[AI] 最终prompt: %s", finalPrompt[:min(100, len(finalPrompt))])

		// 重试机制
		maxRetries := 3
		for retry := 1; retry <= maxRetries; retry++ {
			img, err = sd.Txt2ImgWithConsistency(imgCtx, finalPrompt, characters, sceneContext)
			if err == nil {
				break
			}
			log.Printf("[AI] 第%d格图片生成失败，第%d次重试: %v", i+1, retry, err)

			// 如果是422错误，尝试使用最简化的提示词
			if strings.Contains(err.Error(), "422") && retry == 1 {
				// 使用简化的基础提示词
				simplePrompt := panel + ", high quality, detailed, professional illustration"
				log.Printf("[AI] 尝试简化prompt: %s", simplePrompt)
				img, err = sd.Txt2Img(simplePrompt, map[string]interface{}{
					"width":        512,
					"height":       768,
					"steps":        20,
					"cfg_scale":    7,
					"sampler_name": "DPM++ 2M Karras",
				})
				if err == nil {
					break
				}
			}

			if retry < maxRetries {
				time.Sleep(5 * time.Second)
			}
		}

		if err != nil {
			log.Printf("[AI] 第%d格图片生成最终失败: %v task=%v", i+1, err, task.ID)
			// SD服务不可用时，创建占位符图片
			log.Printf("[AI] SD服务不可用，使用占位符图片")
			placeholderData := []byte("data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNTEyIiBoZWlnaHQ9Ijc2OCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjBmMGYwIi8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIyNCIgZmlsbD0iIzk5OSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPuWbvueJh+WNoOS9jeespuWbvjwvdGV4dD48L3N2Zz4=")
			img = ImageResult{
				Data: placeholderData,
			}
		}

		images[i] = encodeBase64(img.Data)
		task.Progress = 20 + int(float64(i+1)/float64(len(panels))*40)
		_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)
		log.Printf("[AI] 第%d格图片生成完成: task=%v", i+1, task.ID)
	}

	// 4. 生成优化的旁白文本
	log.Printf("[AI] 开始生成旁白文本: task=%v", task.ID)
	narration, err := generateVoiceNarration(panels, ollama)
	if err != nil {
		log.Printf("[AI] 旁白生成失败，使用默认拼接: %v", err)
		// 降级处理：使用改进的默认拼接
		narration = generateDefaultNarration(panels)
	}
	log.Printf("[AI] 旁白文本生成完成，长度: %d字符", len(narration))

	log.Printf("[AI] 开始配音合成: task=%v", task.ID)
	// 5. 配音（TTS）
	audio, err := tts.Synthesize(narration, nil)
	if err != nil {
		task.Status = entity.TaskStatusFailed
		task.Error = "配音生成失败: " + err.Error()
		task.UpdatedAt = time.Now()
		_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)
		log.Printf("[AI] 配音生成失败: %v task=%v", err, task.ID)
		return err
	}
	task.Progress = 70
	_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)
	log.Printf("[AI] 配音合成完成: task=%v", task.ID)

	log.Printf("[AI] 开始视频合成: task=%v", task.ID)
	// 6. 合成动漫视频（FFmpeg）
	videoPath, err := ComposeVideoFromImagesAndAudio(images, audio)
	if err != nil {
		task.Status = entity.TaskStatusFailed
		task.Error = "视频合成失败: " + err.Error()
		task.UpdatedAt = time.Now()
		_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)
		log.Printf("[AI] 视频合成失败: %v task=%v", err, task.ID)
		return err
	}
	task.Progress = 90
	_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)
	log.Printf("[AI] 视频合成完成: task=%v path=%s", task.ID, videoPath)

	// 7. 上传视频到 MinIO（伪代码，需根据你的 MinIO 客户端实现）
	videoURL := "https://minio.example.com/" + videoPath // TODO: 实际应上传并获取外链

	// 8. 写入最终结果
	result := map[string]interface{}{
		"url":    videoURL,
		"images": images,
		"panels": panels,
	}
	b, _ := json.Marshal(result)
	task.Status = entity.TaskStatusCompleted
	task.Progress = 100
	task.Result = string(b)
	task.UpdatedAt = time.Now()
	_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)
	log.Printf("[AI] 任务完成: task=%v", task.ID)
	return nil
}

// splitLines 工具函数 - 按行分割字符串
func splitLines(s string) []string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

// encodeBase64 工具函数
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// ComposeVideoFromImagesAndAudio 伪实现
func ComposeVideoFromImagesAndAudio(images []string, audio []byte) (string, error) {
	tmpDir, err := ioutil.TempDir("", "novel2video_")
	if err != nil {
		return "", err
	}
	// 保存图片
	imgFiles := make([]string, 0, len(images))
	for i, imgBase64 := range images {
		imgData, err := base64.StdEncoding.DecodeString(imgBase64)
		if err != nil {
			return "", err
		}
		imgPath := filepath.Join(tmpDir, fmt.Sprintf("img_%03d.png", i+1))
		if err := ioutil.WriteFile(imgPath, imgData, 0644); err != nil {
			return "", err
		}
		imgFiles = append(imgFiles, imgPath)
	}
	// 保存音频
	audioPath := filepath.Join(tmpDir, "audio.wav")
	if err := ioutil.WriteFile(audioPath, audio, 0644); err != nil {
		return "", err
	}
	// 生成图片列表txt
	listPath := filepath.Join(tmpDir, "images.txt")
	listFile, err := os.Create(listPath)
	if err != nil {
		return "", err
	}
	for _, img := range imgFiles {
		fmt.Fprintf(listFile, "file '%s'\n", img)
	}
	listFile.Close()
	// 合成视频（假设每张图片2秒，音频自动对齐）
	videoPath := filepath.Join(tmpDir, "output.mp4")
	cmd := exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", listPath, "-i", audioPath, "-c:v", "libx264", "-c:a", "aac", "-shortest", "-pix_fmt", "yuv420p", videoPath)
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return videoPath, nil
}

// cleanAndExtractJSON 清理AI输出并提取JSON数组
func cleanAndExtractJSON(rawOutput string) string {
	// 移除常见的思考标签和无关内容
	cleaned := rawOutput

	// 移除<think>...</think>标签及其内容
	thinkRegex := regexp.MustCompile(`(?s)<think>.*?</think>`)
	cleaned = thinkRegex.ReplaceAllString(cleaned, "")

	// 移除其他常见的标记
	cleaned = regexp.MustCompile(`(?i)^(思考|think|分析|analysis)[:：].*?\n`).ReplaceAllString(cleaned, "")
	cleaned = regexp.MustCompile(`(?i)^(回答|answer|结果|result)[:：].*?\n`).ReplaceAllString(cleaned, "")
	cleaned = regexp.MustCompile(`(?i)^(请输出|输出|output).*?\n`).ReplaceAllString(cleaned, "")

	// 首先尝试提取标准JSON数组格式
	jsonRegex := regexp.MustCompile(`(?s)\[\s*"[^"]*"(?:\s*,\s*"[^"]*")*\s*\]`)
	matches := jsonRegex.FindAllString(cleaned, -1)

	// 过滤掉明显的示例格式
	for _, match := range matches {
		if !containsExampleContent(match) && isValidJSONArray(match) {
			return strings.TrimSpace(match)
		}
	}

	// 处理复杂的嵌套结构（如scene, environment等）
	if strings.Contains(cleaned, `"scene"`) || strings.Contains(cleaned, `"environment"`) {
		return extractSimpleDescriptions(cleaned)
	}

	// 如果没找到有效的JSON，尝试提取引号内容并构造JSON
	quoteRegex := regexp.MustCompile(`"([^"]{10,})"`) // 至少10个字符
	quotes := quoteRegex.FindAllStringSubmatch(cleaned, -1)

	if len(quotes) > 0 {
		var items []string
		seen := make(map[string]bool) // 去重
		for _, quote := range quotes {
			if len(quote) > 1 {
				content := strings.TrimSpace(quote[1])
				// 跳过示例内容和重复内容
				if !containsExampleContent(`"`+content+`"`) && !seen[content] && len(content) > 10 {
					items = append(items, content)
					seen[content] = true
				}
			}
		}
		if len(items) > 0 && len(items) <= 15 { // 限制数量
			jsonBytes, _ := json.Marshal(items)
			return string(jsonBytes)
		}
	}

	// 最后尝试按行分割并构造JSON
	lines := strings.Split(cleaned, "\n")
	var items []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 移除序号和标点
		line = regexp.MustCompile(`^\d+[\.、\s]*`).ReplaceAllString(line, "")
		line = regexp.MustCompile(`^[-*•]\s*`).ReplaceAllString(line, "")
		line = strings.Trim(line, " \t\n\r,.;:：。，；")

		if line != "" && len(line) > 10 && !containsExampleContent(line) {
			items = append(items, line)
		}
	}

	if len(items) > 0 && len(items) <= 15 {
		jsonBytes, _ := json.Marshal(items)
		return string(jsonBytes)
	}

	return cleaned
}

// extractSimpleDescriptions 从复杂结构中提取简单描述
func extractSimpleDescriptions(content string) string {
	var descriptions []string

	// 尝试提取scene后面的描述
	sceneRegex := regexp.MustCompile(`"([^"]{20,})"`)
	matches := sceneRegex.FindAllStringSubmatch(content, -1)

	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			desc := strings.TrimSpace(match[1])
			// 过滤掉字段名和示例内容
			if !isFieldName(desc) && !containsExampleContent(desc) && !seen[desc] && len(desc) > 20 {
				descriptions = append(descriptions, desc)
				seen[desc] = true
				if len(descriptions) >= 10 { // 限制数量
					break
				}
			}
		}
	}

	if len(descriptions) > 0 {
		jsonBytes, _ := json.Marshal(descriptions)
		return string(jsonBytes)
	}

	return content
}

// isFieldName 检查是否为字段名
func isFieldName(s string) bool {
	fieldNames := []string{"scene", "environment", "characters", "character", "action", "actions", "lighting", "emotion"}
	s = strings.ToLower(s)
	for _, field := range fieldNames {
		if s == field {
			return true
		}
	}
	return false
}

// isValidJSONArray 检查是否为有效的JSON数组
func isValidJSONArray(s string) bool {
	var arr []string
	return json.Unmarshal([]byte(s), &arr) == nil
}

// containsExampleContent 检查内容是否包含示例格式
func containsExampleContent(content string) bool {
	content = strings.ToLower(content)
	examplePatterns := []string{
		"镜头1描述", "镜头2描述", "镜头3描述",
		"panel1", "panel2", "panel3",
		"描述", "description",
		"镜头描述", "panel description",
	}

	for _, pattern := range examplePatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	return false
}

// extractCharacters 从小说中提取角色信息
func extractCharacters(novel string, ollama *OllamaClient) []CharacterProfile {
	prompt := fmt.Sprintf(`请分析以下小说内容，提取主要角色的名字。

要求：
1. 识别小说中提到的人物名字（最多5个）
2. 如果没有明确的人名，可以根据内容推测角色类型（如"男主角"、"女主角"等）
3. 输出格式为JSON字符串数组，只包含角色名字
4. 不要输出其他解释，直接输出JSON数组

输出格式示例：
["张三", "李四", "王五"]

小说内容：
%s

请输出角色名字JSON数组：`, novel)

	response, err := ollama.Generate(prompt, nil)
	if err != nil {
		log.Printf("[AI] 角色提取失败: %v", err)
		return []CharacterProfile{}
	}

	// 清理和解析JSON
	cleanedResponse := cleanAndExtractJSON(response)
	log.Printf("[AI] 角色解析原始响应: %s", response[:min(200, len(response))])
	log.Printf("[AI] 角色解析清理后: %s", cleanedResponse)

	var rawCharacters []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Appearance  string `json:"appearance"`
	}

	if err := json.Unmarshal([]byte(cleanedResponse), &rawCharacters); err != nil {
		log.Printf("[AI] 角色信息解析失败: %v", err)
		log.Printf("[AI] 尝试解析为字符串数组")

		// 尝试解析为字符串数组
		var characterNames []string
		if err2 := json.Unmarshal([]byte(cleanedResponse), &characterNames); err2 == nil {
			log.Printf("[AI] 成功解析为字符串数组: %v", characterNames)
			// 转换为角色对象
			for _, name := range characterNames {
				if strings.TrimSpace(name) != "" {
					rawCharacters = append(rawCharacters, struct {
						Name        string `json:"name"`
						Description string `json:"description"`
						Appearance  string `json:"appearance"`
					}{
						Name:        name,
						Description: "角色",
						Appearance:  "普通人物",
					})
				}
			}
		} else {
			log.Printf("[AI] 字符串数组解析也失败: %v", err2)
			// 使用默认角色
			rawCharacters = []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Appearance  string `json:"appearance"`
			}{
				{Name: "主角", Description: "故事主角", Appearance: "普通人物"},
			}
		}
	}

	// 转换为CharacterProfile并生成种子
	characters := make([]CharacterProfile, len(rawCharacters))
	for i, char := range rawCharacters {
		// 使用角色名生成固定种子
		hash := md5.Sum([]byte(char.Name))
		seed := int64(hash[0])<<24 | int64(hash[1])<<16 | int64(hash[2])<<8 | int64(hash[3])

		characters[i] = CharacterProfile{
			Name:        char.Name,
			Description: char.Description,
			Appearance:  char.Appearance,
			Seed:        seed,
		}
	}

	return characters
}

// analyzeSceneContext 分析场景上下文
func analyzeSceneContext(novel string, ollama *OllamaClient) SceneContext {
	prompt := fmt.Sprintf(`请分析以下小说的场景信息，提取场景上下文。

要求：
1. 分析主要发生地点
2. 确定时间背景
3. 分析整体氛围和风格
4. 输出JSON格式

输出格式：
{
  "location": "主要地点",
  "time_of_day": "时间（如：深夜、白天等）",
  "weather": "天气情况",
  "style": "画风（如：写实、暗黑、现代等）",
  "color_palette": "主色调",
  "mood": "整体氛围"
}

小说内容：
%s

请输出场景信息JSON：`, novel)

	response, err := ollama.Generate(prompt, nil)
	if err != nil {
		log.Printf("[AI] 场景分析失败: %v", err)
		return getDefaultSceneContext()
	}

	// 清理和解析JSON
	cleanedResponse := cleanAndExtractJSON(response)

	var context SceneContext
	if err := json.Unmarshal([]byte(cleanedResponse), &context); err != nil {
		log.Printf("[AI] 场景信息解析失败: %v", err)
		return getDefaultSceneContext()
	}

	return context
}

// getDefaultSceneContext 获取默认场景上下文
func getDefaultSceneContext() SceneContext {
	return SceneContext{
		Location:     "室内",
		TimeOfDay:    "白天",
		Weather:      "晴朗",
		Style:        "现代写实",
		ColorPalette: "自然色调",
		Mood:         "中性",
	}
}

// buildConsistentPrompt 构建简化的一致性prompt
func buildConsistentPrompt(panel string, characters []CharacterProfile, context SceneContext) string {
	// 直接返回基础分镜描述，让翻译器处理
	// 避免复杂的英文关键词混合导致翻译问题
	return panel
}

// buildEnhancedEnglishPrompt 构建增强的英文提示词
func buildEnhancedEnglishPrompt(translatedPanel string, characters []CharacterProfile, context SceneContext) string {
	// 确保已经是英文提示词
	if !isEnglishPrompt(translatedPanel) {
		log.Printf("[AI] 警告：提示词可能不是英文: %s", translatedPanel[:min(50, len(translatedPanel))])
	}

	// 优化提示词，移除敏感内容
	optimizedPrompt := optimizePromptForSD(translatedPanel)

	// 为本地SD添加更有效的关键词
	// 本地SD更关注艺术质量而非内容安全
	styleKeywords := []string{
		"anime style",
		"illustration",
		"detailed",
		"high quality",
		"masterpiece",
		"best quality",
	}

	// 构建最终提示词
	finalPrompt := optimizedPrompt + ", " + strings.Join(styleKeywords, ", ")

	return finalPrompt
}

// optimizePromptForSD 优化提示词使其适合SD生成
func optimizePromptForSD(prompt string) string {
	// 对于本地SD，主要问题不是敏感词汇，而是提示词质量
	// 重点优化：简化复杂叙事，提高可理解性

	optimized := prompt

	// 1. 移除第一人称和复杂叙事
	optimized = strings.ReplaceAll(optimized, " us.", ".")
	optimized = strings.ReplaceAll(optimized, " us,", ",")
	optimized = strings.ReplaceAll(optimized, " us ", " them ")
	optimized = strings.ReplaceAll(optimized, " we ", " they ")
	optimized = strings.ReplaceAll(optimized, " our ", " their ")

	// 2. 简化复杂的动作序列
	// 将复杂的连续动作简化为主要动作
	actionWords := []string{"turned", "entered", "walked", "ran", "jumped", "looked", "smiled", "sat"}
	actionCount := 0
	for _, action := range actionWords {
		if strings.Contains(strings.ToLower(optimized), action) {
			actionCount++
		}
	}

	// 如果动作过多，简化句子
	if actionCount > 2 && strings.Contains(optimized, ",") {
		parts := strings.Split(optimized, ",")
		// 保留前两个最重要的部分
		if len(parts) > 2 {
			optimized = strings.Join(parts[:2], ",")
		}
	}

	// 3. 优化句子结构，使其更适合SD理解
	// 移除过于复杂的从句
	optimized = strings.ReplaceAll(optimized, " and then ", ", ")
	optimized = strings.ReplaceAll(optimized, " while ", ", ")
	optimized = strings.ReplaceAll(optimized, " when ", ", ")

	// 4. 确保有明确的主体
	if !strings.Contains(strings.ToLower(optimized), "person") &&
		!strings.Contains(strings.ToLower(optimized), "people") &&
		!strings.Contains(strings.ToLower(optimized), "character") &&
		!strings.Contains(strings.ToLower(optimized), "woman") &&
		!strings.Contains(strings.ToLower(optimized), "man") {
		// 如果没有明确的人物主体，添加一个
		if strings.Contains(optimized, "Zhang Fengfeng") {
			optimized = "character " + optimized
		}
	}

	return strings.TrimSpace(optimized)
}

// isEnglishPrompt 检测是否为英文提示词
func isEnglishPrompt(prompt string) bool {
	englishCount := 0
	chineseCount := 0

	for _, r := range prompt {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			englishCount++
		} else if r >= '\u4e00' && r <= '\u9fff' {
			chineseCount++
		}
	}

	total := englishCount + chineseCount
	if total == 0 {
		return true
	}

	return float64(englishCount)/float64(total) > 0.5
}

// processPromptForSD 根据配置处理SD提示词
func processPromptForSD(panel string, characters []CharacterProfile, context SceneContext) string {
	// 检查是否启用翻译
	enableTranslation := strings.ToLower(os.Getenv("ENABLE_PROMPT_TRANSLATION")) == "true"
	useChinesePrompts := strings.ToLower(os.Getenv("SD_USE_CHINESE_PROMPTS")) == "true"

	// 如果配置为使用中文提示词，直接使用中文
	if useChinesePrompts {
		log.Printf("[AI] 配置为使用中文提示词")
		// 为中文提示词添加质量关键词
		chineseKeywords := []string{
			"高质量",
			"详细",
			"杰作",
			"最佳质量",
			"动漫风格",
		}
		return panel + "，" + strings.Join(chineseKeywords, "，")
	}

	// 如果不启用翻译，但也不使用中文，则使用简单的英文关键词
	if !enableTranslation {
		log.Printf("[AI] 翻译已禁用，使用原始提示词")
		// 如果原始提示词是中文，添加中文关键词
		if containsChineseCharacters(panel) {
			chineseKeywords := []string{"高质量", "详细", "动漫风格"}
			return panel + "，" + strings.Join(chineseKeywords, "，")
		}
		// 如果是英文，添加英文关键词
		englishKeywords := []string{"high quality", "detailed", "anime style"}
		return panel + ", " + strings.Join(englishKeywords, ", ")
	}

	// 启用翻译的情况下，进行翻译
	log.Printf("[AI] 翻译已启用，进行提示词翻译")
	translatedPanel, transErr := TranslatePrompt(panel)
	if transErr != nil {
		log.Printf("[AI] 提示词翻译失败，使用原始prompt: %v", transErr)
		translatedPanel = panel
	}

	// 构建增强的英文prompt
	return buildEnhancedEnglishPrompt(translatedPanel, characters, context)
}

// containsChineseCharacters 检查是否包含中文字符
func containsChineseCharacters(s string) bool {
	for _, r := range s {
		if r >= '\u4e00' && r <= '\u9fff' {
			return true
		}
	}
	return false
}

// generateVoiceNarration 生成适合语音的旁白
func generateVoiceNarration(panels []string, ollama *OllamaClient) (string, error) {
	panelsText := strings.Join(panels, "\n")

	prompt := fmt.Sprintf(`请将以下分镜描述转换为适合语音播报的连贯旁白。

要求：
1. 语言自然流畅，适合语音播报
2. 保持故事的连贯性和节奏感
3. 避免过于技术性的描述
4. 增加适当的情感色彩
5. 控制在200字以内

分镜描述：
%s

请输出旁白文本：`, panelsText)

	response, err := ollama.Generate(prompt, nil)
	if err != nil {
		return "", err
	}

	// 清理响应
	narration := strings.TrimSpace(response)
	// 移除可能的思考标签
	narration = regexp.MustCompile(`(?s)<think>.*?</think>`).ReplaceAllString(narration, "")
	narration = strings.TrimSpace(narration)

	return narration, nil
}

// generateDefaultNarration 生成默认的改进旁白
func generateDefaultNarration(panels []string) string {
	if len(panels) == 0 {
		return "故事开始了。"
	}

	// 改进的默认拼接逻辑
	var narration strings.Builder

	for i, panel := range panels {
		// 清理分镜描述，使其更适合语音
		cleaned := strings.TrimSpace(panel)

		// 添加连接词
		if i == 0 {
			narration.WriteString("故事开始，")
		} else if i == len(panels)-1 {
			narration.WriteString("最后，")
		} else {
			connectors := []string{"接着，", "然后，", "此时，", "紧接着，"}
			narration.WriteString(connectors[i%len(connectors)])
		}

		narration.WriteString(cleaned)

		// 添加适当的停顿
		if i < len(panels)-1 {
			narration.WriteString("。")
		} else {
			narration.WriteString("。故事到此结束。")
		}
	}

	return narration.String()
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// isValidPanelContent 验证分镜内容是否有意义和足够数量
func isValidPanelContent(panels []string) bool {
	if len(panels) == 0 {
		return false
	}

	// 检查是否包含无意义的占位符
	invalidPatterns := []string{
		"镜头1描述", "镜头2描述", "镜头3描述",
		"panel1", "panel2", "panel3",
		"分镜1描述", "分镜2描述",
		"示例", "example", "描述",
		"第一个", "第二个", "第三个",
	}

	validCount := 0
	highQualityCount := 0

	for _, panel := range panels {
		panel = strings.TrimSpace(panel)

		// 基本长度检查 - 提高质量要求
		if len(panel) < 20 {
			continue
		}

		isInvalid := false
		panelLower := strings.ToLower(panel)
		for _, pattern := range invalidPatterns {
			if strings.Contains(panelLower, strings.ToLower(pattern)) {
				isInvalid = true
				break
			}
		}

		if !isInvalid {
			validCount++

			// 检查是否是高质量分镜（包含详细描述）
			if len(panel) >= 50 &&
				(strings.Contains(panelLower, "表情") ||
					strings.Contains(panelLower, "眼神") ||
					strings.Contains(panelLower, "背景") ||
					strings.Contains(panelLower, "光") ||
					strings.Contains(panelLower, "阴影") ||
					strings.Contains(panelLower, "特写") ||
					strings.Contains(panelLower, "全景")) {
				highQualityCount++
			}
		}
	}

	// 调整质量要求，更加宽松：
	// 1. 至少要有3个有效分镜
	// 2. 其中至少1个是高质量分镜
	// 3. 总体有效率要达到40%以上
	totalPanels := len(panels)
	validRate := float64(validCount) / float64(totalPanels)

	log.Printf("[AI] 分镜质量检查: 总数=%d, 有效=%d, 高质量=%d, 有效率=%.2f",
		totalPanels, validCount, highQualityCount, validRate)

	return validCount >= 3 && highQualityCount >= 1 && validRate >= 0.4
}

// supplementPanels 补充分镜内容，确保数量和质量
func supplementPanels(existingPanels []string, novel string, needCount int) []string {
	log.Printf("[AI] 开始补充分镜: 现有=%d, 需要补充=%d", len(existingPanels), needCount)

	// 基于小说内容生成补充分镜
	supplementPrompts := []string{
		"主角的特写镜头，表情专注而坚定，眼神中透露出内心的复杂情感，背景虚化突出人物",
		"全景展示故事发生的环境，建筑物和街道的细节清晰可见，光影效果营造氛围",
		"紧张时刻的手部特写，紧握的拳头或颤抖的手指，体现内心的紧张和决心",
		"对话场景的双人镜头，两人面对面站立，表情和肢体语言传达情感交流",
		"环境细节的特写，重要物品或线索的展示，为故事发展提供视觉信息",
		"动作场面的动态镜头，人物移动或奔跑的瞬间，体现紧迫感和动感",
		"情感高潮的面部特写，眼泪、微笑或愤怒的表情，传达强烈的情感冲击",
		"场景转换的过渡镜头，从一个地点到另一个地点的视觉连接",
		"群体场景的广角镜头，多个人物的互动和环境的整体展示",
		"结尾的象征性镜头，寓意深刻的画面，为故事提供完美收尾",
	}

	result := make([]string, len(existingPanels))
	copy(result, existingPanels)

	// 根据需要补充的数量选择合适的分镜
	for i := 0; i < needCount && i < len(supplementPrompts); i++ {
		result = append(result, supplementPrompts[i])
	}

	log.Printf("[AI] 分镜补充完成: 最终数量=%d", len(result))
	return result
}

func UploadVideoToMinio(ctx context.Context, minioClient minio.MinioClient, bucket, videoPath string) (string, error) {
	file, err := os.Open(videoPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return "", err
	}
	objectName := filepath.Base(videoPath)
	contentType := mime.TypeByExtension(filepath.Ext(videoPath))
	if contentType == "" {
		contentType = "video/mp4"
	}
	url, err := minioClient.Upload(ctx, objectName, file, stat.Size(), contentType)
	if err != nil {
		return "", err
	}
	return url, nil
}
