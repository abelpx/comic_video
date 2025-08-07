package user

import (
	"testing"
	"time"

	"comic_video/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockUserRepository 模拟用户仓库
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *entity.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*entity.User, error) {
	args := m.Called(id)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*entity.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*entity.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *entity.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(offset, limit int) ([]*entity.User, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]*entity.User), args.Error(1)
}

func TestUserService_Create(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	user := &entity.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Nickname: "Test User",
		Role:     "user",
		Status:   "active",
	}

	// 测试成功创建用户
	t.Run("successful creation", func(t *testing.T) {
		mockRepo.On("Create", user).Return(nil).Once()
		
		err := service.Create(user)
		
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// 测试创建用户失败
	t.Run("creation failure", func(t *testing.T) {
		mockRepo.On("Create", user).Return(gorm.ErrDuplicatedKey).Once()
		
		err := service.Create(user)
		
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrDuplicatedKey, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	userID := uuid.New()
	expectedUser := &entity.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		Nickname: "Test User",
		Role:     "user",
		Status:   "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 测试成功获取用户
	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.On("GetByID", userID).Return(expectedUser, nil).Once()
		
		user, err := service.GetByID(userID)
		
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	// 测试用户不存在
	t.Run("user not found", func(t *testing.T) {
		mockRepo.On("GetByID", userID).Return((*entity.User)(nil), gorm.ErrRecordNotFound).Once()
		
		user, err := service.GetByID(userID)
		
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetByEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	email := "test@example.com"
	expectedUser := &entity.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    email,
		Nickname: "Test User",
		Role:     "user",
		Status:   "active",
	}

	// 测试成功获取用户
	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.On("GetByEmail", email).Return(expectedUser, nil).Once()
		
		user, err := service.GetByEmail(email)
		
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	// 测试用户不存在
	t.Run("user not found", func(t *testing.T) {
		mockRepo.On("GetByEmail", email).Return((*entity.User)(nil), gorm.ErrRecordNotFound).Once()
		
		user, err := service.GetByEmail(email)
		
		assert.Error(t, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_Update(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	user := &entity.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Nickname: "Updated User",
		Role:     "user",
		Status:   "active",
	}

	// 测试成功更新用户
	t.Run("successful update", func(t *testing.T) {
		mockRepo.On("Update", user).Return(nil).Once()
		
		err := service.Update(user)
		
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// 测试更新失败
	t.Run("update failure", func(t *testing.T) {
		mockRepo.On("Update", user).Return(gorm.ErrRecordNotFound).Once()
		
		err := service.Update(user)
		
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_Delete(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	userID := uuid.New()

	// 测试成功删除用户
	t.Run("successful deletion", func(t *testing.T) {
		mockRepo.On("Delete", userID).Return(nil).Once()
		
		err := service.Delete(userID)
		
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// 测试删除失败
	t.Run("deletion failure", func(t *testing.T) {
		mockRepo.On("Delete", userID).Return(gorm.ErrRecordNotFound).Once()
		
		err := service.Delete(userID)
		
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_List(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	expectedUsers := []*entity.User{
		{
			ID:       uuid.New(),
			Username: "user1",
			Email:    "user1@example.com",
			Nickname: "User 1",
			Role:     "user",
			Status:   "active",
		},
		{
			ID:       uuid.New(),
			Username: "user2",
			Email:    "user2@example.com",
			Nickname: "User 2",
			Role:     "user",
			Status:   "active",
		},
	}

	// 测试成功获取用户列表
	t.Run("successful list", func(t *testing.T) {
		mockRepo.On("List", 0, 10).Return(expectedUsers, nil).Once()
		
		users, err := service.List(0, 10)
		
		assert.NoError(t, err)
		assert.Equal(t, expectedUsers, users)
		assert.Len(t, users, 2)
		mockRepo.AssertExpectations(t)
	})

	// 测试空列表
	t.Run("empty list", func(t *testing.T) {
		mockRepo.On("List", 0, 10).Return([]*entity.User{}, nil).Once()
		
		users, err := service.List(0, 10)
		
		assert.NoError(t, err)
		assert.Empty(t, users)
		mockRepo.AssertExpectations(t)
	})
}

// 基准测试
func BenchmarkUserService_GetByID(b *testing.B) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)
	
	userID := uuid.New()
	user := &entity.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
	}
	
	mockRepo.On("GetByID", userID).Return(user, nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetByID(userID)
	}
}
