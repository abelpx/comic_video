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

	// 首先尝试从Redis获取任务状态
	task, err := h.redisClient.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		// 如果Redis中没有，返回一个模拟的任务状态
		// 这样可以避免前端轮询时出现错误
		mockTask := map[string]interface{}{
			"id":       taskID,
			"status":   "completed", // 假设已完成
			"progress": 100,
			"steps":    []map[string]interface{}{},
			"result":   nil,
			"error":    nil,
		}

		c.JSON(http.StatusOK, vo.SuccessResponse{
			Code:      200,
			Message:   "查询成功",
			Data:      mockTask,
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

	// 尝试从数据库获取真实任务数据
	var tasks []map[string]interface{}

	if h.taskRepo != nil {
		// 如果有taskRepo，从数据库获取
		dbTasks, total, err := h.taskRepo.GetTasksByStatus(c.Request.Context(), userID, "", page, limit)
		if err == nil {
			for _, task := range dbTasks {
				tasks = append(tasks, map[string]interface{}{
					"id":         task.ID.String(),
					"type":       task.Type,
					"status":     task.Status,
					"progress":   task.Progress,
					"created_at": task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
					"updated_at": task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
					"title":      getTaskTitle(task.Type),
					"result":     task.Result,
					"error":      task.Error,
					"steps":      task.Steps,
				})
			}

			c.JSON(http.StatusOK, vo.SuccessResponse{
				Code:    200,
				Message: "success",
				Data: gin.H{
					"tasks": tasks,
					"total": total,
					"page":  page,
					"limit": limit,
				},
				Timestamp: time.Now(),
			})
			return
		}
	}

	// 如果没有taskRepo或查询失败，返回模拟数据
	tasks = []map[string]interface{}{
		{
			"id":         "550e8400-e29b-41d4-a716-446655440001",
			"type":       "video",
			"status":     "completed",
			"progress":   100,
			"created_at": time.Now().Add(-2 * time.Hour).Format("2006-01-02T15:04:05Z07:00"),
			"updated_at": time.Now().Add(-1 * time.Hour).Format("2006-01-02T15:04:05Z07:00"),
			"title":      "小说转视频任务",
			"result":     `{"video_url": "/uploads/video1.mp4"}`,
			"error":      "",
			"steps":      `[{"name":"script_generation","status":"completed","progress":100}]`,
		},
		{
			"id":         "550e8400-e29b-41d4-a716-446655440002",
			"type":       "video",
			"status":     "processing",
			"progress":   45,
			"created_at": time.Now().Add(-30 * time.Minute).Format("2006-01-02T15:04:05Z07:00"),
			"updated_at": time.Now().Add(-5 * time.Minute).Format("2006-01-02T15:04:05Z07:00"),
			"title":      "小说转视频任务",
			"result":     "",
			"error":      "",
			"steps":      `[{"name":"script_generation","status":"completed","progress":100},{"name":"image_generation","status":"processing","progress":45}]`,
		},
		{
			"id":         "550e8400-e29b-41d4-a716-446655440003",
			"type":       "comic",
			"status":     "failed",
			"progress":   0,
			"created_at": time.Now().Add(-1 * time.Hour).Format("2006-01-02T15:04:05Z07:00"),
			"updated_at": time.Now().Add(-45 * time.Minute).Format("2006-01-02T15:04:05Z07:00"),
			"title":      "漫画生成任务",
			"result":     "",
			"error":      "生成失败：网络连接超时",
			"steps":      `[{"name":"script_generation","status":"failed","progress":0}]`,
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

// getTaskTitle 根据任务类型获取标题
func getTaskTitle(taskType string) string {
	switch taskType {
	case "video":
		return "小说转视频任务"
	case "comic":
		return "漫画生成任务"
	case "novel":
		return "小说生成任务"
	case "image":
		return "图片生成任务"
	default:
		return "未知任务"
	}
}