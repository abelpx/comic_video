package handlers

import (
	"net/http"

	"comic_video/internal/domain/dto"
	"comic_video/internal/domain/vo"
	"comic_video/internal/service/video"
	"comic_video/internal/utils"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	service *video.Service
}

func NewVideoHandler(service *video.Service) *VideoHandler {
	return &VideoHandler{service: service}
}

// List 获取视频列表
func (h *VideoHandler) List(c *gin.Context) {
	var req dto.ListVideosRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors: utils.ValidateErrors(err),
		})
		return
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	userID := c.GetString("user_id")
	videos, total, err := h.service.List(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "获取视频列表失败",
			Errors: err.Error(),
		})
		return
	}

	resp := dto.ListVideosResponse{
		Videos:   videos,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取视频列表成功",
		Data:    resp,
	})
}

// Upload 上传视频
func (h *VideoHandler) Upload(c *gin.Context) {
	userID := c.GetString("user_id")
	var req dto.UploadVideoRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors: utils.ValidateErrors(err),
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "未上传文件",
			Errors: err.Error(),
		})
		return
	}

	video, err := h.service.Upload(c.Request.Context(), userID, req, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "上传视频失败",
			Errors: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, vo.SuccessResponse{
		Code:    201,
		Message: "视频上传成功",
		Data:    video,
	})
}

// GetByID 获取视频详情
func (h *VideoHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	video, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{
			Code:    404,
			Message: "视频不存在或无权限访问",
			Errors: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取视频成功",
		Data:    video,
	})
}

// Update 更新视频
func (h *VideoHandler) Update(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")
	var req dto.UpdateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors: utils.ValidateErrors(err),
		})
		return
	}

	video, err := h.service.Update(c.Request.Context(), id, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "更新视频失败",
			Errors: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "视频更新成功",
		Data:    video,
	})
}

// Delete 删除视频
func (h *VideoHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	err := h.service.Delete(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "删除视频失败",
			Errors: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "视频删除成功",
		Data:    nil,
	})
}

// Process 处理视频
func (h *VideoHandler) Process(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	err := h.service.Process(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "处理视频失败",
			Errors: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "视频处理已开始",
		Data:    nil,
	})
}

// GetStatus 获取视频处理状态
func (h *VideoHandler) GetStatus(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	status, err := h.service.GetStatus(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{
			Code:    404,
			Message: "视频不存在或无权限访问",
			Errors: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取状态成功",
		Data:    gin.H{"status": status},
	})
} 