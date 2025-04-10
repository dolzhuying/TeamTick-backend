package service

import (
	"TeamTickBackend/dal/models"
	apperrors "TeamTickBackend/pkg/errors"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mock 实现 ---

// Mock UserDAO
type userMockUserDAO struct {
	mock.Mock
}

func (m *userMockUserDAO) GetByID(ctx context.Context, id int, tx ...*gorm.DB) (*models.User, error) {
	args := m.Called(ctx, id, tx)
	userArg := args.Get(0)
	if userArg == nil {
		return nil, args.Error(1)
	}
	return userArg.(*models.User), args.Error(1)
}

func (m *userMockUserDAO) GetByUsername(ctx context.Context, username string, tx ...*gorm.DB) (*models.User, error) {
	args := m.Called(ctx, username, tx)
	userArg := args.Get(0)
	if userArg == nil {
		return nil, args.Error(1)
	}
	return userArg.(*models.User), args.Error(1)
}

func (m *userMockUserDAO) Create(ctx context.Context, user *models.User, tx ...*gorm.DB) error {
	args := m.Called(ctx, user, tx)
	return args.Error(0)
}

// --- 测试准备 ---

func setupUserServiceTest() (*UserService, *mockUserDAO, *mockTransactionManager) {
	mockUserDao := new(mockUserDAO)
	mockTxManager := new(mockTransactionManager)
	
	userService := NewUserService(
		mockUserDao,
		mockTxManager,
	)
	
	return userService, mockUserDao, mockTxManager
}

// --- GetUserMe 测试 ---

func TestGetUserMe_Success(t *testing.T) {
	userService, mockUserDao, mockTxManager := setupUserServiceTest()
	ctx := context.Background()
	userID := 1
	expectedUser := &models.User{
		UserID:   userID,
		Username: "testuser",
		// 其他字段根据实际需要设置
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockUserDao.On("GetByID", ctx, userID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedUser, nil)

	// 调用函数
	user, err := userService.GetUserMe(ctx, userID)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser, user)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockUserDao.AssertExpectations(t)
}

func TestGetUserMe_UserNotFound(t *testing.T) {
	userService, mockUserDao, mockTxManager := setupUserServiceTest()
	ctx := context.Background()
	userID := 999 // 不存在的ID

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockUserDao.On("GetByID", ctx, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	user, err := userService.GetUserMe(ctx, userID)

	// 断言
	assert.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrUserNotFound))
	assert.Nil(t, user)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockUserDao.AssertExpectations(t)
}
