package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type OllamaClient struct {
	Endpoint string // 例如 http://127.0.0.1:11434
	Model    string // 如 "llama2"、"qwen" 等
}

type ollamaChatReq struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ollamaGenReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaResp struct {
	Response string `json:"response"`
}

func (o *OllamaClient) Chat(messages []Message, opts map[string]interface{}) (string, error) {
	reqBody := ollamaChatReq{Model: o.Model, Messages: messages}
	b, _ := json.Marshal(reqBody)
	resp, err := http.Post(o.Endpoint+"/api/chat", "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Ollama API error: %s", resp.Status)
	}
	var result ollamaResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Response, nil
}

func (o *OllamaClient) Generate(prompt string, opts map[string]interface{}) (string, error) {
	reqBody := ollamaGenReq{Model: o.Model, Prompt: prompt}
	b, _ := json.Marshal(reqBody)
	resp, err := http.Post(o.Endpoint+"/api/generate", "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Ollama API error: %s", resp.Status)
	}
	var result ollamaResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Response, nil
} 