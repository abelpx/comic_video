package handlers

import (
	"net/http"
	"strconv"

	"comic_video/internal/domain/dto"
	"comic_video/internal/domain/vo"
	"comic_video/internal/service/user"
	"comic_video/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *user.Service
}

func NewUserHandler(userService *user.Service) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetByID 根据ID获取用户
func (h *UserHandler) GetByID(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{
			Code:    404,
			Message: "用户不存在",
			Errors: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取用户成功",
		Data:    user,
	})
}

// List 获取用户列表
func (h *UserHandler) List(c *gin.Context) {
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

	users, total, err := h.userService.List(c.Request.Context(), offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "获取用户列表失败",
			Errors: err.Error(),
		})
		return
	}

	response := dto.ListUsersResponse{
		Users:    users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "获取用户列表成功",
		Data:    response,
	})
}

// Update 更新用户
func (h *UserHandler) Update(c *gin.Context) {
	userID := c.Param("id")
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Errors: utils.ValidateErrors(err),
		})
		return
	}

	user, err := h.userService.Update(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "更新用户失败",
			Errors: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "用户更新成功",
		Data:    user,
	})
}

// Delete 删除用户
func (h *UserHandler) Delete(c *gin.Context) {
	userID := c.Param("id")

	err := h.userService.Delete(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: "删除用户失败",
			Errors: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "用户删除成功",
		Data:    nil,
	})
} 