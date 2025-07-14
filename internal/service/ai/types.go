package ai

// ImageResult 图片生成结果
// 可扩展更多字段，如base64、url等
//
type ImageResult struct {
	Data []byte // 图片二进制
	URL  string // 可选，图片URL
}

// Message 对话消息体
//
type Message struct {
	Role    string // user/assistant/system
	Content string
}

// 业务场景建议：
// 1. 漫画生成：TextGen 生成剧情脚本/分镜描述，ImageGen 生成分镜图片
// 2. 小说/推文生成：TextGen 生成长文本/短文案
// 3. 视频转动漫：Audio2Text 识别字幕，TextGen 生成分镜描述，ImageGen 生成动漫画面，TTS 合成配音
// 4. 其它AI能力可按需扩展
// 