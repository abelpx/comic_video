package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/postgres"
	"comic_video/internal/service/script"
	"comic_video/internal/service/character"
	"comic_video/internal/service/scene"
	"comic_video/internal/service/storyboard"
	"comic_video/internal/service/voice"
	"comic_video/internal/service/music"
	"comic_video/internal/service/editing"
	"comic_video/internal/service/publishing"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowEngine 工作流引擎
type WorkflowEngine struct {
	db               *gorm.DB
	workflowRepo     *postgres.WorkflowRepository
	taskRepo         *postgres.WorkflowTaskRepository
	
	// 各个服务
	scriptService     *script.Service
	characterService  *character.Service
	sceneService      *scene.Service
	storyboardService *storyboard.Service
	voiceService      *voice.Service
	musicService      *music.Service
	editingService    *editing.Service
	publishingService *publishing.Service
	
	// 并发控制
	maxConcurrent int
	semaphore     chan struct{}
	mu            sync.RWMutex
	runningTasks  map[uuid.UUID]context.CancelFunc
}

// NewWorkflowEngine 创建工作流引擎
func NewWorkflowEngine(
	db *gorm.DB,
	scriptService *script.Service,
	characterService *character.Service,
	sceneService *scene.Service,
	storyboardService *storyboard.Service,
	voiceService *voice.Service,
	musicService *music.Service,
	editingService *editing.Service,
	publishingService *publishing.Service,
	maxConcurrent int,
) *WorkflowEngine {
	return &WorkflowEngine{
		db:                db,
		workflowRepo:      postgres.NewWorkflowRepository(db),
		taskRepo:          postgres.NewWorkflowTaskRepository(db),
		scriptService:     scriptService,
		characterService:  characterService,
		sceneService:      sceneService,
		storyboardService: storyboardService,
		voiceService:      voiceService,
		musicService:      musicService,
		editingService:    editingService,
		publishingService: publishingService,
		maxConcurrent:     maxConcurrent,
		semaphore:         make(chan struct{}, maxConcurrent),
		runningTasks:      make(map[uuid.UUID]context.CancelFunc),
	}
}

// CreateWorkflow 创建工作流
func (e *WorkflowEngine) CreateWorkflow(ctx context.Context, req *CreateWorkflowRequest) (*entity.Workflow, error) {
	workflow := &entity.Workflow{
		ProjectID:   req.ProjectID,
		UserID:      req.UserID,
		Name:        req.Name,
		Description: req.Description,
		Status:      entity.WorkflowStatusPending,
		CurrentStep: entity.StepTextInput,
		Config:      req.Config,
		Progress:    0,
	}

	if err := e.workflowRepo.Create(ctx, workflow); err != nil {
		return nil, fmt.Errorf("创建工作流失败: %w", err)
	}

	// 创建工作流任务
	if err := e.createWorkflowTasks(ctx, workflow.ID, req.Steps); err != nil {
		return nil, fmt.Errorf("创建工作流任务失败: %w", err)
	}

	return workflow, nil
}

// StartWorkflow 启动工作流
func (e *WorkflowEngine) StartWorkflow(ctx context.Context, workflowID uuid.UUID) error {
	workflow, err := e.workflowRepo.GetByID(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("获取工作流失败: %w", err)
	}

	if workflow.Status != entity.WorkflowStatusPending {
		return fmt.Errorf("工作流状态不正确: %s", workflow.Status)
	}

	// 更新工作流状态
	now := time.Now()
	workflow.Status = entity.WorkflowStatusRunning
	workflow.StartedAt = &now
	if err := e.workflowRepo.Update(ctx, workflow); err != nil {
		return fmt.Errorf("更新工作流状态失败: %w", err)
	}

	// 异步执行工作流
	go e.executeWorkflow(context.Background(), workflow)

	return nil
}

// executeWorkflow 执行工作流
func (e *WorkflowEngine) executeWorkflow(ctx context.Context, workflow *entity.Workflow) {
	log.Printf("[WorkflowEngine] 开始执行工作流: %s", workflow.ID)

	// 获取工作流任务
	tasks, err := e.taskRepo.GetByWorkflowID(ctx, workflow.ID)
	if err != nil {
		log.Printf("[WorkflowEngine] 获取工作流任务失败: %v", err)
		e.markWorkflowFailed(ctx, workflow.ID, err)
		return
	}

	// 按步骤顺序执行任务
	for _, task := range tasks {
		select {
		case <-ctx.Done():
			log.Printf("[WorkflowEngine] 工作流被取消: %s", workflow.ID)
			return
		default:
		}

		// 获取信号量
		e.semaphore <- struct{}{}
		
		// 执行任务
		if err := e.executeTask(ctx, task); err != nil {
			log.Printf("[WorkflowEngine] 任务执行失败: %s, error: %v", task.ID, err)
			e.markWorkflowFailed(ctx, workflow.ID, err)
			<-e.semaphore // 释放信号量
			return
		}

		<-e.semaphore // 释放信号量

		// 更新工作流进度
		e.updateWorkflowProgress(ctx, workflow.ID)
	}

	// 标记工作流完成
	e.markWorkflowCompleted(ctx, workflow.ID)
	log.Printf("[WorkflowEngine] 工作流执行完成: %s", workflow.ID)
}

// executeTask 执行单个任务
func (e *WorkflowEngine) executeTask(ctx context.Context, task *entity.WorkflowTask) error {
	log.Printf("[WorkflowEngine] 开始执行任务: %s, step: %s", task.ID, task.Step)

	// 更新任务状态
	now := time.Now()
	task.Status = entity.WorkflowStatusRunning
	task.StartedAt = &now
	if err := e.taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("更新任务状态失败: %w", err)
	}

	// 根据步骤类型执行相应的服务
	var result interface{}
	var err error

	switch task.Step {
	case entity.StepScriptAdapt:
		result, err = e.executeScriptAdapt(ctx, task)
	case entity.StepCharacterGen:
		result, err = e.executeCharacterGen(ctx, task)
	case entity.StepSceneGen:
		result, err = e.executeSceneGen(ctx, task)
	case entity.StepStoryboard:
		result, err = e.executeStoryboard(ctx, task)
	case entity.StepVoiceGen:
		result, err = e.executeVoiceGen(ctx, task)
	case entity.StepMusicGen:
		result, err = e.executeMusicGen(ctx, task)
	case entity.StepVideoEdit:
		result, err = e.executeVideoEdit(ctx, task)
	case entity.StepPublish:
		result, err = e.executePublish(ctx, task)
	default:
		err = fmt.Errorf("不支持的任务步骤: %s", task.Step)
	}

	// 更新任务结果
	completedAt := time.Now()
	task.CompletedAt = &completedAt

	if err != nil {
		task.Status = entity.WorkflowStatusFailed
		task.Error = err.Error()
		log.Printf("[WorkflowEngine] 任务执行失败: %s, error: %v", task.ID, err)
	} else {
		task.Status = entity.WorkflowStatusCompleted
		task.Progress = 100
		if result != nil {
			outputBytes, _ := json.Marshal(result)
			task.Output = string(outputBytes)
		}
		log.Printf("[WorkflowEngine] 任务执行成功: %s", task.ID)
	}

	return e.taskRepo.Update(ctx, task)
}

// CreateWorkflowRequest 创建工作流请求
type CreateWorkflowRequest struct {
	ProjectID   uuid.UUID                `json:"project_id"`
	UserID      uuid.UUID                `json:"user_id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Config      string                   `json:"config"`
	Steps       []entity.WorkflowStep    `json:"steps"`
}

// createWorkflowTasks 创建工作流任务
func (e *WorkflowEngine) createWorkflowTasks(ctx context.Context, workflowID uuid.UUID, steps []entity.WorkflowStep) error {
	for i, step := range steps {
		task := &entity.WorkflowTask{
			WorkflowID: workflowID,
			Step:       step,
			Name:       string(step),
			Status:     entity.WorkflowStatusPending,
			Progress:   0,
		}

		if err := e.taskRepo.Create(ctx, task); err != nil {
			return fmt.Errorf("创建任务失败 (step: %s): %w", step, err)
		}
	}

	return nil
}

// executeScriptAdapt 执行剧本改编
func (e *WorkflowEngine) executeScriptAdapt(ctx context.Context, task *entity.WorkflowTask) (interface{}, error) {
	// 从任务输入中解析参数
	var input struct {
		NovelText string `json:"novel_text"`
		Title     string `json:"title"`
	}

	if err := json.Unmarshal([]byte(task.Input), &input); err != nil {
		return nil, fmt.Errorf("解析任务输入失败: %w", err)
	}

	// 调用剧本服务
	req := &script.AdaptScriptRequest{
		ProjectID: task.Workflow.ProjectID,
		Title:     input.Title,
		NovelText: input.NovelText,
	}

	result, err := e.scriptService.AdaptNovelToScript(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeCharacterGen 执行角色生成
func (e *WorkflowEngine) executeCharacterGen(ctx context.Context, task *entity.WorkflowTask) (interface{}, error) {
	var input struct {
		ScriptContent string `json:"script_content"`
	}

	if err := json.Unmarshal([]byte(task.Input), &input); err != nil {
		return nil, fmt.Errorf("解析任务输入失败: %w", err)
	}

	req := &character.GenerateCharactersRequest{
		ProjectID:     task.Workflow.ProjectID,
		ScriptContent: input.ScriptContent,
	}

	result, err := e.characterService.GenerateCharacters(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeSceneGen 执行场景生成
func (e *WorkflowEngine) executeSceneGen(ctx context.Context, task *entity.WorkflowTask) (interface{}, error) {
	var input struct {
		ScriptContent string `json:"script_content"`
	}

	if err := json.Unmarshal([]byte(task.Input), &input); err != nil {
		return nil, fmt.Errorf("解析任务输入失败: %w", err)
	}

	req := &scene.GenerateScenesRequest{
		ProjectID:     task.Workflow.ProjectID,
		ScriptContent: input.ScriptContent,
	}

	result, err := e.sceneService.GenerateScenes(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeStoryboard 执行分镜生成
func (e *WorkflowEngine) executeStoryboard(ctx context.Context, task *entity.WorkflowTask) (interface{}, error) {
	var input struct {
		ScriptID      uuid.UUID `json:"script_id"`
		ScriptContent string    `json:"script_content"`
		Title         string    `json:"title"`
	}

	if err := json.Unmarshal([]byte(task.Input), &input); err != nil {
		return nil, fmt.Errorf("解析任务输入失败: %w", err)
	}

	req := &storyboard.GenerateStoryboardRequest{
		ProjectID:     task.Workflow.ProjectID,
		ScriptID:      input.ScriptID,
		Title:         input.Title,
		ScriptContent: input.ScriptContent,
	}

	result, err := e.storyboardService.GenerateStoryboard(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeVoiceGen 执行语音生成
func (e *WorkflowEngine) executeVoiceGen(ctx context.Context, task *entity.WorkflowTask) (interface{}, error) {
	var input struct {
		VoiceRequests []*voice.VoiceRequest `json:"voice_requests"`
	}

	if err := json.Unmarshal([]byte(task.Input), &input); err != nil {
		return nil, fmt.Errorf("解析任务输入失败: %w", err)
	}

	req := &voice.GenerateVoicesRequest{
		ProjectID:     task.Workflow.ProjectID,
		VoiceRequests: input.VoiceRequests,
	}

	result, err := e.voiceService.GenerateVoices(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeMusicGen 执行音乐生成
func (e *WorkflowEngine) executeMusicGen(ctx context.Context, task *entity.WorkflowTask) (interface{}, error) {
	var input struct {
		Prompt   string `json:"prompt"`
		Style    string `json:"style"`
		Mood     string `json:"mood"`
		Tempo    string `json:"tempo"`
		Duration int    `json:"duration"`
	}

	if err := json.Unmarshal([]byte(task.Input), &input); err != nil {
		return nil, fmt.Errorf("解析任务输入失败: %w", err)
	}

	req := &music.GenerateMusicRequest{
		ProjectID: task.Workflow.ProjectID,
		Prompt:    input.Prompt,
		Style:     input.Style,
		Mood:      input.Mood,
		Tempo:     input.Tempo,
		Duration:  input.Duration,
	}

	result, err := e.musicService.GenerateBackgroundMusic(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeVideoEdit 执行视频编辑
func (e *WorkflowEngine) executeVideoEdit(ctx context.Context, task *entity.WorkflowTask) (interface{}, error) {
	var input struct {
		Title             string                        `json:"title"`
		Description       string                        `json:"description"`
		Resolution        string                        `json:"resolution"`
		FrameRate         int                           `json:"frame_rate"`
		EstimatedDuration int                           `json:"estimated_duration"`
		StoryboardFrames  []*editing.FrameRequest       `json:"storyboard_frames"`
		VoiceAudios       []*editing.AudioRequest       `json:"voice_audios"`
		BackgroundMusic   []*editing.MusicRequest       `json:"background_music"`
		Effects           []*editing.EffectRequest      `json:"effects"`
	}

	if err := json.Unmarshal([]byte(task.Input), &input); err != nil {
		return nil, fmt.Errorf("解析任务输入失败: %w", err)
	}

	req := &editing.ComposeVideoRequest{
		ProjectID:         task.Workflow.ProjectID,
		UserID:            task.Workflow.UserID,
		Title:             input.Title,
		Description:       input.Description,
		Resolution:        input.Resolution,
		FrameRate:         input.FrameRate,
		EstimatedDuration: input.EstimatedDuration,
		StoryboardFrames:  input.StoryboardFrames,
		VoiceAudios:       input.VoiceAudios,
		BackgroundMusic:   input.BackgroundMusic,
		Effects:           input.Effects,
	}

	result, err := e.editingService.ComposeVideo(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executePublish 执行发布
func (e *WorkflowEngine) executePublish(ctx context.Context, task *entity.WorkflowTask) (interface{}, error) {
	var input struct {
		VideoTheme              string                           `json:"video_theme"`
		TargetAudience          string                           `json:"target_audience"`
		ContentType             string                           `json:"content_type"`
		Platforms               []*publishing.PlatformConfig     `json:"platforms"`
		AutoGenerateContent     bool                             `json:"auto_generate_content"`
		AutoGenerateTags        bool                             `json:"auto_generate_tags"`
		AutoGenerateThumbnail   bool                             `json:"auto_generate_thumbnail"`
		AutoGenerateSocialContent bool                           `json:"auto_generate_social_content"`
	}

	if err := json.Unmarshal([]byte(task.Input), &input); err != nil {
		return nil, fmt.Errorf("解析任务输入失败: %w", err)
	}

	req := &publishing.PublishVideoRequest{
		ProjectID:               task.Workflow.ProjectID,
		VideoTheme:              input.VideoTheme,
		TargetAudience:          input.TargetAudience,
		ContentType:             input.ContentType,
		Platforms:               input.Platforms,
		AutoGenerateContent:     input.AutoGenerateContent,
		AutoGenerateTags:        input.AutoGenerateTags,
		AutoGenerateThumbnail:   input.AutoGenerateThumbnail,
		AutoGenerateSocialContent: input.AutoGenerateSocialContent,
	}

	result, err := e.publishingService.PublishVideo(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// markWorkflowFailed 标记工作流失败
func (e *WorkflowEngine) markWorkflowFailed(ctx context.Context, workflowID uuid.UUID, err error) {
	workflow, getErr := e.workflowRepo.GetByID(ctx, workflowID)
	if getErr != nil {
		log.Printf("[WorkflowEngine] 获取工作流失败: %v", getErr)
		return
	}

	workflow.Status = entity.WorkflowStatusFailed
	if updateErr := e.workflowRepo.Update(ctx, workflow); updateErr != nil {
		log.Printf("[WorkflowEngine] 更新工作流状态失败: %v", updateErr)
	}

	log.Printf("[WorkflowEngine] 工作流标记为失败: %s, error: %v", workflowID, err)
}

// markWorkflowCompleted 标记工作流完成
func (e *WorkflowEngine) markWorkflowCompleted(ctx context.Context, workflowID uuid.UUID) {
	workflow, err := e.workflowRepo.GetByID(ctx, workflowID)
	if err != nil {
		log.Printf("[WorkflowEngine] 获取工作流失败: %v", err)
		return
	}

	now := time.Now()
	workflow.Status = entity.WorkflowStatusCompleted
	workflow.Progress = 100
	workflow.CompletedAt = &now

	if err := e.workflowRepo.Update(ctx, workflow); err != nil {
		log.Printf("[WorkflowEngine] 更新工作流状态失败: %v", err)
	}

	log.Printf("[WorkflowEngine] 工作流标记为完成: %s", workflowID)
}

// updateWorkflowProgress 更新工作流进度
func (e *WorkflowEngine) updateWorkflowProgress(ctx context.Context, workflowID uuid.UUID) {
	// 获取所有任务
	tasks, err := e.taskRepo.GetByWorkflowID(ctx, workflowID)
	if err != nil {
		log.Printf("[WorkflowEngine] 获取工作流任务失败: %v", err)
		return
	}

	// 计算总进度
	totalProgress := 0
	for _, task := range tasks {
		totalProgress += task.Progress
	}

	averageProgress := totalProgress / len(tasks)

	// 更新工作流进度
	workflow, err := e.workflowRepo.GetByID(ctx, workflowID)
	if err != nil {
		log.Printf("[WorkflowEngine] 获取工作流失败: %v", err)
		return
	}

	workflow.Progress = averageProgress
	if err := e.workflowRepo.Update(ctx, workflow); err != nil {
		log.Printf("[WorkflowEngine] 更新工作流进度失败: %v", err)
	}
}

// GetWorkflow 获取工作流
func (e *WorkflowEngine) GetWorkflow(ctx context.Context, workflowID uuid.UUID) (*entity.Workflow, error) {
	return e.workflowRepo.GetByID(ctx, workflowID)
}

// GetUserWorkflows 获取用户工作流列表
func (e *WorkflowEngine) GetUserWorkflows(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entity.Workflow, int64, error) {
	return e.workflowRepo.GetByUserID(ctx, userID, offset, limit)
}

// GetWorkflowTasks 获取工作流任务列表
func (e *WorkflowEngine) GetWorkflowTasks(ctx context.Context, workflowID uuid.UUID) ([]*entity.WorkflowTask, error) {
	return e.taskRepo.GetByWorkflowID(ctx, workflowID)
}

// CancelWorkflow 取消工作流
func (e *WorkflowEngine) CancelWorkflow(ctx context.Context, workflowID uuid.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 取消正在运行的任务
	if cancelFunc, exists := e.runningTasks[workflowID]; exists {
		cancelFunc()
		delete(e.runningTasks, workflowID)
	}

	// 更新工作流状态
	workflow, err := e.workflowRepo.GetByID(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("获取工作流失败: %w", err)
	}

	workflow.Status = entity.WorkflowStatusCancelled
	return e.workflowRepo.Update(ctx, workflow)
}
