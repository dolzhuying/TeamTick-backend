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

// Mock TaskDAO
type mockTaskDAO struct {
	mock.Mock
}

func (m *mockTaskDAO) Create(ctx context.Context, task *models.Task, tx ...*gorm.DB) error {
	args := m.Called(ctx, task, tx)
	if args.Error(0) == nil {
		task.TaskID = 1
		task.CreatedAt = time.Now()
		task.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *mockTaskDAO) GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.Task, error) {
	args := m.Called(ctx, groupID, tx)
	tasksArg := args.Get(0)
	if tasksArg == nil {
		return nil, args.Error(1)
	}
	return tasksArg.([]*models.Task), args.Error(1)
}

func (m *mockTaskDAO) GetByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.Task, error) {
	args := m.Called(ctx, userID, tx)
	tasksArg := args.Get(0)
	if tasksArg == nil {
		return nil, args.Error(1)
	}
	return tasksArg.([]*models.Task), args.Error(1)
}

func (m *mockTaskDAO) GetActiveTasksByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.Task, error) {
	args := m.Called(ctx, userID, tx)
	tasksArg := args.Get(0)
	if tasksArg == nil {
		return nil, args.Error(1)
	}
	return tasksArg.([]*models.Task), args.Error(1)
}

func (m *mockTaskDAO) GetEndedTasksByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.Task, error) {
	args := m.Called(ctx, userID, tx)
	tasksArg := args.Get(0)
	if tasksArg == nil {
		return nil, args.Error(1)
	}
	return tasksArg.([]*models.Task), args.Error(1)
}

func (m *mockTaskDAO) GetActiveTasksByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.Task, error) {
	args := m.Called(ctx, groupID, tx)
	tasksArg := args.Get(0)
	if tasksArg == nil {
		return nil, args.Error(1)
	}
	return tasksArg.([]*models.Task), args.Error(1)
}

func (m *mockTaskDAO) GetEndedTasksByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.Task, error) {
	args := m.Called(ctx, groupID, tx)
	tasksArg := args.Get(0)
	if tasksArg == nil {
		return nil, args.Error(1)
	}
	return tasksArg.([]*models.Task), args.Error(1)
}

func (m *mockTaskDAO) GetByTaskID(ctx context.Context, taskID int, tx ...*gorm.DB) (*models.Task, error) {
	args := m.Called(ctx, taskID, tx)
	taskArg := args.Get(0)
	if taskArg == nil {
		return nil, args.Error(1)
	}
	return taskArg.(*models.Task), args.Error(1)
}

func (m *mockTaskDAO) UpdateTask(ctx context.Context, taskID int, task *models.Task, tx ...*gorm.DB) error {
	args := m.Called(ctx, taskID, task, tx)
	return args.Error(0)
}

func (m *mockTaskDAO) Delete(ctx context.Context, taskID int, tx ...*gorm.DB) error {
	args := m.Called(ctx, taskID, tx)
	return args.Error(0)
}

// Mock TaskRecordDAO
type mockTaskRecordDAO struct {
	mock.Mock
}

func (m *mockTaskRecordDAO) Create(ctx context.Context, record *models.TaskRecord, tx ...*gorm.DB) error {
	args := m.Called(ctx, record, tx)
	if args.Error(0) == nil {
		record.RecordID = 1
		record.CreatedAt = time.Now()
		record.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *mockTaskRecordDAO) GetByTaskID(ctx context.Context, taskID int, tx ...*gorm.DB) ([]*models.TaskRecord, error) {
	args := m.Called(ctx, taskID, tx)
	recordsArg := args.Get(0)
	if recordsArg == nil {
		return nil, args.Error(1)
	}
	return recordsArg.([]*models.TaskRecord), args.Error(1)
}

func (m *mockTaskRecordDAO) GetByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.TaskRecord, error) {
	args := m.Called(ctx, userID, tx)
	recordsArg := args.Get(0)
	if recordsArg == nil {
		return nil, args.Error(1)
	}
	return recordsArg.([]*models.TaskRecord), args.Error(1)
}

func (m *mockTaskRecordDAO) GetByTaskIDAndUserID(ctx context.Context, taskID, userID int, tx ...*gorm.DB) (*models.TaskRecord, error) {
	args := m.Called(ctx, taskID, userID, tx)
	recordArg := args.Get(0)
	if recordArg == nil {
		return nil, args.Error(1)
	}
	return recordArg.(*models.TaskRecord), args.Error(1)
}

// Mock TransactionManager
type mockTestTransactionManager struct {
	mock.Mock
}

func (m *mockTestTransactionManager) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	args := m.Called(ctx, mock.AnythingOfType("func(*gorm.DB) error"))
	// 执行传入的函数
	if args.Error(0) == nil && fn != nil {
		_ = fn(nil)
	}
	return args.Error(0)
}

// 测试准备
func setupTaskServiceTest() (*TaskService, *mockTaskDAO, *mockTaskRecordDAO, *mockTransactionManager) {
	mockTaskDao := new(mockTaskDAO)
	mockTaskRecordDao := new(mockTaskRecordDAO)
	mockTxManager := new(mockTransactionManager)

	taskService := NewTaskService(
		mockTaskDao,
		mockTaskRecordDao,
		mockTxManager,
	)

	return taskService, mockTaskDao, mockTaskRecordDao, mockTxManager
}

// --- CreateTask 测试 ---

func TestCreateTask_Success(t *testing.T) {
	taskService, mockTaskDao, _, mockTxManager := setupTaskServiceTest()
	ctx := context.Background()

	// 测试数据
	taskName := "测试任务"
	description := "这是一个测试任务"
	groupID := 1
	startTime := time.Now()
	endTime := startTime.Add(24 * time.Hour) // 1天后结束
	latitude := 39.9042
	longitude := 116.4074
	radius := 100
	gps := true
	face := false
	wifi := true
	nfc := false

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("Create", ctx, mock.AnythingOfType("*models.Task"), mock.AnythingOfType("[]*gorm.DB")).Return(nil).Run(func(args mock.Arguments) {
		// 验证传给Create的任务对象
		taskArg := args.Get(1).(*models.Task)
		assert.Equal(t, taskName, taskArg.TaskName)
		assert.Equal(t, description, taskArg.Description)
		assert.Equal(t, groupID, taskArg.GroupID)
		assert.Equal(t, latitude, taskArg.Latitude)
		assert.Equal(t, longitude, taskArg.Longitude)
		assert.Equal(t, radius, taskArg.Radius)
		assert.True(t, taskArg.GPS)
		assert.False(t, taskArg.Face)
		assert.True(t, taskArg.WiFi)
		assert.False(t, taskArg.NFC)
	})

	// 调用函数
	createdTask, err := taskService.CreateTask(ctx, taskName, description, groupID,
		startTime, endTime, latitude, longitude, radius, gps, face, wifi, nfc)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, createdTask)
	assert.Equal(t, taskName, createdTask.TaskName)
	assert.Equal(t, description, createdTask.Description)
	assert.Equal(t, groupID, createdTask.GroupID)
	assert.Equal(t, startTime, createdTask.StartTime)
	assert.Equal(t, endTime, createdTask.EndTime)
	assert.Equal(t, latitude, createdTask.Latitude)
	assert.Equal(t, longitude, createdTask.Longitude)
	assert.Equal(t, radius, createdTask.Radius)
	assert.True(t, createdTask.GPS)
	assert.False(t, createdTask.Face)
	assert.True(t, createdTask.WiFi)
	assert.False(t, createdTask.NFC)
	assert.NotZero(t, createdTask.TaskID)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
}

// --- GetTasksByGroupID 测试 ---

func TestGetTasksByGroupID_All_Success(t *testing.T) {
	taskService, mockTaskDao, _, mockTxManager := setupTaskServiceTest()
	ctx := context.Background()
	groupID := 1

	expectedTasks := []*models.Task{
		{
			TaskID:      1,
			TaskName:    "任务1",
			Description: "描述1",
			GroupID:     groupID,
			StartTime:   time.Now(),
			EndTime:     time.Now().Add(24 * time.Hour),
			Latitude:    39.9042,
			Longitude:   116.4074,
		},
		{
			TaskID:      2,
			TaskName:    "任务2",
			Description: "描述2",
			GroupID:     groupID,
			StartTime:   time.Now().Add(-48 * time.Hour),
			EndTime:     time.Now().Add(-24 * time.Hour),
			Latitude:    39.9042,
			Longitude:   116.4074,
		},
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedTasks, nil)

	// 调用函数 - 获取所有任务
	tasks, err := taskService.GetTasksByGroupID(ctx, groupID)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, tasks)
	assert.Equal(t, expectedTasks, tasks)
	assert.Len(t, tasks, 2)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
}

func TestGetTasksByGroupID_NotFound(t *testing.T) {
	taskService, mockTaskDao, _, mockTxManager := setupTaskServiceTest()
	ctx := context.Background()
	groupID := 999 // 不存在的用户组

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	tasks, err := taskService.GetTasksByGroupID(ctx, groupID)

	// 断言
	assert.Error(t, err)
	assert.Contains(t, err.Error(), appErrors.ErrTaskNotFound.Error())
	assert.Nil(t, tasks)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
}

// --- GetTasksByUserID 测试 ---

func TestGetTasksByUserID_All_Success(t *testing.T) {
	taskService, mockTaskDao, _, mockTxManager := setupTaskServiceTest()
	ctx := context.Background()
	userID := 1

	expectedTasks := []*models.Task{
		{
			TaskID:      1,
			TaskName:    "任务1",
			Description: "描述1",
			GroupID:     1,
			StartTime:   time.Now(),
			EndTime:     time.Now().Add(24 * time.Hour),
		},
		{
			TaskID:      2,
			TaskName:    "任务2",
			Description: "描述2",
			GroupID:     2,
			StartTime:   time.Now().Add(-48 * time.Hour),
			EndTime:     time.Now().Add(-24 * time.Hour),
		},
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("GetByUserID", ctx, userID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedTasks, nil)

	// 调用函数 - 获取所有任务
	tasks, err := taskService.GetTasksByUserID(ctx, userID)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, tasks)
	assert.Equal(t, expectedTasks, tasks)
	assert.Len(t, tasks, 2)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
}

func TestGetTasksByUserID_NotFound(t *testing.T) {
	taskService, mockTaskDao, _, mockTxManager := setupTaskServiceTest()
	ctx := context.Background()
	userID := 999 // 不存在的用户

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("GetByUserID", ctx, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	tasks, err := taskService.GetTasksByUserID(ctx, userID)

	// 断言
	assert.Error(t, err)
	assert.Contains(t, err.Error(), appErrors.ErrTaskNotFound.Error())
	assert.Nil(t, tasks)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
}

// --- GetTaskByTaskID 测试 ---

func TestGetTaskByTaskID_Success(t *testing.T) {
	taskService, mockTaskDao, _, mockTxManager := setupTaskServiceTest()
	ctx := context.Background()
	taskID := 1

	expectedTask := &models.Task{
		TaskID:      taskID,
		TaskName:    "测试任务",
		Description: "这是一个测试任务",
		GroupID:     1,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(24 * time.Hour),
		Latitude:    39.9042,
		Longitude:   116.4074,
		Radius:      100,
		GPS:         true,
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("GetByTaskID", ctx, taskID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedTask, nil)

	// 调用函数
	task, err := taskService.GetTaskByTaskID(ctx, taskID)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, expectedTask, task)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
}

func TestGetTaskByTaskID_NotFound(t *testing.T) {
	taskService, mockTaskDao, _, mockTxManager := setupTaskServiceTest()
	ctx := context.Background()
	taskID := 999 // 不存在的任务ID

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("GetByTaskID", ctx, taskID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	task, err := taskService.GetTaskByTaskID(ctx, taskID)

	// 断言
	assert.Error(t, err)
	assert.Contains(t, err.Error(), appErrors.ErrTaskNotFound.Error())
	assert.Nil(t, task)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
}

// --- UpdateTask 测试 ---

func TestUpdateTask_Success(t *testing.T) {
	taskService, mockTaskDao, _, mockTxManager := setupTaskServiceTest()
	ctx := context.Background()
	taskID := 1

	// 更新的任务信息
	taskName := "更新后的任务名"
	description := "更新后的描述"
	startTime := time.Now()
	endTime := startTime.Add(48 * time.Hour)
	latitude := 39.9042
	longitude := 116.4074
	radius := 150
	gps := true
	face := true
	wifi := false
	nfc := false

	// 原任务
	originalTask := &models.Task{
		TaskID:      taskID,
		TaskName:    "原任务名",
		Description: "原描述",
		GroupID:     1,
		StartTime:   time.Now().Add(-24 * time.Hour),
		EndTime:     time.Now().Add(24 * time.Hour),
		Latitude:    39.9,
		Longitude:   116.4,
		Radius:      100,
		GPS:         false,
		Face:        false,
	}

	// 更新后的任务
	updatedTask := &models.Task{
		TaskID:      taskID,
		TaskName:    taskName,
		Description: description,
		GroupID:     1,
		StartTime:   startTime,
		EndTime:     endTime,
		Latitude:    latitude,
		Longitude:   longitude,
		Radius:      radius,
		GPS:         gps,
		Face:        face,
		WiFi:        wifi,
		NFC:         nfc,
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("GetByTaskID", ctx, taskID, mock.AnythingOfType("[]*gorm.DB")).Return(originalTask, nil).Once()
	mockTaskDao.On("UpdateTask", ctx, taskID, mock.AnythingOfType("*models.Task"), mock.AnythingOfType("[]*gorm.DB")).Return(nil)
	mockTaskDao.On("GetByTaskID", ctx, taskID, mock.AnythingOfType("[]*gorm.DB")).Return(updatedTask, nil).Once()

	// 调用函数
	result, err := taskService.UpdateTask(ctx, taskID, taskName, description, startTime, endTime,
		latitude, longitude, radius, gps, face, wifi, nfc)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, updatedTask, result)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
}

func TestUpdateTask_NotFound(t *testing.T) {
	taskService, mockTaskDao, _, mockTxManager := setupTaskServiceTest()
	ctx := context.Background()
	taskID := 999 // 不存在的任务ID

	// 更新的任务信息
	taskName := "更新后的任务名"
	description := "更新后的描述"
	startTime := time.Now()
	endTime := startTime.Add(48 * time.Hour)
	latitude := 39.9042
	longitude := 116.4074
	radius := 150
	gps := true
	face := true
	wifi := false
	nfc := false

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("GetByTaskID", ctx, taskID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	result, err := taskService.UpdateTask(ctx, taskID, taskName, description, startTime, endTime,
		latitude, longitude, radius, gps, face, wifi, nfc)

	// 断言
	assert.Error(t, err)
	assert.Contains(t, err.Error(), appErrors.ErrTaskNotFound.Error())
	assert.Nil(t, result)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
}
