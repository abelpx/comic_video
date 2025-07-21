package main

import (
	"fmt"
	"log"
	"comic_video/internal/service/ai"
)

// MockOllamaClient 用于测试的模拟客户端
type MockOllamaClient struct{}

func (m *MockOllamaClient) Generate(prompt string, opts map[string]interface{}) (string, error) {
	// 简单的模拟翻译逻辑
	if contains(prompt, "动作场面中动态镜头的运用") {
		return "Dynamic camera shots in action scenes, capturing moments of character movement or running, showcasing urgency and motion", nil
	}
	
	if contains(prompt, "深夜雨巷") {
		return "Dark rainy alley at midnight with neon lights reflecting on wet pavement", nil
	}
	
	if contains(prompt, "废弃仓库") {
		return "Abandoned warehouse with broken shelves and flickering lights casting shadows", nil
	}
	
	// 默认翻译
	return "Translated content: professional illustration, high quality, detailed", nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func main() {
	fmt.Println("=== 提示词翻译功能测试 ===")
	
	// 创建模拟客户端
	mockOllama := &MockOllamaClient{}
	
	// 初始化翻译器
	ai.InitPromptTranslator(mockOllama)
	
	// 测试用例
	testCases := []struct {
		name   string
		input  string
	}{
		{
			name:  "动作场面描述",
			input: "动作场面中动态镜头的运用，捕捉人物移动或奔跑的瞬间，以展现紧迫感和动感。",
		},
		{
			name:  "场景描述",
			input: "深夜雨巷，霓虹灯在湿润地面反射出斑驳色彩",
		},
		{
			name:  "环境描述", 
			input: "废弃仓库内部，破旧货架投下阴影",
		},
		{
			name:  "英文提示词",
			input: "A beautiful landscape with mountains and rivers, high quality, detailed",
		},
		{
			name:  "简化提示词",
			input: "主角紧张地查看手机, high quality, detailed",
		},
	}
	
	for i, tc := range testCases {
		fmt.Printf("\n--- 测试 %d: %s ---\n", i+1, tc.name)
		fmt.Printf("原始提示词: %s\n", tc.input)
		
		// 执行翻译
		translated, err := ai.TranslatePrompt(tc.input)
		if err != nil {
			log.Printf("翻译失败: %v", err)
			fmt.Printf("翻译结果: [错误] %v\n", err)
		} else {
			fmt.Printf("翻译结果: %s\n", translated)
		}
		
		// 检查是否为英文
		isEnglish := isEnglishPrompt(translated)
		fmt.Printf("是否为英文: %v\n", isEnglish)
	}
	
	fmt.Println("\n=== 测试完成 ===")
}

// isEnglishPrompt 简单检测是否为英文提示词
func isEnglishPrompt(prompt string) bool {
	englishCount := 0
	chineseCount := 0
	
	for _, r := range prompt {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			englishCount++
		} else if r >= '\u4e00' && r <= '\u9fff' {
			chineseCount++
		}
	}
	
	total := englishCount + chineseCount
	if total == 0 {
		return true
	}
	
	return float64(englishCount)/float64(total) > 0.5
}
