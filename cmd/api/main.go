package main

import (
	"context"
	"encoding/json"
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
	"comic_video/internal/service/ai"
	"comic_video/internal/domain/entity"
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

	// 初始化AI能力
	sdClient := &ai.SDClient{Endpoint: cfg.AI.SDEndpoint}
	ollamaClient := &ai.OllamaClient{Endpoint: cfg.AI.OllamaEndpoint, Model: cfg.AI.OllamaModel, ApiKey: cfg.AI.OllamaApiKey}
	ttsClient := &ai.TTSClient{Endpoint: cfg.AI.TTSEndpoint}
	// 可根据需要初始化 WhisperClient 等

	// 初始化通用任务队列
	taskQueue := ai.NewMemoryTaskQueue(100)
	log.Println("[Init] AI任务队列已创建", taskQueue)

	// 启动 NovelToVideo worker
	taskQueue.StartWorker(4, func(task *entity.Task) {
		log.Printf("[Worker] 收到任务: id=%v type=%v status=%v", task.ID, task.Type, task.Status)
		if task.Type == entity.TaskTypeVideo {
			log.Printf("[Worker] 开始处理视频任务: id=%v", task.ID)
			err := ai.ProcessNovelToVideo(
				context.Background(),
				task,
				redisClient,
				sdClient,
				ollamaClient,
				ttsClient,
				cfg.MinIO.BucketName,
			)
			if err == nil {
				log.Printf("[Worker] 视频任务处理完成: id=%v", task.ID)
				// 上传视频到 MinIO
				var result map[string]interface{}
				_ = json.Unmarshal([]byte(task.Result), &result)
				videoPath, _ := result["url"].(string)
				url, err := ai.UploadVideoToMinio(context.Background(), minioClient, cfg.MinIO.BucketName, videoPath)
				if err == nil {
					result["url"] = url
					b, _ := json.Marshal(result)
					task.Result = string(b)
					task.Status = entity.TaskStatusCompleted
					task.Progress = 100
					task.UpdatedAt = time.Now()
					_ = redisClient.SetTaskStatus(context.Background(), task, 24*time.Hour)
					log.Printf("[Worker] 视频已上传MinIO: id=%v url=%v", task.ID, url)
				}
			}
			// 新增：写入Postgres历史记录
			dbIns, err := postgres.NewConnection(&cfg.Database)
			if err == nil {
				dbIns.Create(task)
				log.Printf("[Worker] 任务已写入Postgres: id=%v", task.ID)
			}
		} else {
			log.Printf("[Worker] 不支持的任务类型: id=%v type=%v", task.ID, task.Type)
		}
	})

	// 设置路由
	router := routes.SetupRoutes(
		userService,
		authService,
		videoService,
		templateService,
		renderService,
		projectService,
		materialService,
		redisClient,
		taskQueue, // 新增参数
	)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
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
