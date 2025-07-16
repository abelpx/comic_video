package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/redis"
	"comic_video/internal/service/ai"
	"log"
)

type AIHandler struct {
	redisClient *redis.Client
	queue      ai.TaskQueue
}

func NewAIHandler(redisClient *redis.Client, queue ai.TaskQueue) *AIHandler {
	return &AIHandler{redisClient: redisClient, queue: queue}
}

// NovelToVideo 提交一键生成动漫视频任务
func (h *AIHandler) NovelToVideo(c *gin.Context) {
	var req struct{ Novel string `json:"novel"` }
	if err := c.ShouldBindJSON(&req); err != nil || req.Novel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	params, _ := json.Marshal(req)
	task := &entity.Task{
		ID:        uuid.New(),
		Type:      entity.TaskTypeVideo,
		Status:    entity.TaskStatusPending,
		Progress:  0,
		Params:    string(params),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = h.redisClient.SetTaskStatus(c.Request.Context(), task, 24*time.Hour)
	_ = h.queue.Enqueue(task)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "任务已提交", "task_id": task.ID})
}

// GenerateNovel 提交AI生成小说任务
func (h *AIHandler) GenerateNovel(c *gin.Context) {
	var req struct {
		NovelPrompt string `json:"novel_prompt"`
		Title      string `json:"title"`
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
		Title      string `json:"title"`
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