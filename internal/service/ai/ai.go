package ai

// ImageGen 文生图/图生图能力
// opts 可扩展如 steps/width/height/seed 等
// 返回图片结果
//
type ImageGen interface {
	Txt2Img(prompt string, opts map[string]interface{}) (ImageResult, error)
	Img2Img(image []byte, prompt string, opts map[string]interface{}) (ImageResult, error)
}

// TextGen 文本生成/对话能力
// opts 可扩展如 temperature/model 等
//
type TextGen interface {
	Chat(messages []Message, opts map[string]interface{}) (string, error)
	Generate(prompt string, opts map[string]interface{}) (string, error)
}

// Audio2Text 音频转文本能力
//
type Audio2Text interface {
	Transcribe(audio []byte, opts map[string]interface{}) (string, error)
}

// TTS 文本转语音能力
//
type TTS interface {
	Synthesize(text string, opts map[string]interface{}) ([]byte, error)
} 