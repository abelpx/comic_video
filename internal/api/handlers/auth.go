package handlers

import (
	"net/http"

	"comic_video/internal/domain/dto"
	"comic_video/internal/domain/vo"
	"comic_video/internal/service/auth"
	"comic_video/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *auth.Service
}

func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "参数错误",
			Errors:  utils.ValidateErrors(err),
		})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "注册成功",
		Data:    user,
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "参数错误",
			Errors:  utils.ValidateErrors(err),
		})
		return
	}

	token, user, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{
			Code:    401,
			Message: "用户名或密码错误",
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "登录成功",
		Data: gin.H{
			"token": token,
			"user":  user,
		},
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{
			Code:    401,
			Message: "未授权",
		})
		return
	}

	err := h.authService.Logout(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "登出成功",
	})
}

// GetProfile 获取用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{
			Code:    401,
			Message: "未授权",
		})
		return
	}

	user, err := h.authService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "success",
		Data:    user,
	})
}

// UpdateProfile 更新用户信息
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{
			Code:    401,
			Message: "未授权",
		})
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{
			Code:    400,
			Message: "参数错误",
			Errors:  utils.ValidateErrors(err),
		})
		return
	}

	user, err := h.authService.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.SuccessResponse{
		Code:    200,
		Message: "更新成功",
		Data:    user,
	})
} 