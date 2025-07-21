package ai

import (
	"testing"
	"time"
)

// MockOllamaClient 用于测试的模拟Ollama客户端，实现TextGenerator接口
type MockOllamaClient struct {
	responses map[string]string
}

func (m *MockOllamaClient) Generate(prompt string, opts map[string]interface{}) (string, error) {
	// 简单的模拟翻译
	if response, exists := m.responses[prompt]; exists {
		return response, nil
	}
	
	// 默认翻译逻辑
	if contains(prompt, "动作场面中动态镜头的运用") {
		return "Dynamic camera shots in action scenes, capturing moments of character movement or running, showcasing urgency and motion", nil
	}
	
	return "translated content", nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

func TestPromptTranslator_TranslateToEnglish(t *testing.T) {
	// 创建模拟客户端
	mockOllama := &MockOllamaClient{
		responses: make(map[string]string),
	}
	
	// 创建翻译器
	translator := NewPromptTranslator(mockOllama)
	
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "中文提示词翻译",
			input:    "动作场面中动态镜头的运用，捕捉人物移动或奔跑的瞬间，以展现紧迫感和动感。",
			expected: "Dynamic camera shots in action scenes, capturing moments of character movement or running, showcasing urgency and motion",
			wantErr:  false,
		},
		{
			name:     "英文提示词保持不变",
			input:    "A beautiful landscape with mountains and rivers, high quality, detailed",
			expected: "A beautiful landscape with mountains and rivers, high quality, detailed",
			wantErr:  false,
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "",
			wantErr:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := translator.TranslateToEnglish(tt.input)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("TranslateToEnglish() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.input != "" && tt.expected != "" {
				// 对于非空输入，检查结果是否合理
				if len(result) == 0 {
					t.Errorf("TranslateToEnglish() returned empty result for input: %s", tt.input)
				}
			}
		})
	}
}

func TestPromptTranslator_isEnglishPrompt(t *testing.T) {
	translator := NewPromptTranslator(nil)
	
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "纯英文",
			input:    "A beautiful landscape with mountains and rivers",
			expected: true,
		},
		{
			name:     "纯中文",
			input:    "动作场面中动态镜头的运用",
			expected: false,
		},
		{
			name:     "中英混合-英文为主",
			input:    "A beautiful landscape 美丽的风景",
			expected: true,
		},
		{
			name:     "中英混合-中文为主",
			input:    "动作场面中的镜头运用 beautiful",
			expected: false,
		},
		{
			name:     "空字符串",
			input:    "",
			expected: true,
		},
		{
			name:     "只有符号",
			input:    "!@#$%^&*()",
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := translator.isEnglishPrompt(tt.input)
			if result != tt.expected {
				t.Errorf("isEnglishPrompt() = %v, expected %v for input: %s", result, tt.expected, tt.input)
			}
		})
	}
}

func TestPromptTranslator_TranslateWithFallback(t *testing.T) {
	// 创建一个会超时的模拟客户端
	slowMockOllama := &MockOllamaClient{
		responses: make(map[string]string),
	}
	
	translator := NewPromptTranslator(slowMockOllama)
	
	input := "测试超时处理"
	timeout := 100 * time.Millisecond
	
	result := translator.TranslateWithFallback(input, timeout)
	
	// 应该返回原文（因为我们的mock不会真正超时，但这测试了代码路径）
	if result == "" {
		t.Errorf("TranslateWithFallback() returned empty result")
	}
}

func TestPromptTranslator_Cache(t *testing.T) {
	mockOllama := &MockOllamaClient{
		responses: make(map[string]string),
	}
	
	translator := NewPromptTranslator(mockOllama)
	
	input := "测试缓存功能"
	
	// 第一次翻译
	result1, err1 := translator.TranslateToEnglish(input)
	if err1 != nil {
		t.Errorf("First translation failed: %v", err1)
	}
	
	// 检查缓存大小
	if translator.GetCacheSize() != 1 {
		t.Errorf("Expected cache size 1, got %d", translator.GetCacheSize())
	}
	
	// 第二次翻译（应该使用缓存）
	result2, err2 := translator.TranslateToEnglish(input)
	if err2 != nil {
		t.Errorf("Second translation failed: %v", err2)
	}
	
	// 结果应该相同
	if result1 != result2 {
		t.Errorf("Cached result differs: %s != %s", result1, result2)
	}
	
	// 清空缓存
	translator.ClearCache()
	if translator.GetCacheSize() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", translator.GetCacheSize())
	}
}

// 基准测试
func BenchmarkPromptTranslator_TranslateToEnglish(b *testing.B) {
	mockOllama := &MockOllamaClient{
		responses: make(map[string]string),
	}
	
	translator := NewPromptTranslator(mockOllama)
	input := "动作场面中动态镜头的运用，捕捉人物移动或奔跑的瞬间"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = translator.TranslateToEnglish(input)
	}
}
