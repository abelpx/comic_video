package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"comic_video/internal/domain/dto"
	"comic_video/internal/domain/vo"
	"comic_video/internal/service/render"
)

// RenderHandler 渲染处理器
type RenderHandler struct {
	renderService render.Service
}

// NewRenderHandler 创建渲染处理器实例
func NewRenderHandler(renderService render.Service) *RenderHandler {
	return &RenderHandler{
		renderService: renderService,
	}
}

// CreateRender 创建渲染任务
// @Summary 创建渲染任务
// @Description 为指定项目创建新的渲染任务
// @Tags 渲染
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param request body dto.CreateRenderRequest true "创建渲染任务请求"
// @Success 200 {object} vo.SuccessResponse{data=dto.RenderResponse}
// @Failure 400 {object} vo.ErrorResponse
// @Failure 401 {object} vo.ErrorResponse
// @Failure 500 {object} vo.ErrorResponse
// @Router /api/renders [post]
func (h *RenderHandler) CreateRender(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{
			Code:      401,
			Message:   "用户未认证",
			Timestamp: time.Now(),
		})
		return
	}

	var req dto.CreateRenderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:      400,
			Message:   "请求参数错误: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	// 设置默认值
	if req.Format == "" {
		req.Format = "mp4"
	}
	if req.Quality == "" {
		req.Quality = "medium"
	}

	render, err := h.renderService.CreateRender(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:      500,
			Message:   "创建渲染任务失败: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:      200,
		Message:   "渲染任务创建成功",
		Data:      render,
		Timestamp: time.Now(),
	})
}

// GetRender 获取渲染任务详情
// @Summary 获取渲染任务详情
// @Description 获取指定渲染任务的详细信息
// @Tags 渲染
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "渲染任务ID"
// @Success 200 {object} vo.SuccessResponse{data=dto.RenderResponse}
// @Failure 400 {object} vo.ErrorResponse
// @Failure 401 {object} vo.ErrorResponse
// @Failure 404 {object} vo.ErrorResponse
// @Failure 500 {object} vo.ErrorResponse
// @Router /api/renders/{id} [get]
func (h *RenderHandler) GetRender(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{
			Code:      401,
			Message:   "用户未认证",
			Timestamp: time.Now(),
		})
		return
	}

	renderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:      400,
			Message:   "无效的渲染任务ID",
			Timestamp: time.Now(),
		})
		return
	}

	render, err := h.renderService.GetRender(c.Request.Context(), userID.(uuid.UUID), renderID)
	if err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{
			Code:      404,
			Message:   "渲染任务不存在: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:      200,
		Message:   "获取渲染任务成功",
		Data:      render,
		Timestamp: time.Now(),
	})
}

// ListRenders 获取渲染任务列表
// @Summary 获取渲染任务列表
// @Description 获取用户的渲染任务列表，支持分页和筛选
// @Tags 渲染
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页数量，默认为20"
// @Param project_id query string false "项目ID筛选"
// @Param status query string false "状态筛选"
// @Success 200 {object} vo.SuccessResponse{data=dto.ListRendersResponse}
// @Failure 400 {object} vo.ErrorResponse
// @Failure 401 {object} vo.ErrorResponse
// @Failure 500 {object} vo.ErrorResponse
// @Router /api/renders [get]
func (h *RenderHandler) ListRenders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{
			Code:      401,
			Message:   "用户未认证",
			Timestamp: time.Now(),
		})
		return
	}

	var req dto.ListRendersRequest

	// 解析分页参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		} else {
			req.Page = 1
		}
	} else {
		req.Page = 1
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = pageSize
		} else {
			req.PageSize = 20
		}
	} else {
		req.PageSize = 20
	}

	// 解析筛选参数
	req.ProjectID = c.Query("project_id")
	req.Status = c.Query("status")

	renders, err := h.renderService.ListRenders(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:      500,
			Message:   "获取渲染任务列表失败: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:      200,
		Message:   "获取渲染任务列表成功",
		Data:      renders,
		Timestamp: time.Now(),
	})
}

// DeleteRender 删除渲染任务
// @Summary 删除渲染任务
// @Description 删除指定的渲染任务
// @Tags 渲染
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "渲染任务ID"
// @Success 200 {object} vo.SuccessResponse
// @Failure 400 {object} vo.ErrorResponse
// @Failure 401 {object} vo.ErrorResponse
// @Failure 404 {object} vo.ErrorResponse
// @Failure 500 {object} vo.ErrorResponse
// @Router /api/renders/{id} [delete]
func (h *RenderHandler) DeleteRender(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{
			Code:      401,
			Message:   "用户未认证",
			Timestamp: time.Now(),
		})
		return
	}

	renderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:      400,
			Message:   "无效的渲染任务ID",
			Timestamp: time.Now(),
		})
		return
	}

	err = h.renderService.DeleteRender(c.Request.Context(), userID.(uuid.UUID), renderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:      500,
			Message:   "删除渲染任务失败: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:      200,
		Message:   "渲染任务删除成功",
		Data:      nil,
		Timestamp: time.Now(),
	})
}

// GetRenderStatus 获取渲染状态
// @Summary 获取渲染状态
// @Description 获取指定渲染任务的当前状态和进度
// @Tags 渲染
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "渲染任务ID"
// @Success 200 {object} vo.SuccessResponse{data=dto.RenderStatusResponse}
// @Failure 400 {object} vo.ErrorResponse
// @Failure 401 {object} vo.ErrorResponse
// @Failure 404 {object} vo.ErrorResponse
// @Failure 500 {object} vo.ErrorResponse
// @Router /api/renders/{id}/status [get]
func (h *RenderHandler) GetRenderStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{
			Code:      401,
			Message:   "用户未认证",
			Timestamp: time.Now(),
		})
		return
	}

	renderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:      400,
			Message:   "无效的渲染任务ID",
			Timestamp: time.Now(),
		})
		return
	}

	status, err := h.renderService.GetRenderStatus(c.Request.Context(), userID.(uuid.UUID), renderID)
	if err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{
			Code:      404,
			Message:   "获取渲染状态失败: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:      200,
		Message:   "获取渲染状态成功",
		Data:      status,
		Timestamp: time.Now(),
	})
}

// DownloadRender 下载渲染结果
// @Summary 下载渲染结果
// @Description 获取渲染结果的下载链接
// @Tags 渲染
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "渲染任务ID"
// @Success 200 {object} vo.SuccessResponse{data=dto.DownloadRenderResponse}
// @Failure 400 {object} vo.ErrorResponse
// @Failure 401 {object} vo.ErrorResponse
// @Failure 404 {object} vo.ErrorResponse
// @Failure 500 {object} vo.ErrorResponse
// @Router /api/renders/{id}/download [get]
func (h *RenderHandler) DownloadRender(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{
			Code:      401,
			Message:   "用户未认证",
			Timestamp: time.Now(),
		})
		return
	}

	renderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:      400,
			Message:   "无效的渲染任务ID",
			Timestamp: time.Now(),
		})
		return
	}

	download, err := h.renderService.DownloadRender(c.Request.Context(), userID.(uuid.UUID), renderID)
	if err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{
			Code:      404,
			Message:   "获取下载链接失败: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:      200,
		Message:   "获取下载链接成功",
		Data:      download,
		Timestamp: time.Now(),
	})
} 