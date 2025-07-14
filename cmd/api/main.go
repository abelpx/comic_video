package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"comic_video/internal/api/routes"
	"comic_video/internal/config"
	"comic_video/internal/repository/minio"
	"comic_video/internal/repository/postgres"
	"comic_video/internal/repository/redis"
	"comic_video/internal/service/auth"
	"comic_video/internal/service/material"
	"comic_video/internal/service/project"
	"comic_video/internal/service/render"
	"comic_video/internal/service/template"
	"comic_video/internal/service/user"
	"comic_video/internal/service/video"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 设置Gin模式
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 连接数据库
	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库表
	if err := postgres.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 连接Redis
	redisAddr := cfg.Redis.Host + ":" + cfg.Redis.Port
	redisClient, err := redis.NewClient(redisAddr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// 初始化MinIO
	minioClient := minio.NewClient(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKeyID,
		cfg.MinIO.SecretAccessKey,
		cfg.MinIO.BucketName,
		"", // publicHost 暂无
		cfg.MinIO.UseSSL,
	)

	// 初始化仓库
	userRepo := postgres.NewUserRepository(db)
	var projectRepo = postgres.NewProjectRepository(db)
	var projectShareRepo = postgres.NewProjectShareRepository(db)
	videoRepo := postgres.NewVideoRepository(db)
	templateRepo := postgres.NewTemplateRepository(db)
	var materialRepo = postgres.NewMaterialRepository(db)
	renderRepo := postgres.NewRenderRepository(db)

	// 初始化服务
	authService := auth.NewService(userRepo, redisClient, &cfg.JWT)
	userService := user.NewService(userRepo)
	projectService := project.NewService(projectRepo, projectShareRepo)
	videoService := video.NewService(videoRepo, minioClient)
	templateService := template.NewService(templateRepo)
	materialService := material.NewService(materialRepo, minioClient)
	// 初始化渲染队列
	queue := render.NewMemoryRenderQueue(100)

	renderService := render.NewService(
		renderRepo,
		projectRepo,
		materialRepo,
		minioClient,
		cfg.MinIO.BucketName,
		queue, // 注入队列
	)

	queue.StartWorker(4, renderService)

	// 设置路由
	router := routes.SetupRoutes(
		userService,
		authService,
		videoService,
		templateService,
		renderService,
		projectService,
		materialService,
		redisClient, // 新增参数
	)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// 启动服务器
	go func() {
		log.Printf("Starting server on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
