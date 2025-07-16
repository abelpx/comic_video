package ai

var imageGenRegistry = make(map[string]ImageGen)
var textGenRegistry = make(map[string]TextGen)
var audio2TextRegistry = make(map[string]Audio2Text)
var ttsRegistry = make(map[string]TTS)

func RegisterImageGen(name string, gen ImageGen) {
	imageGenRegistry[name] = gen
}
func GetImageGen(name string) ImageGen {
	return imageGenRegistry[name]
}
func RegisterTextGen(name string, gen TextGen) {
	textGenRegistry[name] = gen
}
func GetTextGen(name string) TextGen {
	return textGenRegistry[name]
}
func RegisterAudio2Text(name string, gen Audio2Text) {
	audio2TextRegistry[name] = gen
}
func GetAudio2Text(name string) Audio2Text {
	return audio2TextRegistry[name]
}
func RegisterTTS(name string, gen TTS) {
	ttsRegistry[name] = gen
}
func GetTTS(name string) TTS {
	return ttsRegistry[name]
} 