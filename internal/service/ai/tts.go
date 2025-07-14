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