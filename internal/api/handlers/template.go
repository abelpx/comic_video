package handlers

import (
	"net/http"
	"comic_video/internal/domain/dto"
	"comic_video/internal/domain/vo"
	"comic_video/internal/service/template"
	"comic_video/internal/utils"
	"github.com/gin-gonic/gin"
)

type TemplateHandler struct {
	service *template.Service
}

func NewTemplateHandler(service *template.Service) *TemplateHandler {
	return &TemplateHandler{service: service}
}

// List 获取模板列表
func (h *TemplateHandler) List(c *gin.Context) {
	var req dto.ListTemplatesRequest
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
	templates, total, err := h.service.List(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "获取模板列表失败",
			Errors: err.Error(),
		})
		return
	}
	resp := dto.ListTemplatesResponse{
		Templates: templates,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}
	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取模板列表成功",
		Data:    resp,
	})
}

// GetByID 获取模板详情
func (h *TemplateHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	template, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{
			Code:    404,
			Message: "模板不存在",
			Errors: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取模板成功",
		Data:    template,
	})
}

// Create 创建模板
func (h *TemplateHandler) Create(c *gin.Context) {
	var req dto.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors: utils.ValidateErrors(err),
		})
		return
	}
	template, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "创建模板失败",
			Errors: err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, vo.SuccessResponse{
		Code:    201,
		Message: "模板创建成功",
		Data:    template,
	})
}

// Update 更新模板
func (h *TemplateHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors: utils.ValidateErrors(err),
		})
		return
	}
	template, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "更新模板失败",
			Errors: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "模板更新成功",
		Data:    template,
	})
}

// Delete 删除模板
func (h *TemplateHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "删除模板失败",
			Errors: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "模板删除成功",
		Data:    nil,
	})
}

// Apply 应用模板
func (h *TemplateHandler) Apply(c *gin.Context) {
	id := c.Param("id")
	var req dto.ApplyTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors: utils.ValidateErrors(err),
		})
		return
	}
	resp, err := h.service.Apply(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "应用模板失败",
			Errors: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "模板应用成功",
		Data:    resp,
	})
} 