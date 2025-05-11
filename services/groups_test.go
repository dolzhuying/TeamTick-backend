package service

// import (
// 	"TeamTickBackend/dal/models"
// 	appErrors "TeamTickBackend/pkg/errors"
// 	"context"
// 	"errors"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/redis/go-redis/v9"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"gorm.io/gorm"
// )

// // --- Mock 实现 ---

// // Mock TransactionManager
// type mockTransactionManager struct {
// 	mock.Mock
// }

// func (m *mockTransactionManager) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
// 	args := m.Called(ctx, fn)
// 	return args.Error(0)
// }

// // Mock GroupDAO
// type mockGroupDAO struct {
// 	mock.Mock
// }

// func (m *mockGroupDAO) Create(ctx context.Context, group *models.Group, tx ...*gorm.DB) error {
// 	args := m.Called(ctx, group, tx)
// 	// 模拟创建时分配ID和时间
// 	if args.Error(0) == nil {
// 		group.GroupID = 1
// 		group.CreatedAt = time.Now()
// 		group.UpdatedAt = time.Now()
// 	}
// 	return args.Error(0)
// }

// func (m *mockGroupDAO) GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) (*models.Group, error) {
// 	args := m.Called(ctx, groupID, tx)
// 	groupArg := args.Get(0)
// 	if groupArg == nil {
// 		return nil, args.Error(1)
// 	}
// 	return groupArg.(*models.Group), args.Error(1)
// }

// func (m *mockGroupDAO) GetGroupsByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.Group, error) {
// 	args := m.Called(ctx, userID, tx)
// 	groupsArg := args.Get(0)
// 	if groupsArg == nil {
// 		return nil, args.Error(1)
// 	}
// 	return groupsArg.([]*models.Group), args.Error(1)
// }

// func (m *mockGroupDAO) UpdateMessage(ctx context.Context, groupID int, groupName, description string, tx ...*gorm.DB) error {
// 	args := m.Called(ctx, groupID, groupName, description, tx)
// 	return args.Error(0)
// }

// func (m *mockGroupDAO) UpdateMemberNum(ctx context.Context, groupID int, isAdd bool, tx ...*gorm.DB) error {
// 	args := m.Called(ctx, groupID, isAdd, tx)
// 	return args.Error(0)
// }

// func (m *mockGroupDAO) GetGroupsByUserIDAndfilter(ctx context.Context, userID int, filter string, tx ...*gorm.DB) ([]*models.Group, error) {
// 	args := m.Called(ctx, userID, filter, tx)
// 	groupsArg := args.Get(0)
// 	if groupsArg == nil {
// 		return nil, args.Error(1)
// 	}
// 	return groupsArg.([]*models.Group), args.Error(1)
// }

// // 添加Delete方法
// func (m *mockGroupDAO) Delete(ctx context.Context, groupID int, tx ...*gorm.DB) error {
// 	args := m.Called(ctx, groupID, tx)
// 	return args.Error(0)
// }

// // Mock GroupMemberDAO
// type mockGroupMemberDAO struct {
// 	mock.Mock
// }

// func (m *mockGroupMemberDAO) Create(ctx context.Context, member *models.GroupMember, tx ...*gorm.DB) error {
// 	args := m.Called(ctx, member, tx)
// 	if args.Error(0) == nil {
// 		member.CreatedAt = time.Now()
// 	}
// 	return args.Error(0)
// }

// func (m *mockGroupMemberDAO) GetMemberByGroupIDAndUserID(ctx context.Context, groupID, userID int, tx ...*gorm.DB) (*models.GroupMember, error) {
// 	args := m.Called(ctx, groupID, userID, tx)
// 	memberArg := args.Get(0)
// 	if memberArg == nil {
// 		return nil, args.Error(1)
// 	}
// 	return memberArg.(*models.GroupMember), args.Error(1)
// }

// func (m *mockGroupMemberDAO) GetMembersByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.GroupMember, error) {
// 	args := m.Called(ctx, groupID, tx)
// 	membersArg := args.Get(0)
// 	if membersArg == nil {
// 		return nil, args.Error(1)
// 	}
// 	return membersArg.([]*models.GroupMember), args.Error(1)
// }

// func (m *mockGroupMemberDAO) Delete(ctx context.Context, groupID, userID int, tx ...*gorm.DB) error {
// 	args := m.Called(ctx, groupID, userID, tx)
// 	return args.Error(0)
// }

// // Mock JoinApplicationDAO
// type mockJoinApplicationDAO struct {
// 	mock.Mock
// }

// func (m *mockJoinApplicationDAO) Create(ctx context.Context, application *models.JoinApplication, tx ...*gorm.DB) error {
// 	args := m.Called(ctx, application, tx)
// 	if args.Error(0) == nil {
// 		application.Status = "pending"
// 		application.CreatedAt = time.Now()
// 		application.UpdatedAt = time.Now()
// 	}
// 	return args.Error(0)
// }

// func (m *mockJoinApplicationDAO) GetByGroupIDAndStatus(ctx context.Context, groupID int, status string, tx ...*gorm.DB) ([]*models.JoinApplication, error) {
// 	args := m.Called(ctx, groupID, status, tx)
// 	applicationsArg := args.Get(0)
// 	if applicationsArg == nil {
// 		return nil, args.Error(1)
// 	}
// 	return applicationsArg.([]*models.JoinApplication), args.Error(1)
// }

// // 添加缺失的GetByUserID方法
// func (m *mockJoinApplicationDAO) GetByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.JoinApplication, error) {
// 	args := m.Called(ctx, userID, tx)
// 	applicationsArg := args.Get(0)
// 	if applicationsArg == nil {
// 		return nil, args.Error(1)
// 	}
// 	return applicationsArg.([]*models.JoinApplication), args.Error(1)
// }

// // 添加缺失的UpdateStatus方法
// func (m *mockJoinApplicationDAO) UpdateStatus(ctx context.Context, applicationID int, status string, tx ...*gorm.DB) error {
// 	args := m.Called(ctx, applicationID, status, tx)
// 	return args.Error(0)
// }

// // 添加UpdateRejectReason方法
// func (m *mockJoinApplicationDAO) UpdateRejectReason(ctx context.Context, requestID int, rejectReason string, tx ...*gorm.DB) error {
// 	args := m.Called(ctx, requestID, rejectReason, tx)
// 	return args.Error(0)
// }

// // 添加GetByGroupIDAndUserID方法
// func (m *mockJoinApplicationDAO) GetByGroupIDAndUserID(ctx context.Context, groupID int, userID int, tx ...*gorm.DB) (*models.JoinApplication, error) {
// 	args := m.Called(ctx, groupID, userID, tx)
// 	applicationArg := args.Get(0)
// 	if applicationArg == nil {
// 		return nil, args.Error(1)
// 	}
// 	return applicationArg.(*models.JoinApplication), args.Error(1)
// }

// // 添加缺失的GetByGroupID方法
// func (m *mockJoinApplicationDAO) GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.JoinApplication, error) {
// 	args := m.Called(ctx, groupID, tx)
// 	applicationsArg := args.Get(0)
// 	if applicationsArg == nil {
// 		return nil, args.Error(1)
// 	}
// 	return applicationsArg.([]*models.JoinApplication), args.Error(1)
// }

// // Mock GroupMemberRedisDAO
// type mockGroupMemberRedisDAO struct {
// 	mock.Mock
// }

// func (m *mockGroupMemberRedisDAO) GetMembersByGroupID(ctx context.Context, groupID int, tx ...*redis.Client) ([]*models.GroupMember, error) {
// 	args := m.Called(ctx, groupID, tx)
// 	membersArg := args.Get(0)
// 	if membersArg == nil {
// 		return nil, args.Error(1)
// 	}
// 	return membersArg.([]*models.GroupMember), args.Error(1)
// }

// func (m *mockGroupMemberRedisDAO) SetMembersByGroupID(ctx context.Context, groupID int, members []*models.GroupMember) error {
// 	args := m.Called(ctx, groupID, members)
// 	return args.Error(0)
// }

// func (m *mockGroupMemberRedisDAO) DeleteCacheByGroupID(ctx context.Context, groupID int) error {
// 	args := m.Called(ctx, groupID)
// 	return args.Error(0)
// }

// func (m *mockGroupMemberRedisDAO) GetMemberByGroupIDAndUserID(ctx context.Context, groupID, userID int, tx ...*redis.Client) (*models.GroupMember, error) {
// 	args := m.Called(ctx, groupID, userID, tx)
// 	memberArg := args.Get(0)
// 	if memberArg == nil {
// 		return nil, args.Error(1)
// 	}
// 	return memberArg.(*models.GroupMember), args.Error(1)
// }

// func (m *mockGroupMemberRedisDAO) SetMemberByGroupIDAndUserID(ctx context.Context, member *models.GroupMember) error {
// 	args := m.Called(ctx, member)
// 	return args.Error(0)
// }

// func (m *mockGroupMemberRedisDAO) DeleteCacheByGroupIDAndUserID(ctx context.Context, groupID, userID int) error {
// 	args := m.Called(ctx, groupID, userID)
// 	return args.Error(0)
// }

// // Mock GroupRedisDAO
// type mockGroupRedisDAO struct {
// 	mock.Mock
// }

// func (m *mockGroupRedisDAO) GetByGroupID(ctx context.Context, groupID int, tx ...*redis.Client) (*models.Group, error) {
// 	args := m.Called(ctx, groupID, tx)
// 	groupArg := args.Get(0)
// 	if groupArg == nil {
// 		return nil, args.Error(1)
// 	}
// 	return groupArg.(*models.Group), args.Error(1)
// }

// func (m *mockGroupRedisDAO) SetByGroupID(ctx context.Context, groupID int, group *models.Group) error {
// 	args := m.Called(ctx, groupID, group)
// 	return args.Error(0)
// }

// func (m *mockGroupRedisDAO) DeleteCacheByGroupID(ctx context.Context, groupID int) error {
// 	args := m.Called(ctx, groupID)
// 	return args.Error(0)
// }

// // --- 测试准备 ---

// func setupGroupServiceTest() (*GroupsService, *mockGroupDAO, *mockGroupMemberDAO, *mockJoinApplicationDAO, *mockTransactionManager, *mockGroupRedisDAO, *mockGroupMemberRedisDAO) {
// 	mockGroupDao := new(mockGroupDAO)
// 	mockGroupMemberDao := new(mockGroupMemberDAO)
// 	mockJoinApplicationDao := new(mockJoinApplicationDAO)
// 	mockTxManager := new(mockTransactionManager)
// 	mockGroupRedisDao := new(mockGroupRedisDAO)
// 	mockGroupMemberRedisDao := new(mockGroupMemberRedisDAO)

// 	groupsService := NewGroupsService(
// 		mockGroupDao,
// 		mockGroupMemberDao,
// 		mockJoinApplicationDao,
// 		mockTxManager,
// 		mockGroupRedisDao,
// 		mockGroupMemberRedisDao,
// 	)

// 	return groupsService, mockGroupDao, mockGroupMemberDao, mockJoinApplicationDao, mockTxManager, mockGroupRedisDao, mockGroupMemberRedisDao
// }

// // --- CreateGroup 测试 ---

// func TestCreateGroup_Success(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager, mockGroupRedisDao, _ := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupName := "测试群组"
// 	description := "这是一个测试群组"
// 	creatorName := "admin"
// 	creatorID := 1

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupDao.On("Create", ctx, mock.AnythingOfType("*models.Group"), mock.AnythingOfType("[]*gorm.DB")).Return(nil).Run(func(args mock.Arguments) {
// 		// 验证传给Create的群组对象
// 		groupArg := args.Get(1).(*models.Group)
// 		assert.Equal(t, groupName, groupArg.GroupName)
// 		assert.Equal(t, description, groupArg.Description)
// 		assert.Equal(t, creatorID, groupArg.CreatorID)
// 		assert.Equal(t, creatorName, groupArg.CreatorName)
// 	})
// 	mockGroupMemberDao.On("Create", ctx, mock.AnythingOfType("*models.GroupMember"), mock.AnythingOfType("[]*gorm.DB")).Return(nil).Run(func(args mock.Arguments) {
// 		// 验证传给Create的群组成员对象
// 		memberArg := args.Get(1).(*models.GroupMember)
// 		assert.Equal(t, 1, memberArg.GroupID) // 模拟ID为1
// 		assert.Equal(t, creatorID, memberArg.UserID)
// 		assert.Equal(t, creatorName, memberArg.Username)
// 		assert.Equal(t, "admin", memberArg.Role)
// 	})
// 	mockGroupRedisDao.On("SetByGroupID", ctx, mock.AnythingOfType("int"), mock.AnythingOfType("*models.Group")).Return(nil)

// 	// 调用函数
// 	createdGroup, err := groupsService.CreateGroup(ctx, groupName, description, creatorName, creatorID)

// 	// 断言
// 	assert.NoError(t, err)
// 	assert.NotNil(t, createdGroup)
// 	assert.Equal(t, groupName, createdGroup.GroupName)
// 	assert.Equal(t, description, createdGroup.Description)
// 	assert.Equal(t, creatorID, createdGroup.CreatorID)
// 	assert.Equal(t, creatorName, createdGroup.CreatorName)
// 	assert.NotZero(t, createdGroup.GroupID)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupRedisDao.AssertExpectations(t)
// }

// // --- GetGroupByGroupID 测试 ---

// func TestGetGroupByGroupID_Success(t *testing.T) {
// 	groupsService, mockGroupDao, _, _, mockTxManager, mockGroupRedisDao, _ := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	expectedGroup := &models.Group{
// 		GroupID:     groupID,
// 		GroupName:   "测试群组",
// 		Description: "这是一个测试群组",
// 		CreatorID:   1,
// 		CreatorName: "admin",
// 		MemberNum:   1,
// 	}

// 	// Mock期望
// 	mockGroupRedisDao.On("GetByGroupID", ctx, groupID, mock.Anything).Return(nil, redis.Nil) // 模拟缓存未命中
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedGroup, nil)
// 	mockGroupRedisDao.On("SetByGroupID", ctx, groupID, expectedGroup).Return(nil) // 模拟缓存写入

// 	// 调用函数
// 	group, err := groupsService.GetGroupByGroupID(ctx, groupID)

// 	// 断言
// 	assert.NoError(t, err)
// 	assert.NotNil(t, group)
// 	assert.Equal(t, expectedGroup, group)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// 	mockGroupRedisDao.AssertExpectations(t)
// }

// func TestGetGroupByGroupID_NotFound(t *testing.T) {
// 	groupsService, mockGroupDao, _, _, mockTxManager, mockGroupRedisDao, _ := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 999 // 不存在的ID

// 	// Mock期望
// 	mockGroupRedisDao.On("GetByGroupID", ctx, groupID, mock.Anything).Return(nil, redis.Nil) // 模拟缓存未命中
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

// 	// 调用函数
// 	group, err := groupsService.GetGroupByGroupID(ctx, groupID)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrGroupNotFound))
// 	assert.Nil(t, group)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// 	mockGroupRedisDao.AssertExpectations(t)
// }

// // --- GetGroupsByUserID 测试 ---

// func TestGetGroupsByUserID_Success(t *testing.T) {
// 	groupsService, mockGroupDao, _, _, mockTxManager, _, _ := setupGroupServiceTest()
// 	ctx := context.Background()
// 	userID := 1
// 	expectedGroups := []*models.Group{
// 		{
// 			GroupID:     1,
// 			GroupName:   "群组1",
// 			Description: "描述1",
// 			CreatorID:   1,
// 			CreatorName: "admin",
// 			MemberNum:   2,
// 		},
// 		{
// 			GroupID:     2,
// 			GroupName:   "群组2",
// 			Description: "描述2",
// 			CreatorID:   2,
// 			CreatorName: "user2",
// 			MemberNum:   3,
// 		},
// 	}

// 	// Mock期望 - 不带filter参数的情况
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupDao.On("GetGroupsByUserID", ctx, userID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedGroups, nil)

// 	// 调用函数 - 不带filter参数
// 	groups, err := groupsService.GetGroupsByUserID(ctx, userID)

// 	// 断言
// 	assert.NoError(t, err)
// 	assert.NotNil(t, groups)
// 	assert.Equal(t, expectedGroups, groups)
// 	assert.Len(t, groups, 2)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// }

// func TestGetGroupsByUserID_WithFilter_Success(t *testing.T) {
// 	groupsService, mockGroupDao, _, _, mockTxManager, _, _ := setupGroupServiceTest()
// 	ctx := context.Background()
// 	userID := 1
// 	filter := "created"
// 	expectedGroups := []*models.Group{
// 		{
// 			GroupID:     1,
// 			GroupName:   "已创建群组1",
// 			Description: "描述1",
// 			CreatorID:   1,
// 			CreatorName: "admin",
// 			MemberNum:   2,
// 		},
// 	}

// 	// Mock期望 - 带filter参数的情况
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupDao.On("GetGroupsByUserIDAndfilter", ctx, userID, filter, mock.AnythingOfType("[]*gorm.DB")).Return(expectedGroups, nil)

// 	// 调用函数 - 带filter参数
// 	groups, err := groupsService.GetGroupsByUserID(ctx, userID, filter)

// 	// 断言
// 	assert.NoError(t, err)
// 	assert.NotNil(t, groups)
// 	assert.Equal(t, expectedGroups, groups)
// 	assert.Len(t, groups, 1)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// }

// func TestGetGroupsByUserID_NotFound(t *testing.T) {
// 	groupsService, mockGroupDao, _, _, mockTxManager, _, _ := setupGroupServiceTest()
// 	ctx := context.Background()
// 	userID := 999 // 不存在的用户

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupDao.On("GetGroupsByUserID", ctx, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

// 	// 调用函数
// 	groups, err := groupsService.GetGroupsByUserID(ctx, userID)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrGroupNotFound))
// 	assert.Nil(t, groups)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// }

// func TestGetGroupsByUserID_WithFilter_NotFound(t *testing.T) {
// 	groupsService, mockGroupDao, _, _, mockTxManager, _, _ := setupGroupServiceTest()
// 	ctx := context.Background()
// 	userID := 999 // 不存在的用户
// 	filter := "joined"

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupDao.On("GetGroupsByUserIDAndfilter", ctx, userID, filter, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

// 	// 调用函数
// 	groups, err := groupsService.GetGroupsByUserID(ctx, userID, filter)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrGroupNotFound))
// 	assert.Nil(t, groups)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// }

// // --- UpdateGroup 测试 ---

// func TestUpdateGroup_Success(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager, mockGroupRedisDao, mockGroupMemberRedisDao := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	operatorID := 1 // 添加操作者ID（管理员）
// 	groupName := "更新后的群组名"
// 	description := "更新后的描述"
// 	updatedGroup := &models.Group{
// 		GroupID:     groupID,
// 		GroupName:   groupName,
// 		Description: description,
// 		CreatorID:   1,
// 		CreatorName: "admin",
// 		MemberNum:   1,
// 	}
// 	adminMember := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "admin",
// 		Role:     "admin",
// 	}
// 	members := []*models.GroupMember{adminMember}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(adminMember, nil)
// 	mockGroupDao.On("UpdateMessage", ctx, groupID, groupName, description, mock.AnythingOfType("[]*gorm.DB")).Return(nil)
// 	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(updatedGroup, nil)
// 	mockGroupMemberDao.On("GetMembersByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(members, nil)

// 	// Mock Redis操作
// 	mockGroupRedisDao.On("DeleteCacheByGroupID", ctx, groupID).Return(nil)
// 	mockGroupMemberRedisDao.On("DeleteCacheByGroupID", ctx, groupID).Return(nil)
// 	mockGroupMemberRedisDao.On("DeleteCacheByGroupIDAndUserID", ctx, groupID, operatorID).Return(nil)

// 	// 调用函数
// 	result, err := groupsService.UpdateGroup(ctx, groupID, operatorID, groupName, description)

// 	// 断言
// 	assert.NoError(t, err)
// 	assert.NotNil(t, result)
// 	assert.Equal(t, updatedGroup, result)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupRedisDao.AssertExpectations(t)
// 	mockGroupMemberRedisDao.AssertExpectations(t)
// }

// func TestUpdateGroup_PermissionDenied(t *testing.T) {
// 	groupsService, _, mockGroupMemberDao, _, mockTxManager, _, _ := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	operatorID := 2 // 非管理员用户
// 	groupName := "更新后的群组名"
// 	description := "更新后的描述"
// 	member := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "regular_member",
// 		Role:     "member", // 非管理员角色
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(member, nil)

// 	// 调用函数
// 	result, err := groupsService.UpdateGroup(ctx, groupID, operatorID, groupName, description)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrRolePermissionDenied) ||
// 		strings.Contains(err.Error(), appErrors.ErrRolePermissionDenied.Error()))
// 	assert.Nil(t, result)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// }

// func TestUpdateGroup_NotMember(t *testing.T) {
// 	groupsService, _, mockGroupMemberDao, _, mockTxManager, _, _ := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	operatorID := 999 // 非成员用户
// 	groupName := "更新后的群组名"
// 	description := "更新后的描述"

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

// 	// 调用函数
// 	result, err := groupsService.UpdateGroup(ctx, groupID, operatorID, groupName, description)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrGroupMemberNotFound) ||
// 		strings.Contains(err.Error(), appErrors.ErrGroupMemberNotFound.Error()) ||
// 		strings.Contains(err.Error(), appErrors.ErrRolePermissionDenied.Error()))
// 	assert.Nil(t, result)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// }

// // --- CheckMemberPermission 测试 ---

// func TestCheckMemberPermission_AdminSuccess(t *testing.T) {
// 	groupsService, _, mockGroupMemberDao, _, _, _, mockGroupMemberRedisDao := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	userID := 1
// 	member := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   userID,
// 		Username: "admin",
// 		Role:     "admin",
// 	}

// 	// Mock期望
// 	mockGroupMemberRedisDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.Anything).Return(nil, redis.Nil) // 缓存未命中
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(member, nil)
// 	mockGroupMemberRedisDao.On("SetMemberByGroupIDAndUserID", ctx, member).Return(nil) // 缓存写入

// 	// 调用函数
// 	err := groupsService.CheckMemberPermission(ctx, groupID, userID)

// 	// 断言
// 	assert.NoError(t, err)

// 	// 验证mock调用
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupMemberRedisDao.AssertExpectations(t)
// }

// func TestCheckMemberPermission_NotAdmin(t *testing.T) {
// 	groupsService, _, mockGroupMemberDao, _, _, _, mockGroupMemberRedisDao := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	userID := 2
// 	member := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   userID,
// 		Username: "user",
// 		Role:     "member", // 非管理员
// 	}

// 	// Mock期望
// 	mockGroupMemberRedisDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.Anything).Return(nil, redis.Nil) // 缓存未命中
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(member, nil)
// 	mockGroupMemberRedisDao.On("SetMemberByGroupIDAndUserID", ctx, member).Return(nil) // 缓存写入

// 	// 调用函数
// 	err := groupsService.CheckMemberPermission(ctx, groupID, userID)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrRolePermissionDenied))

// 	// 验证mock调用
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupMemberRedisDao.AssertExpectations(t)
// }

// func TestCheckMemberPermission_MemberNotFound(t *testing.T) {
// 	groupsService, _, mockGroupMemberDao, _, _, _, mockGroupMemberRedisDao := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	userID := 999 // 不存在的成员

// 	// Mock期望
// 	mockGroupMemberRedisDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.Anything).Return(nil, redis.Nil) // 缓存未命中
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

// 	// 调用函数
// 	err := groupsService.CheckMemberPermission(ctx, groupID, userID)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrGroupMemberNotFound))

// 	// 验证mock调用
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupMemberRedisDao.AssertExpectations(t)
// }

// // --- RemoveMemberFromGroup 测试 ---

// func TestRemoveMemberFromGroup_Success(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	userID := 2
// 	operatorID := 1 // 管理员
// 	adminMember := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "admin",
// 		Role:     "admin",
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(adminMember, nil)
// 	mockGroupMemberDao.On("Delete", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil)
// 	mockGroupDao.On("UpdateMemberNum", ctx, groupID, false, mock.AnythingOfType("[]*gorm.DB")).Return(nil)

// 	// 调用函数
// 	err := groupsService.RemoveMemberFromGroup(ctx, groupID, userID, operatorID)

// 	// 断言
// 	assert.NoError(t, err)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// }

// func TestRemoveMemberFromGroup_NoPermission(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	userID := 2
// 	operatorID := 3 // 非管理员
// 	member := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "user3",
// 		Role:     "member", // 非管理员角色
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(member, nil)

// 	// 调用函数
// 	err := groupsService.RemoveMemberFromGroup(ctx, groupID, userID, operatorID)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), appErrors.ErrRolePermissionDenied.Error())

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupDao.AssertNotCalled(t, "UpdateMemberNum", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
// }

// // --- GetMembersByGroupID 测试 ---

// func TestGetMembersByGroupID_Success(t *testing.T) {
// 	groupsService, _, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	expectedMembers := []*models.GroupMember{
// 		{
// 			GroupID:  groupID,
// 			UserID:   1,
// 			Username: "admin",
// 			Role:     "admin",
// 		},
// 		{
// 			GroupID:  groupID,
// 			UserID:   2,
// 			Username: "member1",
// 			Role:     "member",
// 		},
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMembersByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedMembers, nil)

// 	// 调用函数
// 	members, err := groupsService.GetMembersByGroupID(ctx, groupID)

// 	// 断言
// 	assert.NoError(t, err)
// 	assert.NotNil(t, members)
// 	assert.Equal(t, expectedMembers, members)
// 	assert.Len(t, members, 2)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// }

// func TestGetMembersByGroupID_GroupNotFound(t *testing.T) {
// 	groupsService, _, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 999 // 不存在的群组

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMembersByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

// 	// 调用函数
// 	members, err := groupsService.GetMembersByGroupID(ctx, groupID)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrGroupNotFound))
// 	assert.Nil(t, members)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// }

// // --- CreateJoinApplication 测试 ---

// func TestCreateJoinApplication_Success(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, mockJoinApplicationDao, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	userID := 2
// 	username := "applicant"
// 	reason := "我想加入这个群组"
// 	group := &models.Group{
// 		GroupID:     groupID,
// 		GroupName:   "测试群组",
// 		Description: "这是一个测试群组",
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(group, nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)
// 	mockJoinApplicationDao.On("Create", ctx, mock.AnythingOfType("*models.JoinApplication"), mock.AnythingOfType("[]*gorm.DB")).Return(nil).Run(func(args mock.Arguments) {
// 		appArg := args.Get(1).(*models.JoinApplication)
// 		assert.Equal(t, groupID, appArg.GroupID)
// 		assert.Equal(t, userID, appArg.UserID)
// 		assert.Equal(t, username, appArg.Username)
// 		assert.Equal(t, reason, appArg.Reason)
// 	})

// 	// 调用函数
// 	application, err := groupsService.CreateJoinApplication(ctx, groupID, userID, username, reason)

// 	// 断言
// 	assert.NoError(t, err)
// 	assert.NotNil(t, application)
// 	assert.Equal(t, groupID, application.GroupID)
// 	assert.Equal(t, userID, application.UserID)
// 	assert.Equal(t, username, application.Username)
// 	assert.Equal(t, reason, application.Reason)
// 	assert.Equal(t, "pending", application.Status)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockJoinApplicationDao.AssertExpectations(t)
// }

// func TestCreateJoinApplication_GroupNotFound(t *testing.T) {
// 	groupsService, mockGroupDao, _, _, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 999 // 不存在的群组
// 	userID := 2
// 	username := "applicant"
// 	reason := "我想加入这个群组"

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

// 	// 调用函数
// 	application, err := groupsService.CreateJoinApplication(ctx, groupID, userID, username, reason)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrGroupNotFound))
// 	assert.Nil(t, application)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// }

// func TestCreateJoinApplication_AlreadyMember(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	userID := 2
// 	username := "member"
// 	reason := "我想加入这个群组"
// 	group := &models.Group{
// 		GroupID:     groupID,
// 		GroupName:   "测试群组",
// 		Description: "这是一个测试群组",
// 	}
// 	existingMember := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   userID,
// 		Username: username,
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(group, nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(existingMember, nil)

// 	// 调用函数
// 	application, err := groupsService.CreateJoinApplication(ctx, groupID, userID, username, reason)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrGroupMemberAlreadyExists))
// 	assert.Nil(t, application)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// }

// // --- GetJoinApplicationsByGroupID 测试 ---

// func TestGetJoinApplicationsByGroupID_Success(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, mockJoinApplicationDao, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	operatorID := 1 // 管理员
// 	adminMember := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "admin",
// 		Role:     "admin",
// 	}
// 	group := &models.Group{
// 		GroupID:     groupID,
// 		GroupName:   "测试群组",
// 		Description: "这是一个测试群组",
// 	}
// 	expectedApplications := []*models.JoinApplication{
// 		{
// 			RequestID: 1,
// 			GroupID:   groupID,
// 			UserID:    2,
// 			Username:  "applicant1",
// 			Reason:    "我想加入这个群组",
// 			Status:    "pending",
// 		},
// 		{
// 			RequestID: 2,
// 			GroupID:   groupID,
// 			UserID:    3,
// 			Username:  "applicant2",
// 			Reason:    "请允许我加入",
// 			Status:    "pending",
// 		},
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(adminMember, nil)
// 	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(group, nil)
// 	mockJoinApplicationDao.On("GetByGroupIDAndStatus", ctx, groupID, "pending", mock.AnythingOfType("[]*gorm.DB")).Return(expectedApplications, nil)

// 	// 调用函数
// 	applications, err := groupsService.GetJoinApplicationsByGroupID(ctx, groupID, operatorID)

// 	// 断言
// 	assert.NoError(t, err)
// 	assert.NotNil(t, applications)
// 	assert.Equal(t, expectedApplications, applications)
// 	assert.Len(t, applications, 2)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// 	mockJoinApplicationDao.AssertExpectations(t)
// }

// func TestGetJoinApplicationsByGroupID_NoPermission(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	operatorID := 2 // 非管理员
// 	member := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "user",
// 		Role:     "member", // 非管理员角色
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(member, nil)

// 	// 调用函数
// 	applications, err := groupsService.GetJoinApplicationsByGroupID(ctx, groupID, operatorID)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), appErrors.ErrRolePermissionDenied.Error())
// 	assert.Nil(t, applications)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupDao.AssertNotCalled(t, "GetByGroupID", mock.Anything, mock.Anything, mock.Anything)
// }

// func TestGetJoinApplicationsByGroupID_GroupNotFound(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	operatorID := 1 // 管理员
// 	adminMember := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "admin",
// 		Role:     "admin",
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(adminMember, nil)
// 	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

// 	// 调用函数
// 	applications, err := groupsService.GetJoinApplicationsByGroupID(ctx, groupID, operatorID)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrGroupNotFound))
// 	assert.Nil(t, applications)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// }

// func TestGetJoinApplicationsByGroupID_NoApplications(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, mockJoinApplicationDao, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	operatorID := 1 // 管理员
// 	adminMember := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "admin",
// 		Role:     "admin",
// 	}
// 	group := &models.Group{
// 		GroupID:     groupID,
// 		GroupName:   "测试群组",
// 		Description: "这是一个测试群组",
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(adminMember, nil)
// 	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(group, nil)
// 	mockJoinApplicationDao.On("GetByGroupIDAndStatus", ctx, groupID, "pending", mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

// 	// 调用函数
// 	applications, err := groupsService.GetJoinApplicationsByGroupID(ctx, groupID, operatorID)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrJoinApplicationNotFound))
// 	assert.Nil(t, applications)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// 	mockJoinApplicationDao.AssertExpectations(t)
// }

// func TestGetJoinApplicationsByGroupID_AllFilter(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, mockJoinApplicationDao, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	operatorID := 1 // 管理员
// 	filter := "all"
// 	adminMember := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "admin",
// 		Role:     "admin",
// 	}
// 	group := &models.Group{
// 		GroupID:     groupID,
// 		GroupName:   "测试群组",
// 		Description: "这是一个测试群组",
// 	}
// 	expectedApplications := []*models.JoinApplication{
// 		{
// 			RequestID: 1,
// 			GroupID:   groupID,
// 			UserID:    2,
// 			Username:  "applicant1",
// 			Reason:    "我想加入这个群组",
// 			Status:    "pending",
// 		},
// 		{
// 			RequestID: 2,
// 			GroupID:   groupID,
// 			UserID:    3,
// 			Username:  "applicant2",
// 			Reason:    "请允许我加入",
// 			Status:    "accepted",
// 		},
// 		{
// 			RequestID: 3,
// 			GroupID:   groupID,
// 			UserID:    4,
// 			Username:  "applicant3",
// 			Reason:    "申请加入",
// 			Status:    "rejected",
// 		},
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(adminMember, nil)
// 	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(group, nil)
// 	mockJoinApplicationDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedApplications, nil)

// 	// 调用函数
// 	applications, err := groupsService.GetJoinApplicationsByGroupID(ctx, groupID, operatorID, filter)

// 	// 断言
// 	assert.NoError(t, err)
// 	assert.NotNil(t, applications)
// 	assert.Equal(t, expectedApplications, applications)
// 	assert.Len(t, applications, 3)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// 	mockJoinApplicationDao.AssertExpectations(t)
// }

// func TestGetJoinApplicationsByGroupID_StatusFilter(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, mockJoinApplicationDao, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	operatorID := 1      // 管理员
// 	filter := "accepted" // 指定状态过滤
// 	adminMember := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "admin",
// 		Role:     "admin",
// 	}
// 	group := &models.Group{
// 		GroupID:     groupID,
// 		GroupName:   "测试群组",
// 		Description: "这是一个测试群组",
// 	}
// 	expectedApplications := []*models.JoinApplication{
// 		{
// 			RequestID: 2,
// 			GroupID:   groupID,
// 			UserID:    3,
// 			Username:  "applicant2",
// 			Reason:    "请允许我加入",
// 			Status:    "accepted",
// 		},
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(adminMember, nil)
// 	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(group, nil)
// 	mockJoinApplicationDao.On("GetByGroupIDAndStatus", ctx, groupID, filter, mock.AnythingOfType("[]*gorm.DB")).Return(expectedApplications, nil)

// 	// 调用函数
// 	applications, err := groupsService.GetJoinApplicationsByGroupID(ctx, groupID, operatorID, filter)

// 	// 断言
// 	assert.NoError(t, err)
// 	assert.NotNil(t, applications)
// 	assert.Equal(t, expectedApplications, applications)
// 	assert.Len(t, applications, 1)
// 	assert.Equal(t, "accepted", applications[0].Status)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// 	mockJoinApplicationDao.AssertExpectations(t)
// }

// // --- RejectJoinApplication 测试 ---

// func TestRejectJoinApplication_Success(t *testing.T) {
// 	groupsService, _, mockGroupMemberDao, mockJoinApplicationDao, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	userID := 2
// 	operatorID := 1 // 管理员
// 	requestID := 1
// 	username := "applicant"
// 	rejectReason := "不符合要求"
// 	adminMember := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "admin",
// 		Role:     "admin",
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(adminMember, nil)
// 	mockJoinApplicationDao.On("UpdateStatus", ctx, requestID, "rejected", mock.AnythingOfType("[]*gorm.DB")).Return(nil)
// 	mockJoinApplicationDao.On("UpdateRejectReason", ctx, requestID, rejectReason, mock.AnythingOfType("[]*gorm.DB")).Return(nil)

// 	// 调用函数
// 	err := groupsService.RejectJoinApplication(ctx, groupID, userID, operatorID, requestID, username, rejectReason)

// 	// 断言
// 	assert.NoError(t, err)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockJoinApplicationDao.AssertExpectations(t)
// }

// func TestRejectJoinApplication_PermissionDenied(t *testing.T) {
// 	groupsService, _, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	userID := 2
// 	operatorID := 3 // 非管理员
// 	requestID := 1
// 	username := "applicant"
// 	rejectReason := "不符合要求"
// 	member := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "regularUser",
// 		Role:     "member", // 非管理员角色
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(member, nil)

// 	// 调用函数
// 	err := groupsService.RejectJoinApplication(ctx, groupID, userID, operatorID, requestID, username, rejectReason)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrRolePermissionDenied) ||
// 		strings.Contains(err.Error(), appErrors.ErrRolePermissionDenied.Error()))

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// }

// // --- DeleteGroup 测试 ---

// func TestDeleteGroup_Success(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	operatorID := 1 // 管理员
// 	adminMember := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "admin",
// 		Role:     "admin",
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(adminMember, nil)
// 	mockGroupDao.On("Delete", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(nil)

// 	// 调用函数
// 	err := groupsService.DeleteGroup(ctx, groupID, operatorID)

// 	// 断言
// 	assert.NoError(t, err)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// }

// func TestDeleteGroup_PermissionDenied(t *testing.T) {
// 	groupsService, _, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	operatorID := 2 // 非管理员
// 	member := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "regularUser",
// 		Role:     "member", // 非管理员角色
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(member, nil)

// 	// 调用函数
// 	err := groupsService.DeleteGroup(ctx, groupID, operatorID)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrRolePermissionDenied) ||
// 		strings.Contains(err.Error(), appErrors.ErrRolePermissionDenied.Error()))

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// }

// // --- GetUserGroupStatus 测试 ---

// func TestGetUserGroupStatus_Success(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, mockJoinApplicationDao, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	userID := 2
// 	group := &models.Group{
// 		GroupID:     groupID,
// 		GroupName:   "测试群组",
// 		Description: "这是一个测试群组",
// 	}
// 	application := &models.JoinApplication{
// 		RequestID:    1,
// 		GroupID:      groupID,
// 		UserID:       userID,
// 		Username:     "user",
// 		Reason:       "想加入",
// 		Status:       "pending",
// 		RejectReason: "",
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(group, nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)
// 	mockJoinApplicationDao.On("GetByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(application, nil)

// 	// 调用函数
// 	status, requestID, err := groupsService.GetUserGroupStatus(ctx, groupID, userID)

// 	// 断言
// 	assert.NoError(t, err)
// 	assert.Equal(t, application.Status, status)
// 	assert.Equal(t, application.RequestID, requestID)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockJoinApplicationDao.AssertExpectations(t)
// }

// func TestGetUserGroupStatus_NotFound(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, mockJoinApplicationDao, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	userID := 2
// 	group := &models.Group{
// 		GroupID:     groupID,
// 		GroupName:   "测试群组",
// 		Description: "这是一个测试群组",
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(group, nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)
// 	mockJoinApplicationDao.On("GetByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

// 	// 调用函数
// 	status, requestID, err := groupsService.GetUserGroupStatus(ctx, groupID, userID)

// 	// 断言
// 	assert.NoError(t, err) // 根据实现，这种情况应该不报错而是返回"none"状态
// 	assert.Equal(t, "none", status)
// 	assert.Equal(t, 0, requestID)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockJoinApplicationDao.AssertExpectations(t)
// }

// func TestGetUserGroupStatus_Member(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	userID := 2
// 	group := &models.Group{
// 		GroupID:     groupID,
// 		GroupName:   "测试群组",
// 		Description: "这是一个测试群组",
// 	}
// 	member := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   userID,
// 		Username: "user",
// 		Role:     "member", // 普通成员
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(group, nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(member, nil)

// 	// 调用函数
// 	status, requestID, err := groupsService.GetUserGroupStatus(ctx, groupID, userID)

// 	// 断言
// 	assert.NoError(t, err)
// 	assert.Equal(t, member.Role, status) // 返回角色为状态
// 	assert.Equal(t, 0, requestID)        // 组成员的requestID应为0

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// }

// // --- ApproveJoinApplication 测试 ---

// func TestApproveJoinApplication_Success(t *testing.T) {
// 	groupsService, mockGroupDao, mockGroupMemberDao, mockJoinApplicationDao, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	userID := 2
// 	operatorID := 1 // 管理员
// 	requestID := 1
// 	username := "applicant"
// 	adminMember := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "admin",
// 		Role:     "admin",
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(adminMember, nil)
// 	mockGroupDao.On("UpdateMemberNum", ctx, groupID, true, mock.AnythingOfType("[]*gorm.DB")).Return(nil)
// 	mockGroupMemberDao.On("Create", ctx, mock.AnythingOfType("*models.GroupMember"), mock.AnythingOfType("[]*gorm.DB")).Return(nil).Run(func(args mock.Arguments) {
// 		memberArg := args.Get(1).(*models.GroupMember)
// 		assert.Equal(t, groupID, memberArg.GroupID)
// 		assert.Equal(t, userID, memberArg.UserID)
// 		assert.Equal(t, username, memberArg.Username)
// 	})
// 	mockJoinApplicationDao.On("UpdateStatus", ctx, requestID, "accepted", mock.AnythingOfType("[]*gorm.DB")).Return(nil)

// 	// 调用函数
// 	err := groupsService.ApproveJoinApplication(ctx, groupID, userID, operatorID, requestID, username)

// 	// 断言
// 	assert.NoError(t, err)

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// 	mockGroupDao.AssertExpectations(t)
// 	mockJoinApplicationDao.AssertExpectations(t)
// }

// func TestApproveJoinApplication_PermissionDenied(t *testing.T) {
// 	groupsService, _, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
// 	ctx := context.Background()
// 	groupID := 1
// 	userID := 2
// 	operatorID := 3 // 非管理员
// 	requestID := 1
// 	username := "applicant"
// 	member := &models.GroupMember{
// 		GroupID:  groupID,
// 		UserID:   operatorID,
// 		Username: "regularUser",
// 		Role:     "member", // 非管理员角色
// 	}

// 	// Mock期望
// 	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
// 	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(member, nil)

// 	// 调用函数
// 	err := groupsService.ApproveJoinApplication(ctx, groupID, userID, operatorID, requestID, username)

// 	// 断言
// 	assert.Error(t, err)
// 	assert.True(t, errors.Is(err, appErrors.ErrRolePermissionDenied) ||
// 		strings.Contains(err.Error(), appErrors.ErrRolePermissionDenied.Error()))

// 	// 验证mock调用
// 	mockTxManager.AssertExpectations(t)
// 	mockGroupMemberDao.AssertExpectations(t)
// }
