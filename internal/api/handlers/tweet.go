package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"comic_video/internal/service/tweet"
)

// TweetHandler 推文处理器
type TweetHandler struct {
	tweetService    *tweet.Service
	templateService *tweet.TemplateService
	db              *gorm.DB
}

// NewTweetHandler 创建推文处理器
func NewTweetHandler(tweetService *tweet.Service, templateService *tweet.TemplateService, db *gorm.DB) *TweetHandler {
	return &TweetHandler{
		tweetService:    tweetService,
		templateService: templateService,
		db:              db,
	}
}

// SaveTweet 保存推文
func (h *TweetHandler) SaveTweet(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	var req tweet.SaveTweetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 验证必填字段
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "推文内容不能为空"})
		return
	}

	savedTweet, err := h.tweetService.SaveTweet(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		log.Printf("[TweetHandler] 保存推文失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存推文失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "推文保存成功",
		"data":    savedTweet,
	})
}

// UpdateTweet 更新推文
func (h *TweetHandler) UpdateTweet(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	tweetIDStr := c.Param("id")
	tweetID, err := uuid.Parse(tweetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的推文ID"})
		return
	}

	var req tweet.UpdateTweetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	updatedTweet, err := h.tweetService.UpdateTweet(c.Request.Context(), userID.(uuid.UUID), tweetID, &req)
	if err != nil {
		log.Printf("[TweetHandler] 更新推文失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新推文失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "推文更新成功",
		"data":    updatedTweet,
	})
}

// GetTweet 获取推文详情
func (h *TweetHandler) GetTweet(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	tweetIDStr := c.Param("id")
	tweetID, err := uuid.Parse(tweetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的推文ID"})
		return
	}

	tweetData, err := h.tweetService.GetTweet(c.Request.Context(), userID.(uuid.UUID), tweetID)
	if err != nil {
		log.Printf("[TweetHandler] 获取推文失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取推文失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取推文成功",
		"data":    tweetData,
	})
}

// ListTweets 获取推文列表
func (h *TweetHandler) ListTweets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 解析查询参数
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	req := &tweet.ListTweetsRequest{
		Status: status,
		Limit:  limit,
		Offset: offset,
	}

	response, err := h.tweetService.ListTweets(c.Request.Context(), userID.(uuid.UUID), req)
	if err != nil {
		log.Printf("[TweetHandler] 获取推文列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取推文列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取推文列表成功",
		"data":    response,
	})
}

// SearchTweets 搜索推文
func (h *TweetHandler) SearchTweets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	keyword := c.Query("keyword")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	req := &tweet.SearchTweetsRequest{
		Keyword: keyword,
		Limit:   limit,
		Offset:  offset,
	}

	response, err := h.tweetService.SearchTweets(c.Request.Context(), userID.(uuid.UUID), req)
	if err != nil {
		log.Printf("[TweetHandler] 搜索推文失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "搜索推文失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "搜索推文成功",
		"data":    response,
	})
}

// DeleteTweet 删除推文
func (h *TweetHandler) DeleteTweet(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	tweetIDStr := c.Param("id")
	tweetID, err := uuid.Parse(tweetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的推文ID"})
		return
	}

	err = h.tweetService.DeleteTweet(c.Request.Context(), userID.(uuid.UUID), tweetID)
	if err != nil {
		log.Printf("[TweetHandler] 删除推文失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除推文失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "推文删除成功",
	})
}

// PublishTweet 发布推文
func (h *TweetHandler) PublishTweet(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	tweetIDStr := c.Param("id")
	tweetID, err := uuid.Parse(tweetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的推文ID"})
		return
	}

	err = h.tweetService.PublishTweet(c.Request.Context(), userID.(uuid.UUID), tweetID)
	if err != nil {
		log.Printf("[TweetHandler] 发布推文失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "发布推文失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "推文发布成功",
	})
}

// GetTweetStats 获取推文统计
func (h *TweetHandler) GetTweetStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	stats, err := h.tweetService.GetTweetStats(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		log.Printf("[TweetHandler] 获取推文统计失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取推文统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取推文统计成功",
		"data":    stats,
	})
}

// CreateTemplate 创建推文模板
func (h *TweetHandler) CreateTemplate(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	var req tweet.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 验证必填字段
	if req.Name == "" || req.Template == "" || req.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "模板名称、内容和分类不能为空"})
		return
	}

	template, err := h.templateService.CreateTemplate(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		log.Printf("[TweetHandler] 创建模板失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建模板失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "模板创建成功",
		"data":    template,
	})
}

// UpdateTemplate 更新推文模板
func (h *TweetHandler) UpdateTemplate(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	templateIDStr := c.Param("id")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的模板ID"})
		return
	}

	var req tweet.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	template, err := h.templateService.UpdateTemplate(c.Request.Context(), userID.(uuid.UUID), templateID, &req)
	if err != nil {
		log.Printf("[TweetHandler] 更新模板失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新模板失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "模板更新成功",
		"data":    template,
	})
}

// GetTemplate 获取推文模板详情
func (h *TweetHandler) GetTemplate(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的模板ID"})
		return
	}

	template, err := h.templateService.GetTemplate(c.Request.Context(), templateID)
	if err != nil {
		log.Printf("[TweetHandler] 获取模板失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取模板失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取模板成功",
		"data":    template,
	})
}

// ListTemplates 获取推文模板列表
func (h *TweetHandler) ListTemplates(c *gin.Context) {
	category := c.Query("category")
	platform := c.Query("platform")
	isPublicStr := c.Query("is_public")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	var isPublic *bool
	if isPublicStr != "" {
		val := isPublicStr == "true"
		isPublic = &val
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	req := &tweet.ListTemplatesRequest{
		Category: category,
		Platform: platform,
		IsPublic: isPublic,
		Limit:    limit,
		Offset:   offset,
	}

	response, err := h.templateService.ListTemplates(c.Request.Context(), req)
	if err != nil {
		log.Printf("[TweetHandler] 获取模板列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取模板列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取模板列表成功",
		"data":    response,
	})
}

// SearchTemplates 搜索推文模板
func (h *TweetHandler) SearchTemplates(c *gin.Context) {
	keyword := c.Query("keyword")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	req := &tweet.SearchTemplatesRequest{
		Keyword: keyword,
		Limit:   limit,
		Offset:  offset,
	}

	response, err := h.templateService.SearchTemplates(c.Request.Context(), req)
	if err != nil {
		log.Printf("[TweetHandler] 搜索模板失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "搜索模板失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "搜索模板成功",
		"data":    response,
	})
}

// UseTemplate 使用模板生成推文
func (h *TweetHandler) UseTemplate(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的模板ID"})
		return
	}

	var req struct {
		Variables map[string]string `json:"variables"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	content, err := h.templateService.UseTemplate(c.Request.Context(), templateID, req.Variables)
	if err != nil {
		log.Printf("[TweetHandler] 使用模板失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "使用模板失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "模板使用成功",
		"data": gin.H{
			"content": content,
		},
	})
}

// DeleteTemplate 删除推文模板
func (h *TweetHandler) DeleteTemplate(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	templateIDStr := c.Param("id")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的模板ID"})
		return
	}

	err = h.templateService.DeleteTemplate(c.Request.Context(), userID.(uuid.UUID), templateID)
	if err != nil {
		log.Printf("[TweetHandler] 删除模板失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除模板失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "模板删除成功",
	})
}

// GetPopularTemplates 获取热门模板
func (h *TweetHandler) GetPopularTemplates(c *gin.Context) {
	platform := c.Query("platform")
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	templates, err := h.templateService.GetPopularTemplates(c.Request.Context(), platform, limit)
	if err != nil {
		log.Printf("[TweetHandler] 获取热门模板失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取热门模板失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取热门模板成功",
		"data":    templates,
	})
}

// GetTemplateCategories 获取模板分类
func (h *TweetHandler) GetTemplateCategories(c *gin.Context) {
	categories, err := h.templateService.GetTemplateCategories(c.Request.Context())
	if err != nil {
		log.Printf("[TweetHandler] 获取模板分类失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取模板分类失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取模板分类成功",
		"data":    categories,
	})
}
