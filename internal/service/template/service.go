package template

import (
	"context"
	"encoding/json"
	"errors"

	"comic_video/internal/domain/dto"
	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/postgres"

	"github.com/google/uuid"
)

type Service struct {
	repo *postgres.TemplateRepository
}

func NewService(repo *postgres.TemplateRepository) *Service {
	return &Service{repo: repo}
}

// Create 创建模板
func (s *Service) Create(ctx context.Context, req dto.CreateTemplateRequest) (*dto.TemplateResponse, error) {
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, err
	}
	template := &entity.Template{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		Thumbnail:   req.Thumbnail,
		Preview:     req.Preview,
		Config:      string(configJSON),
		IsPublic:    req.IsPublic,
		IsPremium:   req.IsPremium,
		Status:      "active",
	}
	err = s.repo.Create(ctx, template)
	if err != nil {
		return nil, err
	}
	return toTemplateResponse(template), nil
}

// GetByID 获取模板详情
func (s *Service) GetByID(ctx context.Context, id string) (*dto.TemplateResponse, error) {
	templateUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("无效的模板ID")
	}
	template, err := s.repo.GetByID(ctx, templateUUID)
	if err != nil {
		return nil, err
	}
	return toTemplateResponse(template), nil
}

// List 获取模板列表
func (s *Service) List(ctx context.Context, req dto.ListTemplatesRequest) ([]*dto.TemplateResponse, int64, error) {
	offset := (req.Page - 1) * req.PageSize
	filter := make(map[string]interface{})
	if req.Category != "" {
		filter["category"] = req.Category
	}
	if req.IsPublic != nil {
		filter["is_public"] = *req.IsPublic
	}
	if req.IsPremium != nil {
		filter["is_premium"] = *req.IsPremium
	}
	// 关键词模糊搜索可后续扩展

	templates, total, err := s.repo.List(ctx, offset, req.PageSize, filter)
	if err != nil {
		return nil, 0, err
	}

	var responses []*dto.TemplateResponse
	for _, t := range templates {
		responses = append(responses, toTemplateResponse(t))
	}
	return responses, total, nil
}

// Update 更新模板
func (s *Service) Update(ctx context.Context, id string, req dto.UpdateTemplateRequest) (*dto.TemplateResponse, error) {
	templateUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("无效的模板ID")
	}
	template, err := s.repo.GetByID(ctx, templateUUID)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Description != "" {
		template.Description = req.Description
	}
	if req.Category != "" {
		template.Category = req.Category
	}
	if req.Tags != "" {
		template.Tags = req.Tags
	}
	if req.Thumbnail != "" {
		template.Thumbnail = req.Thumbnail
	}
	if req.Preview != "" {
		template.Preview = req.Preview
	}
	if req.Config != nil {
		b, err := json.Marshal(req.Config)
		if err == nil {
			template.Config = string(b)
		}
	}
	if req.IsPublic != nil {
		template.IsPublic = *req.IsPublic
	}
	if req.IsPremium != nil {
		template.IsPremium = *req.IsPremium
	}
	err = s.repo.Update(ctx, template)
	if err != nil {
		return nil, err
	}
	return toTemplateResponse(template), nil
}

// Delete 删除模板
func (s *Service) Delete(ctx context.Context, id string) error {
	templateUUID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("无效的模板ID")
	}
	return s.repo.Delete(ctx, templateUUID)
}

// Apply 应用模板（当前仅返回成功，后续可扩展）
func (s *Service) Apply(ctx context.Context, id string, req dto.ApplyTemplateRequest) (*dto.ApplyTemplateResponse, error) {
	// TODO: 实现模板应用到项目的逻辑
	return &dto.ApplyTemplateResponse{
		Success: true,
		Message: "模板应用成功",
	}, nil
}

// 工具函数：转为响应结构体
func toTemplateResponse(t *entity.Template) *dto.TemplateResponse {
	var config map[string]interface{}
	_ = json.Unmarshal([]byte(t.Config), &config)
	return &dto.TemplateResponse{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Category:    t.Category,
		Tags:        t.Tags,
		Thumbnail:   t.Thumbnail,
		Preview:     t.Preview,
		Config:      config,
		IsPublic:    t.IsPublic,
		IsPremium:   t.IsPremium,
		DownloadCount: t.DownloadCount,
		Rating:      t.Rating,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
} 