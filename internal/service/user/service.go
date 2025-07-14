package user

import (
	"context"
	"errors"

	"comic_video/internal/domain/dto"
	"comic_video/internal/repository/postgres"

	"github.com/google/uuid"
)

type Service struct {
	userRepo *postgres.UserRepository
}

func NewService(userRepo *postgres.UserRepository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

// GetByID 根据ID获取用户
func (s *Service) GetByID(ctx context.Context, userID string) (*dto.UserResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	user, err := s.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// List 获取用户列表
func (s *Service) List(ctx context.Context, offset, limit int) ([]*dto.UserResponse, int64, error) {
	users, total, err := s.userRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	var responses []*dto.UserResponse
	for _, user := range users {
		responses = append(responses, &dto.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Nickname:  user.Nickname,
			Avatar:    user.Avatar,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	return responses, total, nil
}

// Update 更新用户
func (s *Service) Update(ctx context.Context, userID string, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	user, err := s.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	// 更新用户信息
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// Delete 删除用户
func (s *Service) Delete(ctx context.Context, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	return s.userRepo.Delete(ctx, userUUID)
} 