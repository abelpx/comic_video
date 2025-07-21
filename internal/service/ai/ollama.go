package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OllamaClient struct {
	Endpoint string // 例如 http://127.0.0.1:11434
	Model    string // 如 "llama2"、"qwen" 等
	ApiKey   string // 新增，支持 API Key
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
	Done     bool   `json:"done"`
}

func (o *OllamaClient) Chat(messages []Message, opts map[string]interface{}) (string, error) {
	reqBody := ollamaChatReq{Model: o.Model, Messages: messages}
	b, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", o.Endpoint+"/api/chat", bytes.NewReader(b))
	if o.ApiKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.ApiKey)
	}
	resp, err := http.DefaultClient.Do(req)
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
	req, _ := http.NewRequest("POST", o.Endpoint+"/api/generate", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if o.ApiKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.ApiKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Ollama API error: %s", resp.Status)
	}

	// Ollama可能返回流式响应，需要逐行解析
	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)

	for decoder.More() {
		var result ollamaResp
		if err := decoder.Decode(&result); err != nil {
			// 如果解析失败，尝试读取剩余内容作为普通文本
			remaining, _ := io.ReadAll(decoder.Buffered())
			if len(remaining) > 0 {
				fullResponse.WriteString(string(remaining))
			}
			break
		}
		fullResponse.WriteString(result.Response)
		if result.Done {
			break
		}
	}

	response := strings.TrimSpace(fullResponse.String())
	if response == "" {
		return "", fmt.Errorf("empty response from Ollama")
	}

	return response, nil
}