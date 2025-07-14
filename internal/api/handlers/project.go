package handlers

import (
	"net/http"
	"strconv"

	"comic_video/internal/domain/dto"
	"comic_video/internal/domain/vo"
	"comic_video/internal/service/project"
	"comic_video/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projectService *project.Service
}

func NewProjectHandler(projectService *project.Service) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// Create 创建项目
func (h *ProjectHandler) Create(c *gin.Context) {
	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors:  utils.ValidateErrors(err),
		})
		return
	}

	userID := c.GetString("user_id")
	project, err := h.projectService.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "创建项目失败",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, vo.SuccessResponse{
		Code:    201,
		Message: "项目创建成功",
		Data:    project,
	})
}

// GetByID 根据ID获取项目
func (h *ProjectHandler) GetByID(c *gin.Context) {
	projectID := c.Param("id")
	userID := c.GetString("user_id")

	project, err := h.projectService.GetByID(c.Request.Context(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{
			Code:    404,
			Message: "项目不存在或无权限访问",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取项目成功",
		Data:    project,
	})
}

// List 获取项目列表
func (h *ProjectHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	userID := c.GetString("user_id")

	projects, total, err := h.projectService.List(c.Request.Context(), userID, offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "获取项目列表失败",
			Errors:  err.Error(),
		})
		return
	}

	response := dto.ListProjectsResponse{
		Projects: projects,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取项目列表成功",
		Data:    response,
	})
}

// Update 更新项目
func (h *ProjectHandler) Update(c *gin.Context) {
	projectID := c.Param("id")
	var req dto.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors:  utils.ValidateErrors(err),
		})
		return
	}

	userID := c.GetString("user_id")
	project, err := h.projectService.Update(c.Request.Context(), projectID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "更新项目失败",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "项目更新成功",
		Data:    project,
	})
}

// Delete 删除项目
func (h *ProjectHandler) Delete(c *gin.Context) {
	projectID := c.Param("id")
	userID := c.GetString("user_id")

	err := h.projectService.Delete(c.Request.Context(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "删除项目失败",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "项目删除成功",
		Data:    nil,
	})
}

// Share 分享项目
func (h *ProjectHandler) Share(c *gin.Context) {
	projectID := c.Param("id")
	var req dto.ShareProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors:  utils.ValidateErrors(err),
		})
		return
	}

	userID := c.GetString("user_id")
	share, err := h.projectService.Share(c.Request.Context(), projectID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "分享项目失败",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "项目分享成功",
		Data:    share,
	})
}

// CheckShare 校验分享token和密码，返回项目信息
func (h *ProjectHandler) CheckShare(c *gin.Context) {
	token := c.Param("token")
	var req struct{ Password string `json:"password"` }
	_ = c.ShouldBindJSON(&req)

	share, err := h.projectService.CheckShareToken(c.Request.Context(), token, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{
			Code:    401,
			Message: "分享校验失败",
			Errors:  err.Error(),
		})
		return
	}
	// 获取项目信息
	project, err := h.projectService.GetByID(c.Request.Context(), share.ProjectID.String(), share.CreatedBy.String())
	if err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{
			Code:    404,
			Message: "项目不存在",
			Errors:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "分享校验成功",
		Data:    project,
	})
}

// CancelShare 取消/失效分享
func (h *ProjectHandler) CancelShare(c *gin.Context) {
	shareID := c.Param("share_id")
	id, err := uuid.Parse(shareID)
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "无效的分享ID",
			Errors:  err.Error(),
		})
		return
	}
	err = h.projectService.ExpireShare(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "取消分享失败",
			Errors:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "分享已取消/失效",
		Data:    nil,
	})
} 