package main

import (
	"fmt"
	"strings"
)

// optimizePromptForLocalSD 专门为本地SD优化提示词
func optimizePromptForLocalSD(prompt string) string {
	optimized := prompt
	
	// 1. 移除第一人称和复杂叙事
	optimized = strings.ReplaceAll(optimized, " us.", ".")
	optimized = strings.ReplaceAll(optimized, " us,", ",")
	optimized = strings.ReplaceAll(optimized, " us ", " them ")
	optimized = strings.ReplaceAll(optimized, " we ", " they ")
	optimized = strings.ReplaceAll(optimized, " our ", " their ")
	
	// 2. 简化复杂的动作序列
	actionWords := []string{"turned", "entered", "walked", "ran", "jumped", "looked", "smiled", "sat"}
	actionCount := 0
	for _, action := range actionWords {
		if strings.Contains(strings.ToLower(optimized), action) {
			actionCount++
		}
	}
	
	// 如果动作过多，简化句子
	if actionCount > 2 && strings.Contains(optimized, ",") {
		parts := strings.Split(optimized, ",")
		if len(parts) > 2 {
			optimized = strings.Join(parts[:2], ",")
		}
	}
	
	// 3. 优化句子结构
	optimized = strings.ReplaceAll(optimized, " and then ", ", ")
	optimized = strings.ReplaceAll(optimized, " while ", ", ")
	optimized = strings.ReplaceAll(optimized, " when ", ", ")
	
	// 4. 确保有明确的主体
	if !strings.Contains(strings.ToLower(optimized), "person") && 
	   !strings.Contains(strings.ToLower(optimized), "people") &&
	   !strings.Contains(strings.ToLower(optimized), "character") &&
	   !strings.Contains(strings.ToLower(optimized), "woman") &&
	   !strings.Contains(strings.ToLower(optimized), "man") {
		if strings.Contains(optimized, "Zhang Fengfeng") {
			optimized = "character " + optimized
		}
	}
	
	return strings.TrimSpace(optimized)
}

// buildLocalSDPrompt 构建适合本地SD的完整提示词
func buildLocalSDPrompt(optimizedPrompt string) string {
	styleKeywords := []string{
		"anime style",
		"illustration",
		"detailed", 
		"high quality",
		"masterpiece",
		"best quality",
	}
	
	return optimizedPrompt + ", " + strings.Join(styleKeywords, ", ")
}

func main() {
	fmt.Println("🎨 本地SD提示词优化测试")
	fmt.Println("==================================================")
	
	fmt.Println("💡 本地SD特点分析:")
	fmt.Println("✅ 无内容审查 - 敏感词汇不是问题")
	fmt.Println("✅ 完全控制 - 可以生成任何内容")
	fmt.Println("❌ 理解能力有限 - 复杂叙事是主要问题")
	fmt.Println("❌ 需要清晰指导 - 模糊描述效果差")
	
	// 测试用例
	testCases := []struct {
		name        string
		original    string
		issues      []string
		expectation string
	}{
		{
			name: "您的问题提示词",
			original: "Northern Myanmar people armed with guns surrounded the group. Zhang Fengfeng smirked, turned, and entered the small building, guns aimed at us., high quality, detailed",
			issues: []string{"复杂叙事", "多个动作", "第一人称", "句子过长"},
			expectation: "应该大幅简化",
		},
		{
			name: "简单敏感词测试",
			original: "people with guns in Myanmar, high quality, detailed",
			issues: []string{"无问题"},
			expectation: "本地SD应该能正常生成",
		},
		{
			name: "复杂动作序列",
			original: "character walked, turned, jumped, ran, smiled, sat down, looked around",
			issues: []string{"动作过多", "缺乏焦点"},
			expectation: "需要简化动作",
		},
		{
			name: "理想的简单描述",
			original: "Zhang Fengfeng smiling, anime style",
			issues: []string{"无问题"},
			expectation: "应该完美生成",
		},
	}
	
	fmt.Println("\n📝 提示词优化测试:")
	
	for i, tc := range testCases {
		fmt.Printf("\n--- 测试 %d: %s ---\n", i+1, tc.name)
		fmt.Printf("原始提示词: %s\n", tc.original)
		fmt.Printf("存在问题: %s\n", strings.Join(tc.issues, ", "))
		fmt.Printf("预期: %s\n", tc.expectation)
		
		// 执行优化
		optimized := optimizePromptForLocalSD(tc.original)
		fmt.Printf("优化后: %s\n", optimized)
		
		// 构建最终提示词
		final := buildLocalSDPrompt(optimized)
		fmt.Printf("最终提示词: %s\n", final)
		
		// 分析改进
		improvements := []string{}
		
		// 检查动作简化
		originalActions := countActions(tc.original)
		optimizedActions := countActions(optimized)
		if optimizedActions < originalActions {
			improvements = append(improvements, fmt.Sprintf("✅ 动作简化 (%d→%d)", originalActions, optimizedActions))
		}
		
		// 检查第一人称移除
		if strings.Contains(tc.original, " us") && !strings.Contains(optimized, " us") {
			improvements = append(improvements, "✅ 移除第一人称")
		}
		
		// 检查长度优化
		if len(optimized) < len(tc.original) {
			improvements = append(improvements, "✅ 简化长度")
		}
		
		// 检查主体明确性
		if hasSubject(final) && !hasSubject(tc.original) {
			improvements = append(improvements, "✅ 明确主体")
		}
		
		// 检查风格关键词
		if strings.Contains(final, "masterpiece") {
			improvements = append(improvements, "✅ 添加质量关键词")
		}
		
		if len(improvements) > 0 {
			fmt.Printf("改进效果: %s\n", strings.Join(improvements, ", "))
		} else {
			fmt.Printf("改进效果: 保持原有质量\n")
		}
		
		// 长度对比
		fmt.Printf("长度变化: %d → %d → %d 字符\n", 
			len(tc.original), len(optimized), len(final))
		
		// 复杂度评估
		complexity := assessComplexity(final)
		fmt.Printf("复杂度评估: %s\n", complexity)
	}
	
	fmt.Println("\n🎯 本地SD优化策略:")
	fmt.Println("1. ✅ 保留敏感词汇 - 本地SD无内容限制")
	fmt.Println("2. ✅ 简化复杂叙事 - 提高理解度")
	fmt.Println("3. ✅ 明确视觉主体 - 确保焦点清晰")
	fmt.Println("4. ✅ 添加质量关键词 - 提升生成质量")
	fmt.Println("5. ✅ 移除第一人称 - 适配图像生成")
	
	fmt.Println("\n💡 建议:")
	fmt.Println("• 如果您的本地SD仍然无法生成，问题可能在于:")
	fmt.Println("  1. 提示词过于复杂，SD模型理解困难")
	fmt.Println("  2. SD模型版本较老，理解能力有限")
	fmt.Println("  3. 系统资源不足，生成过程中断")
	fmt.Println("  4. SD配置参数不当")
	fmt.Println("• 建议使用简单、清晰的描述，避免复杂叙事")
	fmt.Println("• 可以尝试不同的SD模型或调整参数")
}

// countActions 计算动作词数量
func countActions(text string) int {
	actionWords := []string{"turned", "entered", "walked", "ran", "jumped", "looked", "smiled", "sat", "moved", "went"}
	count := 0
	lowerText := strings.ToLower(text)
	for _, action := range actionWords {
		if strings.Contains(lowerText, action) {
			count++
		}
	}
	return count
}

// hasSubject 检查是否有明确的主体
func hasSubject(text string) bool {
	subjects := []string{"person", "people", "character", "woman", "man", "girl", "boy"}
	lowerText := strings.ToLower(text)
	for _, subject := range subjects {
		if strings.Contains(lowerText, subject) {
			return true
		}
	}
	return false
}

// assessComplexity 评估提示词复杂度
func assessComplexity(text string) string {
	score := 0
	
	// 长度评分
	if len(text) > 150 {
		score += 2
	} else if len(text) > 100 {
		score += 1
	}
	
	// 逗号数量（句子复杂度）
	commaCount := strings.Count(text, ",")
	score += commaCount / 3
	
	// 动作数量
	actionCount := countActions(text)
	score += actionCount / 2
	
	// 复杂度评级
	if score <= 1 {
		return "简单 - 适合SD生成"
	} else if score <= 3 {
		return "中等 - 可能需要优化"
	} else {
		return "复杂 - 建议进一步简化"
	}
}
