 package service

import (
	"TeamTickBackend/dal/models"
	appErrors "TeamTickBackend/pkg/errors"
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
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

// Mock TaskRecordRedisDAO
type mockTaskRecordRedisDAO struct {
	mock.Mock
}

func (m *mockTaskRecordRedisDAO) Create(ctx context.Context, record *models.TaskRecord, tx ...*redis.Client) error {
	args := m.Called(ctx, record, tx)
	return args.Error(0)
}

func (m *mockTaskRecordRedisDAO) GetByTaskID(ctx context.Context, taskID int, tx ...*redis.Client) ([]*models.TaskRecord, error) {
	args := m.Called(ctx, taskID, tx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TaskRecord), args.Error(1)
}

func (m *mockTaskRecordRedisDAO) SetByTaskID(ctx context.Context, taskID int, records []*models.TaskRecord) error {
	args := m.Called(ctx, taskID, records)
	return args.Error(0)
}

func (m *mockTaskRecordRedisDAO) GetByUserID(ctx context.Context, userID int, tx ...*redis.Client) ([]*models.TaskRecord, error) {
	args := m.Called(ctx, userID, tx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TaskRecord), args.Error(1)
}

func (m *mockTaskRecordRedisDAO) SetByUserID(ctx context.Context, userID int, records []*models.TaskRecord) error {
	args := m.Called(ctx, userID, records)
	return args.Error(0)
}

func (m *mockTaskRecordRedisDAO) GetByTaskIDAndUserID(ctx context.Context, taskID, userID int, tx ...*redis.Client) (*models.TaskRecord, error) {
	args := m.Called(ctx, taskID, userID, tx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaskRecord), args.Error(1)
}

func (m *mockTaskRecordRedisDAO) SetTaskIDAndUserID(ctx context.Context, record *models.TaskRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *mockTaskRecordRedisDAO) DeleteCache(ctx context.Context, taskID, userID int) error {
	args := m.Called(ctx, taskID, userID)
	return args.Error(0)
}

// Mock TaskRedisDAO
type mockTaskRedisDAO struct {
	mock.Mock
}

func (m *mockTaskRedisDAO) GetByGroupID(ctx context.Context, groupID int, tx ...*redis.Client) ([]*models.Task, error) {
	args := m.Called(ctx, groupID, tx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Task), args.Error(1)
}

func (m *mockTaskRedisDAO) SetByGroupID(ctx context.Context, groupID int, tasks []*models.Task) error {
	args := m.Called(ctx, groupID, tasks)
	return args.Error(0)
}

func (m *mockTaskRedisDAO) GetByTaskID(ctx context.Context, taskID int, tx ...*redis.Client) (*models.Task, error) {
	args := m.Called(ctx, taskID, tx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *mockTaskRedisDAO) SetByTaskID(ctx context.Context, taskID int, task *models.Task) error {
	args := m.Called(ctx, taskID, task)
	return args.Error(0)
}

func (m *mockTaskRedisDAO) GetByUserID(ctx context.Context, userID int, tx ...*redis.Client) ([]*models.Task, error) {
	args := m.Called(ctx, userID, tx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Task), args.Error(1)
}

func (m *mockTaskRedisDAO) SetByUserID(ctx context.Context, userID int, tasks []*models.Task) error {
	args := m.Called(ctx, userID, tasks)
	return args.Error(0)
}

func (m *mockTaskRedisDAO) DeleteCacheByTaskID(ctx context.Context, taskID int) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}

func (m *mockTaskRedisDAO) DeleteCacheByGroupID(ctx context.Context, groupID int) error {
	args := m.Called(ctx, groupID)
	return args.Error(0)
}

func (m *mockTaskRedisDAO) DeleteCacheByUserID(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// 测试准备
func setupTaskServiceTest() (*TaskService, *mockTaskDAO, *mockTaskRecordDAO, *mockTaskRecordRedisDAO, *mockTaskRedisDAO, *mockTransactionManager, *mockGroupDAO) {
	mockTaskDao := new(mockTaskDAO)
	mockTaskRecordDao := new(mockTaskRecordDAO)
	mockTaskRecordRedisDao := new(mockTaskRecordRedisDAO)
	mockTaskRedisDao := new(mockTaskRedisDAO)
	mockTxManager := new(mockTransactionManager)
	mockGroupDao := new(mockGroupDAO)
	mockGroupMemberDao := new(mockGroupMemberDAO)

	taskService := NewTaskService(
		mockTaskDao,
		mockTaskRecordDao,
		mockTaskRecordRedisDao,
		mockTaskRedisDao,
		mockTxManager,
		mockGroupDao,
		mockGroupMemberDao,
	)

	return taskService, mockTaskDao, mockTaskRecordDao, mockTaskRecordRedisDao, mockTaskRedisDao, mockTxManager, mockGroupDao
}

// --- CreateTask 测试 ---

func TestCreateTask_Success(t *testing.T) {
	taskService, mockTaskDao, _, _, mockTaskRedisDao, mockTxManager, _ := setupTaskServiceTest()
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
	ssid := "test_ssid"
	bssid := "test_bssid"
	tagId := "test_tag_id"
	tagName := "test_tag_name"

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

	// Mock Redis相关期望
	mockTaskRedisDao.On("SetByTaskID", ctx, mock.AnythingOfType("int"), mock.AnythingOfType("*models.Task")).Return(nil)
	mockTaskRedisDao.On("GetByGroupID", ctx, groupID, mock.Anything).Return(([]*models.Task)(nil), nil)
	mockTaskRedisDao.On("SetByGroupID", ctx, groupID, mock.AnythingOfType("[]*models.Task")).Return(nil)

	// 调用函数
	createdTask, err := taskService.CreateTask(ctx, taskName, description, groupID,
		startTime, endTime, latitude, longitude, radius, gps, face, wifi, nfc, ssid, bssid, tagId, tagName)

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
	mockTaskRedisDao.AssertExpectations(t)
}

// --- GetTasksByGroupID 测试 ---

func TestGetTasksByGroupID_All_Success(t *testing.T) {
	taskService, mockTaskDao, _, _, mockTaskRedisDao, mockTxManager, _ := setupTaskServiceTest()
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
	mockTaskRedisDao.On("GetByGroupID", ctx, groupID, mock.Anything).Return(nil, redis.Nil) // 模拟缓存未命中
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedTasks, nil)
	mockTaskRedisDao.On("SetByGroupID", ctx, groupID, expectedTasks).Return(nil) // 模拟缓存写入

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
	mockTaskRedisDao.AssertExpectations(t)
}

func TestGetTasksByGroupID_NotFound(t *testing.T) {
	taskService, mockTaskDao, _, _, mockTaskRedisDao, mockTxManager, _ := setupTaskServiceTest()
	ctx := context.Background()
	groupID := 999 // 不存在的用户组

	// Mock期望
	mockTaskRedisDao.On("GetByGroupID", ctx, groupID, mock.Anything).Return(nil, redis.Nil) // 模拟缓存未命中
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
	mockTaskRedisDao.AssertExpectations(t)
}

// --- GetTasksByUserID 测试 ---

func TestGetTasksByUserID_All_Success(t *testing.T) {
	taskService, mockTaskDao, _, _, _, mockTxManager, _ := setupTaskServiceTest()
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
	taskService, mockTaskDao, _, _, _, mockTxManager, _ := setupTaskServiceTest()
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
	taskService, mockTaskDao, _, _, mockTaskRedisDao, mockTxManager, _ := setupTaskServiceTest()
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
	mockTaskRedisDao.On("GetByTaskID", ctx, taskID, mock.Anything).Return(nil, redis.Nil) // 模拟缓存未命中
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockTaskDao.On("GetByTaskID", ctx, taskID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedTask, nil)
	mockTaskRedisDao.On("SetByTaskID", ctx, taskID, expectedTask).Return(nil) // 模拟缓存写入

	// 调用函数
	task, err := taskService.GetTaskByTaskID(ctx, taskID)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, expectedTask, task)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
	mockTaskRedisDao.AssertExpectations(t)
}

func TestGetTaskByTaskID_NotFound(t *testing.T) {
	taskService, mockTaskDao, _, _, mockTaskRedisDao, mockTxManager, _ := setupTaskServiceTest()
	ctx := context.Background()
	taskID := 999 // 不存在的任务ID

	// Mock期望
	mockTaskRedisDao.On("GetByTaskID", ctx, taskID, mock.Anything).Return(nil, redis.Nil) // 模拟缓存未命中
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
	mockTaskRedisDao.AssertExpectations(t)
}

// --- UpdateTask 测试 ---

func TestUpdateTask_Success(t *testing.T) {
	taskService, mockTaskDao, _, _, mockTaskRedisDao, mockTxManager, _ := setupTaskServiceTest()
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

	// Mock Redis操作
	mockTaskRedisDao.On("DeleteCacheByTaskID", ctx, taskID).Return(nil)
	mockTaskRedisDao.On("DeleteCacheByGroupID", ctx, updatedTask.GroupID).Return(nil)

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
	mockTaskRedisDao.AssertExpectations(t)
}

func TestUpdateTask_NotFound(t *testing.T) {
	taskService, mockTaskDao, _, _, _, mockTxManager, _ := setupTaskServiceTest()
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

// --- CheckInTask 测试 ---
func TestCheckInTask_Success(t *testing.T) {
	taskService, mockTaskDao, mockTaskRecordDao, mockTaskRecordRedisDao, _, mockTxManager, mockGroupDao := setupTaskServiceTest()
	ctx := context.Background()

	// 测试数据
	taskID := 1
	userID := 1
	groupID := 1
	latitude := 39.9042
	longitude := 116.4074
	signedInTime := time.Now()

	mockedTask := &models.Task{
		TaskID:    taskID,
		TaskName:  "签到任务1",
		GroupID:   groupID,
		StartTime: time.Now().Add(-24 * time.Hour), //确保任务已开始
		EndTime:   time.Now().Add(24 * time.Hour),  //确保任务未结束
	}
	mockedGroup := &models.Group{
		GroupID:   groupID,
		GroupName: "测试组",
	}
	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*gorm.DB) error)
		// 在此处模拟事务内的操作，确保fn能成功执行
		mockTaskDao.On("GetByTaskID", ctx, taskID, mock.AnythingOfType("[]*gorm.DB")).Return(mockedTask, nil)
		mockTaskRecordDao.On("GetByTaskIDAndUserID", ctx, taskID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)
		mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(mockedGroup, nil)
		mockTaskRecordDao.On("Create", ctx, mock.AnythingOfType("*models.TaskRecord"), mock.AnythingOfType("[]*gorm.DB")).Return(nil).Run(func(args mock.Arguments) {
			recordArg := args.Get(1).(*models.TaskRecord)
			recordArg.RecordID = 123
			recordArg.TaskName = mockedTask.TaskName
			recordArg.GroupName = mockedGroup.GroupName
			recordArg.CreatedAt = time.Now()
			recordArg.UpdatedAt = time.Now()
		})
		_ = fn(nil) // Simulate successful transaction
	})

	// Mock Redis DAO calls (outside transaction)
	mockTaskRecordRedisDao.On("SetTaskIDAndUserID", ctx, mock.AnythingOfType("*models.TaskRecord")).Return(nil)
	mockTaskRecordRedisDao.On("GetByTaskID", ctx, taskID, mock.Anything).Return(([]*models.TaskRecord)(nil), nil).Once() // For appending to task list
	mockTaskRecordRedisDao.On("SetByTaskID", ctx, taskID, mock.AnythingOfType("[]*models.TaskRecord")).Return(nil)
	mockTaskRecordRedisDao.On("GetByUserID", ctx, userID, mock.Anything).Return(([]*models.TaskRecord)(nil), nil).Once() // For appending to user list
	mockTaskRecordRedisDao.On("SetByUserID", ctx, userID, mock.AnythingOfType("[]*models.TaskRecord")).Return(nil)

	// 调用函数
	createdRecord, err := taskService.CheckInTask(ctx, taskID, userID, latitude, longitude, signedInTime)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, createdRecord)
	assert.Equal(t, taskID, createdRecord.TaskID)
	assert.Equal(t, userID, createdRecord.UserID)
	assert.Equal(t, mockedTask.TaskName, createdRecord.TaskName)    // Check if TaskName is populated
	assert.Equal(t, mockedGroup.GroupName, createdRecord.GroupName) // Check if GroupName is populated
	assert.Equal(t, latitude, createdRecord.Latitude)
	assert.Equal(t, longitude, createdRecord.Longitude)
	assert.NotZero(t, createdRecord.RecordID)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockTaskDao.AssertExpectations(t)
	mockTaskRecordDao.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
	mockTaskRecordRedisDao.AssertExpectations(t)
}
