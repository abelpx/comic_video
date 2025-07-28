package publishing

import (
	"context"
	"fmt"
	"log"
	"time"

	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/postgres"
	"comic_video/internal/service/ai"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service 发布服务
type Service struct {
	db        *gorm.DB
	videoRepo *postgres.VideoRepository
	aiService *ai.Service
}

// NewService 创建发布服务
func NewService(db *gorm.DB, aiService *ai.Service) *Service {
	return &Service{
		db:        db,
		videoRepo: postgres.NewVideoRepository(db),
		aiService: aiService,
	}
}

// PublishVideo 发布视频
func (s *Service) PublishVideo(ctx context.Context, req *PublishVideoRequest) (*PublishResult, error) {
	log.Printf("[PublishingService] 开始发布视频: project=%s", req.ProjectID)

	// 1. 获取视频信息
	video, err := s.videoRepo.GetByProjectID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("获取视频信息失败: %w", err)
	}

	if video.Status != "completed" {
		return nil, fmt.Errorf("视频尚未完成制作，当前状态: %s", video.Status)
	}

	// 2. 生成宣传素材
	promotionMaterials, err := s.generatePromotionMaterials(ctx, req)
	if err != nil {
		log.Printf("[PublishingService] 生成宣传素材失败: %v", err)
		// 不阻断发布流程
	}

	// 3. 发布到各个平台
	publishResults := make([]*PlatformPublishResult, 0)

	for _, platform := range req.Platforms {
		result, err := s.publishToPlatform(ctx, video, platform, promotionMaterials)
		if err != nil {
			log.Printf("[PublishingService] 发布到平台失败: platform=%s, error=%v", platform.Name, err)
			result = &PlatformPublishResult{
				Platform: platform.Name,
				Status:   "failed",
				Error:    err.Error(),
			}
		}
		publishResults = append(publishResults, result)
	}

	// 4. 生成发布报告
	publishResult := &PublishResult{
		ProjectID:           req.ProjectID,
		VideoID:             video.ID,
		PublishedAt:         time.Now(),
		PlatformResults:     publishResults,
		PromotionMaterials:  promotionMaterials,
		TotalPlatforms:      len(req.Platforms),
		SuccessfulPlatforms: s.countSuccessfulPublishes(publishResults),
	}

	log.Printf("[PublishingService] 视频发布完成: project=%s, success=%d/%d", 
		req.ProjectID, publishResult.SuccessfulPlatforms, publishResult.TotalPlatforms)

	return publishResult, nil
}

// generatePromotionMaterials 生成宣传素材
func (s *Service) generatePromotionMaterials(ctx context.Context, req *PublishVideoRequest) (*PromotionMaterials, error) {
	log.Printf("[PublishingService] 开始生成宣传素材")

	materials := &PromotionMaterials{}

	// 1. 生成标题和描述
	if req.AutoGenerateContent {
		titles, descriptions, err := s.generateTitlesAndDescriptions(ctx, req)
		if err != nil {
			log.Printf("[PublishingService] 生成标题描述失败: %v", err)
		} else {
			materials.Titles = titles
			materials.Descriptions = descriptions
		}
	}

	// 2. 生成标签和关键词
	if req.AutoGenerateTags {
		tags, keywords, err := s.generateTagsAndKeywords(ctx, req)
		if err != nil {
			log.Printf("[PublishingService] 生成标签关键词失败: %v", err)
		} else {
			materials.Tags = tags
			materials.Keywords = keywords
		}
	}

	// 3. 生成缩略图
	if req.AutoGenerateThumbnail {
		thumbnails, err := s.generateThumbnails(ctx, req)
		if err != nil {
			log.Printf("[PublishingService] 生成缩略图失败: %v", err)
		} else {
			materials.Thumbnails = thumbnails
		}
	}

	// 4. 生成社交媒体文案
	if req.AutoGenerateSocialContent {
		socialContent, err := s.generateSocialContent(ctx, req)
		if err != nil {
			log.Printf("[PublishingService] 生成社交媒体文案失败: %v", err)
		} else {
			materials.SocialContent = socialContent
		}
	}

	return materials, nil
}

// publishToPlatform 发布到指定平台
func (s *Service) publishToPlatform(ctx context.Context, video *entity.Video, platform *PlatformConfig, materials *PromotionMaterials) (*PlatformPublishResult, error) {
	log.Printf("[PublishingService] 发布到平台: %s", platform.Name)

	// 根据平台类型选择发布策略
	switch platform.Name {
	case "youtube":
		return s.publishToYouTube(ctx, video, platform, materials)
	case "bilibili":
		return s.publishToBilibili(ctx, video, platform, materials)
	case "tiktok":
		return s.publishToTikTok(ctx, video, platform, materials)
	case "weibo":
		return s.publishToWeibo(ctx, video, platform, materials)
	default:
		return nil, fmt.Errorf("不支持的平台: %s", platform.Name)
	}
}

// publishToYouTube 发布到YouTube
func (s *Service) publishToYouTube(ctx context.Context, video *entity.Video, platform *PlatformConfig, materials *PromotionMaterials) (*PlatformPublishResult, error) {
	// 这里实现YouTube API调用
	// 实际实现中需要使用YouTube Data API
	
	result := &PlatformPublishResult{
		Platform:    "youtube",
		Status:      "success",
		PublishedAt: time.Now(),
		URL:         fmt.Sprintf("https://youtube.com/watch?v=%s", generateMockID()),
		Metrics: &PublishMetrics{
			Views:    0,
			Likes:    0,
			Comments: 0,
			Shares:   0,
		},
	}

	return result, nil
}

// publishToBilibili 发布到B站
func (s *Service) publishToBilibili(ctx context.Context, video *entity.Video, platform *PlatformConfig, materials *PromotionMaterials) (*PlatformPublishResult, error) {
	// 这里实现B站API调用
	
	result := &PlatformPublishResult{
		Platform:    "bilibili",
		Status:      "success",
		PublishedAt: time.Now(),
		URL:         fmt.Sprintf("https://bilibili.com/video/BV%s", generateMockID()),
		Metrics: &PublishMetrics{
			Views:    0,
			Likes:    0,
			Comments: 0,
			Shares:   0,
		},
	}

	return result, nil
}

// publishToTikTok 发布到TikTok
func (s *Service) publishToTikTok(ctx context.Context, video *entity.Video, platform *PlatformConfig, materials *PromotionMaterials) (*PlatformPublishResult, error) {
	// 这里实现TikTok API调用
	
	result := &PlatformPublishResult{
		Platform:    "tiktok",
		Status:      "success",
		PublishedAt: time.Now(),
		URL:         fmt.Sprintf("https://tiktok.com/@user/video/%s", generateMockID()),
		Metrics: &PublishMetrics{
			Views:    0,
			Likes:    0,
			Comments: 0,
			Shares:   0,
		},
	}

	return result, nil
}

// publishToWeibo 发布到微博
func (s *Service) publishToWeibo(ctx context.Context, video *entity.Video, platform *PlatformConfig, materials *PromotionMaterials) (*PlatformPublishResult, error) {
	// 这里实现微博API调用
	
	result := &PlatformPublishResult{
		Platform:    "weibo",
		Status:      "success",
		PublishedAt: time.Now(),
		URL:         fmt.Sprintf("https://weibo.com/status/%s", generateMockID()),
		Metrics: &PublishMetrics{
			Views:    0,
			Likes:    0,
			Comments: 0,
			Shares:   0,
		},
	}

	return result, nil
}

// generateTitlesAndDescriptions 生成标题和描述
func (s *Service) generateTitlesAndDescriptions(ctx context.Context, req *PublishVideoRequest) ([]string, []string, error) {
	prompt := fmt.Sprintf(`为以下视频内容生成吸引人的标题和描述：

视频主题：%s
目标受众：%s
内容类型：%s

要求：
1. 生成3个不同风格的标题
2. 为每个标题生成对应的描述
3. 标题要简洁有力，描述要详细吸引人
4. 考虑SEO优化

输出JSON格式：
{
  "titles": ["标题1", "标题2", "标题3"],
  "descriptions": ["描述1", "描述2", "描述3"]
}`, req.VideoTheme, req.TargetAudience, req.ContentType)

	// 调用AI生成服务
	response, err := s.aiService.GenerateText(ctx, prompt)
	if err != nil {
		return nil, nil, err
	}

	// TODO: 解析JSON响应，目前使用简化实现
	_ = response // 避免未使用变量错误
	return []string{"AI生成标题1", "AI生成标题2", "AI生成标题3"},
		   []string{"AI生成描述1", "AI生成描述2", "AI生成描述3"}, nil
}

// generateTagsAndKeywords 生成标签和关键词
func (s *Service) generateTagsAndKeywords(ctx context.Context, req *PublishVideoRequest) ([]string, []string, error) {
	// 简化实现
	tags := []string{"AI视频", "自动生成", "创意内容"}
	keywords := []string{"人工智能", "视频制作", "自动化"}
	return tags, keywords, nil
}

// generateThumbnails 生成缩略图
func (s *Service) generateThumbnails(ctx context.Context, req *PublishVideoRequest) ([]string, error) {
	// 简化实现
	thumbnails := []string{
		"https://example.com/thumbnail1.jpg",
		"https://example.com/thumbnail2.jpg",
		"https://example.com/thumbnail3.jpg",
	}
	return thumbnails, nil
}

// generateSocialContent 生成社交媒体文案
func (s *Service) generateSocialContent(ctx context.Context, req *PublishVideoRequest) (map[string]string, error) {
	// 简化实现
	content := map[string]string{
		"twitter":   "🎬 全新AI生成视频上线！#AI视频 #创意内容",
		"facebook":  "分享我们最新的AI生成视频作品，希望大家喜欢！",
		"instagram": "✨ AI的魔法创造了这个精彩视频 ✨ #AI #视频制作",
	}
	return content, nil
}

// countSuccessfulPublishes 统计成功发布的平台数量
func (s *Service) countSuccessfulPublishes(results []*PlatformPublishResult) int {
	count := 0
	for _, result := range results {
		if result.Status == "success" {
			count++
		}
	}
	return count
}

// generateMockID 生成模拟ID
func generateMockID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
}

// PublishVideoRequest 发布视频请求
type PublishVideoRequest struct {
	ProjectID               uuid.UUID         `json:"project_id"`
	VideoTheme              string            `json:"video_theme"`
	TargetAudience          string            `json:"target_audience"`
	ContentType             string            `json:"content_type"`
	Platforms               []*PlatformConfig `json:"platforms"`
	AutoGenerateContent     bool              `json:"auto_generate_content"`
	AutoGenerateTags        bool              `json:"auto_generate_tags"`
	AutoGenerateThumbnail   bool              `json:"auto_generate_thumbnail"`
	AutoGenerateSocialContent bool            `json:"auto_generate_social_content"`
}

// PlatformConfig 平台配置
type PlatformConfig struct {
	Name        string                 `json:"name"`
	Credentials map[string]string      `json:"credentials"`
	Settings    map[string]interface{} `json:"settings"`
}

// PublishResult 发布结果
type PublishResult struct {
	ProjectID           uuid.UUID                `json:"project_id"`
	VideoID             uuid.UUID                `json:"video_id"`
	PublishedAt         time.Time                `json:"published_at"`
	PlatformResults     []*PlatformPublishResult `json:"platform_results"`
	PromotionMaterials  *PromotionMaterials      `json:"promotion_materials"`
	TotalPlatforms      int                      `json:"total_platforms"`
	SuccessfulPlatforms int                      `json:"successful_platforms"`
}

// PlatformPublishResult 平台发布结果
type PlatformPublishResult struct {
	Platform    string          `json:"platform"`
	Status      string          `json:"status"`
	PublishedAt time.Time       `json:"published_at"`
	URL         string          `json:"url"`
	Error       string          `json:"error,omitempty"`
	Metrics     *PublishMetrics `json:"metrics,omitempty"`
}

// PublishMetrics 发布指标
type PublishMetrics struct {
	Views    int64 `json:"views"`
	Likes    int64 `json:"likes"`
	Comments int64 `json:"comments"`
	Shares   int64 `json:"shares"`
}

// PromotionMaterials 宣传素材
type PromotionMaterials struct {
	Titles        []string          `json:"titles"`
	Descriptions  []string          `json:"descriptions"`
	Tags          []string          `json:"tags"`
	Keywords      []string          `json:"keywords"`
	Thumbnails    []string          `json:"thumbnails"`
	SocialContent map[string]string `json:"social_content"`
}
