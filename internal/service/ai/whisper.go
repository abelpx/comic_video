package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"mime/multipart"
)

type WhisperClient struct {
	Endpoint string // 例如 http://127.0.0.1:9000
}

func (w *WhisperClient) Transcribe(audio []byte, opts map[string]interface{}) (string, error) {
	// 假设API为 /asr，multipart上传
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("audio", "audio.wav")
	part.Write(audio)
	writer.Close()
	
	req, _ := http.NewRequest("POST", w.Endpoint+"/asr", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Whisper API error: %s", resp.Status)
	}
	var result struct{ Text string `json:"text"` }
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Text, nil
} 