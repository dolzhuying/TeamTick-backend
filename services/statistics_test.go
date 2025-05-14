package service

import (
	"context"
	"os"
	"testing"
	"time"

	"TeamTickBackend/dal/models"
	"TeamTickBackend/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Mock StatisticsDAO
type mockStatisticsDAO struct {
	mock.Mock
}

func (m *mockStatisticsDAO) GetAllGroups(ctx context.Context, tx ...*gorm.DB) ([]*models.Group, error) {
	args := m.Called(ctx, tx)
	groupsArg := args.Get(0)
	if groupsArg == nil {
		return nil, args.Error(1)
	}
	return groupsArg.([]*models.Group), args.Error(1)
}

func (m *mockStatisticsDAO) GetGroupSignInSuccess(ctx context.Context, groupID int, startTime, endTime time.Time, tx ...*gorm.DB) ([]*models.TaskRecord, error) {
	args := m.Called(ctx, groupID, startTime, endTime, tx)
	recordsArg := args.Get(0)
	if recordsArg == nil {
		return nil, args.Error(1)
	}
	return recordsArg.([]*models.TaskRecord), args.Error(1)
}

func (m *mockStatisticsDAO) GetGroupSignInException(ctx context.Context, groupID int, startTime, endTime time.Time, tx ...*gorm.DB) ([]*models.CheckApplication, error) {
	args := m.Called(ctx, groupID, startTime, endTime, tx)
	recordsArg := args.Get(0)
	if recordsArg == nil {
		return nil, args.Error(1)
	}
	return recordsArg.([]*models.CheckApplication), args.Error(1)
}

func (m *mockStatisticsDAO) GetGroupSignInAbsent(ctx context.Context, groupID int, startTime, endTime time.Time, tx ...*gorm.DB) ([]*models.AbsentRecord, error) {
	args := m.Called(ctx, groupID, startTime, endTime, tx)
	recordsArg := args.Get(0)
	if recordsArg == nil {
		return nil, args.Error(1)
	}
	return recordsArg.([]*models.AbsentRecord), args.Error(1)
}

func (m *mockStatisticsDAO) GetMemberSignInSuccessNum(ctx context.Context, groupID, userID int, startTime, endTime time.Time, tx ...*gorm.DB) (int, error) {
	args := m.Called(ctx, groupID, userID, startTime, endTime, tx)
	return args.Int(0), args.Error(1)
}

func (m *mockStatisticsDAO) GetMemberSignInExceptionNum(ctx context.Context, groupID, userID int, startTime, endTime time.Time, tx ...*gorm.DB) (int, error) {
	args := m.Called(ctx, groupID, userID, startTime, endTime, tx)
	return args.Int(0), args.Error(1)
}

func (m *mockStatisticsDAO) GetMemberSignInAbsentNum(ctx context.Context, groupID, userID int, startTime, endTime time.Time, tx ...*gorm.DB) (int, error) {
	args := m.Called(ctx, groupID, userID, startTime, endTime, tx)
	return args.Int(0), args.Error(1)
}

func TestMain(m *testing.M) {
	// 初始化 logger
	if err := logger.InitLogger(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func setupStatisticsServiceTest() (*StatisticsService, *mockStatisticsDAO, *mockTransactionManager, *mockGroupDAO, *mockGroupMemberDAO) {
	mockStatisticsDao := new(mockStatisticsDAO)
	mockTxManager := new(mockTransactionManager)
	mockGroupDao := new(mockGroupDAO)
	mockGroupMemberDao := new(mockGroupMemberDAO)

	statisticsService := NewStatisticsService(
		mockStatisticsDao,
		mockGroupDao,
		mockTxManager,
		mockGroupMemberDao,
	)

	return statisticsService, mockStatisticsDao, mockTxManager, mockGroupDao, mockGroupMemberDao
}

func TestGenerateGroupSignInStatisticsExcel_Success(t *testing.T) {
	statisticsService, mockStatisticsDao, mockTxManager, mockGroupDao, mockGroupMemberDao := setupStatisticsServiceTest()
	ctx := context.Background()
	groupID := 1
	startTime := time.Now().AddDate(0, 0, -7)
	endTime := time.Now()

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(&models.Group{
		GroupID:   groupID,
		GroupName: "测试群组",
	}, nil)
	mockGroupMemberDao.On("GetMembersByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return([]*models.GroupMember{
		{
			GroupID:  groupID,
			UserID:   1,
			Username: "user1",
		},
		{
			GroupID:  groupID,
			UserID:   2,
			Username: "user2",
		},
	}, nil)

	// Mock GetGroupMemberSignInStatistics 的调用
	mockStatisticsDao.On("GetMemberSignInSuccessNum", ctx, groupID, 1, startTime, endTime, mock.Anything).Return(5, nil)
	mockStatisticsDao.On("GetMemberSignInExceptionNum", ctx, groupID, 1, startTime, endTime, mock.Anything).Return(1, nil)
	mockStatisticsDao.On("GetMemberSignInAbsentNum", ctx, groupID, 1, startTime, endTime, mock.Anything).Return(0, nil)
	mockStatisticsDao.On("GetMemberSignInSuccessNum", ctx, groupID, 2, startTime, endTime, mock.Anything).Return(3, nil)
	mockStatisticsDao.On("GetMemberSignInExceptionNum", ctx, groupID, 2, startTime, endTime, mock.Anything).Return(0, nil)
	mockStatisticsDao.On("GetMemberSignInAbsentNum", ctx, groupID, 2, startTime, endTime, mock.Anything).Return(2, nil)

	// 调用函数
	excelData, err := statisticsService.GenerateGroupSignInStatisticsExcel(ctx, []int{groupID}, startTime, endTime, []string{"success", "absent", "exception"})

	// 断言
	assert.NoError(t, err)
	assert.NotEmpty(t, excelData)
	assert.Contains(t, excelData, "statistics_")
	assert.Contains(t, excelData, ".xlsx")

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
	mockGroupMemberDao.AssertExpectations(t)
	mockStatisticsDao.AssertExpectations(t)
}

func TestGenerateGroupSignInStatisticsExcel_DAOError(t *testing.T) {
	statisticsService, _, _, mockGroupDao, _ := setupStatisticsServiceTest()
	ctx := context.Background()
	groupID := 1
	startTime := time.Now().AddDate(0, 0, -7)
	endTime := time.Now()

	// Mock期望
	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	excelData, err := statisticsService.GenerateGroupSignInStatisticsExcel(ctx, []int{groupID}, startTime, endTime, []string{"success", "absent", "exception"})

	// 断言
	assert.Error(t, err)
	assert.Empty(t, excelData)
	assert.Contains(t, err.Error(), "用户组不存在")

	// 验证mock调用
	mockGroupDao.AssertExpectations(t)
}
