package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"encoding/base64"
)

type SDClient struct {
	Endpoint string // 例如 http://127.0.0.1:7860
}

func (s *SDClient) Txt2Img(prompt string, opts map[string]interface{}) (ImageResult, error) {
	// 组装请求体
	body := map[string]interface{}{
		"prompt": prompt,
	}
	for k, v := range opts {
		body[k] = v
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(s.Endpoint+"/sdapi/v1/txt2img", "application/json", bytes.NewReader(b))
	if err != nil {
		return ImageResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ImageResult{}, fmt.Errorf("SD API error: %s", resp.Status)
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

func (s *SDClient) Img2Img(image []byte, prompt string, opts map[string]interface{}) (ImageResult, error) {
	// 可扩展，暂未实现
	return ImageResult{}, fmt.Errorf("not implemented")
}

// decodeBase64Image 解码base64图片
func decodeBase64Image(b64 string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(b64)
} 