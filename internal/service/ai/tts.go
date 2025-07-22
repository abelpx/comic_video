package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type TTSClient struct {
	Endpoint string // 例如 http://127.0.0.1:50021
}

func (t *TTSClient) Synthesize(text string, opts map[string]interface{}) ([]byte, error) {
	body := map[string]interface{}{
		"text": text,
	}
	for k, v := range opts {
		body[k] = v
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(t.Endpoint+"/tts", "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("TTS API error: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// GenerateVoice 生成语音（为新的AI服务提供的包装方法）
func (t *TTSClient) GenerateVoice(text, voiceModel, language string, speed, pitch, volume float64, emotion string) ([]byte, float64, error) {
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
		return nil, 0, err
	}

	// 估算音频时长（简单估算：每个字符约0.1秒）
	duration := float64(len(text)) * 0.1

	return audioData, duration, nil
}
