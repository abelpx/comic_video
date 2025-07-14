package project

import (
	"context"
	"errors"
	"encoding/json"
	"fmt"
	"time"

	"comic_video/internal/domain/dto"
	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/postgres"

	"github.com/google/uuid"
)

type Service struct {
	projectRepo       postgres.ProjectRepository
	projectShareRepo  postgres.ProjectShareRepository // 新增
}

func NewService(projectRepo postgres.ProjectRepository, projectShareRepo postgres.ProjectShareRepository) *Service {
	return &Service{
		projectRepo:      projectRepo,
		projectShareRepo: projectShareRepo,
	}
}

// Create 创建项目
func (s *Service) Create(ctx context.Context, userID string, req dto.CreateProjectRequest) (*dto.ProjectResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	configJSON := "{}"
	if req.Settings != nil {
		b, err := json.Marshal(req.Settings)
		if err == nil {
			configJSON = string(b)
		}
	}

	project := &entity.Project{
		Name:        req.Name,
		Description: req.Description,
		UserID:      userUUID,
		Status:      "draft",
		Config:      configJSON,
	}

	err = s.projectRepo.Create(ctx, project)
	if err != nil {
		return nil, err
	}

	var settings map[string]interface{}
	_ = json.Unmarshal([]byte(project.Config), &settings)

	return &dto.ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		UserID:      project.UserID,
		Status:      project.Status,
		Settings:    settings,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}, nil
}

// GetByID 根据ID获取项目
func (s *Service) GetByID(ctx context.Context, projectID string, userID string) (*dto.ProjectResponse, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("无效的项目ID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	project, err := s.projectRepo.GetByID(ctx, projectUUID)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if project.UserID != userUUID {
		return nil, errors.New("无权限访问此项目")
	}

	var settings map[string]interface{}
	_ = json.Unmarshal([]byte(project.Config), &settings)

	return &dto.ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		UserID:      project.UserID,
		Status:      project.Status,
		Settings:    settings,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}, nil
}

// List 获取用户的项目列表
func (s *Service) List(ctx context.Context, userID string, offset, limit int) ([]*dto.ProjectResponse, int64, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, errors.New("无效的用户ID")
	}

	projects, total, err := s.projectRepo.ListByUserID(ctx, userUUID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	var responses []*dto.ProjectResponse
	for _, project := range projects {
		var settings map[string]interface{}
		_ = json.Unmarshal([]byte(project.Config), &settings)
		responses = append(responses, &dto.ProjectResponse{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			UserID:      project.UserID,
			Status:      project.Status,
			Settings:    settings,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
		})
	}

	return responses, total, nil
}

// Update 更新项目
func (s *Service) Update(ctx context.Context, projectID string, userID string, req dto.UpdateProjectRequest) (*dto.ProjectResponse, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("无效的项目ID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	project, err := s.projectRepo.GetByID(ctx, projectUUID)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if project.UserID != userUUID {
		return nil, errors.New("无权限修改此项目")
	}

	// 更新项目信息
	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.Settings != nil {
		b, err := json.Marshal(req.Settings)
		if err == nil {
			project.Config = string(b)
		}
	}

	err = s.projectRepo.Update(ctx, project)
	if err != nil {
		return nil, err
	}

	var settings map[string]interface{}
	_ = json.Unmarshal([]byte(project.Config), &settings)

	return &dto.ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		UserID:      project.UserID,
		Status:      project.Status,
		Settings:    settings,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}, nil
}

// Delete 删除项目
func (s *Service) Delete(ctx context.Context, projectID string, userID string) error {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("无效的项目ID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	project, err := s.projectRepo.GetByID(ctx, projectUUID)
	if err != nil {
		return err
	}

	// 检查权限
	if project.UserID != userUUID {
		return errors.New("无权限删除此项目")
	}

	return s.projectRepo.Delete(ctx, projectUUID)
}

// Share 分享项目
func (s *Service) Share(ctx context.Context, projectID string, userID string, req dto.ShareProjectRequest) (*dto.ShareResponse, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("无效的项目ID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	project, err := s.projectRepo.GetByID(ctx, projectUUID)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if project.UserID != userUUID {
		return nil, errors.New("无权限分享此项目")
	}

	// 生成分享token
	shareToken := uuid.New().String()
	var passwordHash string
	if req.Password != "" {
		// 简单hash，可替换为更安全的hash
		passwordHash = fmt.Sprintf("%x", uuid.NewSHA1(uuid.NameSpaceOID, []byte(req.Password)))
	}
	share := &entity.ProjectShare{
		ProjectID: projectUUID,
		Token:     shareToken,
		Password:  passwordHash,
		ExpiresAt: req.ExpiresAt,
		CreatedBy: userUUID,
		Status:    "active",
	}
	if err := s.projectShareRepo.Create(ctx, share); err != nil {
		return nil, err
	}
	shareURL := "/share/" + shareToken
	return &dto.ShareResponse{
		ShareURL:  shareURL,
		Token:     shareToken,
		ExpiresAt: req.ExpiresAt,
	}, nil
}

// CheckShareToken 校验分享token和密码
func (s *Service) CheckShareToken(ctx context.Context, token string, password string) (*entity.ProjectShare, error) {
	share, err := s.projectShareRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, errors.New("分享不存在或已失效")
	}
	if share.Status != "active" {
		return nil, errors.New("分享已失效")
	}
	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		_ = s.projectShareRepo.ExpireByID(ctx, share.ID)
		return nil, errors.New("分享已过期")
	}
	if share.Password != "" {
		inputHash := fmt.Sprintf("%x", uuid.NewSHA1(uuid.NameSpaceOID, []byte(password)))
		if inputHash != share.Password {
			return nil, errors.New("密码错误")
		}
	}
	return share, nil
}

// ExpireShare 失效分享
func (s *Service) ExpireShare(ctx context.Context, shareID uuid.UUID) error {
	return s.projectShareRepo.ExpireByID(ctx, shareID)
} 