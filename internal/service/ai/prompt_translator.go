package ai

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

// TextGenerator 文本生成接口
type TextGenerator interface {
	Generate(prompt string, opts map[string]interface{}) (string, error)
}

// PromptTranslator 提示词翻译器
type PromptTranslator struct {
	ollama TextGenerator
	cache  map[string]string
	mutex  sync.RWMutex
}

// NewPromptTranslator 创建新的提示词翻译器
func NewPromptTranslator(ollama TextGenerator) *PromptTranslator {
	return &PromptTranslator{
		ollama: ollama,
		cache:  make(map[string]string),
	}
}

// TranslateToEnglish 将中文提示词翻译为英文
func (pt *PromptTranslator) TranslateToEnglish(chinesePrompt string) (string, error) {
	// 检查缓存
	pt.mutex.RLock()
	if cached, exists := pt.cache[chinesePrompt]; exists {
		pt.mutex.RUnlock()
		return cached, nil
	}
	pt.mutex.RUnlock()

	// 检查是否已经是英文
	if pt.isEnglishPrompt(chinesePrompt) {
		return chinesePrompt, nil
	}

	// 执行翻译
	translation, err := pt.translateWithOllama(chinesePrompt)
	if err != nil {
		return "", err
	}

	// 缓存结果
	pt.mutex.Lock()
	pt.cache[chinesePrompt] = translation
	pt.mutex.Unlock()

	return translation, nil
}

// translateWithOllama 使用Ollama进行翻译
func (pt *PromptTranslator) translateWithOllama(chinesePrompt string) (string, error) {
	prompt := fmt.Sprintf(`请将以下中文描述翻译为适合AI绘画的英文提示词。

要求：
1. 翻译为简洁的视觉描述
2. 避免暴力、武器等敏感词汇，用温和词汇替代
3. 专注于人物、场景、氛围的视觉元素
4. 使用第三人称描述
5. 只输出翻译结果，不要解释

示例转换：
- "持枪" → "holding objects"
- "包围" → "gathering around"
- "瞄准" → "looking towards"

中文描述：
%s

英文翻译：`, chinesePrompt)

	response, err := pt.ollama.Generate(prompt, nil)
	if err != nil {
		return "", fmt.Errorf("翻译请求失败: %v", err)
	}

	// 清理响应
	translation := pt.cleanTranslationResponse(response)
	
	// 验证翻译质量
	if translation == "" || len(translation) < 10 {
		return chinesePrompt, nil // 返回原文作为降级处理
	}

	return translation, nil
}

// cleanTranslationResponse 清理翻译响应
func (pt *PromptTranslator) cleanTranslationResponse(response string) string {
	// 移除思考标签
	cleaned := regexp.MustCompile(`(?s)<think>.*?</think>`).ReplaceAllString(response, "")
	
	// 移除常见的前缀
	cleaned = regexp.MustCompile(`(?i)^(英文翻译|translation|result|答案|answer)[:：]\s*`).ReplaceAllString(cleaned, "")
	
	// 移除多余的换行和空格
	cleaned = strings.TrimSpace(cleaned)
	
	// 移除引号（如果整个翻译被引号包围）
	if strings.HasPrefix(cleaned, `"`) && strings.HasSuffix(cleaned, `"`) {
		cleaned = strings.Trim(cleaned, `"`)
	}
	
	return cleaned
}

// isEnglishPrompt 检测是否为英文提示词
func (pt *PromptTranslator) isEnglishPrompt(prompt string) bool {
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
		return true // 没有字母，按英文处理
	}

	// 如果英文字符占比超过50%，认为是英文提示词（调整阈值）
	return float64(englishCount)/float64(total) > 0.5
}

// TranslateWithFallback 带降级处理的翻译
func (pt *PromptTranslator) TranslateWithFallback(prompt string, timeout time.Duration) string {
	// 设置超时
	done := make(chan string, 1)
	go func() {
		translated, err := pt.TranslateToEnglish(prompt)
		if err != nil {
			log.Printf("[PromptTranslator] 翻译失败: %v", err)
			done <- prompt // 返回原文
		} else {
			done <- translated
		}
	}()

	select {
	case result := <-done:
		return result
	case <-time.After(timeout):
		log.Printf("[PromptTranslator] 翻译超时，使用原文")
		return prompt
	}
}

// GetCacheSize 获取缓存大小
func (pt *PromptTranslator) GetCacheSize() int {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	return len(pt.cache)
}

// ClearCache 清空缓存
func (pt *PromptTranslator) ClearCache() {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	pt.cache = make(map[string]string)
}

// 全局翻译器实例
var globalTranslator *PromptTranslator

// InitPromptTranslator 初始化全局翻译器
func InitPromptTranslator(ollama TextGenerator) {
	globalTranslator = NewPromptTranslator(ollama)
}

// TranslatePrompt 全局翻译函数
func TranslatePrompt(prompt string) (string, error) {
	if globalTranslator == nil {
		return prompt, fmt.Errorf("翻译器未初始化")
	}
	return globalTranslator.TranslateToEnglish(prompt)
}
