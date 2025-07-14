package handlers

import (
	"net/http"

	"comic_video/internal/domain/dto"
	"comic_video/internal/domain/vo"
	"comic_video/internal/service/material"
	"comic_video/internal/utils"

	"github.com/gin-gonic/gin"
)

type MaterialHandler struct {
	service *material.Service
}

func NewMaterialHandler(service *material.Service) *MaterialHandler {
	return &MaterialHandler{service: service}
}

// List 获取素材列表
func (h *MaterialHandler) List(c *gin.Context) {
	var req dto.ListMaterialsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors:  utils.ValidateErrors(err),
		})
		return
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}
	materials, total, err := h.service.List(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "获取素材列表失败",
			Errors:  err.Error(),
		})
		return
	}
	resp := dto.ListMaterialsResponse{
		Materials: materials,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}
	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取素材列表成功",
		Data:    resp,
	})
}

// GetByID 获取素材详情
func (h *MaterialHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	material, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{
			Code:    404,
			Message: "素材不存在",
			Errors:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取素材成功",
		Data:    material,
	})
}

// Upload 上传素材
func (h *MaterialHandler) Upload(c *gin.Context) {
	userID := c.GetString("user_id")
	var req dto.UploadMaterialRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors:  utils.ValidateErrors(err),
		})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "未上传文件",
			Errors:  err.Error(),
		})
		return
	}
	material, err := h.service.Upload(c.Request.Context(), userID, req, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "上传素材失败",
			Errors:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, vo.SuccessResponse{
		Code:    201,
		Message: "素材上传成功",
		Data:    material,
	})
}

// Update 更新素材
func (h *MaterialHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors:  utils.ValidateErrors(err),
		})
		return
	}
	material, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "更新素材失败",
			Errors:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "素材更新成功",
		Data:    material,
	})
}

// Delete 删除素材
func (h *MaterialHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "删除素材失败",
			Errors:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "素材删除成功",
		Data:    nil,
	})
} 