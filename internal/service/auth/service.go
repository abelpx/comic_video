package auth

import (
	"context"
	"errors"
	"time"

	"comic_video/internal/config"
	"comic_video/internal/domain/dto"
	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/postgres"
	"comic_video/internal/repository/redis"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userRepo *postgres.UserRepository
	redis    *redis.Client
	config   *config.JWTConfig
}

func NewService(userRepo *postgres.UserRepository, redis *redis.Client, jwtConfig *config.JWTConfig) *Service {
	return &Service{
		userRepo: userRepo,
		redis:    redis,
		config:   jwtConfig,
	}
}

// Register 用户注册
func (s *Service) Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error) {
	// 检查用户名是否已存在
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	existingUser, err = s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("邮箱已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &entity.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Role:     "user",
		Status:   "active",
	}

	err = s.userRepo.Create(ctx, user)
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

// Login 用户登录
func (s *Service) Login(ctx context.Context, req dto.LoginRequest) (string, *dto.UserResponse, error) {
	// 获取用户
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != "active" {
		return "", nil, errors.New("账户已被禁用")
	}

	// 生成JWT令牌
	token, err := s.generateToken(user.ID.String())
	if err != nil {
		return "", nil, err
	}

	// 将令牌存储到Redis
	err = s.redis.Set(ctx, "token:"+user.ID.String(), token, time.Duration(s.config.Expire)*time.Second)
	if err != nil {
		return "", nil, err
	}

	return token, &dto.UserResponse{
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

// Logout 用户登出
func (s *Service) Logout(ctx context.Context, userID string) error {
	// 从Redis中删除令牌
	return s.redis.Del(ctx, "token:"+userID)
}

// GetProfile 获取用户信息
func (s *Service) GetProfile(ctx context.Context, userID string) (*dto.UserResponse, error) {
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

// UpdateProfile 更新用户信息
func (s *Service) UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) (*dto.UserResponse, error) {
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

// ValidateToken 验证令牌
func (s *Service) ValidateToken(ctx context.Context, tokenString string) (string, error) {
	// 解析JWT令牌
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["user_id"].(string)
		
		// 检查令牌是否在Redis中
		exists, err := s.redis.Exists(ctx, "token:"+userID)
		if err != nil || !exists {
			return "", errors.New("令牌已失效")
		}

		return userID, nil
	}

	return "", errors.New("无效的令牌")
}

// generateToken 生成JWT令牌
func (s *Service) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(s.config.Expire) * time.Second).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.SecretKey))
} 