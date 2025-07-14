package material

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"comic_video/internal/domain/dto"
	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/minio"
	"comic_video/internal/repository/postgres"

	"github.com/google/uuid"
)

type Service struct {
	repo  postgres.MaterialRepository
	minio minio.MinioClient
}

func NewService(repo postgres.MaterialRepository, minio minio.MinioClient) *Service {
	return &Service{
		repo:  repo,
		minio: minio,
	}
}

// Upload 上传素材
func (s *Service) Upload(ctx context.Context, userID string, req dto.UploadMaterialRequest, fileHeader *multipart.FileHeader) (*dto.MaterialResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	objectName := fmt.Sprintf("%s/%d_%s", userUUID.String(), time.Now().UnixNano(), fileHeader.Filename)
	url, err := s.minio.Upload(ctx, objectName, file, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	material := &entity.Material{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Type:        req.Type,
		FileName:    fileHeader.Filename,
		FilePath:    objectName,
		FileSize:    fileHeader.Size,
		Format:      getFileExt(fileHeader.Filename),
		Tags:        req.Tags,
		IsPublic:    req.IsPublic,
		IsPremium:   req.IsPremium,
		Status:      "active",
	}

	err = s.repo.Create(ctx, material)
	if err != nil {
		// 回滚MinIO
		_ = s.minio.Delete(ctx, objectName)
		return nil, err
	}

	return toMaterialResponse(material, url), nil
}

// GetByID 获取素材详情
func (s *Service) GetByID(ctx context.Context, id string) (*dto.MaterialResponse, error) {
	materialUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("无效的素材ID")
	}
	material, err := s.repo.GetByID(ctx, materialUUID)
	if err != nil {
		return nil, err
	}
	url := s.minio.GetURL(material.FilePath)
	return toMaterialResponse(material, url), nil
}

// List 获取素材列表
func (s *Service) List(ctx context.Context, req dto.ListMaterialsRequest) ([]*dto.MaterialResponse, int64, error) {
	offset := (req.Page - 1) * req.PageSize
	filter := make(map[string]interface{})
	if req.Category != "" {
		filter["category"] = req.Category
	}
	if req.Type != "" {
		filter["type"] = req.Type
	}
	if req.IsPublic != nil {
		filter["is_public"] = *req.IsPublic
	}
	if req.IsPremium != nil {
		filter["is_premium"] = *req.IsPremium
	}
	// 关键词模糊搜索可后续扩展

	materials, total, err := s.repo.List(ctx, offset, req.PageSize, filter)
	if err != nil {
		return nil, 0, err
	}

	var responses []*dto.MaterialResponse
	for _, m := range materials {
		url := s.minio.GetURL(m.FilePath)
		responses = append(responses, toMaterialResponse(m, url))
	}
	return responses, total, nil
}

// Update 更新素材
func (s *Service) Update(ctx context.Context, id string, req dto.UpdateMaterialRequest) (*dto.MaterialResponse, error) {
	materialUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("无效的素材ID")
	}
	material, err := s.repo.GetByID(ctx, materialUUID)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		material.Name = req.Name
	}
	if req.Description != "" {
		material.Description = req.Description
	}
	if req.IsPublic != nil {
		material.IsPublic = *req.IsPublic
	}
	if req.IsPremium != nil {
		material.IsPremium = *req.IsPremium
	}
	if req.Tags != "" {
		material.Tags = req.Tags
	}
	err = s.repo.Update(ctx, material)
	if err != nil {
		return nil, err
	}
	url := s.minio.GetURL(material.FilePath)
	return toMaterialResponse(material, url), nil
}

// Delete 删除素材
func (s *Service) Delete(ctx context.Context, id string) error {
	materialUUID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("无效的素材ID")
	}
	material, err := s.repo.GetByID(ctx, materialUUID)
	if err != nil {
		return err
	}
	// 先删MinIO
	err = s.minio.Delete(ctx, material.FilePath)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, materialUUID)
}

// 工具函数：转为响应结构体
func toMaterialResponse(m *entity.Material, url string) *dto.MaterialResponse {
	return &dto.MaterialResponse{
		ID:            m.ID,
		Name:          m.Name,
		Description:   m.Description,
		Category:      m.Category,
		Type:          m.Type,
		FileName:      m.FileName,
		FilePath:      url,
		FileSize:      m.FileSize,
		Duration:      m.Duration,
		Width:         m.Width,
		Height:        m.Height,
		Format:        m.Format,
		Thumbnail:     m.Thumbnail,
		Tags:          m.Tags,
		IsPublic:      m.IsPublic,
		IsPremium:     m.IsPremium,
		DownloadCount: m.DownloadCount,
		Rating:        m.Rating,
		Status:        m.Status,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// 获取文件扩展名
func getFileExt(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
} 