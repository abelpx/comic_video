package render

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"comic_video/internal/domain/dto"
	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/minio"
	"comic_video/internal/repository/postgres"
	"encoding/json"
	"io"
	"net/http"
)

// Service 渲染服务接口
type Service interface {
	CreateRender(ctx context.Context, userID uuid.UUID, req *dto.CreateRenderRequest) (*dto.RenderResponse, error)
	GetRender(ctx context.Context, userID, renderID uuid.UUID) (*dto.RenderResponse, error)
	ListRenders(ctx context.Context, userID uuid.UUID, req *dto.ListRendersRequest) (*dto.ListRendersResponse, error)
	DeleteRender(ctx context.Context, userID, renderID uuid.UUID) error
	GetRenderStatus(ctx context.Context, userID, renderID uuid.UUID) (*dto.RenderStatusResponse, error)
	DownloadRender(ctx context.Context, userID, renderID uuid.UUID) (*dto.DownloadRenderResponse, error)
	ProcessRender(ctx context.Context, renderID uuid.UUID) error
}

// Effect 定义特效/滤镜/转场等
// Type: filter/transition/effect，Name: 滤镜或特效名，Params: 相关参数
// 示例：{"type":"filter","name":"grayscale"}，{"type":"transition","name":"fade","duration":1}
type Effect struct {
	Type   string                 `json:"type"`   // filter/transition/effect
	Name   string                 `json:"name"`   // 滤镜/特效/转场名
	Params map[string]interface{} `json:"params"` // 额外参数，如{"duration":1}
}

// Clip 片段，支持特效/滤镜/转场
type Clip struct {
	MaterialID string   `json:"material_id"`
	Start      float64  `json:"start"`
	End        float64  `json:"end"`
	Effects    []Effect `json:"effects"` // 支持多个特效/滤镜/转场
}

// Track 轨道，支持多类型
type Track struct {
	Type  string `json:"type"` // video/audio/image
	Clips []Clip `json:"clips"`
}

// ProjectConfig 项目配置，支持多轨道/分辨率/帧率等
type ProjectConfig struct {
	Tracks     []Track `json:"tracks"`
	Resolution string  `json:"resolution"`
	FrameRate  int     `json:"frame_rate"`
	// 可扩展：全局特效等
}

// service 渲染服务实现
type service struct {
	renderRepo   postgres.RenderRepository
	projectRepo  postgres.ProjectRepository
	materialRepo postgres.MaterialRepository
	minioClient  minio.MinioClient
	outputDir    string
	queue        RenderQueue // 新增：渲染任务队列
}

// NewService 创建渲染服务实例
func NewService(
	renderRepo postgres.RenderRepository,
	projectRepo postgres.ProjectRepository,
	materialRepo postgres.MaterialRepository,
	minioClient minio.MinioClient,
	outputDir string,
	queue RenderQueue, // 新增参数
) Service {
	return &service{
		renderRepo:   renderRepo,
		projectRepo:  projectRepo,
		materialRepo: materialRepo,
		minioClient:  minioClient,
		outputDir:    outputDir,
		queue:        queue,
	}
}

// CreateRender 创建渲染任务
func (s *service) CreateRender(ctx context.Context, userID uuid.UUID, req *dto.CreateRenderRequest) (*dto.RenderResponse, error) {
	// 验证项目是否存在且属于当前用户
	project, err := s.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在: %w", err)
	}
	if project.UserID != userID {
		return nil, fmt.Errorf("无权访问此项目")
	}

	// 创建渲染任务
	render := &entity.Render{
		UserID:     userID,
		ProjectID:  req.ProjectID,
		Name:       req.Name,
		Status:     "pending",
		Progress:   0,
		Quality:    req.Quality,
		Format:     req.Format,
		Resolution: req.Resolution,
	}

	if err := s.renderRepo.Create(ctx, render); err != nil {
		return nil, fmt.Errorf("创建渲染任务失败: %w", err)
	}

	// 自动入队，异步处理
	if s.queue != nil {
		_ = s.queue.Enqueue(render.ID)
	}

	return s.entityToResponse(render), nil
}

// GetRender 获取渲染任务详情
func (s *service) GetRender(ctx context.Context, userID, renderID uuid.UUID) (*dto.RenderResponse, error) {
	render, err := s.renderRepo.GetByID(ctx, renderID)
	if err != nil {
		return nil, fmt.Errorf("渲染任务不存在: %w", err)
	}

	if render.UserID != userID {
		return nil, fmt.Errorf("无权访问此渲染任务")
	}

	return s.entityToResponse(render), nil
}

// ListRenders 获取渲染任务列表
func (s *service) ListRenders(ctx context.Context, userID uuid.UUID, req *dto.ListRendersRequest) (*dto.ListRendersResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	var renders []*entity.Render
	var total int64
	var err error

	if req.ProjectID != "" {
		projectID, err := uuid.Parse(req.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("无效的项目ID: %w", err)
		}
		renders, total, err = s.renderRepo.GetByProjectID(ctx, projectID, req.Page, req.PageSize)
	} else {
		renders, total, err = s.renderRepo.GetByUserID(ctx, userID, req.Page, req.PageSize)
	}

	if err != nil {
		return nil, fmt.Errorf("获取渲染任务列表失败: %w", err)
	}

	// 过滤状态
	if req.Status != "" {
		filteredRenders := make([]*entity.Render, 0)
		for _, render := range renders {
			if render.Status == req.Status {
				filteredRenders = append(filteredRenders, render)
			}
		}
		renders = filteredRenders
	}

	responses := make([]*dto.RenderResponse, len(renders))
	for i, render := range renders {
		responses[i] = s.entityToResponse(render)
	}

	return &dto.ListRendersResponse{
		Renders:  responses,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// DeleteRender 删除渲染任务
func (s *service) DeleteRender(ctx context.Context, userID, renderID uuid.UUID) error {
	render, err := s.renderRepo.GetByID(ctx, renderID)
	if err != nil {
		return fmt.Errorf("渲染任务不存在: %w", err)
	}

	if render.UserID != userID {
		return fmt.Errorf("无权删除此渲染任务")
	}

	// 如果渲染已完成，删除输出文件
	if render.Status == "completed" && render.OutputPath != "" {
		if err := s.minioClient.DeleteObject(ctx, "renders", render.OutputPath); err != nil {
			fmt.Printf("删除输出文件失败: %v\n", err)
		}
	}

	return s.renderRepo.Delete(ctx, renderID)
}

// GetRenderStatus 获取渲染状态
func (s *service) GetRenderStatus(ctx context.Context, userID, renderID uuid.UUID) (*dto.RenderStatusResponse, error) {
	render, err := s.renderRepo.GetByID(ctx, renderID)
	if err != nil {
		return nil, fmt.Errorf("渲染任务不存在: %w", err)
	}

	if render.UserID != userID {
		return nil, fmt.Errorf("无权访问此渲染任务")
	}

	return &dto.RenderStatusResponse{
		Status:   render.Status,
		Progress: render.Progress,
		Error:    render.Error,
	}, nil
}

// DownloadRender 获取渲染结果下载链接
func (s *service) DownloadRender(ctx context.Context, userID, renderID uuid.UUID) (*dto.DownloadRenderResponse, error) {
	render, err := s.renderRepo.GetByID(ctx, renderID)
	if err != nil {
		return nil, fmt.Errorf("渲染任务不存在: %w", err)
	}

	if render.UserID != userID {
		return nil, fmt.Errorf("无权访问此渲染任务")
	}

	if render.Status != "completed" {
		return nil, fmt.Errorf("渲染任务尚未完成")
	}

	if render.OutputPath == "" {
		return nil, fmt.Errorf("输出文件不存在")
	}

	// 生成预签名下载URL
	url, err := s.minioClient.PresignedGetObject(ctx, "renders", render.OutputPath, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("生成下载链接失败: %w", err)
	}

	return &dto.DownloadRenderResponse{
		URL: url.String(),
	}, nil
}

// ProcessRender 处理渲染任务（同步方式）
func (s *service) ProcessRender(ctx context.Context, renderID uuid.UUID) error {
	render, err := s.renderRepo.GetByID(ctx, renderID)
	if err != nil {
		return fmt.Errorf("获取渲染任务失败: %w", err)
	}

	// 更新状态为处理中
	if err := s.renderRepo.UpdateStatus(ctx, renderID, "processing", 0, ""); err != nil {
		return fmt.Errorf("更新渲染状态失败: %w", err)
	}

	// 获取项目信息
	project, err := s.projectRepo.GetByID(ctx, render.ProjectID)
	if err != nil {
		return s.handleRenderError(ctx, renderID, fmt.Sprintf("获取项目信息失败: %v", err))
	}

	// 创建临时工作目录
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("render_%s_*", renderID.String()))
	if err != nil {
		return s.handleRenderError(ctx, renderID, fmt.Sprintf("创建临时目录失败: %v", err))
	}
	defer os.RemoveAll(tempDir)

	// 下载项目素材
	if err := s.downloadProjectMaterials(ctx, project, tempDir); err != nil {
		return s.handleRenderError(ctx, renderID, fmt.Sprintf("下载项目素材失败: %v", err))
	}

	// 更新进度
	if err := s.renderRepo.UpdateStatus(ctx, renderID, "processing", 30, ""); err != nil {
		return fmt.Errorf("更新渲染进度失败: %w", err)
	}

	// 生成FFmpeg命令
	outputFile := filepath.Join(tempDir, fmt.Sprintf("output.%s", s.getFileExtension(render.Format)))
	ffmpegCmd, err := s.buildFFmpegCommand(project, tempDir, outputFile, render)
	if err != nil {
		return s.handleRenderError(ctx, renderID, fmt.Sprintf("构建FFmpeg命令失败: %v", err))
	}

	// 更新进度
	if err := s.renderRepo.UpdateStatus(ctx, renderID, "processing", 50, ""); err != nil {
		return fmt.Errorf("更新渲染进度失败: %w", err)
	}

	// 执行FFmpeg命令
	if err := s.executeFFmpeg(ffmpegCmd); err != nil {
		return s.handleRenderError(ctx, renderID, fmt.Sprintf("FFmpeg执行失败: %v", err))
	}

	// 更新进度
	if err := s.renderRepo.UpdateStatus(ctx, renderID, "processing", 80, ""); err != nil {
		return fmt.Errorf("更新渲染进度失败: %w", err)
	}

	// 上传输出文件到MinIO
	outputPath := fmt.Sprintf("%s/%s.%s", render.UserID.String(), renderID.String(), s.getFileExtension(render.Format))
	if err := s.uploadOutputFile(ctx, outputFile, outputPath); err != nil {
		return s.handleRenderError(ctx, renderID, fmt.Sprintf("上传输出文件失败: %v", err))
	}

	// 获取文件信息
	fileInfo, err := os.Stat(outputFile)
	if err != nil {
		return s.handleRenderError(ctx, renderID, fmt.Sprintf("获取文件信息失败: %v", err))
	}

	// 获取视频时长
	duration, err := s.getVideoDuration(outputFile)
	if err != nil {
		duration = 0 // 如果获取失败，设为0
	}

	// 更新渲染任务为完成状态
	render.OutputPath = outputPath
	render.OutputSize = fileInfo.Size()
	render.Duration = duration
	render.Status = "completed"
	render.Progress = 100

	if err := s.renderRepo.Update(ctx, render); err != nil {
		return fmt.Errorf("更新渲染任务失败: %w", err)
	}

	return nil
}

// downloadProjectMaterials 下载项目所有用到的素材到本地
func (s *service) downloadProjectMaterials(ctx context.Context, project *entity.Project, tempDir string) error {
	var config ProjectConfig
	if err := json.Unmarshal([]byte(project.Config), &config); err != nil {
		return err
	}
	materialsDir := filepath.Join(tempDir, "materials")
	if err := os.MkdirAll(materialsDir, 0755); err != nil {
		return err
	}
	materialSet := make(map[string]struct{})
	for _, track := range config.Tracks {
		for _, clip := range track.Clips {
			materialSet[clip.MaterialID] = struct{}{}
		}
	}
	for materialID := range materialSet {
		id, err := uuid.Parse(materialID)
		if err != nil {
			return err
		}
		material, err := s.materialRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		localPath := filepath.Join(materialsDir, material.ID.String()+filepath.Ext(material.FileName))
		if err := s.downloadFromMinio(ctx, material.FilePath, localPath); err != nil {
			return err
		}
	}
	return nil
}

// downloadFromMinio 下载单个素材
func (s *service) downloadFromMinio(ctx context.Context, objectPath, localPath string) error {
	url, err := s.minioClient.PresignedURL(ctx, objectPath, 10*time.Minute)
	if err != nil {
		return err
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

// buildClipFilter 根据 effects 生成 FFmpeg filter 字符串
func buildClipFilter(effects []Effect, clipDuration float64) string {
	var filters []string
	for _, eff := range effects {
		switch eff.Type {
		case "filter":
			switch eff.Name {
			case "grayscale":
				filters = append(filters, "hue=s=0")
			case "boxblur":
				filters = append(filters, "boxblur=2:1")
			case "negate":
				filters = append(filters, "negate")
			}
		case "transition":
			// 仅支持 clip 首尾淡入淡出
			if eff.Name == "fade" {
				dur := 1.0
				if v, ok := eff.Params["duration"]; ok {
					if f, ok := v.(float64); ok {
						dur = f
					}
				}
				if pos, ok := eff.Params["position"]; ok && pos == "in" {
					filters = append(filters, "fade=t=in:st=0:d="+formatFloat(dur))
				} else if pos == "out" {
					filters = append(filters, "fade=t=out:st="+formatFloat(clipDuration-dur)+":d="+formatFloat(dur))
				}
			}
		}
	}
	return strings.Join(filters, ",")
}

// formatFloat 保证小数点格式
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// buildFFmpegCommand 根据项目轨道智能生成FFmpeg命令，支持clip.effects
func (s *service) buildFFmpegCommand(project *entity.Project, tempDir, outputFile string, render *entity.Render) (*exec.Cmd, error) {
	var config ProjectConfig
	if err := json.Unmarshal([]byte(project.Config), &config); err != nil {
		return nil, err
	}
	materialsDir := filepath.Join(tempDir, "materials")

	// 仅支持单视频轨道，逐clip处理
	var inputArgs []string
	var filterCmds []string
	var concatInputs []string
	clipIdx := 0
	for _, track := range config.Tracks {
		if track.Type != "video" && track.Type != "image" {
			continue
		}
		for _, clip := range track.Clips {
			inputFile := filepath.Join(materialsDir, clip.MaterialID+".mp4")
			if track.Type == "image" {
				inputFile = filepath.Join(materialsDir, clip.MaterialID+".jpg")
			}
			inputArgs = append(inputArgs, "-i", inputFile)
			filter := buildClipFilter(clip.Effects, clip.End-clip.Start)
			labelIn := "[" + strconv.Itoa(clipIdx) + ":v]"
			labelOut := "[v" + strconv.Itoa(clipIdx) + "]"
			if filter != "" {
				filterCmds = append(filterCmds, labelIn+filter+labelOut)
				concatInputs = append(concatInputs, labelOut)
			} else {
				concatInputs = append(concatInputs, labelIn)
			}
			clipIdx++
		}
	}
	// 拼接所有片段
	var args []string
	args = append(args, inputArgs...)
	if len(filterCmds) > 0 {
		filterStr := strings.Join(filterCmds, ";")
		if len(concatInputs) > 1 {
			filterStr += ";" + strings.Join(concatInputs, "") + "concat=n=" + strconv.Itoa(len(concatInputs)) + ":v=1:a=0[vout]"
			args = append(args, "-filter_complex", filterStr, "-map", "[vout]")
		} else {
			args = append(args, "-filter_complex", filterStr, "-map", concatInputs[0])
		}
	} else if len(concatInputs) > 1 {
		// 无滤镜，仅拼接
		filterStr := strings.Join(concatInputs, "") + "concat=n=" + strconv.Itoa(len(concatInputs)) + ":v=1:a=0[vout]"
		args = append(args, "-filter_complex", filterStr, "-map", "[vout]")
	}
	args = append(args, "-c:v", "libx264", "-preset", "fast", "-crf", "23", outputFile)
	cmd := exec.Command("ffmpeg", args...)
	cmd.Dir = tempDir
	return cmd, nil
}

// executeFFmpeg 执行FFmpeg命令
func (s *service) executeFFmpeg(cmd *exec.Cmd) error {
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg执行失败: %v, 输出: %s", err, string(output))
	}
	return nil
}

// uploadOutputFile 上传输出文件到MinIO
func (s *service) uploadOutputFile(ctx context.Context, localPath, objectPath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return s.minioClient.PutObject(ctx, "renders", objectPath, file, "video/mp4")
}

// getVideoDuration 获取视频时长
func (s *service) getVideoDuration(filePath string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries", "format=duration", "-of", "csv=p=0", filePath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

// getFileExtension 根据格式获取文件扩展名
func (s *service) getFileExtension(format string) string {
	switch format {
	case "mp4":
		return "mp4"
	case "avi":
		return "avi"
	case "mov":
		return "mov"
	case "mkv":
		return "mkv"
	default:
		return "mp4"
	}
}

// handleRenderError 处理渲染错误
func (s *service) handleRenderError(ctx context.Context, renderID uuid.UUID, errorMsg string) error {
	if err := s.renderRepo.UpdateStatus(ctx, renderID, "failed", 0, errorMsg); err != nil {
		return fmt.Errorf("更新渲染错误状态失败: %w", err)
	}
	return fmt.Errorf(errorMsg)
}

// entityToResponse 将实体转换为响应DTO
func (s *service) entityToResponse(render *entity.Render) *dto.RenderResponse {
	return &dto.RenderResponse{
		ID:          render.ID,
		UserID:      render.UserID,
		ProjectID:   render.ProjectID,
		Name:        render.Name,
		Status:      render.Status,
		Progress:    render.Progress,
		OutputPath:  render.OutputPath,
		OutputSize:  render.OutputSize,
		Duration:    render.Duration,
		Resolution:  render.Resolution,
		Format:      render.Format,
		Quality:     render.Quality,
		Error:       render.Error,
		StartedAt:   render.StartedAt,
		CompletedAt: render.CompletedAt,
		CreatedAt:   render.CreatedAt,
		UpdatedAt:   render.UpdatedAt,
	}
} 