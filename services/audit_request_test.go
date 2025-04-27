package service

import (
	"TeamTickBackend/dal/models"
	appErrors "TeamTickBackend/pkg/errors"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mock 实现 ---

// Mock CheckApplicationDAO
type mockCheckApplicationDAO struct {
	mock.Mock
}

func (m *mockCheckApplicationDAO) Create(ctx context.Context, application *models.CheckApplication, tx ...*gorm.DB) error {
	args := m.Called(ctx, application, tx)
	if args.Error(0) == nil {
		application.ID = 1
		application.Status = "pending"
		application.CreatedAt = time.Now()
		application.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *mockCheckApplicationDAO) GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.CheckApplication, error) {
	args := m.Called(ctx, groupID, tx)
	applicationsArg := args.Get(0)
	if applicationsArg == nil {
		return nil, args.Error(1)
	}
	return applicationsArg.([]*models.CheckApplication), args.Error(1)
}

func (m *mockCheckApplicationDAO) GetByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.CheckApplication, error) {
	args := m.Called(ctx, userID, tx)
	applicationsArg := args.Get(0)
	if applicationsArg == nil {
		return nil, args.Error(1)
	}
	return applicationsArg.([]*models.CheckApplication), args.Error(1)
}

func (m *mockCheckApplicationDAO) Update(ctx context.Context, status string, requestID int, tx ...*gorm.DB) error {
	args := m.Called(ctx, status, requestID, tx)
	return args.Error(0)
}

func (m *mockCheckApplicationDAO) GetByTaskIDAndUserID(ctx context.Context, taskID int, userID int, tx ...*gorm.DB) (*models.CheckApplication, error) {
	args := m.Called(ctx, taskID, userID, tx)
	applicationArg := args.Get(0)
	if applicationArg == nil {
		return nil, args.Error(1)
	}
	return applicationArg.(*models.CheckApplication), args.Error(1)
}

func (m *mockCheckApplicationDAO) GetByID(ctx context.Context, id int, tx ...*gorm.DB) (*models.CheckApplication, error) {
	args := m.Called(ctx, id, tx)
	applicationArg := args.Get(0)
	if applicationArg == nil {
		return nil, args.Error(1)
	}
	return applicationArg.(*models.CheckApplication), args.Error(1)
}

// 测试准备
func setupAuditRequestServiceTest() (*AuditRequestService, *mockCheckApplicationDAO, *mockTaskDAO, *mockTaskRecordDAO, *mockGroupDAO, *mockTransactionManager) {
	mockCheckApplicationDao := new(mockCheckApplicationDAO)
	mockTaskDao := new(mockTaskDAO)
	mockTaskRecordDao := new(mockTaskRecordDAO)
	mockGroupDao := new(mockGroupDAO)
	mockTxManager := new(mockTransactionManager)

	auditRequestService := NewAuditRequestService(
		mockTxManager,
		mockCheckApplicationDao,
		mockTaskRecordDao,
		mockTaskDao,
		mockGroupDao,
	)

	return auditRequestService, mockCheckApplicationDao, mockTaskDao, mockTaskRecordDao, mockGroupDao, mockTxManager
}

// --- GetAuditRequestListByUserID 测试 ---

func TestGetAuditRequestListByUserID_Success(t *testing.T) {
	auditRequestService, mockCheckApplicationDao, _, _, _, mockTxManager := setupAuditRequestServiceTest()
	ctx := context.Background()
	userID := 1

	expectedRequests := []*models.CheckApplication{
		{
			ID:            1,
			TaskID:        1,
			TaskName:      "任务1",
			UserID:        userID,
			Username:      "user1",
			Status:        "pending",
			Reason:        "网络问题",
			AdminID:       2,
			AdminUsername: "admin1",
			RequestAt:     time.Now(),
		},
		{
			ID:            2,
			TaskID:        2,
			TaskName:      "任务2",
			UserID:        userID,
			Username:      "user1",
			Status:        "approved",
			Reason:        "位置异常",
			AdminID:       2,
			AdminUsername: "admin1",
			RequestAt:     time.Now(),
		},
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockCheckApplicationDao.On("GetByUserID", ctx, userID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedRequests, nil)

	// 调用函数
	requests, err := auditRequestService.GetAuditRequestListByUserID(ctx, userID)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, requests)
	assert.Equal(t, expectedRequests, requests)
	assert.Len(t, requests, 2)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockCheckApplicationDao.AssertExpectations(t)
}

func TestGetAuditRequestListByUserID_NotFound(t *testing.T) {
	auditRequestService, mockCheckApplicationDao, _, _, _, mockTxManager := setupAuditRequestServiceTest()
	ctx := context.Background()
	userID := 999 // 不存在的用户ID

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockCheckApplicationDao.On("GetByUserID", ctx, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	requests, err := auditRequestService.GetAuditRequestListByUserID(ctx, userID)

	// 断言
	assert.Error(t, err)
	assert.Contains(t, err.Error(), appErrors.ErrAuditRequestNotFound.Error())
	assert.Nil(t, requests)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockCheckApplicationDao.AssertExpectations(t)
}

// --- GetAuditRequestByGroupID 测试 ---

func TestGetAuditRequestByGroupID_Success(t *testing.T) {
	auditRequestService, mockCheckApplicationDao, _, _, _, mockTxManager := setupAuditRequestServiceTest()
	ctx := context.Background()
	groupID := 1

	expectedRequests := []*models.CheckApplication{
		{
			ID:            1,
			TaskID:        1,
			TaskName:      "任务1",
			UserID:        1,
			Username:      "user1",
			Status:        "pending",
			Reason:        "网络问题",
			AdminID:       2,
			AdminUsername: "admin1",
			RequestAt:     time.Now(),
		},
		{
			ID:            2,
			TaskID:        1,
			TaskName:      "任务1",
			UserID:        3,
			Username:      "user3",
			Status:        "pending",
			Reason:        "位置异常",
			AdminID:       2,
			AdminUsername: "admin1",
			RequestAt:     time.Now(),
		},
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockCheckApplicationDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedRequests, nil)

	// 调用函数
	requests, err := auditRequestService.GetAuditRequestByGroupID(ctx, groupID)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, requests)
	assert.Equal(t, expectedRequests, requests)
	assert.Len(t, requests, 2)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockCheckApplicationDao.AssertExpectations(t)
}

func TestGetAuditRequestByGroupID_NotFound(t *testing.T) {
	auditRequestService, mockCheckApplicationDao, _, _, _, mockTxManager := setupAuditRequestServiceTest()
	ctx := context.Background()
	groupID := 999 // 不存在的组ID

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockCheckApplicationDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	requests, err := auditRequestService.GetAuditRequestByGroupID(ctx, groupID)

	// 断言
	assert.Error(t, err)
	assert.Contains(t, err.Error(), appErrors.ErrAuditRequestNotFound.Error())
	assert.Nil(t, requests)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockCheckApplicationDao.AssertExpectations(t)
}

// --- CreateAuditRequest 测试 ---

func TestCreateAuditRequest_Success(t *testing.T) {
	auditRequestService, mockCheckApplicationDao, mockTaskDao, _, mockGroupDao, mockTxManager := setupAuditRequestServiceTest()
	ctx := context.Background()

	// 测试数据
	taskID := 1
	userID := 1
	username := "user1"
	reason := "网络问题导致无法正常签到"
	image := "base64encodedimage"

	// 预期的任务和组
	task := &models.Task{
		TaskID:      taskID,
		TaskName:    "测试任务",
		Description: "这是一个测试任务",
		GroupID:     1,
	}

	group := &models.Group{
		GroupID:     1,
		GroupName:   "测试组",
		CreatorID:   2,
		CreatorName: "admin1",
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("GetByTaskID", ctx, taskID, mock.AnythingOfType("[]*gorm.DB")).Return(task, nil)
	mockCheckApplicationDao.On("GetByTaskIDAndUserID", ctx, taskID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)
	mockGroupDao.On("GetByGroupID", ctx, task.GroupID, mock.AnythingOfType("[]*gorm.DB")).Return(group, nil)
	mockCheckApplicationDao.On("Create", ctx, mock.AnythingOfType("*models.CheckApplication"), mock.AnythingOfType("[]*gorm.DB")).Return(nil).Run(func(args mock.Arguments) {
		application := args.Get(1).(*models.CheckApplication)
		application.ID = 1
		application.Status = "pending"
		application.CreatedAt = time.Now()
		application.UpdatedAt = time.Now()
	})

	// 调用函数
	createdRequest, err := auditRequestService.CreateAuditRequest(ctx, taskID, userID, username, reason, image)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, createdRequest)
	assert.Equal(t, taskID, createdRequest.TaskID)
	assert.Equal(t, task.TaskName, createdRequest.TaskName)
	assert.Equal(t, userID, createdRequest.UserID)
	assert.Equal(t, username, createdRequest.Username)
	assert.Equal(t, reason, createdRequest.Reason)
	assert.Equal(t, image, createdRequest.Image)
	assert.Equal(t, group.CreatorID, createdRequest.AdminID)
	assert.Equal(t, group.CreatorName, createdRequest.AdminUsername)
	assert.Equal(t, "pending", createdRequest.Status)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
	mockCheckApplicationDao.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
}

func TestCreateAuditRequest_TaskNotFound(t *testing.T) {
	auditRequestService, _, mockTaskDao, _, _, mockTxManager := setupAuditRequestServiceTest()
	ctx := context.Background()

	// 测试数据
	taskID := 999 // 不存在的任务ID
	userID := 1
	username := "user1"
	reason := "网络问题导致无法正常签到"
	image := "base64encodedimage"

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("GetByTaskID", ctx, taskID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	createdRequest, err := auditRequestService.CreateAuditRequest(ctx, taskID, userID, username, reason, image)

	// 断言
	assert.Error(t, err)
	assert.Contains(t, err.Error(), appErrors.ErrTaskNotFound.Error())
	assert.Nil(t, createdRequest)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
}

func TestCreateAuditRequest_RequestAlreadyExists(t *testing.T) {
	auditRequestService, mockCheckApplicationDao, mockTaskDao, _, _, mockTxManager := setupAuditRequestServiceTest()
	ctx := context.Background()

	// 测试数据
	taskID := 1
	userID := 1
	username := "user1"
	reason := "网络问题导致无法正常签到"
	image := "base64encodedimage"

	// 预期的任务和已存在的申请
	task := &models.Task{
		TaskID:      taskID,
		TaskName:    "测试任务",
		Description: "这是一个测试任务",
		GroupID:     1,
	}

	existingRequest := &models.CheckApplication{
		ID:        1,
		TaskID:    taskID,
		UserID:    userID,
		Status:    "pending",
		RequestAt: time.Now(),
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("GetByTaskID", ctx, taskID, mock.AnythingOfType("[]*gorm.DB")).Return(task, nil)
	mockCheckApplicationDao.On("GetByTaskIDAndUserID", ctx, taskID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(existingRequest, nil)

	// 调用函数
	createdRequest, err := auditRequestService.CreateAuditRequest(ctx, taskID, userID, username, reason, image)

	// 断言
	assert.Error(t, err)
	assert.Contains(t, err.Error(), appErrors.ErrAuditRequestAlreadyExists.Error())
	assert.Nil(t, createdRequest)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
	mockCheckApplicationDao.AssertExpectations(t)
}

func TestCreateAuditRequest_GroupNotFound(t *testing.T) {
	auditRequestService, mockCheckApplicationDao, mockTaskDao, _, mockGroupDao, mockTxManager := setupAuditRequestServiceTest()
	ctx := context.Background()

	// 测试数据
	taskID := 1
	userID := 1
	username := "user1"
	reason := "网络问题导致无法正常签到"
	image := "base64encodedimage"

	// 预期的任务
	task := &models.Task{
		TaskID:      taskID,
		TaskName:    "测试任务",
		Description: "这是一个测试任务",
		GroupID:     999, // 不存在的组ID
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("GetByTaskID", ctx, taskID, mock.AnythingOfType("[]*gorm.DB")).Return(task, nil)
	mockCheckApplicationDao.On("GetByTaskIDAndUserID", ctx, taskID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)
	mockGroupDao.On("GetByGroupID", ctx, task.GroupID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	createdRequest, err := auditRequestService.CreateAuditRequest(ctx, taskID, userID, username, reason, image)

	// 断言
	assert.Error(t, err)
	assert.Contains(t, err.Error(), appErrors.ErrGroupNotFound.Error())
	assert.Nil(t, createdRequest)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
	mockCheckApplicationDao.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
}

// --- UpdateAuditRequest 测试 ---

func TestUpdateAuditRequest_Approve_Success(t *testing.T) {
	auditRequestService, mockCheckApplicationDao, _, mockTaskRecordDao, _, mockTxManager := setupAuditRequestServiceTest()
	ctx := context.Background()
	requestID := 1

	// 预期的申请
	request := &models.CheckApplication{
		ID:            requestID,
		TaskID:        1,
		TaskName:      "测试任务",
		UserID:        1,
		Username:      "user1",
		Status:        "pending",
		Reason:        "网络问题",
		AdminID:       2,
		AdminUsername: "admin1",
		RequestAt:     time.Now(),
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockCheckApplicationDao.On("GetByID", ctx, requestID, mock.AnythingOfType("[]*gorm.DB")).Return(request, nil)
	mockCheckApplicationDao.On("Update", ctx, "approved", requestID, mock.AnythingOfType("[]*gorm.DB")).Return(nil)
	mockTaskRecordDao.On("Create", ctx, mock.AnythingOfType("*models.TaskRecord"), mock.AnythingOfType("[]*gorm.DB")).Return(nil)

	// 调用函数
	err := auditRequestService.UpdateAuditRequest(ctx, requestID, "approve")

	// 断言
	assert.NoError(t, err)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockCheckApplicationDao.AssertExpectations(t)
	mockTaskRecordDao.AssertExpectations(t)
}

func TestUpdateAuditRequest_Reject_Success(t *testing.T) {
	auditRequestService, mockCheckApplicationDao, _, _, _, mockTxManager := setupAuditRequestServiceTest()
	ctx := context.Background()
	requestID := 1

	// 预期的申请
	request := &models.CheckApplication{
		ID:            requestID,
		TaskID:        1,
		TaskName:      "测试任务",
		UserID:        1,
		Username:      "user1",
		Status:        "pending",
		Reason:        "网络问题",
		AdminID:       2,
		AdminUsername: "admin1",
		RequestAt:     time.Now(),
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockCheckApplicationDao.On("GetByID", ctx, requestID, mock.AnythingOfType("[]*gorm.DB")).Return(request, nil)
	mockCheckApplicationDao.On("Update", ctx, "rejected", requestID, mock.AnythingOfType("[]*gorm.DB")).Return(nil)

	// 调用函数
	err := auditRequestService.UpdateAuditRequest(ctx, requestID, "reject")

	// 断言
	assert.NoError(t, err)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockCheckApplicationDao.AssertExpectations(t)
}

func TestUpdateAuditRequest_NotFound(t *testing.T) {
	auditRequestService, mockCheckApplicationDao, _, _, _, mockTxManager := setupAuditRequestServiceTest()
	ctx := context.Background()
	requestID := 999 // 不存在的申请ID

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockCheckApplicationDao.On("GetByID", ctx, requestID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	err := auditRequestService.UpdateAuditRequest(ctx, requestID, "approve")

	// 断言
	assert.Error(t, err)
	assert.Contains(t, err.Error(), appErrors.ErrAuditRequestNotFound.Error())

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockCheckApplicationDao.AssertExpectations(t)
}
