package routes

import (
	"comic_video/internal/api/handlers"
	"comic_video/internal/api/middleware"
	"comic_video/internal/service/auth"
	"comic_video/internal/service/project"
	"comic_video/internal/service/template"
	"comic_video/internal/service/user"
	"comic_video/internal/service/video"
	"comic_video/internal/service/render"
	"comic_video/internal/service/material"
	"comic_video/internal/repository/redis"
	"comic_video/internal/service/ai"
	"comic_video/internal/service/workflow"
	"comic_video/internal/service/tweet"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes 设置路由
func SetupRoutes(
	userService *user.Service,
	authService *auth.Service,
	videoService *video.Service,
	templateService *template.Service,
	renderService render.Service,
	projectService *project.Service,
	materialService *material.Service,
	redisClient *redis.Client,
	taskQueue ai.TaskQueue,
	workflowEngine *workflow.WorkflowEngine, // 新增工作流引擎参数
	aiService *ai.Service, // 新增AI服务参数
	db *gorm.DB, // 新增数据库连接参数
) *gin.Engine {
	router := gin.Default()

	// 添加中间件
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// API版本分组
	v1 := router.Group("/api/v1")

	// 认证相关路由
	authHandler := handlers.NewAuthHandler(authService)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", middleware.AuthMiddleware(authService), authHandler.Logout)
		auth.GET("/profile", middleware.AuthMiddleware(authService), authHandler.GetProfile)
		auth.PUT("/profile", middleware.AuthMiddleware(authService), authHandler.UpdateProfile)
	}

	// 用户相关路由
	userHandler := handlers.NewUserHandler(userService)
	users := v1.Group("/users")
	users.Use(middleware.AuthMiddleware(authService))
	{
		users.GET("/", userHandler.List)
		users.GET("/:id", userHandler.GetByID)
		users.PUT("/:id", userHandler.Update)
		users.DELETE("/:id", userHandler.Delete)
	}

	// 项目相关路由
	projectHandler := handlers.NewProjectHandler(projectService)
	projects := v1.Group("/projects")
	projects.Use(middleware.AuthMiddleware(authService))
	{
		projects.GET("/", projectHandler.List)
		projects.POST("/", projectHandler.Create)
		projects.GET("/:id", projectHandler.GetByID)
		projects.PUT("/:id", projectHandler.Update)
		projects.DELETE("/:id", projectHandler.Delete)
		projects.POST("/:id/share", projectHandler.Share)
	}

	// 项目分享相关路由
	v1.POST("/share/check/:token", projectHandler.CheckShare) // 校验分享（无需登录）
	v1.POST("/share/cancel/:share_id", middleware.AuthMiddleware(authService), projectHandler.CancelShare) // 取消分享（需登录）

	// 视频相关路由
	videoHandler := handlers.NewVideoHandler(videoService)
	videos := v1.Group("/videos")
	videos.Use(middleware.AuthMiddleware(authService))
	{
		videos.GET("/", videoHandler.List)
		videos.POST("/upload", videoHandler.Upload)
		videos.GET("/:id", videoHandler.GetByID)
		videos.PUT("/:id", videoHandler.Update)
		videos.DELETE("/:id", videoHandler.Delete)
		videos.POST("/:id/process", videoHandler.Process)
		videos.GET("/:id/status", videoHandler.GetStatus)
	}

	// 模板相关路由
	templateHandler := handlers.NewTemplateHandler(templateService)
	templates := v1.Group("/templates")
	{
		templates.GET("/", templateHandler.List)
		templates.GET("/:id", templateHandler.GetByID)
		templates.POST("/", middleware.AuthMiddleware(authService), templateHandler.Create)
		templates.PUT("/:id", middleware.AuthMiddleware(authService), templateHandler.Update)
		templates.DELETE("/:id", middleware.AuthMiddleware(authService), templateHandler.Delete)
		templates.POST("/:id/apply", middleware.AuthMiddleware(authService), templateHandler.Apply)
	}

	// 渲染相关路由
	renderHandler := handlers.NewRenderHandler(renderService)
	renders := v1.Group("/renders")
	renders.Use(middleware.AuthMiddleware(authService))
	{
		renders.GET("/", renderHandler.ListRenders)
		renders.POST("/", renderHandler.CreateRender)
		renders.GET("/:id", renderHandler.GetRender)
		renders.DELETE("/:id", renderHandler.DeleteRender)
		renders.GET("/:id/status", renderHandler.GetRenderStatus)
		renders.GET("/:id/download", renderHandler.DownloadRender)
	}

	// 素材相关路由
	materialHandler := handlers.NewMaterialHandler(materialService)
	materials := v1.Group("/materials")
	{
		materials.GET("/", materialHandler.List)
		materials.GET("/:id", materialHandler.GetByID)
		materials.POST("/upload", middleware.AuthMiddleware(authService), materialHandler.Upload)
		materials.PUT("/:id", middleware.AuthMiddleware(authService), materialHandler.Update)
		materials.DELETE("/:id", middleware.AuthMiddleware(authService), materialHandler.Delete)
	}

	// 通用任务进度查询API
	taskHandler := handlers.NewTaskHandler(redisClient, nil) // 暂时传nil，后续可以添加真实的taskRepo
	v1.GET("/task/:id/status", taskHandler.GetTaskStatus)
	v1.GET("/tasks", taskHandler.GetUserTasks) // 获取用户任务列表

	// AI 相关路由
	aiHandler := handlers.NewAIHandler(redisClient, taskQueue, aiService, db)
	ai := v1.Group("/ai")
	{
		ai.POST("/novel-to-video", aiHandler.NovelToVideo)
		ai.POST("/generate-novel", aiHandler.GenerateNovel)
		ai.POST("/generate-tweet", aiHandler.GenerateTweet)     // 新增：推文生成
		ai.POST("/novel-to-tweet", aiHandler.NovelToTweet)      // 新增：小说转推文
		ai.POST("/novel-to-all", aiHandler.NovelToAll)
		ai.GET("/quota", aiHandler.GetUserQuota)        // 获取用户配额
		ai.GET("/usage-stats", aiHandler.GetUsageStats) // 获取使用统计
	}

	// TTS 相关路由
	ttsHandler := handlers.NewTTSHandler(aiService.GetTTSClient())
	tts := v1.Group("/tts")
	{
		// 公开接口（无需认证）
		tts.GET("/health", ttsHandler.HealthCheck)           // 健康检查
		tts.GET("/info", ttsHandler.GetServiceInfo)          // 服务信息
		tts.GET("/voices", ttsHandler.GetVoices)             // 获取可用语音
		tts.GET("/config", ttsHandler.GetConfig)             // 获取配置信息
		tts.GET("/test", ttsHandler.TestVoice)               // 测试语音生成

		// 需要认证的接口
		authTTS := tts.Group("")
		authTTS.Use(middleware.AuthMiddleware(authService))
		{
			authTTS.POST("/generate", ttsHandler.GenerateVoice) // 生成语音
		}
	}

	// 推文相关路由
	tweetService := tweet.NewService(aiService, db)
	templateTweetService := tweet.NewTemplateService(db)
	tweetHandler := handlers.NewTweetHandler(tweetService, templateTweetService, db)

	tweets := v1.Group("/tweets")
	tweets.Use(middleware.AuthMiddleware(authService)) // 推文功能需要认证
	{
		// 推文管理
		tweets.POST("", tweetHandler.SaveTweet)                    // 保存推文
		tweets.GET("", tweetHandler.ListTweets)                    // 获取推文列表
		tweets.GET("/search", tweetHandler.SearchTweets)           // 搜索推文
		tweets.GET("/stats", tweetHandler.GetTweetStats)           // 获取推文统计
		tweets.GET("/:id", tweetHandler.GetTweet)                  // 获取推文详情
		tweets.PUT("/:id", tweetHandler.UpdateTweet)               // 更新推文
		tweets.DELETE("/:id", tweetHandler.DeleteTweet)            // 删除推文
		tweets.POST("/:id/publish", tweetHandler.PublishTweet)     // 发布推文
	}

	// 推文模板相关路由
	tweetTemplates := v1.Group("/tweet-templates")
	{
		// 公开模板（无需认证）
		tweetTemplates.GET("", tweetHandler.ListTemplates)              // 获取模板列表
		tweetTemplates.GET("/search", tweetHandler.SearchTemplates)     // 搜索模板
		tweetTemplates.GET("/popular", tweetHandler.GetPopularTemplates) // 获取热门模板
		tweetTemplates.GET("/categories", tweetHandler.GetTemplateCategories) // 获取模板分类
		tweetTemplates.GET("/:id", tweetHandler.GetTemplate)            // 获取模板详情
		tweetTemplates.POST("/:id/use", tweetHandler.UseTemplate)       // 使用模板

		// 需要认证的模板操作
		authTemplates := tweetTemplates.Group("")
		authTemplates.Use(middleware.AuthMiddleware(authService))
		{
			authTemplates.POST("", tweetHandler.CreateTemplate)     // 创建模板
			authTemplates.PUT("/:id", tweetHandler.UpdateTemplate)  // 更新模板
			authTemplates.DELETE("/:id", tweetHandler.DeleteTemplate) // 删除模板
		}
	}

	// 工作流相关路由
	workflowHandler := handlers.NewWorkflowHandler(workflowEngine)
	workflows := v1.Group("/workflows")
	workflows.Use(middleware.AuthMiddleware(authService)) // 需要认证
	{
		workflows.POST("/", workflowHandler.CreateWorkflow)
		workflows.POST("/:id/start", workflowHandler.StartWorkflow)
		workflows.GET("/:id", workflowHandler.GetWorkflow)
		workflows.GET("/", workflowHandler.ListWorkflows)
		workflows.GET("/:id/tasks", workflowHandler.GetWorkflowTasks)
		workflows.GET("/:id/progress", workflowHandler.GetWorkflowProgress)
		workflows.POST("/:id/cancel", workflowHandler.CancelWorkflow)
		workflows.POST("/novel-to-video", workflowHandler.NovelToVideoWorkflow) // 完整工作流
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "VidCraft Studio API",
		})
	})

	return router
} 