package tweet

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"comic_video/internal/service/ai"
)

// Service 推文生成服务
type Service struct {
	aiService *ai.Service
}

// NewService 创建推文生成服务
func NewService(aiService *ai.Service) *Service {
	return &Service{
		aiService: aiService,
	}
}

// TweetGenerationRequest 推文生成请求
type TweetGenerationRequest struct {
	Topic     string `json:"topic"`
	Style     string `json:"style"`     // 正式/幽默/营销/情感等
	Length    int    `json:"length"`    // 字符长度限制
	Platform  string `json:"platform"`  // 平台：twitter/weibo/douyin等
	Hashtags  []string `json:"hashtags"` // 指定话题标签
	Audience  string `json:"audience"`  // 目标受众
}

// NovelToTweetRequest 小说转推文请求
type NovelToTweetRequest struct {
	Novel     string   `json:"novel"`
	Count     int      `json:"count"`     // 生成数量
	Style     string   `json:"style"`     // 推文风格
	Platform  string   `json:"platform"`  // 目标平台
	Angles    []string `json:"angles"`    // 推广角度
}

// TweetResult 推文生成结果
type TweetResult struct {
	Content   string   `json:"content"`
	Hashtags  []string `json:"hashtags"`
	Length    int      `json:"length"`
	Platform  string   `json:"platform"`
	Theme     string   `json:"theme"`
	Quality   float64  `json:"quality"`   // 质量评分
}

// NovelTweetResult 小说推文生成结果
type NovelTweetResult struct {
	Tweets    []TweetResult `json:"tweets"`
	Summary   string        `json:"summary"`   // 小说摘要
	Keywords  []string      `json:"keywords"`  // 关键词
	Themes    []string      `json:"themes"`    // 主题
}

// GenerateTweet 生成单条推文
func (s *Service) GenerateTweet(ctx context.Context, req *TweetGenerationRequest) (*TweetResult, error) {
	log.Printf("[TweetService] 开始生成推文: topic=%s, style=%s", req.Topic, req.Style)

	// 构建推文生成提示词
	prompt := s.buildTweetPrompt(req)

	// 调用AI生成
	response, err := s.aiService.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI生成推文失败: %w", err)
	}

	// 解析和优化推文
	tweet := s.processTweetResponse(response, req)

	// 质量评估
	quality := s.evaluateTweetQuality(tweet, req)

	result := &TweetResult{
		Content:  tweet,
		Hashtags: s.extractHashtags(tweet),
		Length:   len(tweet),
		Platform: req.Platform,
		Theme:    req.Topic,
		Quality:  quality,
	}

	log.Printf("[TweetService] 推文生成完成: length=%d, quality=%.2f", result.Length, result.Quality)
	return result, nil
}

// GenerateNovelTweets 从小说生成多条推文
func (s *Service) GenerateNovelTweets(ctx context.Context, req *NovelToTweetRequest) (*NovelTweetResult, error) {
	log.Printf("[TweetService] 开始生成小说推文: count=%d, style=%s", req.Count, req.Style)

	// 1. 分析小说内容
	analysis, err := s.analyzeNovelContent(ctx, req.Novel)
	if err != nil {
		return nil, fmt.Errorf("分析小说内容失败: %w", err)
	}

	// 2. 生成多角度推文
	tweets, err := s.generateMultiAngleTweets(ctx, req, analysis)
	if err != nil {
		return nil, fmt.Errorf("生成多角度推文失败: %w", err)
	}

	result := &NovelTweetResult{
		Tweets:   tweets,
		Summary:  analysis.Summary,
		Keywords: analysis.Keywords,
		Themes:   analysis.Themes,
	}

	log.Printf("[TweetService] 小说推文生成完成: count=%d", len(result.Tweets))
	return result, nil
}

// buildTweetPrompt 构建推文生成提示词
func (s *Service) buildTweetPrompt(req *TweetGenerationRequest) string {
	platformLimits := map[string]int{
		"twitter": 280,
		"weibo":   140,
		"douyin":  55,
	}

	limit := req.Length
	if limit == 0 {
		if platformLimit, exists := platformLimits[req.Platform]; exists {
			limit = platformLimit
		} else {
			limit = 280 // 默认
		}
	}

	prompt := fmt.Sprintf(`请生成一条关于"%s"的优质社交媒体推文。

要求：
1. 风格：%s
2. 平台：%s
3. 字符限制：%d字符以内
4. 内容要有吸引力，能引发互动
5. 语言自然流畅，符合社交媒体特点
6. 包含2-3个相关话题标签`, req.Topic, req.Style, req.Platform, limit)

	if len(req.Hashtags) > 0 {
		prompt += fmt.Sprintf("\n7. 必须包含这些话题标签：%s", strings.Join(req.Hashtags, ", "))
	}

	if req.Audience != "" {
		prompt += fmt.Sprintf("\n8. 目标受众：%s", req.Audience)
	}

	prompt += "\n\n请直接输出推文内容，不需要额外说明："

	return prompt
}

// processTweetResponse 处理AI响应
func (s *Service) processTweetResponse(response string, req *TweetGenerationRequest) string {
	// 清理响应
	tweet := strings.TrimSpace(response)
	
	// 移除可能的引号
	tweet = strings.Trim(tweet, "\"'")
	
	// 确保长度限制
	if req.Length > 0 && len(tweet) > req.Length {
		tweet = s.truncateTweet(tweet, req.Length)
	}

	return tweet
}

// truncateTweet 智能截断推文
func (s *Service) truncateTweet(tweet string, maxLength int) string {
	if len(tweet) <= maxLength {
		return tweet
	}

	// 尝试在句号、感叹号或问号处截断
	punctuation := []string{"。", "！", "？", ".", "!", "?"}
	for i := maxLength - 1; i > maxLength/2; i-- {
		for _, p := range punctuation {
			if i < len(tweet) && string(tweet[i]) == p {
				return tweet[:i+1]
			}
		}
	}

	// 如果找不到合适的截断点，在最后一个空格处截断
	for i := maxLength - 1; i > maxLength/2; i-- {
		if i < len(tweet) && tweet[i] == ' ' {
			return tweet[:i] + "..."
		}
	}

	// 最后的选择：硬截断
	return tweet[:maxLength-3] + "..."
}

// extractHashtags 提取话题标签
func (s *Service) extractHashtags(tweet string) []string {
	re := regexp.MustCompile(`#[\w\u4e00-\u9fa5]+`)
	matches := re.FindAllString(tweet, -1)
	
	var hashtags []string
	for _, match := range matches {
		hashtags = append(hashtags, strings.TrimPrefix(match, "#"))
	}
	
	return hashtags
}

// evaluateTweetQuality 评估推文质量
func (s *Service) evaluateTweetQuality(tweet string, req *TweetGenerationRequest) float64 {
	score := 0.0
	
	// 长度评分（适中长度得分更高）
	length := len(tweet)
	if length >= 50 && length <= 200 {
		score += 0.3
	} else if length >= 20 && length <= 280 {
		score += 0.2
	}
	
	// 话题标签评分
	hashtags := s.extractHashtags(tweet)
	if len(hashtags) >= 1 && len(hashtags) <= 3 {
		score += 0.2
	}
	
	// 互动元素评分（问号、感叹号等）
	if strings.Contains(tweet, "？") || strings.Contains(tweet, "?") {
		score += 0.1
	}
	if strings.Contains(tweet, "！") || strings.Contains(tweet, "!") {
		score += 0.1
	}
	
	// 表情符号评分
	emojiPattern := regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]`)
	if emojiPattern.MatchString(tweet) {
		score += 0.1
	}
	
	// 关键词匹配评分
	if strings.Contains(strings.ToLower(tweet), strings.ToLower(req.Topic)) {
		score += 0.2
	}
	
	return score
}

// NovelAnalysis 小说分析结果
type NovelAnalysis struct {
	Summary     string   `json:"summary"`
	Keywords    []string `json:"keywords"`
	Themes      []string `json:"themes"`
	Characters  []string `json:"characters"`
	Highlights  []string `json:"highlights"`
	Emotions    []string `json:"emotions"`
}

// analyzeNovelContent 分析小说内容
func (s *Service) analyzeNovelContent(ctx context.Context, novel string) (*NovelAnalysis, error) {
	prompt := fmt.Sprintf(`请分析以下小说内容，提取关键信息用于推文创作：

小说内容：
%s

请输出JSON格式：
{
  "summary": "简短摘要（50字以内）",
  "keywords": ["关键词1", "关键词2", "关键词3"],
  "themes": ["主题1", "主题2"],
  "characters": ["主要角色1", "主要角色2"],
  "highlights": ["亮点1", "亮点2", "亮点3"],
  "emotions": ["情感标签1", "情感标签2"]
}`, novel)

	response, err := s.aiService.GenerateText(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	var analysis NovelAnalysis
	cleanedResponse := s.cleanJSONResponse(response)
	if err := json.Unmarshal([]byte(cleanedResponse), &analysis); err != nil {
		// 如果解析失败，返回默认分析
		return &NovelAnalysis{
			Summary:    "精彩的故事内容",
			Keywords:   []string{"小说", "故事", "精彩"},
			Themes:     []string{"人生", "成长"},
			Characters: []string{"主角"},
			Highlights: []string{"情节精彩", "人物生动"},
			Emotions:   []string{"感动", "震撼"},
		}, nil
	}

	return &analysis, nil
}

// generateMultiAngleTweets 生成多角度推文
func (s *Service) generateMultiAngleTweets(ctx context.Context, req *NovelToTweetRequest, analysis *NovelAnalysis) ([]TweetResult, error) {
	angles := []string{"情节亮点", "角色魅力", "主题深度", "情感共鸣", "阅读体验"}
	if len(req.Angles) > 0 {
		angles = req.Angles
	}

	var tweets []TweetResult
	count := req.Count
	if count > len(angles) {
		count = len(angles)
	}

	for i := 0; i < count; i++ {
		angle := angles[i%len(angles)]
		
		tweetReq := &TweetGenerationRequest{
			Topic:    fmt.Sprintf("%s - %s", analysis.Summary, angle),
			Style:    req.Style,
			Platform: req.Platform,
			Length:   280,
		}

		tweet, err := s.GenerateTweet(ctx, tweetReq)
		if err != nil {
			log.Printf("[TweetService] 生成推文失败: angle=%s, error=%v", angle, err)
			continue
		}

		tweet.Theme = angle
		tweets = append(tweets, *tweet)
	}

	return tweets, nil
}

// cleanJSONResponse 清理JSON响应
func (s *Service) cleanJSONResponse(response string) string {
	response = strings.ReplaceAll(response, "```json", "")
	response = strings.ReplaceAll(response, "```", "")
	response = strings.TrimSpace(response)

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start != -1 && end != -1 && end > start {
		return response[start : end+1]
	}

	return response
}
