package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	apperrors "TeamTickBackend/pkg/errors"
	"context"
	"errors"

	"gorm.io/gorm"
)

type GroupsService struct {
	groupDao           dao.GroupDAO
	groupMemberDao     dao.GroupMemberDAO
	joinApplicationDao dao.JoinApplicationDAO
	transactionManager dao.TransactionManager
}

func NewGroupsService(
	groupDao dao.GroupDAO,
	groupMemberDao dao.GroupMemberDAO,
	joinApplicationDao dao.JoinApplicationDAO,
	transactionManager dao.TransactionManager,
) *GroupsService {

	return &GroupsService{
		groupDao:           groupDao,
		groupMemberDao:     groupMemberDao,
		joinApplicationDao: joinApplicationDao,
		transactionManager: transactionManager,
	}
}

// 创建用户组
func (s *GroupsService) CreateGroup(ctx context.Context, groupName, description, creatorName string, creatorID int) (*models.Group, error) {
	var createdGroup models.Group

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		group := models.Group{
			GroupName:   groupName,
			Description: description,
			CreatorID:   creatorID,
			CreatorName: creatorName,
		}
		//创建用户组
		if err := s.groupDao.Create(ctx, &group, tx); err != nil {
			return apperrors.ErrGroupCreationFailed.WithError(err)
		}
		//添加用户组管理员
		if err := s.groupMemberDao.Create(ctx, &models.GroupMember{
			GroupID:  group.GroupID,
			UserID:   creatorID,
			Username: creatorName,
			Role:     "admin",
		}, tx); err != nil {
			return apperrors.ErrGroupMemberCreationFailed.WithError(err)
		}
		createdGroup = group
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &createdGroup, nil
}

// 通过GroupID查询用户组信息
func (s *GroupsService) GetGroupByGroupID(ctx context.Context, groupID int) (*models.Group, error) {
	var group models.Group
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		Group, err := s.groupDao.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrGroupNotFound
			}
			return apperrors.ErrDatabaseOperation.WithError(err)
		}
		group = *Group
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// 查询特定用户所创建或加入的所有用户组
func (s *GroupsService) GetGroupsByUserID(ctx context.Context, userID int) ([]*models.Group, error) {
	var groups []*models.Group
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		Groups, err := s.groupDao.GetGroupsByUserID(ctx, userID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrGroupNotFound
			}
			return apperrors.ErrDatabaseOperation.WithError(err)
		}
		groups = Groups
		return nil
	})
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// 更新用户组信息
func (s *GroupsService) UpdateGroup(ctx context.Context, groupID int, groupName, description string) (*models.Group, error) {
	var updatedGroup models.Group
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//更新用户组信息
		if err := s.groupDao.UpdateMessage(ctx, groupID, groupName, description, tx); err != nil {
			return apperrors.ErrGroupUpdateFailed.WithError(err)
		}
		//查询更新后的用户组信息
		group, err := s.groupDao.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			return apperrors.ErrDatabaseOperation.WithError(err)
		}
		updatedGroup = *group
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &updatedGroup, nil
}

// 检查用户组成员权限
func (s *GroupsService) CheckMemberPermission(ctx context.Context, groupID, userID int) error {
	member, err := s.groupMemberDao.GetMemberByGroupIDAndUserID(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrGroupMemberNotFound
		}
		return apperrors.ErrDatabaseOperation.WithError(err)
	}
	if member.Role == "admin" {
		return nil
	}
	return apperrors.ErrRolePermissionDenied
}

// 添加用户到用户组
// MVP版本申请加入暂时为直接加入，不需要审批，直接调用该函数
// 但是迭代版本需要审批，审批通过则会执行 往用户-用户组表添加组员，更新用户组成员数量，更新申请表中记录的状态三个行为，需要放在一个事务中
func (s *GroupsService) AddMemberToGroup(ctx context.Context, groupID, userID, operatorID int, username string) (*models.GroupMember, error) {
	var member models.GroupMember
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查操作员权限
		// if err:=s.CheckMemberPermission(ctx,groupID,operatorID);err!=nil{
		// 	return apperrors.ErrRolePermissionDenied.WithError(err)
		// }
		//检查用户组是否存在
		existMember, err := s.groupMemberDao.GetMemberByGroupIDAndUserID(ctx, groupID, userID, tx)
		if err == nil && existMember != nil {
			return apperrors.ErrGroupMemberAlreadyExists
		}
		newMember := models.GroupMember{
			GroupID:  groupID,
			UserID:   userID,
			Username: username,
		}
		//创建用户组成员
		if err := s.groupMemberDao.Create(ctx, &newMember, tx); err != nil {
			return apperrors.ErrGroupMemberCreationFailed.WithError(err)
		}
		//更新用户组成员数量
		if err := s.groupDao.UpdateMemberNum(ctx, groupID, true, tx); err != nil {
			return apperrors.ErrGroupUpdateFailed.WithError(err)
		}
		member = newMember
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// 删除用户组中的用户
func (s *GroupsService) RemoveMemberFromGroup(ctx context.Context, groupID, userID, operatorID int) error {
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查操作员权限
		if err := s.CheckMemberPermission(ctx, groupID, operatorID); err != nil {
			return apperrors.ErrRolePermissionDenied.WithError(err)
		}
		//删除用户组成员
		if err := s.groupMemberDao.Delete(ctx, groupID, userID, tx); err != nil {
			return apperrors.ErrGroupMemberDeletionFailed.WithError(err)
		}
		//更新用户组成员数量
		if err := s.groupDao.UpdateMemberNum(ctx, groupID, false, tx); err != nil {
			return apperrors.ErrGroupUpdateFailed.WithError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// 查询用户组中的所有成员
// 关于返回值，是否需要在查询成员表来获取成员信息，然后返回成员信息列表？（涉及多次数据库操作）
func (s *GroupsService) GetMembersByGroupID(ctx context.Context, groupID int) ([]*models.GroupMember, error) {
	var members []*models.GroupMember
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		groupMembers, err := s.groupMemberDao.GetMembersByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrGroupNotFound
			}
			return apperrors.ErrDatabaseOperation.WithError(err)
		}
		members = groupMembers
		return nil
	})
	if err != nil {
		return nil, err
	}
	return members, nil
}

// 创建用户申请加入记录(返回值？是否需要返回申请记录)
func (s *GroupsService) CreateJoinApplication(ctx context.Context, groupID, userID int, username, reason string) (*models.JoinApplication, error) {
	var application models.JoinApplication
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查用户组是否存在
		_, err := s.groupDao.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrGroupNotFound
			}
			return apperrors.ErrDatabaseOperation.WithError(err)
		}
		//检查用户是否已加入用户组
		member, err := s.groupMemberDao.GetMemberByGroupIDAndUserID(ctx, groupID, userID, tx)
		if err == nil && member != nil {
			return apperrors.ErrGroupMemberAlreadyExists
		}
		//创建申请记录
		newApplication := models.JoinApplication{
			GroupID:  groupID,
			UserID:   userID,
			Username: username,
			Reason:   reason,
		}
		if err := s.joinApplicationDao.Create(ctx, &newApplication, tx); err != nil {
			return apperrors.ErrJoinApplicationCreationFailed.WithError(err)
		}
		application = newApplication
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &application, nil
}

// 查看用户组加入申请列表
// 关于返回值，是否需要查询用户列表返回所有用户信息？（多次数据库操作）
func (s *GroupsService) GetJoinApplicationsByGroupID(ctx context.Context, groupID, operatorID int) ([]*models.JoinApplication, error) {
	var applications []*models.JoinApplication
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查操作员权限
		if err := s.CheckMemberPermission(ctx, groupID, operatorID); err != nil {
			return apperrors.ErrRolePermissionDenied.WithError(err)
		}
		//检查用户组是否存在
		_, err := s.groupDao.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrGroupNotFound
			}
			return apperrors.ErrDatabaseOperation.WithError(err)
		}

		//查询待审批的申请记录
		existApplications, err := s.joinApplicationDao.GetByGroupIDAndStatus(ctx, groupID, "pending", tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrJoinApplicationNotFound
			}
			return apperrors.ErrDatabaseOperation.WithError(err)
		}
		applications = existApplications
		return nil
	})
	if err != nil {
		return nil, err
	}
	return applications, nil
}

// 审批用户组加入申请（迭代，涉及多个操作，AddMemberToGroup提及）
