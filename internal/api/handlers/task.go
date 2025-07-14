package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"comic_video/internal/repository/redis"
	"comic_video/internal/domain/vo"
)

type TaskHandler struct {
	redisClient *redis.Client
}

func NewTaskHandler(redisClient *redis.Client) *TaskHandler {
	return &TaskHandler{redisClient: redisClient}
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