package handlers

import (
	"net/http"
	"strconv"

	"comic_video/internal/domain/entity"
	"comic_video/internal/domain/vo"
	"comic_video/internal/service/workflow"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WorkflowHandler 工作流处理器
type WorkflowHandler struct {
	workflowEngine *workflow.WorkflowEngine
}

// NewWorkflowHandler 创建工作流处理器
func NewWorkflowHandler(workflowEngine *workflow.WorkflowEngine) *WorkflowHandler {
	return &WorkflowHandler{
		workflowEngine: workflowEngine,
	}
}

// CreateWorkflow 创建工作流
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{
			Code:    401,
			Message: "未授权访问",
		})
		return
	}

	var req workflow.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors:  err.Error(),
		})
		return
	}

	// 设置用户ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}
	req.UserID = userUUID

	// 创建工作流
	workflowEntity, err := h.workflowEngine.CreateWorkflow(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "创建工作流失败",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "工作流创建成功",
		Data:    workflowEntity,
	})
}

// StartWorkflow 启动工作流
func (h *WorkflowHandler) StartWorkflow(c *gin.Context) {
	workflowIDStr := c.Param("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "无效的工作流ID",
		})
		return
	}

	// 启动工作流
	if err := h.workflowEngine.StartWorkflow(c.Request.Context(), workflowID); err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "启动工作流失败",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "工作流启动成功",
	})
}

// GetWorkflow 获取工作流详情
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	workflowIDStr := c.Param("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "无效的工作流ID",
		})
		return
	}

	// 获取工作流详情
	workflow, err := h.workflowEngine.GetWorkflow(c.Request.Context(), workflowID)
	if err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{
			Code:    404,
			Message: "工作流不存在",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取工作流成功",
		Data:    workflow,
	})
}

// ListWorkflows 获取工作流列表
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{
			Code:    401,
			Message: "未授权访问",
		})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}

	// 获取用户工作流列表
	workflows, total, err := h.workflowEngine.GetUserWorkflows(c.Request.Context(), userUUID, offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "获取工作流列表失败",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取工作流列表成功",
		Data: map[string]interface{}{
			"workflows": workflows,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// NovelToVideoWorkflow 小说转视频完整工作流
func (h *WorkflowHandler) NovelToVideoWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{
			Code:    401,
			Message: "未授权访问",
		})
		return
	}

	var req struct {
		ProjectID   uuid.UUID `json:"project_id" binding:"required"`
		NovelText   string    `json:"novel_text" binding:"required"`
		Title       string    `json:"title" binding:"required"`
		Description string    `json:"description"`
		Settings    struct {
			VideoTheme      string `json:"video_theme"`
			TargetAudience  string `json:"target_audience"`
			ContentType     string `json:"content_type"`
			AutoPublish     bool   `json:"auto_publish"`
		} `json:"settings"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors:  err.Error(),
		})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}

	// 定义完整的工作流步骤
	steps := []entity.WorkflowStep{
		entity.StepScriptAdapt,
		entity.StepCharacterGen,
		entity.StepSceneGen,
		entity.StepStoryboard,
		entity.StepVoiceGen,
		entity.StepMusicGen,
		entity.StepVideoEdit,
	}

	if req.Settings.AutoPublish {
		steps = append(steps, entity.StepPublish)
	}

	// 创建工作流请求
	workflowReq := &workflow.CreateWorkflowRequest{
		ProjectID:   req.ProjectID,
		UserID:      userUUID,
		Name:        "小说转视频工作流",
		Description: "从小说文本到完整视频的自动化制作流程",
		Steps:       steps,
	}

	// 创建工作流
	workflowEntity, err := h.workflowEngine.CreateWorkflow(c.Request.Context(), workflowReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "创建工作流失败",
			Errors:  err.Error(),
		})
		return
	}

	// 启动工作流
	if err := h.workflowEngine.StartWorkflow(c.Request.Context(), workflowEntity.ID); err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "启动工作流失败",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "小说转视频工作流启动成功",
		Data: map[string]interface{}{
			"workflow_id": workflowEntity.ID,
			"status":      workflowEntity.Status,
			"steps":       len(steps),
		},
	})
}

// GetWorkflowTasks 获取工作流任务列表
func (h *WorkflowHandler) GetWorkflowTasks(c *gin.Context) {
	workflowIDStr := c.Param("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "无效的工作流ID",
		})
		return
	}

	tasks, err := h.workflowEngine.GetWorkflowTasks(c.Request.Context(), workflowID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "获取工作流任务失败",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取工作流任务成功",
		Data:    tasks,
	})
}

// GetWorkflowProgress 获取工作流进度
func (h *WorkflowHandler) GetWorkflowProgress(c *gin.Context) {
	workflowIDStr := c.Param("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "无效的工作流ID",
		})
		return
	}

	workflow, err := h.workflowEngine.GetWorkflow(c.Request.Context(), workflowID)
	if err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{
			Code:    404,
			Message: "工作流不存在",
			Errors:  err.Error(),
		})
		return
	}

	tasks, err := h.workflowEngine.GetWorkflowTasks(c.Request.Context(), workflowID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "获取工作流任务失败",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取工作流进度成功",
		Data: map[string]interface{}{
			"workflow":         workflow,
			"tasks":           tasks,
			"overall_progress": workflow.Progress,
		},
	})
}

// CancelWorkflow 取消工作流
func (h *WorkflowHandler) CancelWorkflow(c *gin.Context) {
	workflowIDStr := c.Param("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "无效的工作流ID",
		})
		return
	}

	if err := h.workflowEngine.CancelWorkflow(c.Request.Context(), workflowID); err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "取消工作流失败",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "工作流已取消",
	})
}
