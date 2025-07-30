package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"comic_video/internal/service/ai"
)

// TTSHandler TTS处理器
type TTSHandler struct {
	ttsClient *ai.TTSClient
}

// NewTTSHandler 创建TTS处理器
func NewTTSHandler(ttsClient *ai.TTSClient) *TTSHandler {
	return &TTSHandler{
		ttsClient: ttsClient,
	}
}

// TTSRequest TTS请求结构
type TTSRequest struct {
	Text        string  `json:"text" binding:"required"`
	VoiceModel  string  `json:"voice_model"`
	Language    string  `json:"language"`
	Speed       float64 `json:"speed"`
	Pitch       float64 `json:"pitch"`
	Volume      float64 `json:"volume"`
	Emotion     string  `json:"emotion"`
}

// GenerateVoice 生成语音
func (h *TTSHandler) GenerateVoice(c *gin.Context) {
	var req TTSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 验证文本长度
	if len(req.Text) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文本内容不能为空",
		})
		return
	}

	if len(req.Text) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文本长度不能超过1000个字符",
		})
		return
	}

	// 设置默认值
	if req.VoiceModel == "" {
		req.VoiceModel = "spark_tts_zh"
	}
	if req.Language == "" {
		req.Language = "zh"
	}
	if req.Speed <= 0 {
		req.Speed = 1.0
	}
	if req.Pitch <= 0 {
		req.Pitch = 1.0
	}
	if req.Volume <= 0 {
		req.Volume = 1.0
	}
	if req.Emotion == "" {
		req.Emotion = "neutral"
	}

	// 验证参数范围
	if req.Speed < 0.5 || req.Speed > 2.0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "语速必须在0.5-2.0之间",
		})
		return
	}

	if req.Pitch < 0.5 || req.Pitch > 2.0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "音调必须在0.5-2.0之间",
		})
		return
	}

	if req.Volume < 0.1 || req.Volume > 2.0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "音量必须在0.1-2.0之间",
		})
		return
	}

	// 验证语言支持
	if !ai.IsLanguageSupported(req.Language) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的语言: " + req.Language,
		})
		return
	}

	// 验证情感支持
	if !ai.IsEmotionSupported(req.Emotion) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的情感: " + req.Emotion,
		})
		return
	}

	log.Printf("[TTS] 收到语音生成请求: text_length=%d, voice=%s, language=%s", 
		len(req.Text), req.VoiceModel, req.Language)

	// 调用TTS服务生成语音
	audioData, duration, err := h.ttsClient.GenerateVoice(
		req.Text,
		req.VoiceModel,
		req.Language,
		req.Speed,
		req.Pitch,
		req.Volume,
		req.Emotion,
	)
	if err != nil {
		log.Printf("[TTS] 语音生成失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "语音生成失败: " + err.Error(),
		})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "audio/wav")
	c.Header("Content-Length", strconv.Itoa(len(audioData)))
	c.Header("X-Audio-Duration", strconv.FormatFloat(duration, 'f', 2, 64))
	c.Header("X-Text-Length", strconv.Itoa(len(req.Text)))
	c.Header("X-Voice-Model", req.VoiceModel)
	c.Header("X-Language", req.Language)

	// 返回音频数据
	c.Data(http.StatusOK, "audio/wav", audioData)
}

// GetVoices 获取可用的语音模型
func (h *TTSHandler) GetVoices(c *gin.Context) {
	voices, err := h.ttsClient.GetAvailableVoices()
	if err != nil {
		log.Printf("[TTS] 获取语音模型失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取语音模型失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取语音模型成功",
		"data": gin.H{
			"voices":    voices,
			"languages": ai.SupportedLanguages,
			"emotions":  ai.SupportedEmotions,
		},
	})
}

// GetServiceInfo 获取TTS服务信息
func (h *TTSHandler) GetServiceInfo(c *gin.Context) {
	info, err := h.ttsClient.GetServiceInfo()
	if err != nil {
		log.Printf("[TTS] 获取服务信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取服务信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取服务信息成功",
		"data":    info,
	})
}

// HealthCheck TTS服务健康检查
func (h *TTSHandler) HealthCheck(c *gin.Context) {
	// 检查TTS服务健康状态
	err := h.ttsClient.CheckHealth()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    503,
			"message": "TTS服务不可用: " + err.Error(),
			"status":  "unhealthy",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "TTS服务正常",
		"status":  "healthy",
	})
}

// TestVoice 测试语音生成
func (h *TTSHandler) TestVoice(c *gin.Context) {
	voiceModel := c.DefaultQuery("voice_model", "spark_tts_zh")
	language := c.DefaultQuery("language", "zh")
	
	// 根据语言选择测试文本
	var testText string
	switch language {
	case "zh":
		testText = "你好，这是一个语音合成测试。"
	case "en":
		testText = "Hello, this is a text-to-speech test."
	case "zh-en":
		testText = "你好Hello，这是一个mixed language test。"
	default:
		testText = "Hello, this is a test."
	}

	log.Printf("[TTS] 执行语音测试: voice=%s, language=%s", voiceModel, language)

	// 生成测试语音
	audioData, duration, err := h.ttsClient.GenerateVoice(
		testText,
		voiceModel,
		language,
		1.0, // speed
		1.0, // pitch
		1.0, // volume
		"neutral", // emotion
	)
	if err != nil {
		log.Printf("[TTS] 语音测试失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "语音测试失败: " + err.Error(),
		})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "audio/wav")
	c.Header("Content-Length", strconv.Itoa(len(audioData)))
	c.Header("X-Audio-Duration", strconv.FormatFloat(duration, 'f', 2, 64))
	c.Header("X-Test-Text", testText)
	c.Header("X-Voice-Model", voiceModel)
	c.Header("X-Language", language)

	// 返回测试音频
	c.Data(http.StatusOK, "audio/wav", audioData)
}

// GetConfig 获取TTS配置信息
func (h *TTSHandler) GetConfig(c *gin.Context) {
	config := ai.NewTTSConfig()
	
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取TTS配置成功",
		"data": gin.H{
			"endpoint":         config.Endpoint,
			"timeout":          config.Timeout.String(),
			"max_retries":      config.MaxRetries,
			"default_voice":    config.DefaultVoice,
			"default_language": config.DefaultLanguage,
			"default_speed":    config.DefaultSpeed,
			"default_pitch":    config.DefaultPitch,
			"default_volume":   config.DefaultVolume,
			"supported_languages": ai.SupportedLanguages,
			"supported_emotions":  ai.SupportedEmotions,
		},
	})
}
