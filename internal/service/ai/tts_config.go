package ai

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// TTSConfig TTS服务配置
type TTSConfig struct {
	Endpoint        string        `json:"endpoint"`
	Timeout         time.Duration `json:"timeout"`
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	HealthCheckURL  string        `json:"health_check_url"`
	DefaultVoice    string        `json:"default_voice"`
	DefaultLanguage string        `json:"default_language"`
	DefaultSpeed    float64       `json:"default_speed"`
	DefaultPitch    float64       `json:"default_pitch"`
	DefaultVolume   float64       `json:"default_volume"`
}

// NewTTSConfig 创建TTS配置
func NewTTSConfig() *TTSConfig {
	config := &TTSConfig{
		Endpoint:        getEnvOrDefault("TTS_ENDPOINT", "http://tts:8000"),
		Timeout:         parseDurationOrDefault("TTS_TIMEOUT", "30s"),
		MaxRetries:      parseIntOrDefault("TTS_MAX_RETRIES", 3),
		RetryDelay:      parseDurationOrDefault("TTS_RETRY_DELAY", "2s"),
		DefaultVoice:    getEnvOrDefault("TTS_DEFAULT_VOICE", "spark_tts_zh"),
		DefaultLanguage: getEnvOrDefault("TTS_DEFAULT_LANGUAGE", "zh"),
		DefaultSpeed:    parseFloatOrDefault("TTS_DEFAULT_SPEED", 1.0),
		DefaultPitch:    parseFloatOrDefault("TTS_DEFAULT_PITCH", 1.0),
		DefaultVolume:   parseFloatOrDefault("TTS_DEFAULT_VOLUME", 1.0),
	}
	
	config.HealthCheckURL = config.Endpoint + "/v2/health/ready"
	
	log.Printf("[TTS] 配置初始化完成: endpoint=%s, timeout=%v, max_retries=%d", 
		config.Endpoint, config.Timeout, config.MaxRetries)
	
	return config
}

// NewTTSClientWithConfig 使用配置创建TTS客户端
func NewTTSClientWithConfig(config *TTSConfig) *TTSClient {
	return &TTSClient{
		Endpoint: config.Endpoint,
	}
}

// TTSVoiceOptions 语音选项
type TTSVoiceOptions struct {
	VoiceModel string  `json:"voice_model"`
	Language   string  `json:"language"`
	Speed      float64 `json:"speed"`
	Pitch      float64 `json:"pitch"`
	Volume     float64 `json:"volume"`
	Emotion    string  `json:"emotion"`
}

// NewDefaultVoiceOptions 创建默认语音选项
func NewDefaultVoiceOptions() *TTSVoiceOptions {
	return &TTSVoiceOptions{
		VoiceModel: getEnvOrDefault("TTS_DEFAULT_VOICE", "spark_tts_zh"),
		Language:   getEnvOrDefault("TTS_DEFAULT_LANGUAGE", "zh"),
		Speed:      parseFloatOrDefault("TTS_DEFAULT_SPEED", 1.0),
		Pitch:      parseFloatOrDefault("TTS_DEFAULT_PITCH", 1.0),
		Volume:     parseFloatOrDefault("TTS_DEFAULT_VOLUME", 1.0),
		Emotion:    getEnvOrDefault("TTS_DEFAULT_EMOTION", "neutral"),
	}
}

// ApplyOptions 应用语音选项
func (opts *TTSVoiceOptions) ApplyOptions(customOpts map[string]interface{}) {
	if voiceModel, ok := customOpts["voice_model"].(string); ok && voiceModel != "" {
		opts.VoiceModel = voiceModel
	}
	if language, ok := customOpts["language"].(string); ok && language != "" {
		opts.Language = language
	}
	if speed, ok := customOpts["speed"].(float64); ok && speed > 0 {
		opts.Speed = speed
	}
	if pitch, ok := customOpts["pitch"].(float64); ok && pitch > 0 {
		opts.Pitch = pitch
	}
	if volume, ok := customOpts["volume"].(float64); ok && volume > 0 {
		opts.Volume = volume
	}
	if emotion, ok := customOpts["emotion"].(string); ok && emotion != "" {
		opts.Emotion = emotion
	}
}

// ToMap 转换为map
func (opts *TTSVoiceOptions) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"voice_model": opts.VoiceModel,
		"language":    opts.Language,
		"speed":       opts.Speed,
		"pitch":       opts.Pitch,
		"volume":      opts.Volume,
		"emotion":     opts.Emotion,
	}
}

// 工具函数 (getEnvOrDefault 已在 init.go 中定义)

func parseIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseDurationOrDefault(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	if parsed, err := time.ParseDuration(defaultValue); err == nil {
		return parsed
	}
	return 30 * time.Second
}

// TTSServiceInfo TTS服务信息
type TTSServiceInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Status      string   `json:"status"`
	Models      []string `json:"models"`
	Endpoint    string   `json:"endpoint"`
	LastChecked string   `json:"last_checked"`
}

// GetServiceInfo 获取TTS服务信息
func (t *TTSClient) GetServiceInfo() (*TTSServiceInfo, error) {
	info := &TTSServiceInfo{
		Name:        "Spark TTS",
		Version:     "1.0.0",
		Endpoint:    t.Endpoint,
		LastChecked: time.Now().Format(time.RFC3339),
	}

	// 检查服务状态
	if err := t.CheckHealth(); err != nil {
		info.Status = "unhealthy"
		return info, fmt.Errorf("service unhealthy: %w", err)
	}
	info.Status = "healthy"

	// 获取可用模型
	models, err := t.GetAvailableVoices()
	if err != nil {
		log.Printf("[TTS] 获取模型列表失败: %v", err)
		// 使用默认模型列表
		models = []string{"spark_tts_zh", "spark_tts_en", "spark_tts_mix"}
	}
	info.Models = models

	return info, nil
}

// ValidateVoiceOptions 验证语音选项
func ValidateVoiceOptions(opts *TTSVoiceOptions) error {
	if opts.VoiceModel == "" {
		return fmt.Errorf("voice_model is required")
	}
	if opts.Language == "" {
		return fmt.Errorf("language is required")
	}
	if opts.Speed <= 0 || opts.Speed > 3.0 {
		return fmt.Errorf("speed must be between 0 and 3.0")
	}
	if opts.Pitch <= 0 || opts.Pitch > 2.0 {
		return fmt.Errorf("pitch must be between 0 and 2.0")
	}
	if opts.Volume <= 0 || opts.Volume > 2.0 {
		return fmt.Errorf("volume must be between 0 and 2.0")
	}
	return nil
}

// SupportedLanguages 支持的语言列表
var SupportedLanguages = []string{
	"zh",    // 中文
	"en",    // 英文
	"zh-en", // 中英混合
}

// SupportedEmotions 支持的情感列表
var SupportedEmotions = []string{
	"neutral",  // 中性
	"happy",    // 开心
	"sad",      // 悲伤
	"angry",    // 愤怒
	"surprise", // 惊讶
	"fear",     // 恐惧
	"disgust",  // 厌恶
}

// IsLanguageSupported 检查语言是否支持
func IsLanguageSupported(language string) bool {
	for _, lang := range SupportedLanguages {
		if lang == language {
			return true
		}
	}
	return false
}

// IsEmotionSupported 检查情感是否支持
func IsEmotionSupported(emotion string) bool {
	for _, emo := range SupportedEmotions {
		if emo == emotion {
			return true
		}
	}
	return false
}
