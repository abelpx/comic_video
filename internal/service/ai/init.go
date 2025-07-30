package ai

import (
	"log"
	"os"
)

// Config 结构可从全局配置读取
// 这里只做演示
var DefaultSD = &SDClient{Endpoint: getEnvOrDefault("SD_ENDPOINT", "http://127.0.0.1:7860")}
var DefaultOllama = &OllamaClient{
	Endpoint: getEnvOrDefault("OLLAMA_ENDPOINT", "http://127.0.0.1:11434"),
	Model:    getEnvOrDefault("OLLAMA_MODEL", "llama2"),
}
var DefaultWhisper = &WhisperClient{Endpoint: getEnvOrDefault("WHISPER_ENDPOINT", "http://127.0.0.1:9000")}

// 使用新的TTS配置
var DefaultTTSConfig = NewTTSConfig()
var DefaultTTS = NewTTSClientWithConfig(DefaultTTSConfig)

func InitAIProviders() {
	log.Printf("[AI] 初始化AI服务提供者...")

	// 注册各种AI服务
	RegisterImageGen("sd", DefaultSD)
	RegisterTextGen("ollama", DefaultOllama)
	RegisterAudio2Text("whisper", DefaultWhisper)
	RegisterTTS("spark", DefaultTTS) // 使用spark作为TTS提供者名称

	log.Printf("[AI] AI服务提供者初始化完成")
	log.Printf("[AI] - SD: %s", DefaultSD.Endpoint)
	log.Printf("[AI] - Ollama: %s (model: %s)", DefaultOllama.Endpoint, DefaultOllama.Model)
	log.Printf("[AI] - Whisper: %s", DefaultWhisper.Endpoint)
	log.Printf("[AI] - TTS: %s", DefaultTTS.Endpoint)
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}