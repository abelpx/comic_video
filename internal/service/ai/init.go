package ai

// Config 结构可从全局配置读取
// 这里只做演示
var DefaultSD = &SDClient{Endpoint: "http://127.0.0.1:7860"}
var DefaultOllama = &OllamaClient{Endpoint: "http://127.0.0.1:11434", Model: "llama2"}
var DefaultWhisper = &WhisperClient{Endpoint: "http://127.0.0.1:9000"}
var DefaultTTS = &TTSClient{Endpoint: "http://127.0.0.1:50021"}

func InitAIProviders() {
	RegisterImageGen("sd", DefaultSD)
	RegisterTextGen("ollama", DefaultOllama)
	RegisterAudio2Text("whisper", DefaultWhisper)
	RegisterTTS("edge", DefaultTTS)
	// 可扩展更多
} 