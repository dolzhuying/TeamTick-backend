package service

import (
	"TeamTickBackend/dal/models"
	appErrors "TeamTickBackend/pkg/errors"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mock 实现 ---

// Mock GroupDAO
type mockGroupDAO struct {
	mock.Mock
}

func (m *mockGroupDAO) Create(ctx context.Context, group *models.Group, tx ...*gorm.DB) error {
	args := m.Called(ctx, group, tx)
	// 模拟创建时分配ID和时间
	if args.Error(0) == nil {
		group.GroupID = 1
		group.CreatedAt = time.Now()
		group.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *mockGroupDAO) GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) (*models.Group, error) {
	args := m.Called(ctx, groupID, tx)
	groupArg := args.Get(0)
	if groupArg == nil {
		return nil, args.Error(1)
	}
	return groupArg.(*models.Group), args.Error(1)
}

func (m *mockGroupDAO) GetGroupsByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.Group, error) {
	args := m.Called(ctx, userID, tx)
	groupsArg := args.Get(0)
	if groupsArg == nil {
		return nil, args.Error(1)
	}
	return groupsArg.([]*models.Group), args.Error(1)
}

func (m *mockGroupDAO) UpdateMessage(ctx context.Context, groupID int, groupName, description string, tx ...*gorm.DB) error {
	args := m.Called(ctx, groupID, groupName, description, tx)
	return args.Error(0)
}

func (m *mockGroupDAO) UpdateMemberNum(ctx context.Context, groupID int, isAdd bool, tx ...*gorm.DB) error {
	args := m.Called(ctx, groupID, isAdd, tx)
	return args.Error(0)
}

// Mock GroupMemberDAO
type mockGroupMemberDAO struct {
	mock.Mock
}

func (m *mockGroupMemberDAO) Create(ctx context.Context, member *models.GroupMember, tx ...*gorm.DB) error {
	args := m.Called(ctx, member, tx)
	if args.Error(0) == nil {
		member.CreatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *mockGroupMemberDAO) GetMemberByGroupIDAndUserID(ctx context.Context, groupID, userID int, tx ...*gorm.DB) (*models.GroupMember, error) {
	args := m.Called(ctx, groupID, userID, tx)
	memberArg := args.Get(0)
	if memberArg == nil {
		return nil, args.Error(1)
	}
	return memberArg.(*models.GroupMember), args.Error(1)
}

func (m *mockGroupMemberDAO) GetMembersByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.GroupMember, error) {
	args := m.Called(ctx, groupID, tx)
	membersArg := args.Get(0)
	if membersArg == nil {
		return nil, args.Error(1)
	}
	return membersArg.([]*models.GroupMember), args.Error(1)
}

func (m *mockGroupMemberDAO) Delete(ctx context.Context, groupID, userID int, tx ...*gorm.DB) error {
	args := m.Called(ctx, groupID, userID, tx)
	return args.Error(0)
}

// Mock JoinApplicationDAO
type mockJoinApplicationDAO struct {
	mock.Mock
}

func (m *mockJoinApplicationDAO) Create(ctx context.Context, application *models.JoinApplication, tx ...*gorm.DB) error {
	args := m.Called(ctx, application, tx)
	if args.Error(0) == nil {
		application.Status = "pending"
		application.CreatedAt = time.Now()
		application.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *mockJoinApplicationDAO) GetByGroupIDAndStatus(ctx context.Context, groupID int, status string, tx ...*gorm.DB) ([]*models.JoinApplication, error) {
	args := m.Called(ctx, groupID, status, tx)
	applicationsArg := args.Get(0)
	if applicationsArg == nil {
		return nil, args.Error(1)
	}
	return applicationsArg.([]*models.JoinApplication), args.Error(1)
}

// 添加缺失的GetByUserID方法
func (m *mockJoinApplicationDAO) GetByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.JoinApplication, error) {
	args := m.Called(ctx, userID, tx)
	applicationsArg := args.Get(0)
	if applicationsArg == nil {
		return nil, args.Error(1)
	}
	return applicationsArg.([]*models.JoinApplication), args.Error(1)
}

// 添加缺失的UpdateStatus方法
func (m *mockJoinApplicationDAO) UpdateStatus(ctx context.Context, applicationID int, status string, tx ...*gorm.DB) error {
	args := m.Called(ctx, applicationID, status, tx)
	return args.Error(0)
}

// --- 测试准备 ---

func setupGroupServiceTest() (*GroupsService, *mockGroupDAO, *mockGroupMemberDAO, *mockJoinApplicationDAO, *mockTransactionManager) {
	mockGroupDao := new(mockGroupDAO)
	mockGroupMemberDao := new(mockGroupMemberDAO)
	mockJoinApplicationDao := new(mockJoinApplicationDAO)
	mockTxManager := new(mockTransactionManager)

	groupsService := NewGroupsService(
		mockGroupDao,
		mockGroupMemberDao,
		mockJoinApplicationDao,
		mockTxManager,
	)

	return groupsService, mockGroupDao, mockGroupMemberDao, mockJoinApplicationDao, mockTxManager
}

// --- CreateGroup 测试 ---

func TestCreateGroup_Success(t *testing.T) {
	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupName := "测试群组"
	description := "这是一个测试群组"
	creatorName := "admin"
	creatorID := 1

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupDao.On("Create", ctx, mock.AnythingOfType("*models.Group"), mock.AnythingOfType("[]*gorm.DB")).Return(nil).Run(func(args mock.Arguments) {
		// 验证传给Create的群组对象
		groupArg := args.Get(1).(*models.Group)
		assert.Equal(t, groupName, groupArg.GroupName)
		assert.Equal(t, description, groupArg.Description)
		assert.Equal(t, creatorID, groupArg.CreatorID)
		assert.Equal(t, creatorName, groupArg.CreatorName)
	})
	mockGroupMemberDao.On("Create", ctx, mock.AnythingOfType("*models.GroupMember"), mock.AnythingOfType("[]*gorm.DB")).Return(nil).Run(func(args mock.Arguments) {
		// 验证传给Create的群组成员对象
		memberArg := args.Get(1).(*models.GroupMember)
		assert.Equal(t, 1, memberArg.GroupID) // 模拟ID为1
		assert.Equal(t, creatorID, memberArg.UserID)
		assert.Equal(t, creatorName, memberArg.Username)
		assert.Equal(t, "admin", memberArg.Role)
	})

	// 调用函数
	createdGroup, err := groupsService.CreateGroup(ctx, groupName, description, creatorName, creatorID)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, createdGroup)
	assert.Equal(t, groupName, createdGroup.GroupName)
	assert.Equal(t, description, createdGroup.Description)
	assert.Equal(t, creatorID, createdGroup.CreatorID)
	assert.Equal(t, creatorName, createdGroup.CreatorName)
	assert.NotZero(t, createdGroup.GroupID)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
	mockGroupMemberDao.AssertExpectations(t)
}

// --- GetGroupByGroupID 测试 ---

func TestGetGroupByGroupID_Success(t *testing.T) {
	groupsService, mockGroupDao, _, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	expectedGroup := &models.Group{
		GroupID:     groupID,
		GroupName:   "测试群组",
		Description: "这是一个测试群组",
		CreatorID:   1,
		CreatorName: "admin",
		MemberNum:   1,
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedGroup, nil)

	// 调用函数
	group, err := groupsService.GetGroupByGroupID(ctx, groupID)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, group)
	assert.Equal(t, expectedGroup, group)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
}

func TestGetGroupByGroupID_NotFound(t *testing.T) {
	groupsService, mockGroupDao, _, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 999 // 不存在的ID

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	group, err := groupsService.GetGroupByGroupID(ctx, groupID)

	// 断言
	assert.Error(t, err)
	assert.True(t, errors.Is(err, appErrors.ErrGroupNotFound))
	assert.Nil(t, group)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
}

// --- GetGroupsByUserID 测试 ---

func TestGetGroupsByUserID_Success(t *testing.T) {
	groupsService, mockGroupDao, _, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	userID := 1
	expectedGroups := []*models.Group{
		{
			GroupID:     1,
			GroupName:   "群组1",
			Description: "描述1",
			CreatorID:   1,
			CreatorName: "admin",
			MemberNum:   2,
		},
		{
			GroupID:     2,
			GroupName:   "群组2",
			Description: "描述2",
			CreatorID:   2,
			CreatorName: "user2",
			MemberNum:   3,
		},
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupDao.On("GetGroupsByUserID", ctx, userID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedGroups, nil)

	// 调用函数
	groups, err := groupsService.GetGroupsByUserID(ctx, userID)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, groups)
	assert.Equal(t, expectedGroups, groups)
	assert.Len(t, groups, 2)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
}

func TestGetGroupsByUserID_NotFound(t *testing.T) {
	groupsService, mockGroupDao, _, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	userID := 999 // 不存在的用户

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupDao.On("GetGroupsByUserID", ctx, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	groups, err := groupsService.GetGroupsByUserID(ctx, userID)

	// 断言
	assert.Error(t, err)
	assert.True(t, errors.Is(err, appErrors.ErrGroupNotFound))
	assert.Nil(t, groups)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
}

// --- UpdateGroup 测试 ---

func TestUpdateGroup_Success(t *testing.T) {
	groupsService, mockGroupDao, _, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	groupName := "更新后的群组名"
	description := "更新后的描述"
	updatedGroup := &models.Group{
		GroupID:     groupID,
		GroupName:   groupName,
		Description: description,
		CreatorID:   1,
		CreatorName: "admin",
		MemberNum:   1,
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupDao.On("UpdateMessage", ctx, groupID, groupName, description, mock.AnythingOfType("[]*gorm.DB")).Return(nil)
	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(updatedGroup, nil)

	// 调用函数
	result, err := groupsService.UpdateGroup(ctx, groupID, groupName, description)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, updatedGroup, result)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
}

// --- CheckMemberPermission 测试 ---

func TestCheckMemberPermission_AdminSuccess(t *testing.T) {
	groupsService, _, mockGroupMemberDao, _, _ := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	userID := 1
	member := &models.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		Username: "admin",
		Role:     "admin",
	}

	// Mock期望
	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(member, nil)

	// 调用函数
	err := groupsService.CheckMemberPermission(ctx, groupID, userID)

	// 断言
	assert.NoError(t, err)

	// 验证mock调用
	mockGroupMemberDao.AssertExpectations(t)
}

func TestCheckMemberPermission_NotAdmin(t *testing.T) {
	groupsService, _, mockGroupMemberDao, _, _ := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	userID := 2
	member := &models.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		Username: "user",
		Role:     "member", // 非管理员
	}

	// Mock期望
	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(member, nil)

	// 调用函数
	err := groupsService.CheckMemberPermission(ctx, groupID, userID)

	// 断言
	assert.Error(t, err)
	assert.True(t, errors.Is(err, appErrors.ErrRolePermissionDenied))

	// 验证mock调用
	mockGroupMemberDao.AssertExpectations(t)
}

func TestCheckMemberPermission_MemberNotFound(t *testing.T) {
	groupsService, _, mockGroupMemberDao, _, _ := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	userID := 999 // 不存在的成员

	// Mock期望
	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	err := groupsService.CheckMemberPermission(ctx, groupID, userID)

	// 断言
	assert.Error(t, err)
	assert.True(t, errors.Is(err, appErrors.ErrGroupMemberNotFound))

	// 验证mock调用
	mockGroupMemberDao.AssertExpectations(t)
}

// --- AddMemberToGroup 测试 ---

func TestAddMemberToGroup_Success(t *testing.T) {
	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	userID := 2
	operatorID := 1 // 操作者ID（管理员）
	username := "testuser"

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)
	mockGroupMemberDao.On("Create", ctx, mock.AnythingOfType("*models.GroupMember"), mock.AnythingOfType("[]*gorm.DB")).Return(nil).Run(func(args mock.Arguments) {
		memberArg := args.Get(1).(*models.GroupMember)
		assert.Equal(t, groupID, memberArg.GroupID)
		assert.Equal(t, userID, memberArg.UserID)
		assert.Equal(t, username, memberArg.Username)
	})
	mockGroupDao.On("UpdateMemberNum", ctx, groupID, true, mock.AnythingOfType("[]*gorm.DB")).Return(nil)

	// 调用函数
	member, err := groupsService.AddMemberToGroup(ctx, groupID, userID, operatorID, username)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, member)
	assert.Equal(t, groupID, member.GroupID)
	assert.Equal(t, userID, member.UserID)
	assert.Equal(t, username, member.Username)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupMemberDao.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
}

func TestAddMemberToGroup_MemberAlreadyExists(t *testing.T) {
	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	userID := 2
	operatorID := 1
	username := "testuser"
	existingMember := &models.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		Username: username,
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(existingMember, nil)

	// 调用函数
	member, err := groupsService.AddMemberToGroup(ctx, groupID, userID, operatorID, username)

	// 断言
	assert.Error(t, err)
	assert.True(t, errors.Is(err, appErrors.ErrGroupMemberAlreadyExists))
	assert.Nil(t, member)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupMemberDao.AssertExpectations(t)
	mockGroupDao.AssertNotCalled(t, "UpdateMemberNum", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// --- RemoveMemberFromGroup 测试 ---

func TestRemoveMemberFromGroup_Success(t *testing.T) {
	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	userID := 2
	operatorID := 1 // 管理员
	adminMember := &models.GroupMember{
		GroupID:  groupID,
		UserID:   operatorID,
		Username: "admin",
		Role:     "admin",
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(adminMember, nil)
	mockGroupMemberDao.On("Delete", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil)
	mockGroupDao.On("UpdateMemberNum", ctx, groupID, false, mock.AnythingOfType("[]*gorm.DB")).Return(nil)

	// 调用函数
	err := groupsService.RemoveMemberFromGroup(ctx, groupID, userID, operatorID)

	// 断言
	assert.NoError(t, err)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupMemberDao.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
}

func TestRemoveMemberFromGroup_NoPermission(t *testing.T) {
	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	userID := 2
	operatorID := 3 // 非管理员
	member := &models.GroupMember{
		GroupID:  groupID,
		UserID:   operatorID,
		Username: "user3",
		Role:     "member", // 非管理员角色
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(member, nil)

	// 调用函数
	err := groupsService.RemoveMemberFromGroup(ctx, groupID, userID, operatorID)

	// 断言
	assert.Error(t, err)
	assert.Contains(t, err.Error(), appErrors.ErrRolePermissionDenied.Error())

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupMemberDao.AssertExpectations(t)
	mockGroupDao.AssertNotCalled(t, "UpdateMemberNum", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// --- GetMembersByGroupID 测试 ---

func TestGetMembersByGroupID_Success(t *testing.T) {
	groupsService, _, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	expectedMembers := []*models.GroupMember{
		{
			GroupID:  groupID,
			UserID:   1,
			Username: "admin",
			Role:     "admin",
		},
		{
			GroupID:  groupID,
			UserID:   2,
			Username: "member1",
			Role:     "member",
		},
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupMemberDao.On("GetMembersByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(expectedMembers, nil)

	// 调用函数
	members, err := groupsService.GetMembersByGroupID(ctx, groupID)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, members)
	assert.Equal(t, expectedMembers, members)
	assert.Len(t, members, 2)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupMemberDao.AssertExpectations(t)
}

func TestGetMembersByGroupID_GroupNotFound(t *testing.T) {
	groupsService, _, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 999 // 不存在的群组

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupMemberDao.On("GetMembersByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	members, err := groupsService.GetMembersByGroupID(ctx, groupID)

	// 断言
	assert.Error(t, err)
	assert.True(t, errors.Is(err, appErrors.ErrGroupNotFound))
	assert.Nil(t, members)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupMemberDao.AssertExpectations(t)
}

// --- CreateJoinApplication 测试 ---

func TestCreateJoinApplication_Success(t *testing.T) {
	groupsService, mockGroupDao, mockGroupMemberDao, mockJoinApplicationDao, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	userID := 2
	username := "applicant"
	reason := "我想加入这个群组"
	group := &models.Group{
		GroupID:     groupID,
		GroupName:   "测试群组",
		Description: "这是一个测试群组",
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(group, nil)
	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)
	mockJoinApplicationDao.On("Create", ctx, mock.AnythingOfType("*models.JoinApplication"), mock.AnythingOfType("[]*gorm.DB")).Return(nil).Run(func(args mock.Arguments) {
		appArg := args.Get(1).(*models.JoinApplication)
		assert.Equal(t, groupID, appArg.GroupID)
		assert.Equal(t, userID, appArg.UserID)
		assert.Equal(t, username, appArg.Username)
		assert.Equal(t, reason, appArg.Reason)
	})

	// 调用函数
	application, err := groupsService.CreateJoinApplication(ctx, groupID, userID, username, reason)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, application)
	assert.Equal(t, groupID, application.GroupID)
	assert.Equal(t, userID, application.UserID)
	assert.Equal(t, username, application.Username)
	assert.Equal(t, reason, application.Reason)
	assert.Equal(t, "pending", application.Status)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
	mockGroupMemberDao.AssertExpectations(t)
	mockJoinApplicationDao.AssertExpectations(t)
}

func TestCreateJoinApplication_GroupNotFound(t *testing.T) {
	groupsService, mockGroupDao, _, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 999 // 不存在的群组
	userID := 2
	username := "applicant"
	reason := "我想加入这个群组"

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	application, err := groupsService.CreateJoinApplication(ctx, groupID, userID, username, reason)

	// 断言
	assert.Error(t, err)
	assert.True(t, errors.Is(err, appErrors.ErrGroupNotFound))
	assert.Nil(t, application)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
}

func TestCreateJoinApplication_AlreadyMember(t *testing.T) {
	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	userID := 2
	username := "member"
	reason := "我想加入这个群组"
	group := &models.Group{
		GroupID:     groupID,
		GroupName:   "测试群组",
		Description: "这是一个测试群组",
	}
	existingMember := &models.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		Username: username,
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(group, nil)
	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, userID, mock.AnythingOfType("[]*gorm.DB")).Return(existingMember, nil)

	// 调用函数
	application, err := groupsService.CreateJoinApplication(ctx, groupID, userID, username, reason)

	// 断言
	assert.Error(t, err)
	assert.True(t, errors.Is(err, appErrors.ErrGroupMemberAlreadyExists))
	assert.Nil(t, application)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
	mockGroupMemberDao.AssertExpectations(t)
}

// --- GetJoinApplicationsByGroupID 测试 ---

func TestGetJoinApplicationsByGroupID_Success(t *testing.T) {
	groupsService, mockGroupDao, mockGroupMemberDao, mockJoinApplicationDao, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	operatorID := 1 // 管理员
	adminMember := &models.GroupMember{
		GroupID:  groupID,
		UserID:   operatorID,
		Username: "admin",
		Role:     "admin",
	}
	group := &models.Group{
		GroupID:     groupID,
		GroupName:   "测试群组",
		Description: "这是一个测试群组",
	}
	expectedApplications := []*models.JoinApplication{
		{
			RequestID: 1,
			GroupID:   groupID,
			UserID:    2,
			Username:  "applicant1",
			Reason:    "我想加入这个群组",
			Status:    "pending",
		},
		{
			RequestID: 2,
			GroupID:   groupID,
			UserID:    3,
			Username:  "applicant2",
			Reason:    "请允许我加入",
			Status:    "pending",
		},
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(adminMember, nil)
	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(group, nil)
	mockJoinApplicationDao.On("GetByGroupIDAndStatus", ctx, groupID, "pending", mock.AnythingOfType("[]*gorm.DB")).Return(expectedApplications, nil)

	// 调用函数
	applications, err := groupsService.GetJoinApplicationsByGroupID(ctx, groupID, operatorID)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Equal(t, expectedApplications, applications)
	assert.Len(t, applications, 2)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupMemberDao.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
	mockJoinApplicationDao.AssertExpectations(t)
}

func TestGetJoinApplicationsByGroupID_NoPermission(t *testing.T) {
	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	operatorID := 2 // 非管理员
	member := &models.GroupMember{
		GroupID:  groupID,
		UserID:   operatorID,
		Username: "user",
		Role:     "member", // 非管理员角色
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(member, nil)

	// 调用函数
	applications, err := groupsService.GetJoinApplicationsByGroupID(ctx, groupID, operatorID)

	// 断言
	assert.Error(t, err)
	assert.Contains(t, err.Error(), appErrors.ErrRolePermissionDenied.Error())
	assert.Nil(t, applications)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupMemberDao.AssertExpectations(t)
	mockGroupDao.AssertNotCalled(t, "GetByGroupID", mock.Anything, mock.Anything, mock.Anything)
}

func TestGetJoinApplicationsByGroupID_GroupNotFound(t *testing.T) {
	groupsService, mockGroupDao, mockGroupMemberDao, _, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	operatorID := 1 // 管理员
	adminMember := &models.GroupMember{
		GroupID:  groupID,
		UserID:   operatorID,
		Username: "admin",
		Role:     "admin",
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(adminMember, nil)
	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	applications, err := groupsService.GetJoinApplicationsByGroupID(ctx, groupID, operatorID)

	// 断言
	assert.Error(t, err)
	assert.True(t, errors.Is(err, appErrors.ErrGroupNotFound))
	assert.Nil(t, applications)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupMemberDao.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
}

func TestGetJoinApplicationsByGroupID_NoApplications(t *testing.T) {
	groupsService, mockGroupDao, mockGroupMemberDao, mockJoinApplicationDao, mockTxManager := setupGroupServiceTest()
	ctx := context.Background()
	groupID := 1
	operatorID := 1 // 管理员
	adminMember := &models.GroupMember{
		GroupID:  groupID,
		UserID:   operatorID,
		Username: "admin",
		Role:     "admin",
	}
	group := &models.Group{
		GroupID:     groupID,
		GroupName:   "测试群组",
		Description: "这是一个测试群组",
	}

	// Mock期望
	mockTxManager.On("WithTransaction", ctx, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)
	mockGroupMemberDao.On("GetMemberByGroupIDAndUserID", ctx, groupID, operatorID, mock.AnythingOfType("[]*gorm.DB")).Return(adminMember, nil)
	mockGroupDao.On("GetByGroupID", ctx, groupID, mock.AnythingOfType("[]*gorm.DB")).Return(group, nil)
	mockJoinApplicationDao.On("GetByGroupIDAndStatus", ctx, groupID, "pending", mock.AnythingOfType("[]*gorm.DB")).Return(nil, gorm.ErrRecordNotFound)

	// 调用函数
	applications, err := groupsService.GetJoinApplicationsByGroupID(ctx, groupID, operatorID)

	// 断言
	assert.Error(t, err)
	assert.True(t, errors.Is(err, appErrors.ErrJoinApplicationNotFound))
	assert.Nil(t, applications)

	// 验证mock调用
	mockTxManager.AssertExpectations(t)
	mockGroupMemberDao.AssertExpectations(t)
	mockGroupDao.AssertExpectations(t)
	mockJoinApplicationDao.AssertExpectations(t)
}
