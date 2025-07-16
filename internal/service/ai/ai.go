package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/redis"
	"comic_video/internal/repository/minio"
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
	"mime"
	"log"
)

// ProcessNovelToVideo: 小说转动漫视频一键生成主流程
func ProcessNovelToVideo(ctx context.Context, task *entity.Task, redisClient *redis.Client, sd *SDClient, ollama *OllamaClient, tts *TTSClient, minioBucket string) error {
	task.Status = entity.TaskStatusProcessing
	task.Progress = 5
	task.UpdatedAt = time.Now()
	_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)

	log.Printf("[AI] Ollama模型: %s", ollama.Model)
	log.Printf("[AI] Ollama状态: endpoint=%s apikey=%s", ollama.Endpoint, ollama.ApiKey)

	log.Printf("[AI] 开始分镜生成: task=%v", task.ID)
	var req struct{ Novel string `json:"novel"` }
	_ = json.Unmarshal([]byte(task.Params), &req)
	if req.Novel == "" {
		task.Status = entity.TaskStatusFailed
		task.Error = "小说内容为空"
		task.UpdatedAt = time.Now()
		_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)
		log.Printf("[AI] 任务失败: 小说内容为空 task=%v", task.ID)
		return fmt.Errorf("novel empty")
	}

	// 优化prompt，适配中文和结构化输出
	ollamaPrompt := fmt.Sprintf(`你是一个漫画分镜脚本专家。请将以下小说内容拆分为分镜，每一镜一句话，直接输出JSON数组（如 [\"镜头1描述\", \"镜头2描述\", ...] ），不要输出其它内容、不要输出<think>、不要输出解释说明。每个分镜要简洁、具体、有画面感。小说内容如下：%s`, req.Novel)

	var script string
	var err error
	maxRetry := 3
	for retry := 1; retry <= maxRetry; retry++ {
		script, err = ollama.Generate(ollamaPrompt, nil)
		if err != nil {
			log.Printf("[AI] 分镜生成失败: %v task=%v 第%d次", err, task.ID, retry)
			continue
		}
		// 校验输出是否为JSON数组
		var panelsTest []string
		if json.Unmarshal([]byte(script), &panelsTest) == nil && len(panelsTest) > 0 {
			log.Printf("[AI] 分镜生成成功: task=%v panels=%d 第%d次", task.ID, len(panelsTest), retry)
			break
		}
		log.Printf("[AI] 分镜输出不合法，第%d次: %s", retry, script)
		time.Sleep(2 * time.Second)
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
	log.Printf("[AI] 分镜生成完成: task=%v script=%s", task.ID, script)

	// 2. 解析分镜
	var panels []string
	if err := json.Unmarshal([]byte(script), &panels); err != nil || len(panels) == 0 {
		// 容错：若不是JSON数组，按换行分割
		panels = []string{}
		for _, line := range splitLines(script) {
			if line != "" {
				panels = append(panels, line)
			}
		}
		if len(panels) == 0 {
			task.Status = entity.TaskStatusFailed
			task.Error = "分镜解析失败"
			task.UpdatedAt = time.Now()
			_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)
			log.Printf("[AI] 分镜解析失败: task=%v", task.ID)
			return fmt.Errorf("panel parse error")
		}
	}
	log.Printf("[AI] 分镜解析完成: task=%v panels=%d", task.ID, len(panels))

	// 3. 生成每格图片（SD）
	images := make([]string, 0, len(panels))
	for i, panel := range panels {
		log.Printf("[AI] 开始生成第%d格图片: %s", i+1, panel)
		img, err := sd.Txt2Img(panel, nil)
		if err != nil {
			task.Status = entity.TaskStatusFailed
			task.Error = fmt.Sprintf("第%d格图片生成失败: %v", i+1, err)
			task.UpdatedAt = time.Now()
			_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)
			log.Printf("[AI] 第%d格图片生成失败: %v task=%v", i+1, err, task.ID)
			return err
		}
		images = append(images, encodeBase64(img.Data))
		task.Progress = 20 + int(float64(i+1)/float64(len(panels))*40)
		_ = redisClient.SetTaskStatus(ctx, task, 24*time.Hour)
		log.Printf("[AI] 第%d格图片生成完成: task=%v", i+1, task.ID)
	}

	// 4. 合成旁白文本
	narration := ""
	for _, panel := range panels {
		narration += panel + "。"
	}

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

// splitLines 工具函数
func splitLines(s string) []string {
	var lines []string
	for _, line := range []byte(s) {
		if line == '\n' || line == '\r' {
			continue
		}
		lines = append(lines, string(line))
	}
	return lines
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