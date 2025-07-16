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

	"github.com/gin-gonic/gin"
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
	redisClient *redis.Client, // 新增参数
	taskQueue ai.TaskQueue, // 新增参数
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
	taskHandler := handlers.NewTaskHandler(redisClient)
	v1.GET("/task/:id/status", taskHandler.GetTaskStatus)

	// AI 相关路由
	aiHandler := handlers.NewAIHandler(redisClient, taskQueue)
	v1.POST("/ai/novel-to-video", aiHandler.NovelToVideo)
	v1.POST("/ai/generate-novel", aiHandler.GenerateNovel)
	v1.POST("/ai/novel-to-all", aiHandler.NovelToAll)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "VidCraft Studio API",
		})
	})

	return router
} 