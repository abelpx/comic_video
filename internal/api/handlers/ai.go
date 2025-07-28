package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/redis"
	"comic_video/internal/service/ai"
	"comic_video/internal/service/quota"
)

type AIHandler struct {
	redisClient  *redis.Client
	queue        ai.TaskQueue
	quotaManager *quota.QuotaManager
}

func NewAIHandler(redisClient *redis.Client, queue ai.TaskQueue) *AIHandler {
	return &AIHandler{
		redisClient:  redisClient,
		queue:        queue,
		quotaManager: quota.NewQuotaManager(redisClient),
	}
}

// NovelToVideo 提交一键生成动漫视频任务
func (h *AIHandler) NovelToVideo(c *gin.Context) {
	var req struct {
		Novel string `json:"novel"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Novel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 获取用户ID（这里简化处理，实际应该从JWT token中获取）
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "anonymous" // 匿名用户
	}

	// 检查视频生成配额
	if err := h.quotaManager.CheckQuota(c.Request.Context(), userID, quota.QuotaTypeVideo, 1); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "配额不足: " + err.Error(),
			"type":    "quota_exceeded",
		})
		return
	}

	params, _ := json.Marshal(req)
	task := &entity.Task{
		ID:        uuid.New(),
		Type:      entity.TaskTypeVideo,
		Status:    entity.TaskStatusPending,
		Progress:  0,
		Params:    string(params),
		UserID:    userID, // 添加用户ID
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 消费配额
	if err := h.quotaManager.ConsumeQuota(c.Request.Context(), userID, quota.QuotaTypeVideo, 1); err != nil {
		log.Printf("[Quota] 配额消费失败: %v", err)
		// 不阻止任务提交，但记录日志
	}

	_ = h.redisClient.SetTaskStatus(c.Request.Context(), task, 24*time.Hour)
	_ = h.queue.Enqueue(task)

	log.Printf("[Handler] NovelToVideo: 入队任务 id=%s type=%s queue=%p userID=%s", task.ID, task.Type, h.queue, userID)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "任务已提交", "task_id": task.ID})
}

// GenerateNovel 提交AI生成小说任务
func (h *AIHandler) GenerateNovel(c *gin.Context) {
	var req struct {
		NovelPrompt string `json:"novel_prompt"`
		Title       string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.NovelPrompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	params, _ := json.Marshal(map[string]interface{}{
		"novel": req.NovelPrompt,
		"title": req.Title,
	})
	task := &entity.Task{
		ID:        uuid.New(),
		Type:      entity.TaskTypeAI,
		Status:    entity.TaskStatusPending,
		Progress:  0,
		Params:    string(params),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = h.redisClient.SetTaskStatus(c.Request.Context(), task, 24*time.Hour)
	_ = h.queue.Enqueue(task)
	log.Printf("[Handler] GenerateNovel: 入队任务 id=%v type=%v queue=%p", task.ID, task.Type, h.queue)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "任务已提交", "task_id": task.ID})
}

// NovelToAll 一键生成漫画、推文、动漫视频（增强版）
func (h *AIHandler) NovelToAll(c *gin.Context) {
	var req struct {
		NovelPrompt   string `json:"novel_prompt"`
		Title         string `json:"title"`
		Style         string `json:"style"`         // 新增：风格
		VideoFormat   string `json:"video_format"`  // 新增：视频格式
		Quality       string `json:"quality"`       // 新增：质量等级
		GenerateTweet bool   `json:"generate_tweet"` // 新增：是否生成推文
		TargetLength  int    `json:"target_length"`  // 新增：目标时长
		Platform      string `json:"platform"`      // 新增：目标平台
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.NovelPrompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 设置默认值
	if req.Style == "" {
		req.Style = "anime"
	}
	if req.VideoFormat == "" {
		req.VideoFormat = "1080p"
	}
	if req.Quality == "" {
		req.Quality = "high"
	}
	if req.TargetLength == 0 {
		req.TargetLength = 180 // 默认3分钟
	}
	if req.Platform == "" {
		req.Platform = "general"
	}

	params, _ := json.Marshal(map[string]interface{}{
		"novel":         req.NovelPrompt,
		"title":         req.Title,
		"style":         req.Style,
		"video_format":  req.VideoFormat,
		"quality":       req.Quality,
		"generate_tweet": req.GenerateTweet,
		"target_length": req.TargetLength,
		"platform":      req.Platform,
		"enhanced":      true, // 标记为增强版
	})

	task := &entity.Task{
		ID:        uuid.New(),
		Type:      entity.TaskTypeVideo, // 复用video类型
		Status:    entity.TaskStatusPending,
		Progress:  0,
		Params:    string(params),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = h.redisClient.SetTaskStatus(c.Request.Context(), task, 24*time.Hour)
	_ = h.queue.Enqueue(task)
	log.Printf("[Handler] NovelToAll Enhanced: 入队任务 id=%v type=%v queue=%p", task.ID, task.Type, h.queue)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "增强任务已提交",
		"data": gin.H{
			"task_id":       task.ID,
			"video_format":  req.VideoFormat,
			"quality":       req.Quality,
			"target_length": req.TargetLength,
			"platform":      req.Platform,
			"features": []string{
				"高级视频生成",
				"角色动画系统",
				"智能音频处理",
				"平台自动适配",
				"质量自动优化",
			},
		},
	})
}

// GetUserQuota 获取用户配额信息
func (h *AIHandler) GetUserQuota(c *gin.Context) {
	// 获取用户ID
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	// 获取配额信息
	quotas, err := h.quotaManager.GetAllUserQuotas(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取配额信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    quotas,
	})
}

// GetUsageStats 获取用户使用统计
func (h *AIHandler) GetUsageStats(c *gin.Context) {
	// 获取用户ID
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	// 获取使用统计
	stats, err := h.quotaManager.GetUsageStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取使用统计失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    stats,
	})
}

// GenerateTweet 生成推文
func (h *AIHandler) GenerateTweet(c *gin.Context) {
	var req struct {
		Topic string `json:"topic"`
		Style string `json:"style"` // 可选：正式/幽默/营销等
		Length int   `json:"length"` // 可选：字符长度限制
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Topic == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误，主题不能为空"})
		return
	}

	// 设置默认值
	if req.Style == "" {
		req.Style = "正式"
	}
	if req.Length == 0 {
		req.Length = 280 // Twitter默认长度
	}

	// 构建推文生成提示词
	_ = fmt.Sprintf(`请根据以下主题生成一条优质的社交媒体推文：

主题：%s
风格：%s
长度限制：%d字符以内

要求：
1. 内容要吸引人，有趣且有价值
2. 适合社交媒体传播
3. 包含适当的话题标签
4. 语言自然流畅
5. 符合指定的风格和长度要求

请直接输出推文内容，不需要额外说明：`, req.Topic, req.Style, req.Length)

	// 调用AI生成服务（这里需要实现AI服务调用）
	// 暂时返回模拟结果
	tweet := fmt.Sprintf("🌟 关于%s的精彩分享！这个话题真的很有意思，值得大家深入思考。你们觉得呢？#%s #分享", req.Topic, req.Topic)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "推文生成成功",
		"data": gin.H{
			"tweet": tweet,
			"topic": req.Topic,
			"style": req.Style,
			"length": len(tweet),
		},
	})
}

// NovelToTweet 小说转推文
func (h *AIHandler) NovelToTweet(c *gin.Context) {
	var req struct {
		Novel string `json:"novel"`
		Count int    `json:"count"` // 生成推文数量
		Style string `json:"style"` // 推文风格
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Novel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误，小说内容不能为空"})
		return
	}

	// 设置默认值
	if req.Count == 0 {
		req.Count = 3
	}
	if req.Style == "" {
		req.Style = "吸引人"
	}

	// 构建小说转推文提示词
	_ = fmt.Sprintf(`请根据以下小说内容，生成%d条不同角度的推文来推广这部小说：

小说内容：
%s

要求：
1. 每条推文都要有不同的切入角度（如情节亮点、角色魅力、主题深度等）
2. 风格：%s
3. 每条推文280字符以内
4. 包含适当的话题标签
5. 能够激发读者的阅读兴趣

请输出JSON格式：
{
  "tweets": ["推文1", "推文2", "推文3"],
  "themes": ["主题1", "主题2", "主题3"]
}`, req.Count, req.Novel, req.Style)

	// 暂时返回模拟结果
	tweets := []string{
		"📚 这部小说的情节设计太精彩了！每一个转折都让人意想不到，强烈推荐！#小说推荐 #精彩情节",
		"💫 书中的角色塑造真的很棒，每个人物都有自己的故事和成长轨迹 #角色魅力 #好书分享",
		"🎭 这个故事探讨的主题很深刻，读完让人思考良久，值得细细品味 #深度阅读 #思考人生",
	}

	themes := []string{"情节亮点", "角色魅力", "主题深度"}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "小说推文生成成功",
		"data": gin.H{
			"tweets": tweets,
			"themes": themes,
			"count": len(tweets),
		},
	})
}
