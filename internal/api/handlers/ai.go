package handlers

import (
	"encoding/json"
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

// NovelToAll 一键生成漫画、推文、动漫视频
func (h *AIHandler) NovelToAll(c *gin.Context) {
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
		Type:      entity.TaskTypeVideo, // 复用video类型
		Status:    entity.TaskStatusPending,
		Progress:  0,
		Params:    string(params),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = h.redisClient.SetTaskStatus(c.Request.Context(), task, 24*time.Hour)
	_ = h.queue.Enqueue(task)
	log.Printf("[Handler] NovelToAll: 入队任务 id=%v type=%v queue=%p", task.ID, task.Type, h.queue)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "任务已提交", "task_id": task.ID})
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
