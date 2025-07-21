package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"comic_video/internal/repository/redis"
	"comic_video/internal/repository/postgres"
	"comic_video/internal/domain/vo"
)

type TaskHandler struct {
	redisClient *redis.Client
	taskRepo    *postgres.TaskRepository
}

func NewTaskHandler(redisClient *redis.Client, taskRepo *postgres.TaskRepository) *TaskHandler {
	return &TaskHandler{
		redisClient: redisClient,
		taskRepo:    taskRepo,
	}
}

// GetTaskStatus 查询任务进度与状态
// @Summary 查询任务进度
// @Description 查询指定任务ID的进度、状态、错误等
// @Tags 任务
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} vo.SuccessResponse
// @Failure 404 {object} vo.ErrorResponse
// @Router /api/v1/task/{id}/status [get]
func (h *TaskHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	task, err := h.redisClient.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{
			Code:      404,
			Message:   "任务不存在或已过期",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:      200,
		Message:   "查询成功",
		Data:      task,
		Timestamp: time.Now(),
	})
}

// GetUserTasks 获取用户任务列表
func (h *TaskHandler) GetUserTasks(c *gin.Context) {
	// 获取用户ID（这里简化处理，实际应该从JWT token中获取）
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "anonymous" // 匿名用户
	}

	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// 模拟任务数据（因为目前没有真实的数据库数据）
	tasks := []map[string]interface{}{
		{
			"id":         "task-1",
			"type":       "video",
			"status":     "completed",
			"progress":   100,
			"created_at": time.Now().Add(-2 * time.Hour).Format("2006-01-02T15:04:05Z07:00"),
			"title":      "小说转视频任务",
		},
		{
			"id":         "task-2",
			"type":       "video",
			"status":     "processing",
			"progress":   45,
			"created_at": time.Now().Add(-30 * time.Minute).Format("2006-01-02T15:04:05Z07:00"),
			"title":      "小说转视频任务",
		},
		{
			"id":         "task-3",
			"type":       "comic",
			"status":     "failed",
			"progress":   0,
			"created_at": time.Now().Add(-1 * time.Hour).Format("2006-01-02T15:04:05Z07:00"),
			"title":      "漫画生成任务",
		},
	}

	// 返回结果
	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "success",
		Data: gin.H{
			"tasks": tasks,
			"total": len(tasks),
			"page":  page,
			"limit": limit,
		},
		Timestamp: time.Now(),
	})
}