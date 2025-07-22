package main

import (
	"fmt"
	"strings"
)

// optimizePromptForSD 优化提示词使其适合SD生成（复制自修复后的代码）
func optimizePromptForSD(prompt string) string {
	// 敏感词汇替换映射
	sensitiveReplacements := map[string]string{
		"armed with guns":     "holding objects",
		"guns":               "objects",
		"gun":                "object",
		"weapon":             "tool",
		"weapons":            "tools",
		"aimed at":           "looking towards",
		"surrounded":         "gathered around",
		"threatening":        "serious",
		"violence":           "drama",
		"attack":             "approach",
		"fighting":           "interaction",
		"Northern Myanmar":   "people from remote area",
		"Myanmar":            "remote area",
	}
	
	// 执行替换
	optimized := prompt
	for sensitive, replacement := range sensitiveReplacements {
		optimized = strings.ReplaceAll(optimized, sensitive, replacement)
	}
	
	// 移除第一人称和复杂叙事
	optimized = strings.ReplaceAll(optimized, " us.", ".")
	optimized = strings.ReplaceAll(optimized, " us,", ",")
	optimized = strings.ReplaceAll(optimized, " we ", " they ")
	optimized = strings.ReplaceAll(optimized, " our ", " their ")
	
	// 简化复杂句子
	if strings.Contains(optimized, ",") && len(optimized) > 100 {
		// 如果句子太长且复杂，取前半部分
		parts := strings.Split(optimized, ",")
		if len(parts) > 2 {
			optimized = strings.Join(parts[:2], ",")
		}
	}
	
	return strings.TrimSpace(optimized)
}

// buildEnhancedPrompt 构建增强的提示词
func buildEnhancedPrompt(optimizedPrompt string) string {
	styleKeywords := []string{
		"anime style",
		"illustration", 
		"high quality",
		"detailed",
		"safe content",
	}
	
	return optimizedPrompt + ", " + strings.Join(styleKeywords, ", ")
}

func main() {
	fmt.Println("🎨 提示词优化测试")
	fmt.Println("==================================================")
	
	// 测试用例
	testCases := []struct {
		name     string
		original string
		issues   []string
	}{
		{
			name: "您提供的问题提示词",
			original: "Northern Myanmar people armed with guns surrounded the group. Zhang Fengfeng smirked, turned, and entered the small building, guns aimed at us.",
			issues: []string{"敏感词汇", "复杂叙事", "第一人称", "暴力内容"},
		},
		{
			name: "其他敏感内容示例",
			original: "Armed soldiers with weapons attacking the village, violence everywhere",
			issues: []string{"武器描述", "暴力场景"},
		},
		{
			name: "复杂叙事示例", 
			original: "The character ran quickly, jumped over the fence, fought with enemies, and finally escaped to safety",
			issues: []string{"复杂动作序列"},
		},
		{
			name: "正常内容示例",
			original: "A young woman with long hair standing in a beautiful garden",
			issues: []string{"无问题"},
		},
	}
	
	fmt.Println("\n📝 提示词优化测试:")
	
	for i, tc := range testCases {
		fmt.Printf("\n--- 测试 %d: %s ---\n", i+1, tc.name)
		fmt.Printf("原始提示词: %s\n", tc.original)
		fmt.Printf("存在问题: %s\n", strings.Join(tc.issues, ", "))
		
		// 执行优化
		optimized := optimizePromptForSD(tc.original)
		fmt.Printf("优化后: %s\n", optimized)
		
		// 构建最终提示词
		final := buildEnhancedPrompt(optimized)
		fmt.Printf("最终提示词: %s\n", final)
		
		// 分析改进
		improvements := []string{}
		if strings.Contains(tc.original, "gun") && !strings.Contains(optimized, "gun") {
			improvements = append(improvements, "✅ 移除武器词汇")
		}
		if strings.Contains(tc.original, "Myanmar") && !strings.Contains(optimized, "Myanmar") {
			improvements = append(improvements, "✅ 移除地理敏感词")
		}
		if strings.Contains(tc.original, " us") && !strings.Contains(optimized, " us") {
			improvements = append(improvements, "✅ 移除第一人称")
		}
		if len(optimized) < len(tc.original) {
			improvements = append(improvements, "✅ 简化复杂描述")
		}
		if strings.Contains(final, "anime style") {
			improvements = append(improvements, "✅ 添加艺术风格")
		}
		
		if len(improvements) > 0 {
			fmt.Printf("改进效果: %s\n", strings.Join(improvements, ", "))
		} else {
			fmt.Printf("改进效果: 保持原有内容\n")
		}
		
		// 长度对比
		fmt.Printf("长度变化: %d → %d → %d 字符\n", 
			len(tc.original), len(optimized), len(final))
	}
	
	fmt.Println("\n🎯 优化策略总结:")
	fmt.Println("1. ✅ 敏感词汇替换:")
	fmt.Println("   - 'armed with guns' → 'holding objects'")
	fmt.Println("   - 'surrounded' → 'gathered around'")
	fmt.Println("   - 'aimed at' → 'looking towards'")
	fmt.Println("   - 'Northern Myanmar' → 'people from remote area'")
	
	fmt.Println("2. ✅ 叙事简化:")
	fmt.Println("   - 移除第一人称 ('us', 'we', 'our')")
	fmt.Println("   - 简化复杂动作序列")
	fmt.Println("   - 专注于视觉元素")
	
	fmt.Println("3. ✅ 风格增强:")
	fmt.Println("   - 添加 'anime style, illustration'")
	fmt.Println("   - 添加 'high quality, detailed'")
	fmt.Println("   - 添加 'safe content' 确保安全")
	
	fmt.Println("\n🚀 预期效果:")
	fmt.Println("• SD能够理解和生成的安全内容")
	fmt.Println("• 清晰的视觉焦点和构图")
	fmt.Println("• 避免敏感内容被过滤")
	fmt.Println("• 提高图像生成成功率")
	
	fmt.Println("\n✨ 您的问题提示词现在应该能够正常生成图像了！")
}
