package main

import (
	"fmt"
	"log"
	"comic_video/internal/service/ai"
)

// DemoOllamaClient 演示用的Ollama客户端
type DemoOllamaClient struct{}

func (d *DemoOllamaClient) Generate(prompt string, opts map[string]interface{}) (string, error) {
	// 模拟真实的翻译响应
	if contains(prompt, "动作场面中动态镜头的运用") {
		return "Dynamic camera shots in action scenes, capturing moments of character movement or running, showcasing urgency and motion", nil
	}
	
	if contains(prompt, "深夜雨巷") {
		return "Dark rainy alley at midnight, neon lights reflecting on wet pavement creating colorful distortions", nil
	}
	
	if contains(prompt, "废弃仓库") {
		return "Abandoned warehouse interior with broken shelves forming maze-like passages, flickering lights casting web-like shadows", nil
	}
	
	if contains(prompt, "铁笼囚禁") {
		return "Iron cage confinement with rusty bars reflecting moonlight, creating an oppressive atmosphere", nil
	}
	
	if contains(prompt, "主角紧张地查看手机") {
		return "Protagonist nervously checking phone with tense facial expression under dim lighting", nil
	}
	
	// 默认翻译
	return "Professional illustration with high quality details and cinematic composition", nil
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func main() {
	fmt.Println("🎬 漫画视频生成系统 - 提示词翻译增强演示")
	fmt.Println("============================================================")
	
	// 初始化翻译器
	demoOllama := &DemoOllamaClient{}
	ai.InitPromptTranslator(demoOllama)
	
	// 模拟分镜描述
	chinesePanels := []string{
		"动作场面中动态镜头的运用，捕捉人物移动或奔跑的瞬间，以展现紧迫感和动感。",
		"深夜雨巷，霓虹灯在湿润地面反射出斑驳色彩，主角紧张地四处张望",
		"废弃仓库内部，破旧货架构成迷宫般的通道，摇晃的灯泡投下蛛网状阴影",
		"铁笼囚禁场景，锈蚀的铁栏在月光下泛着冷光，营造压抑氛围",
		"主角紧张地查看手机，面部表情紧绷，昏暗灯光下显得格外焦虑",
	}
	
	fmt.Println("\n📝 原始中文分镜描述:")
	for i, panel := range chinesePanels {
		fmt.Printf("%d. %s\n", i+1, panel)
	}
	
	fmt.Println("\n🔄 执行提示词翻译...")
	
	translatedPanels := make([]string, len(chinesePanels))
	for i, panel := range chinesePanels {
		translated, err := ai.TranslatePrompt(panel)
		if err != nil {
			log.Printf("翻译失败: %v", err)
			translatedPanels[i] = panel // 使用原文
		} else {
			translatedPanels[i] = translated
		}
	}
	
	fmt.Println("\n✅ 翻译后的英文提示词:")
	for i, panel := range translatedPanels {
		fmt.Printf("%d. %s\n", i+1, panel)
	}
	
	fmt.Println("\n🎨 模拟发送给Stable Diffusion的最终提示词:")
	for i, panel := range translatedPanels {
		finalPrompt := panel + ", high quality, detailed, professional illustration, cinematic composition"
		fmt.Printf("%d. %s\n", i+1, finalPrompt)
	}
	
	// 展示翻译前后的对比
	fmt.Println("\n📊 翻译效果对比:")
	fmt.Println("============================================================")
	
	for i := 0; i < len(chinesePanels); i++ {
		fmt.Printf("\n场景 %d:\n", i+1)
		fmt.Printf("🇨🇳 中文: %s\n", chinesePanels[i])
		fmt.Printf("🇺🇸 英文: %s\n", translatedPanels[i])
		
		// 检查翻译质量
		if isEnglishPrompt(translatedPanels[i]) {
			fmt.Printf("✅ 翻译成功 - 已转换为英文\n")
		} else {
			fmt.Printf("⚠️  翻译可能不完整 - 仍包含中文字符\n")
		}
	}
	
	fmt.Println("\n🎯 系统优势:")
	fmt.Println("• 自动检测中文提示词并翻译为英文")
	fmt.Println("• 提高Stable Diffusion的理解准确性")
	fmt.Println("• 缓存机制避免重复翻译")
	fmt.Println("• 降级处理确保系统稳定性")
	fmt.Println("• 保持艺术性和技术性词汇的准确性")
	
	fmt.Println("\n🚀 演示完成！系统已准备好处理中文输入并生成高质量图像。")
}

// isEnglishPrompt 检测是否为英文提示词
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
