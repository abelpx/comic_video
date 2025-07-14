package video

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"comic_video/internal/domain/dto"
	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/minio"
	"comic_video/internal/repository/postgres"

	"github.com/google/uuid"
)

type Service struct {
	repo  *postgres.VideoRepository
	minio minio.MinioClient
}

func NewService(repo *postgres.VideoRepository, minio minio.MinioClient) *Service {
	return &Service{
		repo:  repo,
		minio: minio,
	}
}

// Upload 上传视频
func (s *Service) Upload(ctx context.Context, userID string, req dto.UploadVideoRequest, fileHeader *multipart.FileHeader) (*dto.VideoResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 生成唯一文件名
	ext := filepath.Ext(fileHeader.Filename)
	objectName := fmt.Sprintf("videos/%s/%d_%s%s", userUUID.String(), time.Now().UnixNano(), uuid.New().String()[:8], ext)
	
	// 上传到MinIO
	url, err := s.minio.Upload(ctx, objectName, file, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	// 创建视频记录
	video := &entity.Video{
		UserID:       userUUID,
		ProjectID:    req.ProjectID,
		FileName:     objectName,
		OriginalName: fileHeader.Filename,
		FilePath:     objectName,
		FileSize:     fileHeader.Size,
		Format:       strings.TrimPrefix(ext, "."),
		Type:         req.Type,
		Status:       "uploaded",
	}

	err = s.repo.Create(ctx, video)
	if err != nil {
		// 回滚MinIO
		_ = s.minio.Delete(ctx, objectName)
		return nil, err
	}

	return toVideoResponse(video, url), nil
}

// GetByID 获取视频详情
func (s *Service) GetByID(ctx context.Context, id string, userID string) (*dto.VideoResponse, error) {
	videoUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("无效的视频ID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	video, err := s.repo.GetByIDAndUser(ctx, videoUUID, userUUID)
	if err != nil {
		return nil, err
	}

	url := s.minio.GetURL(video.FilePath)
	return toVideoResponse(video, url), nil
}

// List 获取视频列表
func (s *Service) List(ctx context.Context, userID string, req dto.ListVideosRequest) ([]*dto.VideoResponse, int64, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, errors.New("无效的用户ID")
	}

	offset := (req.Page - 1) * req.PageSize
	filter := make(map[string]interface{})
	
	if req.Type != "" {
		filter["type"] = req.Type
	}
	if req.Status != "" {
		filter["status"] = req.Status
	}

	videos, total, err := s.repo.ListByUser(ctx, userUUID, offset, req.PageSize, filter)
	if err != nil {
		return nil, 0, err
	}

	var responses []*dto.VideoResponse
	for _, v := range videos {
		url := s.minio.GetURL(v.FilePath)
		responses = append(responses, toVideoResponse(v, url))
	}

	return responses, total, nil
}

// Update 更新视频
func (s *Service) Update(ctx context.Context, id string, userID string, req dto.UpdateVideoRequest) (*dto.VideoResponse, error) {
	videoUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("无效的视频ID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	video, err := s.repo.GetByIDAndUser(ctx, videoUUID, userUUID)
	if err != nil {
		return nil, err
	}

	if req.Status != "" {
		video.Status = req.Status
	}
	if req.Thumbnail != "" {
		video.Thumbnail = req.Thumbnail
	}

	err = s.repo.Update(ctx, video)
	if err != nil {
		return nil, err
	}

	url := s.minio.GetURL(video.FilePath)
	return toVideoResponse(video, url), nil
}

// Delete 删除视频
func (s *Service) Delete(ctx context.Context, id string, userID string) error {
	videoUUID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("无效的视频ID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	video, err := s.repo.GetByIDAndUser(ctx, videoUUID, userUUID)
	if err != nil {
		return err
	}

	// 先删除MinIO文件
	err = s.minio.Delete(ctx, video.FilePath)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, videoUUID)
}

// Process 处理视频（占位符，后续可扩展转码等功能）
func (s *Service) Process(ctx context.Context, id string, userID string) error {
	videoUUID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("无效的视频ID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	_, err = s.repo.GetByIDAndUser(ctx, videoUUID, userUUID)
	if err != nil {
		return err
	}

	// TODO: 实现视频处理逻辑（转码、生成缩略图等）
	return s.repo.UpdateStatus(ctx, videoUUID, "processing")
}

// GetStatus 获取视频处理状态
func (s *Service) GetStatus(ctx context.Context, id string, userID string) (string, error) {
	videoUUID, err := uuid.Parse(id)
	if err != nil {
		return "", errors.New("无效的视频ID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return "", errors.New("无效的用户ID")
	}

	video, err := s.repo.GetByIDAndUser(ctx, videoUUID, userUUID)
	if err != nil {
		return "", err
	}

	return video.Status, nil
}

// 工具函数：转为响应结构体
func toVideoResponse(v *entity.Video, url string) *dto.VideoResponse {
	return &dto.VideoResponse{
		ID:           v.ID,
		UserID:       v.UserID,
		ProjectID:    v.ProjectID,
		FileName:     v.FileName,
		OriginalName: v.OriginalName,
		FilePath:     url,
		FileSize:     v.FileSize,
		Duration:     v.Duration,
		Width:        v.Width,
		Height:       v.Height,
		Format:       v.Format,
		Codec:        v.Codec,
		Bitrate:      v.Bitrate,
		FPS:          v.FPS,
		Thumbnail:    v.Thumbnail,
		Status:       v.Status,
		Type:         v.Type,
		CreatedAt:    v.CreatedAt,
		UpdatedAt:    v.UpdatedAt,
	}
} 