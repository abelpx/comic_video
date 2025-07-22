package ai

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)



type SDClient struct {
	Endpoint              string // 例如 http://127.0.0.1:7860
	Timeout               time.Duration
	MaxRetries           int
	RetryDelay           time.Duration
	EnableTranslation    bool
	UseChinesePrompts    bool
}

// NewSDClient 创建新的SD客户端
func NewSDClient(endpoint string) *SDClient {
	// 从环境变量读取配置
	timeout := 60 * time.Second
	if timeoutStr := os.Getenv("SD_TIMEOUT"); timeoutStr != "" {
		if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = parsedTimeout
		}
	}

	maxRetries := 3
	if retriesStr := os.Getenv("SD_MAX_RETRIES"); retriesStr != "" {
		if parsedRetries, err := strconv.Atoi(retriesStr); err == nil {
			maxRetries = parsedRetries
		}
	}

	retryDelay := 10 * time.Second
	if delayStr := os.Getenv("SD_RETRY_DELAY"); delayStr != "" {
		if parsedDelay, err := time.ParseDuration(delayStr); err == nil {
			retryDelay = parsedDelay
		}
	}

	enableTranslation := strings.ToLower(os.Getenv("ENABLE_PROMPT_TRANSLATION")) == "true"
	useChinesePrompts := strings.ToLower(os.Getenv("SD_USE_CHINESE_PROMPTS")) == "true"

	return &SDClient{
		Endpoint:              endpoint,
		Timeout:               timeout,
		MaxRetries:           maxRetries,
		RetryDelay:           retryDelay,
		EnableTranslation:    enableTranslation,
		UseChinesePrompts:    useChinesePrompts,
	}
}

func (s *SDClient) Txt2Img(prompt string, opts map[string]interface{}) (ImageResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()
	return s.Txt2ImgWithContext(ctx, prompt, opts)
}

func (s *SDClient) Txt2ImgWithContext(ctx context.Context, prompt string, opts map[string]interface{}) (ImageResult, error) {
	// 处理提示词：根据配置决定是否使用中文
	finalPrompt := s.processPrompt(prompt)

	var lastErr error

	// 重试逻辑
	for attempt := 0; attempt <= s.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[SD] 第%d次重试，等待%v", attempt, s.RetryDelay)
			select {
			case <-time.After(s.RetryDelay):
			case <-ctx.Done():
				return ImageResult{}, ctx.Err()
			}
		}

		result, err := s.attemptGeneration(ctx, finalPrompt, opts)
		if err == nil {
			if attempt > 0 {
				log.Printf("[SD] 第%d次重试成功", attempt)
			}
			return result, nil
		}

		lastErr = err
		log.Printf("[SD] 第%d次尝试失败: %v", attempt+1, err)

		// 如果是上下文取消，不再重试
		if ctx.Err() != nil {
			break
		}
	}

	return ImageResult{}, fmt.Errorf("SD生成失败，已重试%d次: %v", s.MaxRetries, lastErr)
}

// processPrompt 处理提示词
func (s *SDClient) processPrompt(prompt string) string {
	// 如果配置为使用中文提示词，且提示词包含中文，则保持中文
	if s.UseChinesePrompts && containsChinese(prompt) {
		log.Printf("[SD] 使用中文提示词: %s", prompt[:min(50, len(prompt))])
		return prompt
	}

	// 否则使用英文（可能已经翻译过）
	log.Printf("[SD] 使用英文提示词: %s", prompt[:min(50, len(prompt))])
	return prompt
}

// attemptGeneration 尝试生成图像
func (s *SDClient) attemptGeneration(ctx context.Context, prompt string, opts map[string]interface{}) (ImageResult, error) {
	// 组装请求体
	body := map[string]interface{}{
		"prompt": prompt,
	}
	for k, v := range opts {
		body[k] = v
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return ImageResult{}, fmt.Errorf("JSON编码失败: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.Endpoint+"/sdapi/v1/txt2img", bytes.NewReader(jsonData))
	if err != nil {
		return ImageResult{}, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 使用配置的超时时间
	client := &http.Client{
		Timeout: s.Timeout + 30*time.Second, // 客户端超时比context稍长
	}

	log.Printf("[SD] 开始生成图像，超时设置: %v", s.Timeout)
	resp, err := client.Do(req)
	if err != nil {
		return ImageResult{}, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errorMsg := string(bodyBytes)
		if len(errorMsg) > 200 {
			errorMsg = errorMsg[:200] + "..."
		}
		return ImageResult{}, fmt.Errorf("SD API错误 %s: %s", resp.Status, errorMsg)
	}

	var result struct {
		Images []string `json:"images"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ImageResult{}, fmt.Errorf("响应解析失败: %v", err)
	}

	if len(result.Images) == 0 {
		return ImageResult{}, fmt.Errorf("未生成任何图像")
	}

	log.Printf("[SD] 图像生成成功")

	// SD WebUI 返回 base64，解码
	imgData, err := decodeBase64Image(result.Images[0])
	if err != nil {
		return ImageResult{}, err
	}
	return ImageResult{Data: imgData}, nil
}

// containsChinese 检查字符串是否包含中文字符
func containsChinese(s string) bool {
	for _, r := range s {
		if r >= '\u4e00' && r <= '\u9fff' {
			return true
		}
	}
	return false
}



// Txt2ImgWithConsistency 带有角色和场景一致性的图片生成
func (s *SDClient) Txt2ImgWithConsistency(ctx context.Context, prompt string, characters []CharacterProfile, sceneContext SceneContext) (ImageResult, error) {
	// 计算一致性种子
	seed := calculateConsistencySeed(characters, sceneContext)

	// 组装请求体，添加一致性参数（移除不支持的参数）
	body := map[string]interface{}{
		"prompt":       prompt,
		"seed":         seed,
		"width":        512,
		"height":       768,
		"steps":        20,
		"cfg_scale":    7,
		"sampler_name": "DPM++ 2M Karras",
	}

	b, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", s.Endpoint+"/sdapi/v1/txt2img", bytes.NewReader(b))
	if err != nil {
		return ImageResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 3 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return ImageResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// 读取错误响应体
		bodyBytes, _ := io.ReadAll(resp.Body)
		errorMsg := string(bodyBytes)
		if len(errorMsg) > 200 {
			errorMsg = errorMsg[:200] + "..."
		}
		return ImageResult{}, fmt.Errorf("SD API error: %s, response: %s", resp.Status, errorMsg)
	}

	var result struct {
		Images []string `json:"images"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ImageResult{}, err
	}
	if len(result.Images) == 0 {
		return ImageResult{}, fmt.Errorf("no image returned")
	}

	// SD WebUI 返回 base64，解码
	imgData, err := decodeBase64Image(result.Images[0])
	if err != nil {
		return ImageResult{}, err
	}
	return ImageResult{Data: imgData}, nil
}

// calculateConsistencySeed 计算一致性种子
func calculateConsistencySeed(characters []CharacterProfile, sceneContext SceneContext) int64 {
	// 组合所有一致性因素
	var seedSource string

	// 添加角色信息
	for _, char := range characters {
		seedSource += char.Name + char.Appearance
	}

	// 添加场景信息
	seedSource += sceneContext.Location + sceneContext.Style + sceneContext.ColorPalette

	// 生成MD5哈希并转换为种子
	hash := md5.Sum([]byte(seedSource))
	seed := int64(hash[0])<<24 | int64(hash[1])<<16 | int64(hash[2])<<8 | int64(hash[3])

	// 确保种子为正数
	if seed < 0 {
		seed = -seed
	}

	return seed
}

func (s *SDClient) Img2Img(image []byte, prompt string, opts map[string]interface{}) (ImageResult, error) {
	// 可扩展，暂未实现
	return ImageResult{}, fmt.Errorf("not implemented")
}

// decodeBase64Image 解码base64图片
func decodeBase64Image(b64 string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(b64)
} 