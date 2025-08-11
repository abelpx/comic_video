package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"comic_video/internal/service/monitoring"
	"comic_video/internal/service/orchestrator"
	"comic_video/internal/service/quality"
	"comic_video/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIGateway API网关
type APIGateway struct {
	orchestrator    *orchestrator.WorkflowOrchestrator
	qualityController *quality.AdvancedQualityController
	monitor         *monitoring.IntelligentMonitor
	rateLimiter     *RateLimiter
	authManager     *AuthManager
	router          *gin.Engine
}

// NewAPIGateway 创建API网关
func NewAPIGateway(
	orchestrator *orchestrator.WorkflowOrchestrator,
	qualityController *quality.AdvancedQualityController,
	monitor *monitoring.IntelligentMonitor,
) *APIGateway {
	gateway := &APIGateway{
		orchestrator:      orchestrator,
		qualityController: qualityController,
		monitor:          monitor,
		rateLimiter:      NewRateLimiter(),
		authManager:      NewAuthManager(),
	}
	
	gateway.setupRoutes()
	return gateway
}

// setupRoutes 设置路由
func (gw *APIGateway) setupRoutes() {
	gw.router = gin.New()
	
	// 中间件
	gw.router.Use(gw.loggingMiddleware())
	gw.router.Use(gw.corsMiddleware())
	gw.router.Use(gw.rateLimitMiddleware())
	gw.router.Use(gw.authMiddleware())
	gw.router.Use(gw.monitoringMiddleware())
	
	// API版本分组
	v1 := gw.router.Group("/api/v1")
	{
		// 项目管理
		projects := v1.Group("/projects")
		{
			projects.POST("/", gw.createProject)
			projects.GET("/:id", gw.getProject)
			projects.GET("/:id/status", gw.getProjectStatus)
			projects.DELETE("/:id", gw.deleteProject)
		}
		
		// 工作流执行
		workflow := v1.Group("/workflow")
		{
			workflow.POST("/execute", gw.executeWorkflow)
			workflow.GET("/status/:id", gw.getWorkflowStatus)
			workflow.POST("/cancel/:id", gw.cancelWorkflow)
		}
		
		// 质量控制
		quality := v1.Group("/quality")
		{
			quality.POST("/check", gw.qualityCheck)
			quality.GET("/report/:id", gw.getQualityReport)
		}
		
		// 监控和指标
		monitoring := v1.Group("/monitoring")
		{
			monitoring.GET("/metrics", gw.getMetrics)
			monitoring.GET("/health", gw.getHealth)
			monitoring.GET("/alerts", gw.getAlerts)
			monitoring.POST("/alerts", gw.createAlert)
		}
		
		// 系统管理
		system := v1.Group("/system")
		{
			system.GET("/info", gw.getSystemInfo)
			system.GET("/performance", gw.getPerformanceReport)
			system.POST("/maintenance", gw.enterMaintenance)
		}
	}
	
	// WebSocket端点
	gw.router.GET("/ws/progress/:id", gw.handleWebSocket)
	
	// 静态文件服务
	gw.router.Static("/static", "./static")
	gw.router.StaticFile("/", "./static/index.html")
}

// HTTP处理器
func (gw *APIGateway) createProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gw.respondError(c, http.StatusBadRequest, "无效的请求参数", err)
		return
	}
	
	// 验证请求
	if err := gw.validateCreateProjectRequest(&req); err != nil {
		gw.respondError(c, http.StatusBadRequest, "请求验证失败", err)
		return
	}
	
	// 创建项目
	project := &Project{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Novel:       req.Novel,
		Style:       req.Style,
		Status:      "created",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// 记录指标
	monitoring.IncrementCounter(monitoring.MonitoringMetrics.AppProjects, map[string]string{
		"action": "create",
		"style":  req.Style,
	})
	
	gw.respondSuccess(c, project)
}

func (gw *APIGateway) executeWorkflow(c *gin.Context) {
	var req ExecuteWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gw.respondError(c, http.StatusBadRequest, "无效的请求参数", err)
		return
	}
	
	// 创建编排请求
	orchestrationReq := &orchestrator.OrchestrationRequest{
		ProjectID:         req.ProjectID,
		Novel:            req.Novel,
		Style:            req.Style,
		QualityLevel:     req.QualityLevel,
		OptimizationLevel: req.OptimizationLevel,
		CustomOptions:    req.CustomOptions,
		Callbacks: &orchestrator.OrchestrationCallbacks{
			OnProgress: func(stage string, progress float64, message string) {
				// 通过WebSocket发送进度更新
				gw.broadcastProgress(req.ProjectID, stage, progress, message)
			},
			OnStageComplete: func(stage string, result interface{}, duration time.Duration) {
				log.Printf("[APIGateway] 阶段完成: %s, 耗时: %v", stage, duration)
			},
			OnQualityCheck: func(report *quality.ComprehensiveQualityReport) {
				log.Printf("[APIGateway] 质量检查完成: 分数=%.2f", report.OverallScore)
			},
			OnError: func(stage string, err error, canRetry bool) {
				log.Printf("[APIGateway] 阶段错误: %s, 错误: %v, 可重试: %v", stage, err, canRetry)
			},
		},
	}
	
	// 异步执行工作流
	go func() {
		ctx := context.Background()
		startTime := time.Now()
		
		result, err := gw.orchestrator.ExecuteOrchestration(ctx, orchestrationReq)
		duration := time.Since(startTime)
		
		// 记录指标
		monitoring.RecordMetric(monitoring.MonitoringMetrics.AppResponseTime, float64(duration.Milliseconds()), map[string]string{
			"operation": "workflow_execution",
			"success":   strconv.FormatBool(err == nil),
		})
		
		if err != nil {
			log.Printf("[APIGateway] 工作流执行失败: %v", err)
			monitoring.IncrementCounter("app.workflow_errors", map[string]string{
				"project_id": req.ProjectID.String(),
			})
		} else {
			log.Printf("[APIGateway] 工作流执行成功: %s", result.FinalVideoPath)
			monitoring.IncrementCounter("app.workflow_success", map[string]string{
				"project_id": req.ProjectID.String(),
				"quality":    req.QualityLevel,
			})
		}
	}()
	
	gw.respondSuccess(c, map[string]interface{}{
		"message":    "工作流已开始执行",
		"project_id": req.ProjectID,
		"status":     "executing",
	})
}

func (gw *APIGateway) getProjectStatus(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		gw.respondError(c, http.StatusBadRequest, "无效的项目ID", err)
		return
	}
	
	// 获取项目状态（这里应该从数据库获取）
	status := &ProjectStatus{
		ProjectID:     projectID,
		Status:        "executing",
		Progress:      0.65,
		CurrentStage:  "image_generation",
		Message:       "正在生成图像...",
		StartTime:     time.Now().Add(-10 * time.Minute),
		EstimatedEnd:  time.Now().Add(5 * time.Minute),
		LastUpdate:    time.Now(),
	}
	
	gw.respondSuccess(c, status)
}

func (gw *APIGateway) qualityCheck(c *gin.Context) {
	var req QualityCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gw.respondError(c, http.StatusBadRequest, "无效的请求参数", err)
		return
	}
	
	// 创建质量检查请求
	qualityReq := &quality.QualityCheckRequest{
		ProjectID: req.ProjectID,
		VideoPath: req.VideoPath,
		CheckType: req.CheckType,
		Options:   req.Options,
	}
	
	// 执行质量检查
	ctx := context.Background()
	report, err := gw.qualityController.ComprehensiveQualityCheck(ctx, qualityReq)
	if err != nil {
		gw.respondError(c, http.StatusInternalServerError, "质量检查失败", err)
		return
	}
	
	gw.respondSuccess(c, report)
}

func (gw *APIGateway) getMetrics(c *gin.Context) {
	metrics := gw.monitor.GetMetrics()
	gw.respondSuccess(c, metrics)
}

func (gw *APIGateway) getHealth(c *gin.Context) {
	health := gw.monitor.GetSystemHealth()
	
	// 根据健康状态设置HTTP状态码
	statusCode := http.StatusOK
	if health.OverallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}
	
	c.JSON(statusCode, APIResponse{
		Success: health.OverallStatus == "healthy",
		Data:    health,
		Message: fmt.Sprintf("系统状态: %s", health.OverallStatus),
	})
}

func (gw *APIGateway) getAlerts(c *gin.Context) {
	alerts := gw.monitor.GetAlerts()
	gw.respondSuccess(c, alerts)
}

func (gw *APIGateway) getSystemInfo(c *gin.Context) {
	info := &SystemInfo{
		Version:     "1.0.0",
		BuildTime:   "2024-01-01T00:00:00Z",
		Environment: "production",
		Uptime:      time.Since(time.Now().Add(-2 * time.Hour)),
		Features: map[string]bool{
			"quality_control":     true,
			"advanced_animation":  true,
			"character_consistency": true,
			"scene_analysis":      true,
			"monitoring":          true,
		},
	}
	
	gw.respondSuccess(c, info)
}

func (gw *APIGateway) getPerformanceReport(c *gin.Context) {
	report := gw.monitor.GetPerformanceReport()
	gw.respondSuccess(c, report)
}

// 中间件
func (gw *APIGateway) loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

func (gw *APIGateway) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

func (gw *APIGateway) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		if !gw.rateLimiter.Allow(clientIP) {
			gw.respondError(c, http.StatusTooManyRequests, "请求频率过高", nil)
			c.Abort()
			return
		}
		
		c.Next()
	}
}

func (gw *APIGateway) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过健康检查和公开端点
		if c.Request.URL.Path == "/api/v1/monitoring/health" || 
		   c.Request.URL.Path == "/api/v1/system/info" {
			c.Next()
			return
		}
		
		token := c.GetHeader("Authorization")
		if token == "" {
			gw.respondError(c, http.StatusUnauthorized, "缺少认证令牌", nil)
			c.Abort()
			return
		}
		
		// 验证令牌
		userID, err := gw.authManager.ValidateToken(token)
		if err != nil {
			gw.respondError(c, http.StatusUnauthorized, "无效的认证令牌", err)
			c.Abort()
			return
		}
		
		c.Set("user_id", userID)
		c.Next()
	}
}

func (gw *APIGateway) monitoringMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		duration := time.Since(start)
		
		// 记录请求指标
		monitoring.RecordMetric(monitoring.MonitoringMetrics.AppResponseTime, float64(duration.Milliseconds()), map[string]string{
			"method":   c.Request.Method,
			"endpoint": c.Request.URL.Path,
			"status":   strconv.Itoa(c.Writer.Status()),
		})
		
		// 记录吞吐量
		monitoring.IncrementCounter(monitoring.MonitoringMetrics.AppThroughput, map[string]string{
			"method":   c.Request.Method,
			"endpoint": c.Request.URL.Path,
		})
		
		// 记录错误率
		if c.Writer.Status() >= 400 {
			monitoring.IncrementCounter(monitoring.MonitoringMetrics.AppErrorRate, map[string]string{
				"method":   c.Request.Method,
				"endpoint": c.Request.URL.Path,
				"status":   strconv.Itoa(c.Writer.Status()),
			})
		}
	}
}

// 辅助方法
func (gw *APIGateway) respondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Message: "操作成功",
	})
}

func (gw *APIGateway) respondError(c *gin.Context, statusCode int, message string, err error) {
	response := APIResponse{
		Success: false,
		Message: message,
	}
	
	if err != nil {
		response.Error = err.Error()
	}
	
	c.JSON(statusCode, response)
}

func (gw *APIGateway) validateCreateProjectRequest(req *CreateProjectRequest) error {
	if req.Name == "" {
		return fmt.Errorf("项目名称不能为空")
	}
	if req.Novel == "" {
		return fmt.Errorf("小说内容不能为空")
	}
	if len(req.Novel) < 100 {
		return fmt.Errorf("小说内容过短，至少需要100个字符")
	}
	return nil
}

func (gw *APIGateway) broadcastProgress(projectID uuid.UUID, stage string, progress float64, message string) {
	// WebSocket广播实现
	log.Printf("[APIGateway] 进度更新: %s - %.2f%% - %s", stage, progress*100, message)
}

func (gw *APIGateway) handleWebSocket(c *gin.Context) {
	// WebSocket处理实现
	projectIDStr := c.Param("id")
	log.Printf("[APIGateway] WebSocket连接: %s", projectIDStr)
	
	// 这里应该实现WebSocket升级和消息处理
	c.JSON(http.StatusOK, map[string]string{
		"message": "WebSocket endpoint",
		"project": projectIDStr,
	})
}

// Start 启动API网关
func (gw *APIGateway) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("[APIGateway] 启动API网关: %s", addr)
	return gw.router.Run(addr)
}

// 数据结构定义
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message"`
	Error   string      `json:"error,omitempty"`
}

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Novel       string `json:"novel" binding:"required"`
	Style       string `json:"style"`
}

type ExecuteWorkflowRequest struct {
	ProjectID         uuid.UUID              `json:"project_id" binding:"required"`
	Novel            string                 `json:"novel" binding:"required"`
	Style            string                 `json:"style"`
	QualityLevel     string                 `json:"quality_level"`
	OptimizationLevel string                 `json:"optimization_level"`
	CustomOptions    map[string]interface{} `json:"custom_options"`
}

type QualityCheckRequest struct {
	ProjectID uuid.UUID              `json:"project_id" binding:"required"`
	VideoPath string                 `json:"video_path" binding:"required"`
	CheckType string                 `json:"check_type"`
	Options   map[string]interface{} `json:"options"`
}

type Project struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Novel       string    `json:"novel"`
	Style       string    `json:"style"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProjectStatus struct {
	ProjectID     uuid.UUID `json:"project_id"`
	Status        string    `json:"status"`
	Progress      float64   `json:"progress"`
	CurrentStage  string    `json:"current_stage"`
	Message       string    `json:"message"`
	StartTime     time.Time `json:"start_time"`
	EstimatedEnd  time.Time `json:"estimated_end"`
	LastUpdate    time.Time `json:"last_update"`
}

type SystemInfo struct {
	Version     string            `json:"version"`
	BuildTime   string            `json:"build_time"`
	Environment string            `json:"environment"`
	Uptime      time.Duration     `json:"uptime"`
	Features    map[string]bool   `json:"features"`
}
