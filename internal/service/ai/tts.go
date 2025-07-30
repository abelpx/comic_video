package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type TTSClient struct {
	Endpoint string // 例如 http://tts:8000
}

// TritonTTSRequest Triton TTS API请求结构
type TritonTTSRequest struct {
	Inputs []TritonInput `json:"inputs"`
}

type TritonInput struct {
	Name     string      `json:"name"`
	Shape    []int       `json:"shape"`
	DataType string      `json:"datatype"`
	Data     interface{} `json:"data"`
}

// TritonTTSResponse Triton TTS API响应结构
type TritonTTSResponse struct {
	Outputs []TritonOutput `json:"outputs"`
}

type TritonOutput struct {
	Name     string      `json:"name"`
	Shape    []int       `json:"shape"`
	DataType string      `json:"datatype"`
	Data     interface{} `json:"data"`
}

func (t *TTSClient) Synthesize(text string, opts map[string]interface{}) ([]byte, error) {
	// 构建Triton推理请求
	req := TritonTTSRequest{
		Inputs: []TritonInput{
			{
				Name:     "text",
				Shape:    []int{1},
				DataType: "BYTES",
				Data:     []string{text},
			},
		},
	}

	// 添加可选参数
	if voiceModel, ok := opts["voice_model"].(string); ok && voiceModel != "" {
		req.Inputs = append(req.Inputs, TritonInput{
			Name:     "voice_model",
			Shape:    []int{1},
			DataType: "BYTES",
			Data:     []string{voiceModel},
		})
	}

	if language, ok := opts["language"].(string); ok && language != "" {
		req.Inputs = append(req.Inputs, TritonInput{
			Name:     "language",
			Shape:    []int{1},
			DataType: "BYTES",
			Data:     []string{language},
		})
	}

	if speed, ok := opts["speed"].(float64); ok {
		req.Inputs = append(req.Inputs, TritonInput{
			Name:     "speed",
			Shape:    []int{1},
			DataType: "FP32",
			Data:     []float32{float32(speed)},
		})
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	log.Printf("[TTS] 发送请求到 %s/v2/models/spark_tts/infer", t.Endpoint)

	// 发送请求到Triton推理服务器
	resp, err := http.Post(t.Endpoint+"/v2/models/spark_tts/infer", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("TTS request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TTS API error: %s, body: %s", resp.Status, string(body))
	}

	// 解析响应
	var ttsResp TritonTTSResponse
	if err := json.NewDecoder(resp.Body).Decode(&ttsResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	// 提取音频数据
	for _, output := range ttsResp.Outputs {
		if output.Name == "audio" {
			if audioData, ok := output.Data.([]interface{}); ok && len(audioData) > 0 {
				if audioBytes, ok := audioData[0].(string); ok {
					// 如果是base64编码的音频数据，需要解码
					// 这里假设返回的是原始音频字节
					return []byte(audioBytes), nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no audio data found in response")
}

// GenerateVoice 生成语音（为新的AI服务提供的包装方法）
func (t *TTSClient) GenerateVoice(text, voiceModel, language string, speed, pitch, volume float64, emotion string) ([]byte, float64, error) {
	log.Printf("[TTS] 开始生成语音: text_length=%d, voice_model=%s, language=%s", len(text), voiceModel, language)

	// 检查服务健康状态
	if err := t.CheckHealth(); err != nil {
		return nil, 0, fmt.Errorf("TTS service not ready: %w", err)
	}

	opts := map[string]interface{}{
		"voice_model": voiceModel,
		"language":    language,
		"speed":       speed,
		"pitch":       pitch,
		"volume":      volume,
		"emotion":     emotion,
	}

	audioData, err := t.Synthesize(text, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("语音合成失败: %w", err)
	}

	// 估算音频时长（中文：每个字符约0.15秒，英文：每个单词约0.3秒）
	duration := t.estimateDuration(text)

	log.Printf("[TTS] 语音生成完成: audio_size=%d bytes, duration=%.2f seconds", len(audioData), duration)
	return audioData, duration, nil
}

// CheckHealth 检查TTS服务健康状态
func (t *TTSClient) CheckHealth() error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(t.Endpoint + "/v2/health/ready")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("TTS service not ready, status: %s", resp.Status)
	}
	return nil
}

// estimateDuration 估算音频时长
func (t *TTSClient) estimateDuration(text string) float64 {
	// 计算中文字符数
	chineseCount := 0
	englishWordCount := 0

	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			chineseCount++
		}
	}

	// 计算英文单词数（简单估算）
	words := strings.Fields(text)
	for _, word := range words {
		hasEnglish := false
		for _, r := range word {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				hasEnglish = true
				break
			}
		}
		if hasEnglish {
			englishWordCount++
		}
	}

	// 中文：每个字符0.15秒，英文：每个单词0.3秒
	duration := float64(chineseCount)*0.15 + float64(englishWordCount)*0.3

	// 最小时长0.5秒
	if duration < 0.5 {
		duration = 0.5
	}

	return duration
}

// GetAvailableVoices 获取可用的语音模型列表
func (t *TTSClient) GetAvailableVoices() ([]string, error) {
	resp, err := http.Get(t.Endpoint + "/v2/models")
	if err != nil {
		return nil, fmt.Errorf("get models failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get models API error: %s", resp.Status)
	}

	// 这里应该解析实际的模型列表响应
	// 暂时返回一些默认的语音模型
	return []string{
		"spark_tts_zh",
		"spark_tts_en",
		"spark_tts_mix",
	}, nil
}
