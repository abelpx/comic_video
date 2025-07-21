package ai

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)



type SDClient struct {
	Endpoint string // 例如 http://127.0.0.1:7860
}

func (s *SDClient) Txt2Img(prompt string, opts map[string]interface{}) (ImageResult, error) {
	return s.Txt2ImgWithContext(context.Background(), prompt, opts)
}

func (s *SDClient) Txt2ImgWithContext(ctx context.Context, prompt string, opts map[string]interface{}) (ImageResult, error) {
	// 组装请求体
	body := map[string]interface{}{
		"prompt": prompt,
	}
	for k, v := range opts {
		body[k] = v
	}
	b, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", s.Endpoint+"/sdapi/v1/txt2img", bytes.NewReader(b))
	if err != nil {
		return ImageResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 3 * time.Minute, // 设置客户端超时
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