// tests/unit/user_service_test.go
// 用户服务单元测试示例 - 展示 Mock 使用

package unit

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// User 用户模型
type User struct {
	ID       int64
	Username string
	Email    string
}

// UserRepository 用户仓库接口
type UserRepository interface {
	FindByID(id int64) (*User, error)
	FindByEmail(email string) (*User, error)
	Create(user *User) error
	Update(user *User) error
}

// UserService 用户服务
type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(id int64) (*User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) CreateUser(username, email string) (*User, error) {
	// 检查邮箱是否已存在
	existing, err := s.repo.FindByEmail(email)
	if err == nil && existing != nil {
		return nil, errors.New("email already exists")
	}

	user := &User{
		Username: username,
		Email:    email,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// MockUserRepository Mock 实现
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByID(id int64) (*User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) Create(user *User) error {
	args := m.Called(user)
	// 模拟设置 ID
	user.ID = 1
	return args.Error(0)
}

func (m *MockUserRepository) Update(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

// 测试用例
func TestUserService_GetUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	expectedUser := &User{ID: 1, Username: "alice", Email: "alice@example.com"}
	mockRepo.On("FindByID", int64(1)).Return(expectedUser, nil)

	user, err := service.GetUser(1)

	require.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUser_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("FindByID", int64(999)).Return(nil, errors.New("user not found"))

	user, err := service.GetUser(999)

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	// 邮箱不存在
	mockRepo.On("FindByEmail", "bob@example.com").Return(nil, errors.New("not found"))
	// 创建成功
	mockRepo.On("Create", mock.AnythingOfType("*unit.User")).Return(nil)

	user, err := service.CreateUser("bob", "bob@example.com")

	require.NoError(t, err)
	assert.Equal(t, "bob", user.Username)
	assert.Equal(t, "bob@example.com", user.Email)
	assert.Equal(t, int64(1), user.ID)
	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_EmailExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	existingUser := &User{ID: 1, Email: "existing@example.com"}
	mockRepo.On("FindByEmail", "existing@example.com").Return(existingUser, nil)

	user, err := service.CreateUser("newuser", "existing@example.com")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertExpectations(t)
}

// 表驱动 Mock 测试
func TestUserService_GetUser_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		userID      int64
		mockReturn  *User
		mockError   error
		expectError bool
	}{
		{
			name:       "success",
			userID:     1,
			mockReturn: &User{ID: 1, Username: "test"},
			mockError:  nil,
		},
		{
			name:        "not found",
			userID:      999,
			mockReturn:  nil,
			mockError:   errors.New("not found"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			service := NewUserService(mockRepo)

			mockRepo.On("FindByID", tt.userID).Return(tt.mockReturn, tt.mockError)

			user, err := service.GetUser(tt.userID)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.mockReturn, user)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
