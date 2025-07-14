package ai

var (
	imageGenProviders = map[string]ImageGen{}
	textGenProviders  = map[string]TextGen{}
	audio2TextProviders = map[string]Audio2Text{}
	ttsProviders = map[string]TTS{}
)

func RegisterImageGen(name string, provider ImageGen) {
	imageGenProviders[name] = provider
}
func GetImageGen(name string) ImageGen {
	return imageGenProviders[name]
}

func RegisterTextGen(name string, provider TextGen) {
	textGenProviders[name] = provider
}
func GetTextGen(name string) TextGen {
	return textGenProviders[name]
}

func RegisterAudio2Text(name string, provider Audio2Text) {
	audio2TextProviders[name] = provider
}
func GetAudio2Text(name string) Audio2Text {
	return audio2TextProviders[name]
}

func RegisterTTS(name string, provider TTS) {
	ttsProviders[name] = provider
}
func GetTTS(name string) TTS {
	return ttsProviders[name]
} 